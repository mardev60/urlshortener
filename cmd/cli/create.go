package cli

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// longURLFlag stocke la valeur du flag --url
var longURLFlag string

// CreateCmd représente la commande 'create'
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crée une URL courte à partir d'une URL longue.",
	Long: `Cette commande raccourcit une URL longue fournie et affiche le code court généré.

Exemple:
  url-shortener create --url="https://www.google.com/search?q=go+lang"`,
	Run: func(cobraCmd *cobra.Command, args []string) {
		if longURLFlag == "" {
			fmt.Println("Erreur: Le flag --url est requis")
			os.Exit(1)
		}

		// Validation basique du format de l'URL
		if _, err := url.ParseRequestURI(longURLFlag); err != nil {
			fmt.Printf("Erreur: URL invalide: %v\n", err)
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

		// Créer le lien court
		link, err := linkService.CreateLink(longURLFlag)
		if err != nil {
			log.Fatalf("FATAL: Erreur lors de la création du lien: %v", err)
		}

		fullShortURL := fmt.Sprintf("%s/%s", cmd.Cfg.Server.BaseURL, link.ShortCode)
		fmt.Printf("URL courte créée avec succès:\n")
		fmt.Printf("Code: %s\n", link.ShortCode)
		fmt.Printf("URL complète: %s\n", fullShortURL)
	},
}

func init() {
	CreateCmd.Flags().StringVar(&longURLFlag, "url", "", "URL longue à raccourcir")
	CreateCmd.MarkFlagRequired("url")
	cmd.RootCmd.AddCommand(CreateCmd)
}
