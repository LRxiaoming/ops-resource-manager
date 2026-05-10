package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ops-manager/internal/models"
)

type DashboardHandler struct {
	DB *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{DB: db}
}

type Stats struct {
	TotalAssets     int64            `json:"total_assets"`
	AssetByType     map[string]int64 `json:"asset_by_type"`
	AssetByStatus   map[string]int64 `json:"asset_by_status"`
	TotalTickets    int64            `json:"total_tickets"`
	TicketByType    map[string]int64 `json:"ticket_by_type"`
	TicketByStatus  map[string]int64 `json:"ticket_by_status"`
	TicketByPriority map[string]int64 `json:"ticket_by_priority"`
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	stats := Stats{
		AssetByType:    make(map[string]int64),
		AssetByStatus:  make(map[string]int64),
		TicketByType:   make(map[string]int64),
		TicketByStatus: make(map[string]int64),
		TicketByPriority: make(map[string]int64),
	}

	h.DB.Model(&models.Asset{}).Count(&stats.TotalAssets)
	h.DB.Model(&models.Ticket{}).Count(&stats.TotalTickets)

	var typeCount int64
	rows, _ := h.DB.Model(&models.Asset{}).Select("type, count(*) as count").Group("type").Rows()
	for rows.Next() {
		var assetType string
		rows.Scan(&assetType, &typeCount)
		stats.AssetByType[assetType] = typeCount
	}

	rows, _ = h.DB.Model(&models.Asset{}).Select("status, count(*) as count").Group("status").Rows()
	for rows.Next() {
		var status string
		rows.Scan(&status, &typeCount)
		stats.AssetByStatus[status] = typeCount
	}

	rows, _ = h.DB.Model(&models.Ticket{}).Select("type, count(*) as count").Group("type").Rows()
	for rows.Next() {
		var ticketType string
		rows.Scan(&ticketType, &typeCount)
		stats.TicketByType[ticketType] = typeCount
	}

	rows, _ = h.DB.Model(&models.Ticket{}).Select("status, count(*) as count").Group("status").Rows()
	for rows.Next() {
		var status string
		rows.Scan(&status, &typeCount)
		stats.TicketByStatus[status] = typeCount
	}

	rows, _ = h.DB.Model(&models.Ticket{}).Select("priority, count(*) as count").Group("priority").Rows()
	for rows.Next() {
		var priority string
		rows.Scan(&priority, &typeCount)
		stats.TicketByPriority[priority] = typeCount
	}

	c.JSON(http.StatusOK, stats)
}

func (h *DashboardHandler) Charts(c *gin.Context) {
	type Trend struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	var trends []Trend

	rows, _ := h.DB.Model(&models.Ticket{}).
		Select("date(created_at) as date, count(*) as count").
		Group("date(created_at)").
		Order("date desc").
		Limit(30).Rows()

	for rows.Next() {
		var t Trend
		rows.Scan(&t.Date, &t.Count)
		trends = append(trends, t)
	}

	c.JSON(http.StatusOK, gin.H{"trends": trends})
}