package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Role      string    `gorm:"size:20;not null" json:"role"` // admin, applicant, leader, manager
	CreatedAt time.Time `json:"created_at"`
}

type Asset struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"size:20;not null" json:"type"` // server, network
	IP          string    `gorm:"size:50" json:"ip"`
	Hostname    string    `gorm:"size:100" json:"hostname"`
	CPU         string    `gorm:"size:50" json:"cpu"`
	Memory      string    `gorm:"size:50" json:"memory"`
	Disk        string    `gorm:"size:50" json:"disk"`
	Location    string    `gorm:"size:100" json:"location"`
	Responsible string    `gorm:"size:100" json:"responsible"`
	Status      string    `gorm:"size:20;not null" json:"status"` // online, offline, maintenance
	CreatedAt   time.Time `json:"created_at"`
}

type Ticket struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"size:30;not null" json:"type"` // fault, change, resource
	Title       string    `gorm:"size:200;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Priority    string    `gorm:"size:20;not null" json:"priority"` // urgent, important, normal
	Status      string    `gorm:"size:20;not null" json:"status"`   // pending, approved, rejected, closed
	ApplicantID uint      `gorm:"not null" json:"applicant_id"`
	Applicant   User      `gorm:"foreignKey:ApplicantID" json:"applicant,omitempty"`
	HandlerID   *uint     `json:"handler_id"`
	Handler     *User     `gorm:"foreignKey:HandlerID" json:"handler,omitempty"`
	Assets      []Asset   `gorm:"many2many:ticket_assets" json:"assets,omitempty"`
	Approvals   []Approval `gorm:"foreignKey:TicketID" json:"approvals,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Approval struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TicketID  uint      `gorm:"not null" json:"ticket_id"`
	ApproverID uint     `gorm:"not null" json:"approver_id"`
	Approver   User     `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	Level     int       `gorm:"not null" json:"level"` // 1=组长, 2=经理
	Result    string    `gorm:"size:20;not null" json:"result"` // approved, rejected
	Comment   string    `gorm:"type:text" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}