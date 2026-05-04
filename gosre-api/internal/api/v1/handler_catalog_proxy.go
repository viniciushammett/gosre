// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CatalogProxy forwards /api/v1/catalog/* requests to gosre-catalog.
// GOSRE_CATALOG_URL env var sets the upstream base URL (default http://localhost:8081).
type CatalogProxy struct {
	baseURL string
	client  *http.Client
}

// NewCatalogProxy constructs a CatalogProxy.
func NewCatalogProxy() *CatalogProxy {
	u := os.Getenv("GOSRE_CATALOG_URL")
	if u == "" {
		u = "http://localhost:8081"
	}
	return &CatalogProxy{
		baseURL: strings.TrimRight(u, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Proxy strips /catalog from the incoming path and forwards to gosre-catalog.
// Example: /api/v1/catalog/services/:id → GOSRE_CATALOG_URL/api/v1/services/<value>
func (p *CatalogProxy) Proxy(c *gin.Context) {
	// Use the actual resolved path (param values already substituted).
	path := strings.Replace(c.Request.URL.Path, "/catalog", "", 1)
	targetURL := fmt.Sprintf("%s%s", p.baseURL, path)
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		Fail(c, http.StatusBadGateway, "catalog_proxy_error", err.Error())
		return
	}
	for key, values := range c.Request.Header {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		Fail(c, http.StatusBadGateway, "catalog_unreachable", err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()

	for key, values := range resp.Header {
		for _, v := range values {
			c.Header(key, v)
		}
	}
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}
