package main

import (
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"go-mine-stats/src/stats"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {

	config.Config_file = config.LoadConfig()

	if config.SanityCheck(config.Config_file) {
		log.Println("Config passed sanity check.")
	}

	MaxRespLimit := config.Config_file.API.MaxRespLimit
	MaxRespString := strconv.Itoa(MaxRespLimit)
	DefaultRespString := strconv.Itoa(config.Config_file.API.DefaultRespLimit)

	stats.Poll_official = stats.Init_poll_official()

	if _, err := os.Stat("./stats.db"); err != nil {
		db.Monika = db.DbConnect(true)

		stats.CollectAllStats(true)
	} else {
		db.Monika = db.DbConnect(false)

		stats.CollectAllStats(false)
	}

	app := fiber.New()

	api := app.Group("/api")

	v1 := api.Group("/v1")

	v1.Get("/stats/:category", func(c *fiber.Ctx) error { // TODO: Make not awful to look at
		sort := c.Query("sort")
		limit := c.Query("limit", DefaultRespString)
		limitNum, _ := strconv.Atoi(c.Query("limit"))
		if limitNum > MaxRespLimit {
			limit = MaxRespString
		}
		if c.Params("category") != "" {
			if c.Query("stat") == "all" { // Return statistics data in a given category
				switch {
				default: // hopefully prevent SQL injection
					statistic := db.GetStatsForCategory(c.Params("category"), c.Query("world"), "DESC", limit)
					return c.JSON(statistic)
				case sort == "max":
					statistic := db.GetStatsForCategory(c.Params("category"), c.Query("world"), "DESC", limit)
					return c.JSON(statistic)
				case sort == "min":
					statistic := db.GetStatsForCategory(c.Params("category"), c.Query("world"), "ASC", limit)
					return c.JSON(statistic)
				case sort != "":
					return c.SendString("Error: invalid 'sort' parameter.")
				}
			} else if c.Query("uuid") != "" && c.Query("stat") != "" { // Return specific statistic data from player
				statistic, err := db.RetrievePlayerStat(c.Query("uuid"), c.Params("category"), c.Query("stat"), c.Query("world"))
				if err != nil {
					return c.SendString("Error: row not found.")
				}
				return c.JSON(statistic)
			} else if c.Query("sort") != "" && c.Query("stat") != "" { // Return all players' data in specific statistic
				switch {
				default: // hopefully prevent SQL injection
					statistic := db.GetExtrema(c.Params("category"), c.Query("stat"), c.Query("world"), "DESC", limit)
					return c.JSON(statistic)
				case sort == "max":
					statistic := db.GetExtrema(c.Params("category"), c.Query("stat"), c.Query("world"), "DESC", limit)
					return c.JSON(statistic)
				case sort == "min":
					statistic := db.GetExtrema(c.Params("category"), c.Query("stat"), c.Query("world"), "ASC", limit)
					return c.JSON(statistic)
				case sort != "":
					return c.SendString("Error: invalid 'sort' parameter.")
				}
			}
		}
		return c.SendString("Bruh")
	})

	app.Listen(":3000")
}
