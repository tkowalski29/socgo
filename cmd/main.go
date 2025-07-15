package main

import (
	"log"
	"net/http"

	"github.com/tkowalski/socgo/internal/config"
	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/di"
	"github.com/tkowalski/socgo/internal/oauth"
	"github.com/tkowalski/socgo/internal/providers"
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

	srv := server.New(container)
	container.Register("server", srv)

	addr := cfg.GetServerAddr()
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}
