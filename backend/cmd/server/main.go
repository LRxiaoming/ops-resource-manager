package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ops-manager/internal/config"
	"ops-manager/internal/handlers"
	"ops-manager/internal/middleware"
	"ops-manager/internal/models"
)

func main() {
	cfg := config.Load()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.AutoMigrate(&models.User{}, &models.Asset{}, &models.Ticket{}, &models.Approval{})

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	userHandler := handlers.NewUserHandler(db)
	assetHandler := handlers.NewAssetHandler(db)
	ticketHandler := handlers.NewTicketHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			users := protected.Group("/users")
			{
				users.GET("", middleware.RoleMiddleware("admin"), userHandler.List)
				users.GET("/:id", userHandler.Get)
				users.POST("", middleware.RoleMiddleware("admin"), userHandler.Create)
				users.PUT("/:id", middleware.RoleMiddleware("admin"), userHandler.Update)
				users.DELETE("/:id", middleware.RoleMiddleware("admin"), userHandler.Delete)
			}

			assets := protected.Group("/assets")
			{
				assets.GET("", assetHandler.List)
				assets.GET("/:id", assetHandler.Get)
				assets.POST("", middleware.RoleMiddleware("admin", "leader"), assetHandler.Create)
				assets.PUT("/:id", middleware.RoleMiddleware("admin", "leader"), assetHandler.Update)
				assets.DELETE("/:id", middleware.RoleMiddleware("admin"), assetHandler.Delete)
			}

			tickets := protected.Group("/tickets")
			{
				tickets.GET("", ticketHandler.List)
				tickets.GET("/:id", ticketHandler.Get)
				tickets.POST("", ticketHandler.Create)
				tickets.PUT("/:id", ticketHandler.Update)
				tickets.POST("/:id/approve", middleware.RoleMiddleware("leader", "manager"), ticketHandler.Approve)
				tickets.POST("/:id/close", ticketHandler.Close)
			}

			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", dashboardHandler.Stats)
				dashboard.GET("/charts", dashboardHandler.Charts)
			}
		}
	}

	r.Run(":" + cfg.ServerPort)
}