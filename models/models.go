package models

import (
    // "gorm.io/gorm"
)

type Post struct {
    ID        uint   `gorm:"primaryKey" json:"id"`
    Title     string `json:"title"`
    Body      string `json:"body"`
    DateTime  string `gorm:"column:date_time" json:"datetime"`  // Ensure this matches your column name in the DB
    CreatedAt string `json:"created_at"`  // Manually set
    UpdatedAt string `json:"updated_at"`  // Manually set
}



type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique;not null"`
    Username    string `json:"username" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
}


type Message struct {
	UserID  int    `json:"user_id"`
	Content string `json:"content"`
    Fromid  int `json:"from_id" gorm:"column:from_id"`
}