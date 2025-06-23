package workers

import (
	"log"
	"time"

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository"
)

// StartClickWorkers lance un nombre spécifié de workers pour traiter les événements de clic.
// Chaque worker écoute le channel d'événements et enregistre les clics en base de données.
func StartClickWorkers(clickEvents chan models.ClickEvent, clickRepo repository.ClickRepository, workerCount int) {
	log.Printf("Démarrage de %d workers pour le traitement des clics...", workerCount)

	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			log.Printf("Worker %d démarré", workerID)
			for event := range clickEvents {
				processClickEvent(event, clickRepo, workerID)
			}
		}(i)
	}
}

// processClickEvent traite un événement de clic individuel.
func processClickEvent(event models.ClickEvent, clickRepo repository.ClickRepository, workerID int) {
	click := models.Click{
		LinkID:    event.LinkID,
		Timestamp: time.Now(),
		UserAgent: event.UserAgent,
		IPAddress: event.IPAddress,
	}

	if err := clickRepo.CreateClick(&click); err != nil {
		log.Printf("[Worker %d] Erreur lors de l'enregistrement du clic: %v", workerID, err)
		return
	}

	log.Printf("[Worker %d] Clic enregistré avec succès pour le lien ID %d", workerID, event.LinkID)
}
