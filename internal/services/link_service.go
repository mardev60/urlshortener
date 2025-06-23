package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"gorm.io/gorm" // Nécessaire pour la gestion spécifique de gorm.ErrRecordNotFound

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository" // Importe le package repository
)

// Définition du jeu de caractères pour la génération des codes courts.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// LinkService est une structure qui fournit des méthodes pour la logique métier des liens.
type LinkService struct {
	linkRepo repository.LinkRepository
}

// NewLinkService crée et retourne une nouvelle instance de LinkService.
func NewLinkService(linkRepo repository.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}

// GenerateShortCode génère un code court aléatoire d'une longueur spécifiée.
func (s *LinkService) GenerateShortCode(length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("erreur lors de la génération du code court: %w", err)
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

// CreateLink crée un nouveau lien raccourci.
func (s *LinkService) CreateLink(longURL string) (*models.Link, error) {
	var shortCode string
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		code, err := s.GenerateShortCode(6)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la génération du code court: %w", err)
		}

		_, err = s.GetLinkByShortCode(code)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				shortCode = code
				break
			}
			return nil, fmt.Errorf("erreur lors de la vérification du code court: %w", err)
		}

		log.Printf("Le code court '%s' existe déjà, nouvelle tentative (%d/%d)...", code, i+1, maxRetries)
	}

	if shortCode == "" {
		return nil, errors.New("impossible de générer un code court unique après plusieurs tentatives")
	}

	link := &models.Link{
		ShortCode: shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
	}

	err := s.linkRepo.CreateLink(link)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du lien: %w", err)
	}

	return link, nil
}

// GetLinkByShortCode récupère un lien via son code court.
func (s *LinkService) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	link, err := s.linkRepo.GetLinkByShortCode(shortCode)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du lien: %w", err)
	}
	return link, nil
}

// GetLinkStats récupère les statistiques pour un lien donné.
func (s *LinkService) GetLinkStats(shortCode string) (*models.Link, int, error) {
	link, err := s.GetLinkByShortCode(shortCode)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors de la récupération du lien: %w", err)
	}

	clickCount, err := s.linkRepo.CountClicksByLinkID(link.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur lors du comptage des clics: %w", err)
	}

	return link, clickCount, nil
}
