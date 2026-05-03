// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

// Package hash provides a thread-safe consistent hash ring for distributing
// checks across agents. The ring uses agent_id as the node key — never the
// agent IP, which changes on every K8s restart (L-031).
package hash

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

const defaultVirtualNodes = 150

// Ring maps arbitrary string keys to agent IDs using consistent hashing.
// Virtual nodes ensure even distribution even with small agent counts.
type Ring struct {
	mu           sync.RWMutex
	virtualNodes int
	entries      []entry
}

type entry struct {
	hash    uint32
	agentID string
}

// New creates a Ring with 150 virtual nodes per agent.
func New() *Ring {
	return &Ring{virtualNodes: defaultVirtualNodes}
}

// Add inserts an agent into the ring with the configured number of virtual replicas.
func (r *Ring) Add(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.virtualNodes {
		h := fnvHash(fmt.Sprintf("%s-%d", agentID, i))
		r.entries = append(r.entries, entry{hash: h, agentID: agentID})
	}
	sort.Slice(r.entries, func(i, j int) bool {
		return r.entries[i].hash < r.entries[j].hash
	})
}

// Remove deletes an agent and all its virtual nodes from the ring.
func (r *Ring) Remove(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	keep := r.entries[:0]
	for _, e := range r.entries {
		if e.agentID != agentID {
			keep = append(keep, e)
		}
	}
	r.entries = keep
}

// Get returns the agent ID responsible for the given key.
// Returns an empty string if the ring has no agents.
func (r *Ring) Get(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.entries) == 0 {
		return ""
	}
	h := fnvHash(key)
	idx := sort.Search(len(r.entries), func(i int) bool {
		return r.entries[i].hash >= h
	})
	if idx == len(r.entries) {
		idx = 0
	}
	return r.entries[idx].agentID
}

// IsEmpty reports whether the ring contains no agents.
func (r *Ring) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries) == 0
}

func fnvHash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
