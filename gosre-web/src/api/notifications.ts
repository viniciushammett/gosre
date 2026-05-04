import { apiFetch } from "./client";

export type ChannelKind = "slack" | "email" | "webhook";

export interface NotificationChannel {
  id: string;
  project_id: string;
  name: string;
  kind: ChannelKind;
  config: Record<string, string>;
}

export interface NotificationRule {
  id: string;
  project_id: string;
  channel_id: string;
  event_kind: string;
  tag_filter?: string[];
}

export function listChannels(projectId: string): Promise<NotificationChannel[]> {
  return apiFetch<NotificationChannel[]>(
    `/api/v1/notification/channels?project_id=${encodeURIComponent(projectId)}`
  );
}

export function createChannel(body: Omit<NotificationChannel, "id">): Promise<NotificationChannel> {
  return apiFetch<NotificationChannel>("/api/v1/notification/channels", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function deleteChannel(id: string): Promise<void> {
  return apiFetch<void>(`/api/v1/notification/channels/${id}`, { method: "DELETE" });
}

export function listRules(projectId: string): Promise<NotificationRule[]> {
  return apiFetch<NotificationRule[]>(
    `/api/v1/notification/rules?project_id=${encodeURIComponent(projectId)}`
  );
}

export function createRule(body: Omit<NotificationRule, "id">): Promise<NotificationRule> {
  return apiFetch<NotificationRule>("/api/v1/notification/rules", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function deleteRule(id: string): Promise<void> {
  return apiFetch<void>(`/api/v1/notification/rules/${id}`, { method: "DELETE" });
}
