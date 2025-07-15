package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tkowalski/socgo/internal/config"
	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/di"
	"github.com/tkowalski/socgo/internal/oauth"
	"github.com/tkowalski/socgo/internal/providers"
	"github.com/tkowalski/socgo/internal/scheduler"
	"github.com/tkowalski/socgo/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	container := di.NewContainer()
	container.Register("config", cfg)

	dbManager := database.NewManager(cfg.Database.DataDir)
	container.Register("database", dbManager)

	oauthService := oauth.NewService(dbManager)
	container.Register("oauth_service", oauthService)

	providerService := providers.NewProviderService(dbManager, oauthService)
	container.Register("provider_service", providerService)

	// Create and start scheduler
	jobScheduler := scheduler.New(dbManager, providerService)
	container.Register("scheduler", jobScheduler)
	jobScheduler.Start()

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
	jobScheduler.Stop()
	log.Println("Goodbye!")
}
