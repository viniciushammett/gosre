import { clearTokens, getAccessToken, refreshToken } from "./auth";
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

class UnauthorizedError extends Error {
  constructor() {
    super("unauthorized");
  }
}

// Ensures only one refresh call is in-flight at a time.
let _refreshPromise: Promise<void> | null = null;

function ensureRefresh(): Promise<void> {
  if (!_refreshPromise) {
    _refreshPromise = refreshToken().finally(() => {
      _refreshPromise = null;
    });
  }
  return _refreshPromise;
}

async function executeRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getAccessToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (API_KEY) headers["X-API-Key"] = API_KEY;
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: { ...headers, ...init?.headers },
  });

  if (res.status === 204) return undefined as T;
  if (res.status === 401) throw new UnauthorizedError();

  const body = (await res.json()) as Envelope<T>;

  if (!res.ok) {
    throw new Error(body.error?.message ?? res.statusText);
  }
  if (body.error?.message) {
    throw new Error(body.error.message);
  }

  return body.data as T;
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await executeRequest<T>(path, init);
  } catch (err) {
    if (err instanceof UnauthorizedError) {
      try {
        await ensureRefresh();
        return await executeRequest<T>(path, init);
      } catch {
        clearTokens();
        window.location.href = "/login";
        throw new Error("Session expired. Please sign in again.");
      }
    }
    throw err;
  }
}
