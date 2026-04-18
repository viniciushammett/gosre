import { apiFetch, type Result } from "./client";

export function listResults(targetId?: string): Promise<Result[]> {
  const qs = targetId ? `?target_id=${encodeURIComponent(targetId)}` : "";
  return apiFetch<Result[]>(`/api/v1/results${qs}`);
}

export function getResult(id: string): Promise<Result> {
  return apiFetch<Result>(`/api/v1/results/${id}`);
}
