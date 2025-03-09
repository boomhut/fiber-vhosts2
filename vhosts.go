// This file contains the VhostsManager struct and methods to manage sub-apps based on hostnames in a thread-safe manner using RWMutex for locking and unlocking the map of hosts. It also contains a VhostMiddleware function that mounts a specific sub-app based on the hostname. If the hostname is not found, it returns a 404 response. This middleware is intended to be used with the main app to route requests to different sub-apps based on the hostname.
// © 2025 MHJ Wiggers. All rights reserved.
package fibervhosts

import (
	"errors"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	ErrInvalidHostname = errors.New("invalid hostname")
	ErrHostExists      = errors.New("host already exists")
	ErrHostNotFound    = errors.New("host not found")
)

// VhostsManager is a struct that holds a map of hostnames to sub-apps and provides methods to add and retrieve sub-apps based on hostnames in a thread-safe manner using RWMutex for locking and unlocking the map of hosts.
type VhostsManager struct {
	mu         sync.RWMutex
	hosts      map[string]*fiber.App
	wildcards  map[string]*fiber.App
	defaultApp *fiber.App
	enableLog  bool
}

type Config struct {
	DefaultApp       *fiber.App
	EnableLogging    bool
	RecoverFromPanic bool
}

// NewVhostsManager creates a new VhostsManager instance with an empty map of hosts and returns a pointer to it
func NewVhostsManager(config ...Config) *VhostsManager {
	m := &VhostsManager{
		hosts:     make(map[string]*fiber.App),
		wildcards: make(map[string]*fiber.App),
	}

	if len(config) > 0 {
		m.defaultApp = config[0].DefaultApp
		m.enableLog = config[0].EnableLogging
	}

	return m
}

// AddHostname adds a sub-app for a given hostname to the manager
func (m *VhostsManager) AddHostname(hostname string, app *fiber.App) error {
	if hostname == "" {
		return ErrInvalidHostname
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Handle wildcard hostnames
	if strings.HasPrefix(hostname, "*.") {
		suffix := hostname[2:]
		if _, exists := m.wildcards[suffix]; exists {
			return ErrHostExists
		}
		m.wildcards[suffix] = app
		return nil
	}

	if _, exists := m.hosts[hostname]; exists {
		return ErrHostExists
	}

	m.hosts[hostname] = app
	return nil
}

// RemoveHostname removes a sub-app for a given hostname from the manager
func (m *VhostsManager) RemoveHostname(hostname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strings.HasPrefix(hostname, "*.") {
		suffix := hostname[2:]
		if _, exists := m.wildcards[suffix]; !exists {
			return ErrHostNotFound
		}
		delete(m.wildcards, suffix)
		return nil
	}

	if _, exists := m.hosts[hostname]; !exists {
		return ErrHostNotFound
	}

	delete(m.hosts, hostname)
	return nil
}

// GetHostname returns the sub-app for a given hostname if it exists
func (m *VhostsManager) GetHostname(hostname string) (*fiber.App, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	app, exists := m.hosts[hostname]
	return app, exists
}

// GetHostnames returns a list of all hostnames in the manager
func (m *VhostsManager) GetHostnames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hostnames := make([]string, 0, len(m.hosts))
	for hostname := range m.hosts {
		hostnames = append(hostnames, hostname)
	}
	return hostnames
}

// SetDefaultApp sets the default app to be used when no matching hostname is found
func (m *VhostsManager) SetDefaultApp(app *fiber.App) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultApp = app
}

// findMatchingApp finds the sub-app for a given hostname, trying exact match first, then wildcard match, and finally returning the default app if no match is found
func (m *VhostsManager) findMatchingApp(hostname string) *fiber.App {
	// First try exact match
	if app, exists := m.hosts[hostname]; exists {
		return app
	}

	// Then try wildcard match
	parts := strings.Split(hostname, ".")
	if len(parts) > 1 {
		domain := strings.Join(parts[1:], ".")
		if app, exists := m.wildcards[domain]; exists {
			return app
		}
	}

	return m.defaultApp
}

// VhostMiddleware mounts a specific sub-app based on the hostname. If the hostname is not found, it returns a 404 response. This middleware is intended to be used with the main app to route requests to different sub-apps based on the hostname.
func VhostMiddleware(manager *VhostsManager) fiber.Handler {
	// Create recover middleware if enabled
	recoverHandler := recover.New()

	return func(c *fiber.Ctx) error {
		hostname := c.Hostname()

		if manager.enableLog {
			log.Infof("Processing request for hostname: %s", hostname)
		}

		app := manager.findMatchingApp(hostname)
		if app == nil {
			if manager.enableLog {
				log.Warnf("No application found for hostname: %s", hostname)
			}
			return fiber.ErrNotFound
		}

		// Wrap the handler with panic recovery if enabled
		if manager.enableLog {
			app.Use(recoverHandler)
		}

		app.Handler()(c.Context())
		return nil
	}
}
