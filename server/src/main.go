package main

import (
	"flag"
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"go-mine-stats/src/routes"
	"go-mine-stats/src/stats"
	"log"
	"os"
)

func main() {
	// Register flags
	var configPath string
	flag.StringVar(&configPath, "config", "../../config.json", "Path to config.json file.")
	var dbPath string
	flag.StringVar(&dbPath, "dbpath", "./stats.db", "Path to sqlite database.")

	flag.Parse()

	// Read config.json and make it accessible to go-mine-stats.
	config.Config_file = config.LoadConfig(configPath)
	if config.SanityCheck(config.Config_file) {
		log.Println("Config passed sanity check.")
	}

	// Create the object that holds file metadata to check against
	stats.InitFileTracking()

	// Create polling object to perodically check stats
	stats.Poll_official = stats.InitPollOfficial()

	// Create database and log all initial player stats
	if _, err := os.Stat(dbPath); err != nil {
		db.Monika = db.DbConnect(true)

		stats.CollectAllStats(true)
	} else {
		db.Monika = db.DbConnect(false)

		stats.CollectAllStats(false)
	}

	// Begin the webserver
	routes.InitRoutes()
}
