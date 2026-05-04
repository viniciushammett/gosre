import { apiFetch } from "./client";

export type ServiceCriticality = "low" | "medium" | "high" | "critical";
export type DependencyKind = "http" | "grpc" | "database" | "queue" | "generic";
export type EnvironmentKind = "dev" | "staging" | "prod";

export interface CatalogService {
  id: string;
  name: string;
  owner: string;
  criticality: ServiceCriticality;
  runbook_url?: string;
  repo_url?: string;
  project_id?: string;
  created_at: string;
}

export interface Dependency {
  id: string;
  source_service_id: string;
  target_service_id: string;
  kind: DependencyKind;
  created_at: string;
}

export interface Environment {
  id: string;
  name: string;
  project_id: string;
  kind: EnvironmentKind;
  created_at: string;
}

export function listServices(projectId?: string): Promise<CatalogService[]> {
  const qs = projectId ? `?project_id=${encodeURIComponent(projectId)}` : "";
  return apiFetch<CatalogService[]>(`/api/v1/catalog/services${qs}`);
}

export function getService(id: string): Promise<CatalogService> {
  return apiFetch<CatalogService>(`/api/v1/catalog/services/${id}`);
}

export function createService(body: Omit<CatalogService, "id" | "created_at">): Promise<CatalogService> {
  return apiFetch<CatalogService>("/api/v1/catalog/services", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function deleteService(id: string): Promise<void> {
  return apiFetch<void>(`/api/v1/catalog/services/${id}`, { method: "DELETE" });
}

export function listDependenciesBySource(sourceId: string): Promise<Dependency[]> {
  return apiFetch<Dependency[]>(`/api/v1/catalog/dependencies?source=${encodeURIComponent(sourceId)}`);
}

export function listDependenciesByTarget(targetId: string): Promise<Dependency[]> {
  return apiFetch<Dependency[]>(`/api/v1/catalog/dependencies?target=${encodeURIComponent(targetId)}`);
}

export function createDependency(body: Omit<Dependency, "id" | "created_at">): Promise<Dependency> {
  return apiFetch<Dependency>("/api/v1/catalog/dependencies", {
    method: "POST",
    body: JSON.stringify(body),
  });
}
