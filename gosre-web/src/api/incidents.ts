import { apiFetch, type Incident, type IncidentState } from "./client";

export function listIncidents(state?: IncidentState): Promise<Incident[]> {
  const qs = state ? `?state=${encodeURIComponent(state)}` : "";
  return apiFetch<Incident[]>(`/api/v1/incidents${qs}`);
}

export function patchIncident(id: string, state: IncidentState): Promise<Incident> {
  return apiFetch<Incident>(`/api/v1/incidents/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ state }),
  });
}
