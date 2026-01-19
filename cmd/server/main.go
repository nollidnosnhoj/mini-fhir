package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"mini-fhir/internal/api"
	"mini-fhir/internal/fhir/dstu3"
	"mini-fhir/internal/search"
	"mini-fhir/internal/store"
	"mini-fhir/internal/validation"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	fhirVersion := flag.String("fhir-version", "dstu3", "FHIR version (dstu3)")
	seedGlob := flag.String("seed", "", "Seed data glob pattern")
	seedStrict := flag.Bool("seed-strict", true, "Fail on seed validation errors")
	profileCache := flag.String("profile-cache", ".fhir-cache", "Directory for StructureDefinition cache")
	profileCacheTTL := flag.Duration("profile-cache-ttl", 24*time.Hour, "Cache TTL for StructureDefinitions")
	profileCacheVersion := flag.Int("profile-cache-version", validation.CacheVersion, "Cache version for StructureDefinitions")
	flag.Parse()

	if *fhirVersion != "dstu3" {
		log.Fatalf("unsupported fhir-version: %s", *fhirVersion)
	}

	registry := dstu3.NewRegistry()
	profileStore := validation.NewProfileStore(*profileCache, *profileCacheTTL, *profileCacheVersion)
	if err := profileStore.LoadDefaults(context.Background(), registry); err != nil {
		log.Fatalf("profile load failed: %v", err)
	}
	validator := validation.NewValidator(registry, profileStore)
	store := store.NewStore()
	searcher := search.NewSearcher(registry, store)

	if *seedGlob != "" {
		if err := api.LoadSeed(*seedGlob, *seedStrict, registry, validator, store); err != nil {
			log.Fatalf("seed load failed: %v", err)
		}
	}

	e := echo.New()
	e.Server.ReadHeaderTimeout = 10 * time.Second
	e.HideBanner = true
	e.HidePort = true

	api.RegisterRoutes(e, registry, validator, store, searcher)

	go func() {
		log.Printf("listening on %s", *addr)
		if err := e.Start(*addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
