package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/api"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/monitor"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/axellelanca/urlshortener/internal/workers"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// RunServerCmd représente la commande 'run-server' de Cobra.
var RunServerCmd = &cobra.Command{
	Use:   "run-server",
	Short: "Lance le serveur API de raccourcissement d'URLs et les processus de fond.",
	Long: `Cette commande initialise la base de données, configure les APIs,
démarre les workers asynchrones pour les clics et le moniteur d'URLs,
puis lance le serveur HTTP.`,
	Run: func(cobraCmd *cobra.Command, args []string) {
		if cmd.Cfg == nil {
			log.Fatal("FATAL: La configuration n'est pas initialisée")
		}

		// Initialiser la connexion à la base de données SQLite
		db, err := gorm.Open(sqlite.Open(cmd.Cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Échec de la connexion à la base de données: %v", err)
		}

		// Initialiser les repositories
		linkRepo := repository.NewLinkRepository(db)
		clickRepo := repository.NewClickRepository(db)

		log.Println("Repositories initialisés.")

		// Initialiser les services métiers
		linkService := services.NewLinkService(linkRepo)

		log.Println("Services métiers initialisés.")

		// Initialiser le channel des événements de clic et lancer les workers
		clickEvents := make(chan models.ClickEvent, cmd.Cfg.Analytics.BufferSize)
		workers.StartClickWorkers(clickEvents, clickRepo, cmd.Cfg.Analytics.WorkerCount)

		log.Printf("Channel d'événements de clic initialisé avec un buffer de %d. %d worker(s) de clics démarré(s).",
			cmd.Cfg.Analytics.BufferSize, cmd.Cfg.Analytics.WorkerCount)

		// Initialiser et lancer le moniteur d'URLs
		monitorInterval := time.Duration(cmd.Cfg.Monitor.IntervalMinutes) * time.Minute
		urlMonitor := monitor.NewUrlMonitor(linkRepo, monitorInterval)
		go urlMonitor.Start()
		log.Printf("Moniteur d'URLs démarré avec un intervalle de %v.", monitorInterval)

		// Configurer le routeur Gin et les handlers API
		router := gin.Default()
		api.SetupRoutes(router, linkService, cmd.Cfg.Analytics.BufferSize)

		log.Println("Routes API configurées.")

		// Créer le serveur HTTP Gin
		serverAddr := fmt.Sprintf(":%d", cmd.Cfg.Server.Port)
		srv := &http.Server{
			Addr:    serverAddr,
			Handler: router,
		}

		// Gérer l'arrêt gracieux
		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			log.Println("Arrêt du serveur...")
			if err := srv.Close(); err != nil {
				log.Printf("Erreur lors de la fermeture du serveur: %v", err)
			}
		}()

		// Démarrer le serveur
		log.Printf("Serveur démarré sur %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Erreur lors du démarrage du serveur: %v", err)
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(RunServerCmd)
}
