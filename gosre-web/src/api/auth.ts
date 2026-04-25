const AUTH_BASE = (import.meta.env.VITE_AUTH_URL ?? "http://localhost:8081").replace(/\/$/, "");

const TOKEN_KEY = "gosre_access_token";
const REFRESH_KEY = "gosre_refresh_token";

export interface AuthUser {
  user_id: string;
  email: string;
  role: 'viewer' | 'operator' | 'admin' | 'owner';
}

interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY);
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_KEY);
}

function setTokens(pair: TokenPair): void {
  localStorage.setItem(TOKEN_KEY, pair.access_token);
  localStorage.setItem(REFRESH_KEY, pair.refresh_token);
}

async function authPost<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${AUTH_BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (res.status === 204) return undefined as T;
  const json = (await res.json()) as Record<string, unknown>;
  if (!res.ok) {
    throw new Error((json.error as string | undefined) ?? res.statusText);
  }
  return json as T;
}

export async function login(email: string, password: string): Promise<void> {
  const pair = await authPost<TokenPair>("/auth/login", { email, password });
  setTokens(pair);
}

export async function refreshToken(): Promise<void> {
  const rt = getRefreshToken();
  if (!rt) throw new Error("no refresh token");
  const pair = await authPost<TokenPair>("/auth/refresh", { refresh_token: rt });
  setTokens(pair);
}

export async function logout(): Promise<void> {
  const rt = getRefreshToken();
  if (rt) {
    await authPost<void>("/auth/logout", { refresh_token: rt }).catch(() => undefined);
  }
  clearTokens();
}

export async function me(): Promise<AuthUser> {
  const token = getAccessToken();
  const res = await fetch(`${AUTH_BASE}/auth/me`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!res.ok) throw new Error("not authenticated");
  return res.json() as Promise<AuthUser>;
}
