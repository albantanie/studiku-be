package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	case "init":
		if len(os.Args) >= 3 && os.Args[2] == "demo" {
			runSQLSequence(
				"scripts/purge.sql",
				"scripts/migrate_seed.sql",
				"scripts/keep_demo_accounts_only.sql",
			)
			return
		}
		fmt.Printf("unknown command: %s\n\n", strings.Join(os.Args[1:], " "))
		printUsage()
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
	fmt.Println("  go run main.go init demo     Fresh init DB, keep only 4 demo login accounts")
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
	serveFrontend(router, cfg.WebDir)

	log.Printf("REST API listening on :%s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

// serveFrontend menyajikan hasil build React dari satu origin yang sama dengan API.
//
// Aplikasi memakai routing sisi klien, jadi URL seperti /student/courses/3/sessions/5
// tidak punya berkas padanannya di disk. Permintaan yang bukan API dan bukan berkas
// nyata dikembalikan sebagai index.html supaya refresh dan tautan langsung tetap
// membuka halaman yang benar, lalu router di browser yang menentukan tampilannya.
//
// Bila folder build belum ada, fungsi ini tidak melakukan apa-apa: mode pengembangan
// tetap memakai vite dev server yang mem-proxy /api ke sini.
func serveFrontend(router *gin.Engine, webDir string) {
	indexPath := filepath.Join(webDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		log.Printf("frontend build tidak ditemukan di %s, server hanya melayani API", webDir)
		return
	}

	router.Static("/assets", filepath.Join(webDir, "assets"))
	for _, name := range []string{"favicon.ico", "robots.txt", "manifest.json"} {
		path := filepath.Join(webDir, name)
		if _, err := os.Stat(path); err == nil {
			router.StaticFile("/"+name, path)
		}
	}

	router.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path

		// Jalur API harus tetap menjawab 404 sebagai JSON. Tanpa penjagaan ini,
		// salah ketik endpoint akan membalas HTML dan menyesatkan klien.
		if strings.HasPrefix(requestPath, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "endpoint tidak ditemukan"})
			return
		}

		// Berkas statis yang benar-benar ada dilayani apa adanya.
		if requestPath != "/" {
			candidate := filepath.Join(webDir, filepath.Clean(requestPath))
			// filepath.Clean menormalkan "..", lalu dipastikan hasilnya masih di dalam webDir.
			if strings.HasPrefix(candidate, filepath.Clean(webDir)+string(os.PathSeparator)) {
				if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
					c.File(candidate)
					return
				}
			}
		}

		c.File(indexPath)
	})

	log.Printf("menyajikan frontend dari %s", webDir)
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

func runSQLSequence(paths ...string) {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, path := range paths {
		if err := database.RunSQLFile(db, path); err != nil {
			log.Fatalf("failed executing %s: %v", path, err)
		}
		log.Printf("executed %s", path)
	}
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
