package routes

import (
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func InitRoutes() {
	app := fiber.New()

	maxRespLimit := config.Config_file.API.MaxRespLimit
	maxRespString := strconv.Itoa(maxRespLimit)
	defaultRespString := strconv.Itoa(config.Config_file.API.DefaultRespLimit)
	default_world := config.Config_file.API.DefaultWorld

	api := app.Group("/api")

	v1 := api.Group("/v1")

	v1.Get("/stats/:category", func(c *fiber.Ctx) error { // TODO: Make not awful to look at
		sort := c.Query("sort")
		limit := c.Query("limit", defaultRespString)
		limitNum, _ := strconv.Atoi(c.Query("limit"))
		if limitNum > maxRespLimit {
			limit = maxRespString
		}
		if c.Params("category") != "" {
			if c.Query("stat") == "all" { // Return statistics data in a given category
				switch {
				default: // hopefully prevent SQL injection
					statistic := db.GetStatsForCategory(c.Params("category"), c.Query("world", default_world), "DESC", limit)
					return c.JSON(statistic)
				case sort == "max":
					statistic := db.GetStatsForCategory(c.Params("category"), c.Query("world", default_world), "DESC", limit)
					return c.JSON(statistic)
				case sort == "min":
					statistic := db.GetStatsForCategory(c.Params("category"), c.Query("world", default_world), "ASC", limit)
					return c.JSON(statistic)
				case sort != "":
					return c.SendString("Error: invalid 'sort' parameter.")
				}
			} else if c.Query("uuid") != "" && c.Query("stat") != "" { // Return specific statistic data from player
				if c.Query("start_date") != "" && c.Query("end_date") != "" {
					statistic := db.GetStatDateRange(c.Query("uuid"), c.Params("category"), c.Query("stat"), c.Query("world", default_world), c.Query("start_date"), c.Query("end_date"))
					return c.JSON(statistic)
				} else {
					statistic, err := db.RetrievePlayerStat(c.Query("uuid"), c.Params("category"), c.Query("stat"), c.Query("world", default_world))
					if err != nil {
						return c.SendString("Error: row not found.")
					}
					return c.JSON(statistic)
				}
			} else if c.Query("stat") != "" { // Return all players' data in specific statistic
				switch {
				default: // hopefully prevent SQL injection
					statistic := db.GetExtrema(c.Params("category"), c.Query("stat"), c.Query("world", default_world), "DESC", limit)
					return c.JSON(statistic)
				case sort == "max":
					statistic := db.GetExtrema(c.Params("category"), c.Query("stat"), c.Query("world", default_world), "DESC", limit)
					return c.JSON(statistic)
				case sort == "min":
					statistic := db.GetExtrema(c.Params("category"), c.Query("stat"), c.Query("world", default_world), "ASC", limit)
					return c.JSON(statistic)
				case sort != "":
					return c.SendString("Error: invalid 'sort' parameter.")
				}
			}
		}
		return c.SendString("Error: no valid query provided.")
	})

	v1.Get("/stats/:category/summary", func(c *fiber.Ctx) error { // TODO: Make not awful to look at
		sort := c.Query("sort")
		limit := c.Query("limit", defaultRespString)
		limitNum, _ := strconv.Atoi(c.Query("limit"))
		if limitNum > maxRespLimit {
			limit = maxRespString
		}
		if c.Query("stat") == "all" { // Return sum of
			switch {
			default: // hopefully prevent SQL injection
				statistic := db.GetTotalStatsForCategory(c.Params("category"), c.Query("world", default_world), "DESC", limit)
				return c.JSON(statistic)
			case sort == "max":
				statistic := db.GetTotalStatsForCategory(c.Params("category"), c.Query("world", default_world), "DESC", limit)
				return c.JSON(statistic)
			case sort == "min":
				statistic := db.GetTotalStatsForCategory(c.Params("category"), c.Query("world", default_world), "ASC", limit)
				return c.JSON(statistic)
			case sort != "":
				return c.SendString("Error: invalid 'sort' parameter.")
			}
		} else if c.Query("stat") != "" { // Return sum of specific statistic
			statistic, err := db.RetrieveTotalStat(c.Params("category"), c.Query("stat"), c.Query("world", default_world))
			if err != nil {
				return c.SendString("Error: row not found.")
			}
			return c.JSON(statistic)
		} else {
			statistic, err := db.RetrieveTotalCategory(c.Params("category"), c.Query("world", default_world))
			if err != nil {
				return c.SendString("Error: row not found.")
			}
			return c.JSON(statistic)
		}
	})

	app.Listen(":3000")
}
