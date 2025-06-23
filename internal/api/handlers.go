package api

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	// Pour gérer gorm.ErrRecordNotFound
)

// ClickEventsChannel est le channel global utilisé pour envoyer les événements de clic
// aux workers asynchrones. Il est bufferisé pour ne pas bloquer les requêtes de redirection.
var ClickEventsChannel chan models.ClickEvent

// GetClickEventsChannel retourne le channel global des événements de clic.
// Cette fonction est utilisée par les workers pour s'assurer qu'ils utilisent le bon channel.
func GetClickEventsChannel() chan models.ClickEvent {
	return ClickEventsChannel
}

// SetupRoutes configure toutes les routes de l'API Gin et injecte les dépendances nécessaires
func SetupRoutes(router *gin.Engine, linkService *services.LinkService, bufferSize int) {
	// Initialiser le channel avec la taille du buffer configurée
	ClickEventsChannel = make(chan models.ClickEvent, bufferSize)
	log.Printf("[DEBUG] Channel des événements de clic initialisé avec un buffer de %d", bufferSize)

	// Route de Health Check
	router.GET("/health", HealthCheckHandler)

	// Routes de l'API
	v1 := router.Group("/api/v1")
	{
		v1.POST("/links", CreateShortLinkHandler(linkService))
		v1.GET("/links/:shortCode/stats", GetLinkStatsHandler(linkService))
	}

	// Route de Redirection
	router.GET("/:shortCode", RedirectHandler(linkService))
}

// HealthCheckHandler gère la route /health pour vérifier l'état du service
func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// CreateLinkRequest représente le corps de la requête JSON pour la création d'un lien
type CreateLinkRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

// CreateShortLinkHandler gère la création d'une URL courte
func CreateShortLinkHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateLinkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL invalide"})
			return
		}

		link, err := linkService.CreateLink(req.LongURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du lien"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"short_code": link.ShortCode,
			"long_url":   link.LongURL,
		})
	}
}

// RedirectHandler gère la redirection des URLs courtes vers leurs URLs longues
func RedirectHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortCode := c.Param("shortCode")
		log.Printf("[DEBUG] Tentative de redirection pour le code court: %s", shortCode)

		link, err := linkService.GetLinkByShortCode(shortCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Code court non trouvé"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du lien"})
			return
		}

		// Créer un événement de clic
		clickEvent := models.ClickEvent{
			LinkID:    link.ID,
			UserAgent: c.Request.UserAgent(),
			IPAddress: c.ClientIP(),
			Timestamp: time.Now(),
		}

		log.Printf("[DEBUG] Envoi d'un événement de clic pour le lien ID %d (code: %s)", link.ID, shortCode)

		// Envoyer l'événement de clic au channel de manière non bloquante
		select {
		case ClickEventsChannel <- clickEvent:
			log.Printf("[DEBUG] Événement de clic envoyé avec succès pour le lien %s", shortCode)
		default:
			log.Printf("[WARN] Channel de clics plein, événement ignoré pour le lien %s", shortCode)
		}

		c.Redirect(http.StatusFound, link.LongURL)
	}
}

// GetLinkStatsHandler gère la récupération des statistiques d'un lien
func GetLinkStatsHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortCode := c.Param("shortCode")
		log.Printf("[DEBUG] Récupération des statistiques pour le code court: %s", shortCode)

		link, totalClicks, err := linkService.GetLinkStats(shortCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Code court non trouvé"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des statistiques"})
			return
		}

		log.Printf("[DEBUG] Statistiques récupérées pour %s : %d clics", shortCode, totalClicks)

		c.JSON(http.StatusOK, gin.H{
			"short_code":   link.ShortCode,
			"long_url":     link.LongURL,
			"total_clicks": totalClicks,
			"created_at":   link.CreatedAt,
		})
	}
}
