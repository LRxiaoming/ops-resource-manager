package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"

	"ops-manager/internal/models"
)

func main() {
	dsn := "host=localhost user=ops password=ops123 dbname=ops_manager port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	user := models.User{
		Username: "admin",
		Password: string(hash),
		Role:     "admin",
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Println("Admin user created successfully!")
}