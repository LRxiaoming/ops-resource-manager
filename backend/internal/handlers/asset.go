package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ops-manager/internal/models"
)

type AssetHandler struct {
	DB *gorm.DB
}

func NewAssetHandler(db *gorm.DB) *AssetHandler {
	return &AssetHandler{DB: db}
}

func (h *AssetHandler) List(c *gin.Context) {
	var assets []models.Asset
	query := h.DB

	if assetType := c.Query("type"); assetType != "" {
		query = query.Where("type = ?", assetType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Find(&assets)
	c.JSON(http.StatusOK, assets)
}

func (h *AssetHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var asset models.Asset
	if err := h.DB.First(&asset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

type CreateAssetRequest struct {
	Type        string `json:"type" binding:"required"`
	IP          string `json:"ip"`
	Hostname    string `json:"hostname"`
	CPU         string `json:"cpu"`
	Memory      string `json:"memory"`
	Disk        string `json:"disk"`
	Location    string `json:"location"`
	Responsible string `json:"responsible"`
	Status      string `json:"status" binding:"required"`
}

func (h *AssetHandler) Create(c *gin.Context) {
	var req CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	asset := models.Asset{
		Type:        req.Type,
		IP:          req.IP,
		Hostname:    req.Hostname,
		CPU:         req.CPU,
		Memory:      req.Memory,
		Disk:        req.Disk,
		Location:    req.Location,
		Responsible: req.Responsible,
		Status:      req.Status,
	}

	if err := h.DB.Create(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create asset"})
		return
	}

	c.JSON(http.StatusCreated, asset)
}

type UpdateAssetRequest struct {
	Type        string `json:"type"`
	IP          string `json:"ip"`
	Hostname    string `json:"hostname"`
	CPU         string `json:"cpu"`
	Memory      string `json:"memory"`
	Disk        string `json:"disk"`
	Location    string `json:"location"`
	Responsible string `json:"responsible"`
	Status      string `json:"status"`
}

func (h *AssetHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var asset models.Asset
	if err := h.DB.First(&asset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	var req UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Type != "" {
		asset.Type = req.Type
	}
	if req.IP != "" {
		asset.IP = req.IP
	}
	if req.Hostname != "" {
		asset.Hostname = req.Hostname
	}
	if req.CPU != "" {
		asset.CPU = req.CPU
	}
	if req.Memory != "" {
		asset.Memory = req.Memory
	}
	if req.Disk != "" {
		asset.Disk = req.Disk
	}
	if req.Location != "" {
		asset.Location = req.Location
	}
	if req.Responsible != "" {
		asset.Responsible = req.Responsible
	}
	if req.Status != "" {
		asset.Status = req.Status
	}

	h.DB.Save(&asset)
	c.JSON(http.StatusOK, asset)
}

func (h *AssetHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.DB.Delete(&models.Asset{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete asset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset deleted"})
}