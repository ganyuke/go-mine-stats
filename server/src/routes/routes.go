package routes

import (
	"go-mine-stats/src/config"
	"log"

	"github.com/gofiber/fiber/v2"
)

func InitRoutes() {
	app := fiber.New()

	api := app.Group("/api")

	v1 := api.Group("/v1")

	v1.Get("/stats/:category", func(c *fiber.Ctx) error {
		return guidePlayerStatistic(c)
	})

	v1.Get("/stats/:category/summary", func(c *fiber.Ctx) error {
		return guideAggregate(c)
	})

	v1.Get("/stats/:category/top", func(c *fiber.Ctx) error {
		return guideTopStatistic(c)
	})

	v1.Get("/users", func(c *fiber.Ctx) error {
		return guideUsernames(c)
	})

	log.Fatal(app.Listen(config.Config_file.API.Port))
}
