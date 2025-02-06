// This file contains the VhostsManager struct and methods to manage sub-apps based on hostnames in a thread-safe manner using RWMutex for locking and unlocking the map of hosts. It also contains a VhostMiddleware function that mounts a specific sub-app based on the hostname. If the hostname is not found, it returns a 404 response. This middleware is intended to be used with the main app to route requests to different sub-apps based on the hostname.
// © 2025 MHJ Wiggers. All rights reserved.
package fibervhosts

import (
	"sync"

	"github.com/gofiber/fiber/v2"
)

// VhostsManager is a struct that holds a map of hostnames to sub-apps and provides methods to add and retrieve sub-apps based on hostnames in a thread-safe manner using RWMutex for locking and unlocking the map of hosts.
type VhostsManager struct {
	mu    sync.RWMutex
	hosts map[string]*fiber.App
}

// NewVhostsManager creates a new VhostsManager instance with an empty map of hosts and returns a pointer to it
func NewVhostsManager() *VhostsManager {
	return &VhostsManager{
		hosts: make(map[string]*fiber.App),
	}
}

// AddHostname adds a sub-app for a given hostname to the manager
func (m *VhostsManager) AddHostname(hostname string, app *fiber.App) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hosts[hostname] = app
}

// GetHostname returns the sub-app for a given hostname if it exists
func (m *VhostsManager) GetHostname(hostname string) (*fiber.App, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	app, exists := m.hosts[hostname]
	return app, exists
}

// VhostMiddleware mounts a specific sub-app based on the hostname. If the hostname is not found, it returns a 404 response. This middleware is intended to be used with the main app to route requests to different sub-apps based on the hostname.
func VhostMiddleware(manager *VhostsManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		hostname := c.Hostname()
		if app, exists := manager.GetHostname(hostname); exists {
			app.Handler()(c.Context()) // Call sub-app handler
			return nil
		}
		return fiber.ErrNotFound // Return 404 if no matching hostname found
	}
}
