# fiber-vhosts2

[![Go Reference](https://pkg.go.dev/badge/github.com/boomhut/fiber-vhosts2.svg)](https://pkg.go.dev/github.com/boomhut/fiber-vhosts2)
[![Go Report Card](https://goreportcard.com/badge/github.com/boomhut/fiber-vhosts2)](https://goreportcard.com/report/github.com/boomhut/fiber-vhosts2)

## Overview

- [fiber-vhosts2](#fiber-vhosts2)
		- [Overview](#overview)
	- [Description](#description)
	- [Example Usage](#example-usage)
	- [API](#api)
		- [func VhostMiddleware](#func-vhostmiddleware)
		- [type VhostsManager](#type-vhostsmanager)
		- [func NewVhostsManager](#func-newvhostsmanager)
		- [func (\*VhostsManager) AddHostname](#func-vhostsmanager-addhostname)
		- [func (\*VhostsManager) GetHostname](#func-vhostsmanager-gethostname)
	- [License](#license)


## Description

New approach vhosts implementation. This package provides a middleware for the Fiber web framework that allows you to route requests to different sub-apps based on the hostname. It also provides a VhostsManager struct that allows you to add and retrieve sub-apps based on hostnames in a thread-safe manner.

## Example Usage

```go
package main

import (
    "errors"
    "github.com/gofiber/fiber/v2"
    vhosts "github.com/boomhut/fiber-vhosts2"
)


func main() {
	app := fiber.New()
	manager = vhosts.NewVhostsManager()

	// Custom error handler for subApp1 is set in its Fiber config.
	errorHandler1 := func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).SendString("Custom error page for subApp1")
	}

	// Define sub-apps for specific hostnames with their respective ErrorHandler in config.
	subApp1 := fiber.New(
		fiber.Config{
			ErrorHandler: errorHandler1,
		},
	)
	subApp1.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from subApp1.example.com")
	})

	errorHandler2 := func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).SendString("Custom error page for subApp2: " + err.Error())
	}

	subApp2 := fiber.New(
		fiber.Config{
			ErrorHandler: errorHandler2,
		},
	)
	subApp2.Get("/", func(c *fiber.Ctx) error {
		return errors.New("An error occurred in subApp2")
	})

	// Register sub-apps without error handlers in the manager.
	manager.AddSubApp("subapp1.example.com", subApp1)
	manager.AddSubApp("127.0.0.1:3000", subApp2)

	// Use middleware
	app.Use(vhosts.VhostMiddleware(manager))

	// Start server
	app.Listen(":3000")
}

```

## API

### func VhostMiddleware

```go
func VhostMiddleware(manager *VhostsManager) fiber.Handler
```

VhostMiddleware mounts a specific sub-app based on the hostname. If the hostname is not found, it returns a 404 response. This middleware is intended to be used with the main app to route requests to different sub-apps based on the hostname.

### type VhostsManager

VhostsManager is a struct that holds a map of hostnames to sub-apps and provides methods to add and retrieve sub-apps based on hostnames in a thread-safe manner using RWMutex for locking and unlocking the map of hosts.

```go
type VhostsManager struct {
// contains filtered or unexported fields
}
```

### func NewVhostsManager

```go
func NewVhostsManager() *VhostsManager
```

NewVhostsManager creates a new VhostsManager instance with an empty map of hosts and returns a pointer to it.

### func (*VhostsManager) AddHostname

```go
func (m *VhostsManager) AddHostname(hostname string, app *fiber.App)
```

AddHostname adds a sub-app for a given hostname to the manager map. If the hostname already exists in the map, the sub-app is replaced with the new one.

### func (*VhostsManager) GetHostname

```go
func (m *VhostsManager) GetHostname(hostname string) (*fiber.App, bool)
```

GetHostname returns the sub-app for a given hostname if it exists in the manager and a boolean indicating whether the hostname was found or not.

---

## License

© 2025 MHJ Wiggers. All rights reserved.