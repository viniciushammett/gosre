// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import "github.com/gin-gonic/gin"

// Response is the standard API envelope.
type Response struct {
	Data  any    `json:"data"`
	Error *Error `json:"error"`
}

// Error carries a machine-readable code and a human-readable message.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a successful JSON response.
func OK(c *gin.Context, status int, data any) {
	c.JSON(status, Response{Data: data, Error: nil})
}

// Fail writes an error JSON response.
func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{Data: nil, Error: &Error{Code: code, Message: message}})
}
