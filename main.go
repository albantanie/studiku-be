package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"studi-ku-backend/internal/config"
	"studi-ku-backend/internal/database"
	"studi-ku-backend/internal/handlers"
	"studi-ku-backend/internal/repositories"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "rest":
		runREST()
	case "migrate":
		if len(os.Args) >= 3 && os.Args[2] == "seed" {
			runSQL("scripts/migrate_seed.sql")
			return
		}
		fmt.Printf("unknown command: %s\n\n", strings.Join(os.Args[1:], " "))
		printUsage()
	case "purge":
		runSQL("scripts/purge.sql")
	default:
		fmt.Printf("unknown command: %s\n\n", os.Args[1])
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run main.go rest          Start the Gin REST API server")
	fmt.Println("  go run main.go migrate seed  Create schema and seed data")
	fmt.Println("  go run main.go purge         Drop all public database tables")
}

func runREST() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := gin.Default()
	router.Use(cors(cfg.AllowedOrigins))

	repo := repositories.New(db)
	handler := handlers.New(repo)
	handlers.Register(router, handler)

	log.Printf("REST API listening on :%s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

func runSQL(path string) {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.RunSQLFile(db, path); err != nil {
		log.Fatal(err)
	}

	log.Printf("executed %s", path)
}

func cors(allowedOrigins string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, origin := range strings.Split(allowedOrigins, ",") {
		allowed[strings.TrimSpace(origin)] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
