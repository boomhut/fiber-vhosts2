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

// Test RemoveHostname functionality.
func TestVhostsManager_RemoveHostname(t *testing.T) {
	manager := NewVhostsManager()
	app := fiber.New()
	hostname := "example.com"

	// Add and confirm
	err := manager.AddHostname(hostname, app)
	assert.NoError(t, err)
	_, exists := manager.GetHostname(hostname)
	assert.True(t, exists)

	// Remove and confirm
	err = manager.RemoveHostname(hostname)
	assert.NoError(t, err)
	_, exists = manager.GetHostname(hostname)
	assert.False(t, exists)

	// Try to remove non-existent hostname
	err = manager.RemoveHostname("nonexistent.com")
	assert.Equal(t, ErrHostNotFound, err)
}

// Test wildcard hostname functionality.
func TestVhostsManager_WildcardHostname(t *testing.T) {
	manager := NewVhostsManager()
	wildcardApp := fiber.New()
	wildcardApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("wildcard app")
	})

	// Add a wildcard hostname
	err := manager.AddHostname("*.example.com", wildcardApp)
	assert.NoError(t, err)

	// Create main app with middleware
	mainApp := fiber.New()
	mainApp.Use(VhostMiddleware(manager))

	// Test with a subdomain that should match the wildcard
	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "subdomain.example.com"
	resp, err := mainApp.Test(req)
	assert.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "wildcard app", string(body))

	// Test removing wildcard hostname
	err = manager.RemoveHostname("*.example.com")
	assert.NoError(t, err)

	// Try to remove it again
	err = manager.RemoveHostname("*.example.com")
	assert.Equal(t, ErrHostNotFound, err)
}

// Test SetDefaultApp functionality.
func TestVhostsManager_SetDefaultApp(t *testing.T) {
	manager := NewVhostsManager()
	defaultApp := fiber.New()
	defaultApp.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("default app")
	})

	// Set default app
	manager.SetDefaultApp(defaultApp)

	// Create main app with middleware
	mainApp := fiber.New()
	mainApp.Use(VhostMiddleware(manager))

	// Test with a non-existing hostname
	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "nonexistent.com"
	resp, err := mainApp.Test(req)
	assert.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "default app", string(body))
}

// Test error cases for AddHostname.
func TestVhostsManager_AddHostname_Errors(t *testing.T) {
	manager := NewVhostsManager()
	app := fiber.New()

	// Test with empty hostname
	err := manager.AddHostname("", app)
	assert.Equal(t, ErrInvalidHostname, err)

	// Test with duplicate hostname
	hostname := "example.com"
	err = manager.AddHostname(hostname, app)
	assert.NoError(t, err)

	err = manager.AddHostname(hostname, app)
	assert.Equal(t, ErrHostExists, err)

	// Test with duplicate wildcard hostname
	wildcardHostname := "*.example.org"
	err = manager.AddHostname(wildcardHostname, app)
	assert.NoError(t, err)

	err = manager.AddHostname(wildcardHostname, app)
	assert.Equal(t, ErrHostExists, err)
}

// Test NewVhostsManager with config.
func TestNewVhostsManager_WithConfig(t *testing.T) {
	defaultApp := fiber.New()
	config := Config{
		DefaultApp:       defaultApp,
		EnableLogging:    true,
		RecoverFromPanic: true,
	}

	manager := NewVhostsManager(config)

	// Test properties were set correctly
	assert.Equal(t, defaultApp, manager.defaultApp)
	assert.True(t, manager.enableLog)

	// Ensure maps were initialized
	assert.NotNil(t, manager.hosts)
	assert.NotNil(t, manager.wildcards)
}

// Non-existing hostname without default app should return 404.
func TestVhostMiddleware_NoDefaultApp(t *testing.T) {
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

// With logging enabled, the middleware should log the hostname.
func TestVhostMiddleware_Logging(t *testing.T) {
	manager := NewVhostsManager(Config{
		EnableLogging: true,
	})
	mainApp := fiber.New()
	mainApp.Use(VhostMiddleware(manager))

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "example.com"
	resp, err := mainApp.Test(req)
	assert.NoError(t, err)
	resp.Body.Close()

}

// With panic recovery enabled, the middleware should recover from panics.
func TestVhostMiddleware_RecoverFromPanic(t *testing.T) {
	manager := NewVhostsManager(Config{
		RecoverFromPanic: true,
	})
	mainApp := fiber.New()
	mainApp.Use(VhostMiddleware(manager))

	mainApp.Get("/", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "example.com"
	resp, err := mainApp.Test(req)
	assert.NoError(t, err)
	resp.Body.Close()
}

// Test adding longer hostname like "sw.didam.smartest.website"
func TestVhostsManager_AddLongHostname(t *testing.T) {
	manager := NewVhostsManager()
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("subapp response")
	})

	hostname := "sw.didam.smartest.website"
	manager.AddHostname(hostname, app)

	retrieved, exists := manager.GetHostname(hostname)
	assert.True(t, exists)
	assert.Equal(t, app, retrieved)

	// test if app is found for subdomain
	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "sw.didam.smartest.website"
	resp, err := app.Test(req)
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

}
