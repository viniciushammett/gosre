import { apiFetch } from "./client";

export interface Agent {
  id: string;
  hostname: string;
  version: string;
  last_seen: string;
}

export function listAgents(): Promise<Agent[]> {
  return apiFetch<Agent[]>("/api/v1/agents");
}
