// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package check

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/viniciushammett/gosre/gosre-sdk/domain"
)

const defaultExpiryThresholdDays = 14

// TLSChecker validates TLS connectivity and certificate health for a Target.
type TLSChecker struct{}

// NewTLSChecker returns a ready-to-use TLSChecker.
func NewTLSChecker() *TLSChecker {
	return &TLSChecker{}
}

// Execute dials t.Address over TLS, inspects the leaf certificate, and
// returns StatusFail if it expires within the configured threshold.
// It implements domain.Checker.
func (c *TLSChecker) Execute(ctx context.Context, t domain.Target, cfg domain.CheckConfig) domain.Result {
	start := time.Now()

	if ctx.Err() != nil {
		return domain.Result{
			ID:        fmt.Sprintf("%d", start.UnixNano()),
			CheckID:   cfg.ID,
			TargetID:  t.ID,
			Status:    domain.StatusTimeout,
			Duration:  time.Since(start),
			Error:     ctx.Err().Error(),
			Timestamp: time.Now().UTC(),
		}
	}

	host, _, err := net.SplitHostPort(t.Address)
	if err != nil {
		host = t.Address
	}

	tlsCfg := &tls.Config{ServerName: host}
	if cfg.Params["insecure"] == "true" {
		tlsCfg.InsecureSkipVerify = true //nolint:gosec // test-only flag
	}

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{},
		Config:    tlsCfg,
	}

	dialCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	conn, err := dialer.DialContext(dialCtx, "tcp", t.Address)
	duration := time.Since(start)

	if err != nil {
		status := domain.StatusFail
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			status = domain.StatusTimeout
		} else if dialCtx.Err() == context.DeadlineExceeded {
			status = domain.StatusTimeout
		}
		return domain.Result{
			ID:        fmt.Sprintf("%d", start.UnixNano()),
			CheckID:   cfg.ID,
			TargetID:  t.ID,
			Status:    status,
			Duration:  duration,
			Error:     err.Error(),
			Timestamp: time.Now().UTC(),
		}
	}
	defer func() { _ = conn.Close() }()

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return domain.Result{
			ID:        fmt.Sprintf("%d", start.UnixNano()),
			CheckID:   cfg.ID,
			TargetID:  t.ID,
			Status:    domain.StatusFail,
			Duration:  duration,
			Error:     "connection is not a TLS connection",
			Timestamp: time.Now().UTC(),
		}
	}

	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return domain.Result{
			ID:        fmt.Sprintf("%d", start.UnixNano()),
			CheckID:   cfg.ID,
			TargetID:  t.ID,
			Status:    domain.StatusFail,
			Duration:  duration,
			Error:     "no peer certificates returned",
			Timestamp: time.Now().UTC(),
		}
	}

	leaf := certs[0]
	daysRemaining := int(time.Until(leaf.NotAfter).Hours() / 24)

	threshold := defaultExpiryThresholdDays
	if v, ok := cfg.Params["expiry_days"]; ok && v != "" {
		if n, parseErr := strconv.Atoi(v); parseErr == nil {
			threshold = n
		}
	}

	status := domain.StatusOK
	errMsg := ""
	if daysRemaining < threshold {
		status = domain.StatusFail
		errMsg = fmt.Sprintf("certificate expires in %d days (threshold: %d)", daysRemaining, threshold)
	}

	return domain.Result{
		ID:        fmt.Sprintf("%d", start.UnixNano()),
		CheckID:   cfg.ID,
		TargetID:  t.ID,
		Status:    status,
		Duration:  duration,
		Error:     errMsg,
		Timestamp: time.Now().UTC(),
		Metadata: map[string]string{
			"expiry_days": strconv.Itoa(daysRemaining),
			"common_name": leaf.Subject.CommonName,
			"issuer":      leaf.Issuer.CommonName,
		},
	}
}
