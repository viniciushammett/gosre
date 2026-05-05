// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package v1

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/viniciushammett/gosre/gosre-sdk/domain"

	"github.com/viniciushammett/gosre/gosre-api/internal/service"
)

// NotificationHandler handles HTTP requests for notification channel and rule resources.
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler constructs a NotificationHandler.
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// ListChannels handles GET /api/v1/notification/channels?project_id=xxx.
func (h *NotificationHandler) ListChannels(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		Fail(c, http.StatusBadRequest, "project_id_required", "project_id query parameter is required")
		return
	}
	channels, err := h.svc.ListChannels(c.Request.Context(), projectID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, channels)
}

// CreateChannel handles POST /api/v1/notification/channels.
func (h *NotificationHandler) CreateChannel(c *gin.Context) {
	var body domain.NotificationChannel
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	ch, err := h.svc.CreateChannel(c.Request.Context(), body)
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, ch)
}

// GetChannel handles GET /api/v1/notification/channels/:id.
func (h *NotificationHandler) GetChannel(c *gin.Context) {
	ch, err := h.svc.GetChannel(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "channel_not_found", "notification channel not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, ch)
}

// DeleteChannel handles DELETE /api/v1/notification/channels/:id.
func (h *NotificationHandler) DeleteChannel(c *gin.Context) {
	err := h.svc.DeleteChannel(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "channel_not_found", "notification channel not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusNoContent, nil)
}

// ListRules handles GET /api/v1/notification/rules?project_id=xxx.
func (h *NotificationHandler) ListRules(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		Fail(c, http.StatusBadRequest, "project_id_required", "project_id query parameter is required")
		return
	}
	rules, err := h.svc.ListRules(c.Request.Context(), projectID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, rules)
}

// CreateRule handles POST /api/v1/notification/rules.
func (h *NotificationHandler) CreateRule(c *gin.Context) {
	var body domain.NotificationRule
	if err := c.ShouldBindJSON(&body); err != nil {
		Fail(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	r, err := h.svc.CreateRule(c.Request.Context(), body)
	if err != nil {
		Fail(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	OK(c, http.StatusCreated, r)
}

// GetRule handles GET /api/v1/notification/rules/:id.
func (h *NotificationHandler) GetRule(c *gin.Context) {
	r, err := h.svc.GetRule(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "rule_not_found", "notification rule not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusOK, r)
}

// DeleteRule handles DELETE /api/v1/notification/rules/:id.
func (h *NotificationHandler) DeleteRule(c *gin.Context) {
	err := h.svc.DeleteRule(c.Request.Context(), c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		Fail(c, http.StatusNotFound, "rule_not_found", "notification rule not found")
		return
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	OK(c, http.StatusNoContent, nil)
}
