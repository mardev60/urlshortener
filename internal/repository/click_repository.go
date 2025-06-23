package repository

import (
	"fmt"

	"github.com/axellelanca/urlshortener/internal/models"
	"gorm.io/gorm"
)

// ClickRepository est une interface qui définit les méthodes d'accès aux données
// pour les opérations sur les clics. Cette abstraction permet à la couche service
// de rester indépendante de l'implémentation spécifique de la base de données.
type ClickRepository interface {
	CreateClick(click *models.Click) error
	CountClicksByLinkID(linkID uint) (int, error) // Utilisé par LinkService pour les stats
}

// GormClickRepository est l'implémentation de l'interface ClickRepository utilisant GORM.
type GormClickRepository struct {
	db *gorm.DB // Référence à l'instance de la base de données GORM
}

// NewClickRepository crée et retourne une nouvelle instance de GormClickRepository.
// C'est la méthode recommandée pour obtenir un dépôt, garantissant que la connexion à la base de données est injectée.
func NewClickRepository(db *gorm.DB) *GormClickRepository {
	return &GormClickRepository{db: db}
}

// CreateClick insère un nouvel enregistrement de clic dans la base de données.
// Elle reçoit un pointeur vers une structure models.Click et la persiste en utilisant GORM.
func (r *GormClickRepository) CreateClick(click *models.Click) error {
	result := r.db.Create(click)
	if result.Error != nil {
		return fmt.Errorf("erreur lors de la création du clic: %w", result.Error)
	}
	return nil
}

// CountClicksByLinkID compte le nombre total de clics pour un ID de lien donné.
// Cette méthode est utilisée pour fournir des statistiques pour une URL courte.
func (r *GormClickRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64
	result := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("erreur lors du comptage des clics: %w", result.Error)
	}
	return int(count), nil
}
