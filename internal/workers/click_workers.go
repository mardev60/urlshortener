package workers

import (
	"log"

	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/repository"
)

// StartClickWorkers lance un nombre spécifié de workers pour traiter les événements de clic.
// Chaque worker écoute le channel d'événements et enregistre les clics en base de données.
func StartClickWorkers(clickEvents chan models.ClickEvent, clickRepo repository.ClickRepository, workerCount int) {
	log.Printf("[WORKERS] Démarrage de %d workers pour le traitement des clics...", workerCount)

	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			log.Printf("[WORKERS] Worker %d démarré et en attente d'événements", workerID)
			for event := range clickEvents {
				processClickEvent(event, clickRepo, workerID)
			}
		}(i)
	}
}

// processClickEvent traite un événement de clic individuel.
func processClickEvent(event models.ClickEvent, clickRepo repository.ClickRepository, workerID int) {
	log.Printf("[WORKERS] Worker %d : Traitement d'un clic pour le lien ID %d (IP: %s, UA: %s)",
		workerID, event.LinkID, event.IPAddress, event.UserAgent)

	click := models.Click{
		LinkID:    event.LinkID,
		Timestamp: event.Timestamp,
		UserAgent: event.UserAgent,
		IPAddress: event.IPAddress,
	}

	if err := clickRepo.CreateClick(&click); err != nil {
		log.Printf("[WORKERS] Worker %d : ERREUR lors de l'enregistrement du clic pour le lien ID %d : %v",
			workerID, event.LinkID, err)
		return
	}

	log.Printf("[WORKERS] Worker %d : Clic enregistré avec succès pour le lien ID %d (Click ID: %d)",
		workerID, event.LinkID, click.ID)
}
