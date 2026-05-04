import { apiFetch } from "./client";

export interface SLO {
  id: string;
  target_id: string;
  name: string;
  metric: string;
  threshold: number;
  window_seconds: number;
}

export interface BudgetResult {
  slo_id: string;
  target_id: string;
  compliance: number;
  burn_rate_1h: number;
  burn_rate_6h: number;
  burn_rate_24h: number;
  insufficient_data: boolean;
  total_results: number;
}

export function listSLOs(targetId: string): Promise<SLO[]> {
  return apiFetch<SLO[]>(`/api/v1/slos?target_id=${encodeURIComponent(targetId)}`);
}

export function getSLO(id: string): Promise<SLO> {
  return apiFetch<SLO>(`/api/v1/slos/${id}`);
}

export function getSLOBudget(id: string): Promise<BudgetResult> {
  return apiFetch<BudgetResult>(`/api/v1/slos/${id}/budget`);
}

export function createSLO(body: Omit<SLO, "id">): Promise<SLO> {
  return apiFetch<SLO>("/api/v1/slos", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function deleteSLO(id: string): Promise<void> {
  return apiFetch<void>(`/api/v1/slos/${id}`, { method: "DELETE" });
}
