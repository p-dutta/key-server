package router

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"key-server/handler"
	"net/http"
	"os"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {

	api := app.Group(os.Getenv("ROUTE_PREFIX") + "/" + os.Getenv("API_VERSION"))
	drm := api.Group("/drm")

	drm.Get("/health", HealthCheckHandler)
	drm.Post("/key/:contentId/:packageId", handler.GetKey)
	drm.Post("/key", handler.GenerateStaticKey)

}

func HealthCheckHandler(c *fiber.Ctx) error {
	// Create a map representing the JSON response
	response := map[string]string{
		"status":  "ok",
		"message": "Service is healthy",
	}

	// Marshal the map into JSON format
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal server error")
	}

	// Set the Content-Type header and write the JSON response
	c.Set("Content-Type", "application/json")
	return c.Status(http.StatusOK).Send(jsonResponse)
}
