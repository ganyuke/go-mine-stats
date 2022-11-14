package main

import (
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"go-mine-stats/src/stats"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {

	config.Config_file = config.Load_config()

	MaxRespLimit := config.Config_file.API.MaxRespLimit
	MaxRespString := strconv.Itoa(MaxRespLimit)
	DefaultRespString := strconv.Itoa(config.Config_file.API.DefaultRespLimit)

	stats.Poll_official = stats.Init_poll_official()

	if _, err := os.Stat("./stats.db"); err != nil {
		db.Init_db()

		stats.Collect_all_stats(true)
	} else {
		stats.Collect_all_stats(false)
	}

	app := fiber.New(fiber.Config{
		Immutable: true,
	})

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
					statistic := db.Get_stats_for_category(c.Params("category"), c.Query("date"), "DESC", limit)
					return c.JSON(statistic)
				case sort == "max":
					statistic := db.Get_stats_for_category(c.Params("category"), c.Query("date"), "DESC", limit)
					return c.JSON(statistic)
				case sort == "min":
					statistic := db.Get_stats_for_category(c.Params("category"), c.Query("date"), "ASC", limit)
					return c.JSON(statistic)
				case sort != "":
					return c.SendString("Error: invalid 'sort' parameter.")
				}
			} else if c.Query("uuid") != "" && c.Query("stat") != "" { // Return specific statistic data from player
				statistic, err := db.Retrieve_player_stat(c.Query("uuid"), c.Params("category"), c.Query("stat"))
				if err != nil {
					return c.SendString("Error: row not found.")
				}
				return c.JSON(statistic)
			} else if c.Query("sort") != "" && c.Query("stat") != "" { // Return all players' data in specific statistic
				switch {
				default: // hopefully prevent SQL injection
					statistic := db.Get_extrema(c.Params("category"), c.Query("stat"), c.Query("date"), "DESC", limit)
					return c.JSON(statistic)
				case sort == "max":
					statistic := db.Get_extrema(c.Params("category"), c.Query("stat"), c.Query("date"), "DESC", limit)
					return c.JSON(statistic)
				case sort == "min":
					statistic := db.Get_extrema(c.Params("category"), c.Query("stat"), c.Query("date"), "ASC", limit)
					return c.JSON(statistic)
				case sort != "":
					return c.SendString("Error: invalid 'sort' parameter.")
				}
			}
		} else { // Return all statistics in all categories
			return c.SendString("Assume all stats")
		}
		return c.SendString("Bruh")
	})

	app.Listen(":3000")
}
