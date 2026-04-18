import { apiFetch, type Target } from "./client";

export function listTargets(): Promise<Target[]> {
  return apiFetch<Target[]>("/api/v1/targets");
}

export function getTarget(id: string): Promise<Target> {
  return apiFetch<Target>(`/api/v1/targets/${id}`);
}

export function createTarget(body: Target): Promise<Target> {
  return apiFetch<Target>("/api/v1/targets", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateTarget(id: string, body: Target): Promise<Target> {
  return apiFetch<Target>(`/api/v1/targets/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteTarget(id: string): Promise<void> {
  return apiFetch<void>(`/api/v1/targets/${id}`, { method: "DELETE" });
}
