package routes

import (
	"go-mine-stats/src/config"
	"go-mine-stats/src/db"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func guideAggregate(c *fiber.Ctx) error {

	defaultRespString := strconv.Itoa(config.Config_file.API.DefaultRespLimit)
	default_world := config.Config_file.API.DefaultWorld

	limit, err := convertLimit(c.Query("limit", defaultRespString))
	if err != nil {
		return c.SendString("E_LIMIT_PRASE_FAIL")
	}
	sortOrder := convertOrder(c.Query("order"))

	category := c.Params("category")
	statistic := c.Query("stat")
	world := c.Query("world", default_world)

	startDate, _ := parseDate(c.Query("from"))
	endDate, _ := parseDate(c.Query("to", time.Now().Format(time.RFC3339)))

	switch statistic {
	case "all": // Return sum of all stats in a given category (Statistic: "all")
		result := db.GetTotalStatsForCategory(category, world, sortOrder, limit)
		return c.JSON(result)
	case "": // Retrieve total statistic for category (Statistic not specified)
		result, err := db.RetrieveTotalCategory(category, world)
		if err != nil {
			return c.SendString("E_ROW_NOT_FOUND")
		}
		return c.JSON(result)
	default: // (Statistic specified)
		if c.Query("from") != "" { // Return cumulative statistic if given range
			result := db.GetCumulativeStat(category, statistic, world, sortOrder, startDate, endDate)
			return c.JSON(result)
		} else {
			result, err := db.RetrieveTotalStat(category, statistic, world)
			if err != nil { // Return sum of specific statistic
				return c.SendString("E_ROW_NOT_FOUND")
			}
			return c.JSON(result)
		}
	}
}

func guideTopStatistic(c *fiber.Ctx) error {

	defaultRespString := strconv.Itoa(config.Config_file.API.DefaultRespLimit)
	default_world := config.Config_file.API.DefaultWorld

	limit, err := convertLimit(c.Query("limit", defaultRespString))
	if err != nil {
		return c.SendString("E_LIMIT_PRASE_FAIL")
	}
	sortOrder := convertOrder(c.Query("order"))

	category := c.Params("category")
	statistic := c.Query("stat")
	world := c.Query("world", default_world)

	switch statistic {
	case "all": // Return sum of all stats in a given category
		statistic := db.GetStatsForCategory(category, world, sortOrder, limit)
		return c.JSON(statistic)
	case "":
		return c.SendString("E_MISSING_STATISTIC")
	default: // Return the top players in the given statistic
		statistic := db.GetExtrema(category, statistic, world, sortOrder, limit)
		return c.JSON(statistic)
	}
}

func guidePlayerStatistic(c *fiber.Ctx) error {

	default_world := config.Config_file.API.DefaultWorld

	category := c.Params("category")
	statistic := c.Query("stat")
	world := c.Query("world", default_world)
	uuid := c.Query("uuid")

	startDate, _ := parseDate(c.Query("from"))
	endDate, _ := parseDate(c.Query("to", time.Now().Format(time.RFC3339)))

	if uuid == "" {
		return c.SendString("E_MISSING_UUID")
	}

	if statistic == "" {
		return c.SendString("E_MISSING_STATISTIC")
	}

	if config.Config_file.CheckBlacklist(uuid) {
		return c.SendStatus(403)
	}

	if c.Query("from") != "" {
		// If date range specified, do special things
		statistic := db.GetStatDateRange(uuid, category, statistic, world, startDate, endDate)
		return c.JSON(statistic)
	} else {
		// Else, just return the latest statistic value
		statistic, err := db.RetrievePlayerStat(uuid, category, statistic, world)
		if err != nil {
			return c.SendString("E_ROW_NOT_FOUND")
		}
		return c.JSON(statistic)
	}

}

func guideUsernames(c *fiber.Ctx) error {

	uuid := c.Query("uuid")

	if uuid != "" {
		data, err := db.GetUsernameFromUuid(uuid)
		if err != nil {
			return c.SendString("E_UUID_NO_EXIST")
		}
		return c.JSON(data)
	}
	return c.SendString("E_MISSING_UUID.")

}
