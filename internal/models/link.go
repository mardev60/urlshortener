package models

import "time"

// Link représente un lien raccourci dans la base de données.
// Les tags `gorm:"..."` définissent comment GORM doit mapper cette structure à une table SQL.
type Link struct {
	ID        uint      `gorm:"primarykey"`
	ShortCode string    `gorm:"uniqueIndex;size:10;not null"`
	LongURL   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}
