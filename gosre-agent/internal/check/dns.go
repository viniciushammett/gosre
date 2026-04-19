// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package check

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gosre/gosre-sdk/domain"
)

// DNSChecker validates DNS resolution for a Target.
type DNSChecker struct{}

// NewDNSChecker returns a ready-to-use DNSChecker.
func NewDNSChecker() *DNSChecker {
	return &DNSChecker{}
}

// Execute resolves t.Address using the record type from cfg.Params["record_type"].
// Supported types: A (default), AAAA, CNAME, MX. It implements domain.Checker.
func (c *DNSChecker) Execute(ctx context.Context, t domain.Target, cfg domain.CheckConfig) domain.Result {
	start := time.Now()

	recordType := "A"
	if rt, ok := cfg.Params["record_type"]; ok && rt != "" {
		recordType = strings.ToUpper(rt)
	}

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

	resolved, err := resolve(ctx, net.DefaultResolver, t.Address, recordType)
	duration := time.Since(start)

	if err != nil {
		status := domain.StatusFail
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			status = domain.StatusTimeout
		} else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
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

	return domain.Result{
		ID:        fmt.Sprintf("%d", start.UnixNano()),
		CheckID:   cfg.ID,
		TargetID:  t.ID,
		Status:    domain.StatusOK,
		Duration:  duration,
		Timestamp: time.Now().UTC(),
		Metadata:  map[string]string{"resolved": strings.Join(resolved, ",")},
	}
}

func resolve(ctx context.Context, r *net.Resolver, host, recordType string) ([]string, error) {
	switch recordType {
	case "AAAA":
		addrs, err := r.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		var results []string
		for _, a := range addrs {
			if a.IP.To4() == nil {
				results = append(results, a.IP.String())
			}
		}
		return results, nil

	case "CNAME":
		cname, err := r.LookupCNAME(ctx, host)
		if err != nil {
			return nil, err
		}
		return []string{cname}, nil

	case "MX":
		records, err := r.LookupMX(ctx, host)
		if err != nil {
			return nil, err
		}
		results := make([]string, len(records))
		for i, mx := range records {
			results[i] = fmt.Sprintf("%s:%d", mx.Host, mx.Pref)
		}
		return results, nil

	default:
		addrs, err := r.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		return addrs, nil
	}
}
