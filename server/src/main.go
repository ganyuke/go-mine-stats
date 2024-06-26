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
	var migrationApproved bool
	flag.BoolVar(&migrationApproved, "migrate", false, "If a database migration is available, run it.")

	flag.Parse()

	// Read config.json and make it accessible to go-mine-stats.
	config.Config_file = config.LoadConfig(configPath)
	if config.Config_file.CheckSanity() {
		log.Println("Config passed sanity check.")
	}

	if config.Config_file.Scan.Enabled {
		// Create the object that holds file metadata to check against
		stats.InitFileTracking()

		// Create polling object to perodically check stats
		stats.Poll_official = stats.InitPollOfficial()
	}

	// Create database and log all initial player stats
	if _, err := os.Stat(dbPath); err != nil {
		db.Monika = db.DbConnect(true, dbPath)

		if config.Config_file.Scan.Enabled {
			stats.CollectAllStats(true)
		}
	} else {
		db.Monika = db.DbConnect(false, dbPath)

		// Check for migrations
		version := db.CheckPragma()
		if version < 1 {
			log.Println("DB schema is outdated. Please make a backup if you haven't already!")
			if migrationApproved {
				db.RunMigration()
				return
			} else {
				log.Fatal("Migration not approved. Run the flag `-migrate true` once you have created a backup.")
			}
		}

		if config.Config_file.Scan.Enabled {
			stats.CollectAllStats(true)
		}
	}

	// Begin the webserver
	routes.InitRoutes()
}
