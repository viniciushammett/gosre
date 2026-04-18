import type { components } from "../types/api";

export type Target = components["schemas"]["Target"];
export type CheckConfig = components["schemas"]["CheckConfig"];
export type Result = components["schemas"]["Result"];
export type Incident = components["schemas"]["Incident"];
export type TargetType = components["schemas"]["TargetType"];
export type CheckType = components["schemas"]["CheckType"];
export type CheckStatus = components["schemas"]["CheckStatus"];
export type IncidentState = components["schemas"]["IncidentState"];

const BASE_URL = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");
const API_KEY = import.meta.env.VITE_API_KEY ?? "";

interface Envelope<T> {
  data?: T;
  error?: { code?: string; message?: string };
}

export async function apiFetch<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (API_KEY) {
    headers["X-API-Key"] = API_KEY;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: { ...headers, ...init?.headers },
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const body = (await res.json()) as Envelope<T>;

  if (!res.ok) {
    const msg = body.error?.message ?? res.statusText;
    throw new Error(msg);
  }

  if (body.error?.message) {
    throw new Error(body.error.message);
  }

  return body.data as T;
}
