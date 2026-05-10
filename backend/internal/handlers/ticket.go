package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ops-manager/internal/models"
)

type TicketHandler struct {
	DB *gorm.DB
}

func NewTicketHandler(db *gorm.DB) *TicketHandler {
	return &TicketHandler{DB: db}
}

func (h *TicketHandler) List(c *gin.Context) {
	var tickets []models.Ticket
	query := h.DB.Preload("Applicant").Preload("Handler").Preload("Assets")

	if ticketType := c.Query("type"); ticketType != "" {
		query = query.Where("type = ?", ticketType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	query.Find(&tickets)
	c.JSON(http.StatusOK, tickets)
}

func (h *TicketHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var ticket models.Ticket
	if err := h.DB.Preload("Applicant").Preload("Handler").Preload("Assets").Preload("Approvals.Approver").First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

type CreateTicketRequest struct {
	Type        string `json:"type" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Priority    string `json:"priority" binding:"required"`
	AssetIDs    []uint `json:"asset_ids"`
}

func (h *TicketHandler) Create(c *gin.Context) {
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, _ := c.Get("userID")

	ticket := models.Ticket{
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      "pending",
		ApplicantID: userID.(uint),
	}

	if err := h.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	if len(req.AssetIDs) > 0 {
		var assets []models.Asset
		h.DB.Find(&assets, req.AssetIDs)
		h.DB.Model(&ticket).Association("Assets").Replace(assets)
	}

	h.DB.Preload("Applicant").Preload("Handler").Preload("Assets").First(&ticket, ticket.ID)
	c.JSON(http.StatusCreated, ticket)
}

type UpdateTicketRequest struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	HandlerID   *uint  `json:"handler_id"`
	AssetIDs    []uint `json:"asset_ids"`
}

func (h *TicketHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var ticket models.Ticket
	if err := h.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	var req UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Type != "" {
		ticket.Type = req.Type
	}
	if req.Title != "" {
		ticket.Title = req.Title
	}
	if req.Description != "" {
		ticket.Description = req.Description
	}
	if req.Priority != "" {
		ticket.Priority = req.Priority
	}
	if req.HandlerID != nil {
		ticket.HandlerID = req.HandlerID
	}

	h.DB.Save(&ticket)

	if len(req.AssetIDs) > 0 {
		var assets []models.Asset
		h.DB.Find(&assets, req.AssetIDs)
		h.DB.Model(&ticket).Association("Assets").Replace(assets)
	}

	h.DB.Preload("Applicant").Preload("Handler").Preload("Assets").First(&ticket, ticket.ID)
	c.JSON(http.StatusOK, ticket)
}

type ApproveRequest struct {
	Result  string `json:"result" binding:"required"`
	Comment string `json:"comment"`
}

func (h *TicketHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var ticket models.Ticket
	if err := h.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	if ticket.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket is not pending"})
		return
	}

	var req ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, _ := c.Get("userID")
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)

	var level int
	if role == "leader" {
		level = 1
	} else if role == "manager" {
		level = 2
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only leader or manager can approve"})
		return
	}

	approval := models.Approval{
		TicketID:   uint(id),
		ApproverID: userID.(uint),
		Level:      level,
		Result:     req.Result,
		Comment:    req.Comment,
	}

	if err := h.DB.Create(&approval).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create approval"})
		return
	}

	if req.Result == "approved" {
		if level == 2 {
			ticket.Status = "approved"
		} else {
			ticket.Status = "pending"
		}
	} else {
		ticket.Status = "rejected"
	}

	h.DB.Save(&ticket)
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) Close(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var ticket models.Ticket
	if err := h.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	ticket.Status = "closed"
	h.DB.Save(&ticket)

	c.JSON(http.StatusOK, ticket)
}