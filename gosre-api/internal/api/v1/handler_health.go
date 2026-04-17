// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import "github.com/gin-gonic/gin"

// HealthHandler returns the API health status and version.
func HealthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "version": "0.1.0"})
}
