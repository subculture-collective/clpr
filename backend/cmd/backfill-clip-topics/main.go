package main

import (
	"context"
	"flag"
	"log"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
)

func main() {
	batchSize := flag.Int("batch", 250, "clips to classify per batch")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()
	classifier := services.NewTopicClassificationService(repository.NewClipTopicRepository(db.Pool))
	total := 0
	for {
		processed, err := classifier.Backfill(context.Background(), *batchSize)
		if err != nil {
			log.Fatalf("classify topics: %v", err)
		}
		total += processed
		log.Printf("classified %d clips (%d total)", processed, total)
		if processed < *batchSize {
			break
		}
	}
}
