package fibervhosts

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test for AddSubApp and GetSubApp functionality.
func TestVhostsManager_AddGetSubApp(t *testing.T) {
	manager := NewVhostsManager()
	app := fiber.New()
	hostname := "example.com"
	manager.AddHostname(hostname, app)

	retrieved, exists := manager.GetHostname(hostname)
	assert.True(t, exists)
	assert.Equal(t, app, retrieved)
}

// Test VhostMiddleware for an existing hostname.
func TestVhostMiddleware_ExistingHost(t *testing.T) {
	// Create a dummy sub-app with a simple handler.
	subApp := fiber.New()
	subApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("subapp response")
	})

	manager := NewVhostsManager()
	hostname := "example.com"
	manager.AddHostname(hostname, subApp)

	// Create main app with middleware.
	mainApp := fiber.New()
	mainApp.Use(VhostMiddleware(manager))

	// Create test request with Host header set to example.com.
	req := httptest.NewRequest("GET", "/", nil)
	req.Host = hostname
	resp, err := mainApp.Test(req)
	assert.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "subapp response", string(body))
}

// Test VhostMiddleware for a non-existing hostname.
func TestVhostMiddleware_NonExistingHost(t *testing.T) {
	manager := NewVhostsManager()
	mainApp := fiber.New()
	mainApp.Use(VhostMiddleware(manager))

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "nonexistent.com"
	resp, err := mainApp.Test(req)
	assert.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// Test GetHostnames functionality.
func TestVhostsManager_GetHostnames(t *testing.T) {
	manager := NewVhostsManager()
	app1 := fiber.New()
	app2 := fiber.New()
	app3 := fiber.New()
	manager.AddHostname("example1.com", app1)
	manager.AddHostname("example2.com", app2)
	manager.AddHostname("example3.com", app3)

	hostnames := manager.GetHostnames()
	assert.ElementsMatch(t, []string{"example1.com", "example2.com", "example3.com"}, hostnames)
}
