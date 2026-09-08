package main

import (
	"fmt"
	"log"

	"flickey/go-backend/config"
	"flickey/go-backend/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	// 1. Detach media
	if err := database.Exec("UPDATE media SET listing_id = NULL").Error; err != nil {
		log.Printf("error detaching media: %v", err)
	}

	// 2. Delete listing amenities
	if err := database.Exec("DELETE FROM listing_amenities").Error; err != nil {
		log.Printf("error deleting listing_amenities: %v", err)
	}

	// 3. Delete listings
	resListings := database.Exec("DELETE FROM listings")
	if resListings.Error != nil {
		log.Fatalf("error deleting listings: %v", resListings.Error)
	}
	fmt.Printf("Deleted %d listings\n", resListings.RowsAffected)

	// 4. Delete listing drafts
	resDrafts := database.Exec("DELETE FROM listing_drafts")
	if resDrafts.Error != nil {
		log.Fatalf("error deleting listing_drafts: %v", resDrafts.Error)
	}
	fmt.Printf("Deleted %d listing drafts\n", resDrafts.RowsAffected)

	fmt.Println("All listings and drafts successfully cleared!")
}
