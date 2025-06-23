package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// shortCodeFlag stocke la valeur du flag --code
var shortCodeFlag string

// StatsCmd représente la commande 'stats'
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Affiche les statistiques (nombre de clics) pour un lien court.",
	Long: `Cette commande permet de récupérer et d'afficher le nombre total de clics
pour une URL courte spécifique en utilisant son code.

Exemple:
  url-shortener stats --code="xyz123"`,
	Run: func(cobraCmd *cobra.Command, args []string) {
		if shortCodeFlag == "" {
			fmt.Println("Erreur: Le flag --code est requis")
			os.Exit(1)
		}

		if cmd.Cfg == nil {
			log.Fatal("FATAL: La configuration n'est pas initialisée")
		}

		// Initialiser la connexion à la base de données SQLite
		db, err := gorm.Open(sqlite.Open(cmd.Cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("FATAL: Échec de la connexion à la base de données: %v", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la base de données SQL sous-jacente: %v", err)
		}
		defer sqlDB.Close()

		// Initialiser les repositories et services
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// Récupérer les statistiques
		link, totalClicks, err := linkService.GetLinkStats(shortCodeFlag)
		if err != nil {
			log.Fatalf("FATAL: Erreur lors de la récupération des statistiques: %v", err)
		}

		fmt.Printf("Statistiques pour le code court: %s\n", link.ShortCode)
		fmt.Printf("URL longue: %s\n", link.LongURL)
		fmt.Printf("Total de clics: %d\n", totalClicks)
	},
}

func init() {
	StatsCmd.Flags().StringVar(&shortCodeFlag, "code", "", "Code court de l'URL à analyser")
	StatsCmd.MarkFlagRequired("code")
	cmd.RootCmd.AddCommand(StatsCmd)
}
