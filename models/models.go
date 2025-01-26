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
