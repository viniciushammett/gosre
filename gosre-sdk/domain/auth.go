// Copyright 2026 Vinicius Teixeira
// Licensed under the Apache License, Version 2.0

package domain

// Role defines the access level of a User within an Organization.
type Role string

const (
	// RoleViewer can read resources but cannot modify anything.
	RoleViewer Role = "viewer"
	// RoleOperator can execute checks and acknowledge incidents.
	RoleOperator Role = "operator"
	// RoleAdmin can manage targets, checks, SLOs, and team membership.
	RoleAdmin Role = "admin"
	// RoleOwner has full control including billing and organization deletion.
	RoleOwner Role = "owner"
)

// Permission represents a discrete action a Role may be granted.
type Permission string

const (
	// PermViewResults allows reading check results and incidents.
	PermViewResults Permission = "results:read"
	// PermRunChecks allows triggering immediate check executions.
	PermRunChecks Permission = "checks:run"
	// PermManageTargets allows creating, updating, and deleting targets.
	PermManageTargets Permission = "targets:manage"
	// PermManageMembers allows inviting and removing organization members.
	PermManageMembers Permission = "members:manage"
	// PermManageOrg allows full organization configuration including deletion.
	PermManageOrg Permission = "org:manage"
)

// User is an authenticated principal within the GoSRE platform.
// A User belongs to exactly one Organization and carries a Role that
// determines which actions are permitted.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	OrgID string `json:"org_id"`
	Role  Role   `json:"role"`
}
