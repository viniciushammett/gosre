// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/gosre/gosre-exporter/internal/apiclient"
	"github.com/gosre/gosre-sdk/domain"
)

const agentOnlineThreshold = 60 * time.Second

// ensure Collector implements prometheus.Collector at compile time.
var _ prometheus.Collector = (*Collector)(nil)

var histogramBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// Collector implements prometheus.Collector for gosre-api metrics.
type Collector struct {
	client     *apiclient.Client
	logger     *zap.Logger
	apiTimeout time.Duration

	targetUp         *prometheus.Desc
	checkResultTotal *prometheus.Desc
	checkDuration    *prometheus.Desc
	incidentTotal    *prometheus.Desc
	agentUp          *prometheus.Desc

	sloCompliance        *prometheus.Desc
	errorBudgetRemaining *prometheus.Desc
	burnRate1h           *prometheus.Desc
	burnRate6h           *prometheus.Desc
}

// New constructs a Collector backed by the given API client.
// apiTimeout controls the deadline for each scrape's API calls (use poll interval / 2).
func New(client *apiclient.Client, apiTimeout time.Duration, logger *zap.Logger) *Collector {
	return &Collector{
		client:     client,
		logger:     logger,
		apiTimeout: apiTimeout,
		targetUp: prometheus.NewDesc(
			"gosre_target_up",
			"1 if the most recent check result for this target was ok, 0 otherwise.",
			[]string{"target_id", "target_name", "target_type"},
			nil,
		),
		checkResultTotal: prometheus.NewDesc(
			"gosre_check_result_total",
			"Total number of check results by target, check type, and status.",
			[]string{"target_id", "check_type", "status"},
			nil,
		),
		checkDuration: prometheus.NewDesc(
			"gosre_check_duration_seconds",
			"Duration of check executions in seconds.",
			[]string{"target_id", "check_type"},
			nil,
		),
		incidentTotal: prometheus.NewDesc(
			"gosre_incident_total",
			"Number of incidents per target and state.",
			[]string{"target_id", "state"},
			nil,
		),
		agentUp: prometheus.NewDesc(
			"gosre_agent_up",
			"1 if the agent sent a heartbeat in the last 60 seconds, 0 otherwise.",
			[]string{"agent_id", "hostname"},
			nil,
		),
		sloCompliance: prometheus.NewDesc(
			"gosre_slo_compliance",
			"SLO compliance in the observation window (0 to 1). Omitted when insufficient data (L-030).",
			[]string{"slo_id", "target_id"},
			nil,
		),
		errorBudgetRemaining: prometheus.NewDesc(
			"gosre_error_budget_remaining",
			"Fraction of error budget not yet consumed (0 to 1). Omitted when insufficient data.",
			[]string{"slo_id", "target_id"},
			nil,
		),
		burnRate1h: prometheus.NewDesc(
			"gosre_burn_rate_1h",
			"Error budget burn rate over the last 1 hour. Values > 1 indicate exhaustion.",
			[]string{"slo_id", "target_id"},
			nil,
		),
		burnRate6h: prometheus.NewDesc(
			"gosre_burn_rate_6h",
			"Error budget burn rate over the last 6 hours. Values > 1 indicate exhaustion.",
			[]string{"slo_id", "target_id"},
			nil,
		),
	}
}

// Describe sends all metric descriptors to ch.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.targetUp
	ch <- c.checkResultTotal
	ch <- c.checkDuration
	ch <- c.incidentTotal
	ch <- c.agentUp
	ch <- c.sloCompliance
	ch <- c.errorBudgetRemaining
	ch <- c.burnRate1h
	ch <- c.burnRate6h
}

// Collect fetches live data from gosre-api and sends metrics to ch.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.apiTimeout)
	defer cancel()

	targets, err := c.client.ListTargets(ctx)
	if err != nil {
		c.logger.Error("collect: list targets", zap.Error(err))
		return
	}

	checks, err := c.client.ListChecks(ctx)
	if err != nil {
		c.logger.Error("collect: list checks", zap.Error(err))
		return
	}

	results, err := c.client.ListResults(ctx)
	if err != nil {
		c.logger.Error("collect: list results", zap.Error(err))
		return
	}

	incidents, err := c.client.ListIncidents(ctx)
	if err != nil {
		c.logger.Error("collect: list incidents", zap.Error(err))
		return
	}

	agents, err := c.client.ListAgents(ctx)
	if err != nil {
		c.logger.Error("collect: list agents", zap.Error(err))
		return
	}

	// Build check_id → check_type index.
	checkType := make(map[string]string, len(checks))
	for _, ck := range checks {
		checkType[ck.ID] = string(ck.Type)
	}

	// Build target index for label lookup.
	targetByID := make(map[string]domain.Target, len(targets))
	for _, t := range targets {
		targetByID[t.ID] = t
	}

	c.collectTargetUp(ch, targets, results)
	c.collectCheckResults(ch, results, checkType)
	c.collectCheckDuration(ch, results, checkType)
	c.collectIncidents(ch, incidents)
	c.collectAgentUp(ch, agents)
	c.collectSLOMetrics(ctx, ch, targets)
}

func (c *Collector) collectTargetUp(ch chan<- prometheus.Metric, targets []domain.Target, results []domain.Result) {
	// results are ordered timestamp DESC — first occurrence per target is the latest.
	latestStatus := make(map[string]domain.CheckStatus, len(targets))
	for _, r := range results {
		if _, seen := latestStatus[r.TargetID]; !seen {
			latestStatus[r.TargetID] = r.Status
		}
	}

	for _, t := range targets {
		up := 0.0
		if latestStatus[t.ID] == domain.StatusOK {
			up = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.targetUp, prometheus.GaugeValue, up,
			t.ID, t.Name, string(t.Type),
		)
	}
}

func (c *Collector) collectCheckResults(ch chan<- prometheus.Metric, results []domain.Result, checkType map[string]string) {
	type key struct {
		targetID  string
		checkType string
		status    string
	}
	counts := make(map[key]float64)
	for _, r := range results {
		ct := checkType[r.CheckID]
		k := key{r.TargetID, ct, string(r.Status)}
		counts[k]++
	}
	for k, n := range counts {
		ch <- prometheus.MustNewConstMetric(
			c.checkResultTotal, prometheus.CounterValue, n,
			k.targetID, k.checkType, k.status,
		)
	}
}

func (c *Collector) collectCheckDuration(ch chan<- prometheus.Metric, results []domain.Result, checkType map[string]string) {
	type key struct {
		targetID  string
		checkType string
	}
	type stats struct {
		count   uint64
		sum     float64
		buckets map[float64]uint64
	}

	groups := make(map[key]*stats)
	for _, r := range results {
		ct := checkType[r.CheckID]
		k := key{r.TargetID, ct}
		if groups[k] == nil {
			b := make(map[float64]uint64, len(histogramBuckets))
			for _, bound := range histogramBuckets {
				b[bound] = 0
			}
			groups[k] = &stats{buckets: b}
		}
		s := groups[k]
		durationSec := float64(r.Duration) / float64(time.Second)
		s.count++
		s.sum += durationSec
		for _, bound := range histogramBuckets {
			if durationSec <= bound {
				s.buckets[bound]++
			}
		}
	}

	for k, s := range groups {
		ch <- prometheus.MustNewConstHistogram(
			c.checkDuration,
			s.count, s.sum, s.buckets,
			k.targetID, k.checkType,
		)
	}
}

func (c *Collector) collectAgentUp(ch chan<- prometheus.Metric, agents []apiclient.AgentRecord) {
	for _, a := range agents {
		up := 0.0
		if time.Since(a.LastSeen) < agentOnlineThreshold {
			up = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.agentUp, prometheus.GaugeValue, up,
			a.ID, a.Hostname,
		)
	}
}

func (c *Collector) collectSLOMetrics(ctx context.Context, ch chan<- prometheus.Metric, targets []domain.Target) {
	for _, t := range targets {
		slos, err := c.client.ListSLOsByTarget(ctx, t.ID)
		if err != nil {
			c.logger.Warn("collect: list slos", zap.String("target_id", t.ID), zap.Error(err))
			continue
		}
		for _, slo := range slos {
			budget, err := c.client.GetSLOBudget(ctx, slo.ID)
			if err != nil {
				c.logger.Warn("collect: get slo budget", zap.String("slo_id", slo.ID), zap.Error(err))
				continue
			}
			if budget.InsufficientData {
				continue
			}
			remaining := (budget.Compliance - slo.Threshold) / (1 - slo.Threshold)
			if remaining < 0 {
				remaining = 0
			}
			ch <- prometheus.MustNewConstMetric(c.sloCompliance, prometheus.GaugeValue, budget.Compliance, slo.ID, t.ID)
			ch <- prometheus.MustNewConstMetric(c.errorBudgetRemaining, prometheus.GaugeValue, remaining, slo.ID, t.ID)
			ch <- prometheus.MustNewConstMetric(c.burnRate1h, prometheus.GaugeValue, budget.BurnRate1h, slo.ID, t.ID)
			ch <- prometheus.MustNewConstMetric(c.burnRate6h, prometheus.GaugeValue, budget.BurnRate6h, slo.ID, t.ID)
		}
	}
}

func (c *Collector) collectIncidents(ch chan<- prometheus.Metric, incidents []domain.Incident) {
	type key struct {
		targetID string
		state    string
	}
	counts := make(map[key]float64)
	for _, i := range incidents {
		k := key{i.TargetID, string(i.State)}
		counts[k]++
	}
	for k, n := range counts {
		ch <- prometheus.MustNewConstMetric(
			c.incidentTotal, prometheus.GaugeValue, n,
			k.targetID, k.state,
		)
	}
}
