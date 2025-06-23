package models

import "time"

// Click représente un événement de clic sur un lien raccourci.
// GORM utilisera ces tags pour créer la table 'clicks'.
type Click struct {
	ID        uint `gorm:"primaryKey"`
	LinkID    uint `gorm:"index"`
	Link      Link `gorm:"foreignKey:LinkID"`
	Timestamp time.Time
	UserAgent string `gorm:"size:255"`
	IPAddress string `gorm:"size:50"`
}

// ClickEvent représente un événement de clic brut, destiné à être passé via un channel
// Ce n'est pas un modèle GORM direct.
type ClickEvent struct {
	LinkID    uint
	Timestamp time.Time
	UserAgent string
	IPAddress string
}
