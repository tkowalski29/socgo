package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tkowalski/socgo/internal/data/config"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/dependency"
	"github.com/tkowalski/socgo/internal/service/notifications"
	"github.com/tkowalski/socgo/internal/service/oauth"
	"github.com/tkowalski/socgo/internal/service/post"
	"github.com/tkowalski/socgo/internal/service/scheduler"
	"github.com/tkowalski/socgo/internal/service/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Log configuration info
	log.Printf("Server will start on: %s", cfg.GetServerAddr())
	log.Printf("Base URL for OAuth: %s", cfg.Server.BaseURL)

	container := dependency.NewContainer()
	container.Register("config", cfg)

	dbManager := database.NewManager(cfg.Database.DataDir)
	container.Register("database", dbManager)

	oauthService := oauth.NewService(dbManager, cfg)
	container.Register("oauth_service", oauthService)

	providerService := post.NewProviderService(dbManager, oauthService)
	container.Register("provider_service", providerService)

	notificationService := notifications.NewService(dbManager)
	container.Register("notification_service", notificationService)

	// Create and start scheduler
	schedulerService := scheduler.New(dbManager, providerService, notificationService)
	container.Register("scheduler", schedulerService)
	schedulerService.Start()

	srv := server.New(container)
	container.Register("server", srv)

	// Setup graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := cfg.GetServerAddr()
		log.Printf("Starting server on %s", addr)
		if err := http.ListenAndServe(addr, srv); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for shutdown signal
	<-stopChan
	log.Println("Shutting down...")
	schedulerService.Stop()
	log.Println("Goodbye!")
}
