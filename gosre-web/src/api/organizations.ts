import { apiFetch } from "./client";

export interface Organization {
  id: string;
  name: string;
  slug: string;
  created_at: string;
}

export interface Team {
  id: string;
  organization_id: string;
  name: string;
  slug: string;
  created_at: string;
}

export function listOrganizations(): Promise<Organization[]> {
  return apiFetch<Organization[]>("/api/v1/organizations");
}

export function createOrganization(body: { name: string; slug?: string }): Promise<Organization> {
  return apiFetch<Organization>("/api/v1/organizations", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function deleteOrganization(id: string): Promise<void> {
  return apiFetch<void>(`/api/v1/organizations/${id}`, { method: "DELETE" });
}

export function listTeams(orgId: string): Promise<Team[]> {
  return apiFetch<Team[]>(`/api/v1/organizations/${orgId}/teams`);
}

export function createTeam(orgId: string, body: { name: string; slug?: string }): Promise<Team> {
  return apiFetch<Team>(`/api/v1/organizations/${orgId}/teams`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function deleteTeam(orgId: string, teamId: string): Promise<void> {
  return apiFetch<void>(`/api/v1/organizations/${orgId}/teams/${teamId}`, { method: "DELETE" });
}
