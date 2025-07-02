package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"ozbarginscraper.com/helpers"

	"github.com/robfig/cron"
)

const (
	ComputingTopDealsURL = "https://www.ozbargain.com.au/ozbapi/block/ozbdeal_top?dur=30&tid=12"
	ComputingNewDealsURL = "https://www.ozbargain.com.au/ozbapi/block/ozbdeal_new?tid=12&f=1"
	ElectronicsTopDealsURL = "https://www.ozbargain.com.au/ozbapi/block/ozbdeal_top?dur=30&tid=13"
	ElectronicsNewDealsURL = "https://www.ozbargain.com.au/ozbapi/block/ozbdeal_new?tid=13&f=1"
)

// Response structure for API responses
type Response struct {
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// Health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := Response{
		Message:   "OzBargain Scraper is running",
		Timestamp: time.Now(),
	}
	
	json.NewEncoder(w).Encode(response)
}

// Background scraping functions
func ScrapeTopDeals() {
	log.Println("Scraping top deals...")
	compResp, err := http.Get(ComputingTopDealsURL)
	if err != nil {
		log.Printf("Error scraping top deals: %v", err)
		return
	}
	defer compResp.Body.Close()

	deals, err := helpers.ParseDeals(compResp, "Computing Top Deals")
	if err != nil {
		log.Printf("Error parsing deals: %v", err)
		return
	}

	log.Printf("Found %d deals", len(deals))
	helpers.FormatAndPostToDiscord(deals, false)

	elecResp, err := http.Get(ElectronicsTopDealsURL)
	if err != nil {
		log.Printf("Error scraping top deals: %v", err)
		return
	}
	defer elecResp.Body.Close()
	
	deals, err = helpers.ParseDeals(elecResp, "Electronics Top Deals")
	if err != nil {
		log.Printf("Error parsing deals: %v", err)
		return
	}

	log.Printf("Found %d deals", len(deals))
	helpers.FormatAndPostToDiscord(deals, false)
}

func ScrapeNewDeals() {
	log.Println("Scraping new deals...")
	compResp, err := http.Get(ComputingNewDealsURL)
	if err != nil {
		log.Printf("Error scraping new deals: %v", err)
		return
	}
	defer compResp.Body.Close()

	deals, err := helpers.ParseDeals(compResp, "Computing New Deals")
	if err != nil {
		log.Printf("Error parsing deals: %v", err)
		return
	}

	log.Printf("Found %d deals", len(deals))
	helpers.FormatAndPostToDiscord(deals, true)

	elecResp, err := http.Get(ElectronicsNewDealsURL)
	if err != nil {
		log.Printf("Error scraping new deals: %v", err)
		return
	}
	defer elecResp.Body.Close()

	deals, err = helpers.ParseDeals(elecResp, "Electronics New Deals")
	if err != nil {
		log.Printf("Error parsing deals: %v", err)
		return
	}

	log.Printf("Found %d deals", len(deals))
	helpers.FormatAndPostToDiscord(deals, true)
}

// Start scheduled scrapers using cron
func StartScheduledScrapers() {
	c := cron.New()
	
	// Schedule top deals scraper to run daily at 9:00 AM
	err := c.AddFunc("0 0 9 * * *", func() {
		log.Println("Running scheduled top deals scrape...")
		ScrapeTopDeals()
	})
	if err != nil {
		log.Printf("Error scheduling top deals scraper: %v", err)
	}
	
	// Schedule new deals scraper to run every 6 hours
	err = c.AddFunc("0 0 */6 * * *", func() {
		log.Println("Running scheduled new deals scrape...")
		ScrapeNewDeals()
	})
	if err != nil {
		log.Printf("Error scheduling new deals scraper: %v", err)
	}
	
	// Start the cron scheduler
	c.Start()
	
	log.Println("Scheduled scrapers started with cron:")
	log.Println("  - Top deals: Daily at 9:00 AM")
	log.Println("  - New deals: Every 6 hours")
	
	// Run scrapers once immediately on startup
	log.Println("Running initial scrapes...")
	go ScrapeTopDeals()
	go ScrapeNewDeals()
}