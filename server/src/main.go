package main

import (
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"go-mine-stats/src/routes"
	"go-mine-stats/src/stats"
	"log"
	"os"
)

func main() {

	config.Config_file = config.LoadConfig()

	if config.SanityCheck(config.Config_file) {
		log.Println("Config passed sanity check.")
	}

	stats.Poll_official = stats.Init_poll_official()

	if _, err := os.Stat("./stats.db"); err != nil {
		db.Monika = db.DbConnect(true)

		stats.CollectAllStats(true)
	} else {
		db.Monika = db.DbConnect(false)

		stats.CollectAllStats(false)
	}

	routes.InitRoutes()
}
