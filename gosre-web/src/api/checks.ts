import { apiFetch, type CheckConfig, type Result } from "./client";

export function listChecks(): Promise<CheckConfig[]> {
  return apiFetch<CheckConfig[]>("/api/v1/checks");
}

export function createCheck(body: CheckConfig): Promise<CheckConfig> {
  return apiFetch<CheckConfig>("/api/v1/checks", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function runCheck(id: string): Promise<Result> {
  return apiFetch<Result>(`/api/v1/checks/${id}/run`, { method: "POST" });
}
