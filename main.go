package main

import (
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tavo-wasd-gh/webapp/config"
	"github.com/tavo-wasd-gh/webapp/database"
	"github.com/tavo-wasd-gh/webtoolkit/cors"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
	"github.com/tavo-wasd-gh/webtoolkit/views"
)

type App struct {
	Production bool
	Views      map[string]*template.Template
	Log        *logger.Logger
	DB         *sql.DB
	Secret     string
}

//go:embed public/*
var publicFS embed.FS

//go:embed templates/*.html
var viewFS embed.FS

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize config: %v", err))
	}

	views, err := views.Init(viewFS, config.ViewMap, config.FuncMap)
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize views: %v", err))
	}

	db, err := database.Init(cfg.DBConnDvr, cfg.DBConnStr)
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize database: %v", err))
	}
	defer db.Close()

	app := &App{
		Production: cfg.Production,
		Views:      views,
		Log:        &logger.Logger{Enabled: cfg.Debug},
		DB:         db,
		Secret:     cfg.Secret,
	}

	http.HandleFunc("/dashboard", app.handleDashboard)

	staticFiles, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to create sub filesystem: %v", err))
	}

	http.Handle("/", http.FileServer(http.FS(staticFiles)))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		app.Log.Printf("starting on :%s...", cfg.Port)

		if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
			log.Fatalf("%v", logger.Errorf("failed to start server: %v", err))
		}
	}()

	<-stop
}

func (app *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	app.Log.Printf("New dashboard request")

	type Data struct {
		Title   string
		Message string
	}

	data := Data{
		Title:   "Dashboard",
		Message: "Welcome to your dashboard!",
	}

	err := app.Views["dashboard"].Execute(w, data)
	if err != nil {
		app.Log.Errorf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
