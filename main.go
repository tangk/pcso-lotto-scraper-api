package main

import (
	"log"
	"net/http"
	"time"

	"pcso-lotto-scraper-api/internal/api"
	"pcso-lotto-scraper-api/internal/database"
	"pcso-lotto-scraper-api/internal/scraper"

	"github.com/robfig/cron/v3"
)

func main() {
	db := database.InitDB()
	defer db.Close()

	api.SetDB(db)

	// Function to scrape and insert results into the database
	scrapeAndInsert := func() {
		log.Println("Running scrape...")

		fromDate, err := database.QueryLatestDate(db)
		if err != nil || fromDate.IsZero() {
			log.Println("Failed to query latest date:", err)
			return
		}

		var results []scraper.LottoResult

		if fromDate.Equal(time.Now()) {
			log.Println("Scraping daily results...")
			results, err = scraper.FetchDaily()
		} else {
			log.Println("Scraping to latest results...")
			results, err = scraper.FetchToLatest(fromDate)
		}

		if err != nil {
			log.Println("Failed to scrape:", err)
			return
		}

		// Insert results into the database
		for _, result := range results {
			database.InsertLottoResults(db, []map[string]string{
				{
					"gameType":    result.GameType,
					"combination": result.Combination,
					"drawDate":    result.DrawDate,
					"jackpot":     result.Jackpot,
					"winners":     result.Winners,
				},
			})
		}

		log.Println("Scrape and database insertion completed.")
	}

	// Run scraper 10 seconds after the main function starts
	time.AfterFunc(10*time.Second, scrapeAndInsert)

	// Schedule the cron job
	c := cron.New(cron.WithLocation(time.FixedZone("PHT", 8*60*60))) // PH time (UTC+8)
	_, err := c.AddFunc("0-30/5 14 * * *", scrapeAndInsert)
	if err != nil {
		log.Fatal("Failed to schedule cron job:", err)
	}
	_, err = c.AddFunc("0-30/5 16 * * *", scrapeAndInsert)
	if err != nil {
		log.Fatal("Failed to schedule cron job:", err)
	}
	_, err = c.AddFunc("0-45/5 21 * * *", scrapeAndInsert)
	if err != nil {
		log.Fatal("Failed to schedule cron job:", err)
	}

	c.Start()

	// Start REST API server
	http.HandleFunc("/results", api.GetLottoResults)
	http.HandleFunc("/heatmap", api.GetHeatmap)
	http.HandleFunc("/game-types", api.GetGameTypes)
	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
