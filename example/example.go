package main

import (
	fibervhosts "github.com/boomhut/fiber-vhosts2"
	"github.com/gofiber/fiber/v2"
)

// Example usage of the VhostsManager
func main() {
	// Create main app
	app := fiber.New()

	// Create vhosts manager with config
	manager := fibervhosts.NewVhostsManager(fibervhosts.Config{
		EnableLogging:    true,
		RecoverFromPanic: true,
	})

	// Create and add apps for different hostnames
	app1 := fiber.New()
	app1.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("App 1")
	})
	manager.AddHostname("app1.example.com", app1)

	// Add wildcard subdomain support
	wildApp := fiber.New()
	wildApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Wildcard subdomain")
	})
	manager.AddHostname("*.example.com", wildApp)

	// Create default app
	defaultApp := fiber.New()
	defaultApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Default app")
	})
	manager.SetDefaultApp(defaultApp)

	// Use the vhost middleware
	app.Use(fibervhosts.VhostMiddleware(manager))

	// Start server
	app.Listen(":3000")
}
