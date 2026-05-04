import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  listChannels,
  createChannel,
  deleteChannel,
  listRules,
  createRule,
  deleteRule,
  type NotificationChannel,
  type NotificationRule,
  type ChannelKind,
} from "../../api/notifications";
import { listOrganizations } from "../../api/organizations";
import { apiFetch } from "../../api/client";
import { useRole } from "../../hooks/useRole";
import LoadingSpinner from "../../components/LoadingSpinner";
import ErrorMessage from "../../components/ErrorMessage";
import EmptyState from "../../components/EmptyState";

interface Project {
  id: string;
  name: string;
  organization_id: string;
}

const kindIcon: Record<ChannelKind, string> = {
  slack:   "💬",
  email:   "✉️",
  webhook: "🔗",
};

const kindStyle: Record<ChannelKind, string> = {
  slack:   "bg-purple-500/15 text-purple-400 border-purple-500/30",
  email:   "bg-brand/15 text-brand border-brand/30",
  webhook: "bg-orange-500/15 text-orange-400 border-orange-500/30",
};

const EVENT_KINDS = [
  "gosre.incidents.opened",
  "gosre.incidents.resolved",
  "gosre.incidents.acknowledged",
];

const CHANNEL_KINDS: ChannelKind[] = ["slack", "email", "webhook"];

const EMPTY_CHANNEL = { name: "", kind: "slack" as ChannelKind, config: {} as Record<string, string> };
const EMPTY_RULE = { channel_id: "", event_kind: EVENT_KINDS[0], tag_filter: [] as string[] };

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h2 className="text-sm font-medium text-gray-300 mb-3">{title}</h2>
      {children}
    </div>
  );
}

function ConfigFields({ kind, config, onChange }: {
  kind: ChannelKind;
  config: Record<string, string>;
  onChange: (k: string, v: string) => void;
}) {
  if (kind === "slack") {
    return (
      <div className="flex flex-col gap-1">
        <label className="text-xs text-gray-500 uppercase tracking-wider">Webhook URL</label>
        <input
          required
          type="url"
          value={config["webhook_url"] ?? ""}
          onChange={(e) => onChange("webhook_url", e.target.value)}
          placeholder="https://hooks.slack.com/…"
          className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
        />
      </div>
    );
  }
  if (kind === "email") {
    return (
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <label className="text-xs text-gray-500 uppercase tracking-wider">SMTP Host</label>
          <input
            required
            value={config["smtp_host"] ?? ""}
            onChange={(e) => onChange("smtp_host", e.target.value)}
            placeholder="smtp.example.com"
            className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label className="text-xs text-gray-500 uppercase tracking-wider">To</label>
          <input
            required
            type="email"
            value={config["to"] ?? ""}
            onChange={(e) => onChange("to", e.target.value)}
            placeholder="on-call@example.com"
            className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
          />
        </div>
      </div>
    );
  }
  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs text-gray-500 uppercase tracking-wider">URL</label>
      <input
        required
        type="url"
        value={config["url"] ?? ""}
        onChange={(e) => onChange("url", e.target.value)}
        placeholder="https://example.com/webhook"
        className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
      />
    </div>
  );
}

export default function Notifications() {
  const qc = useQueryClient();
  const { hasMinRole } = useRole();

  const [selectedProject, setSelectedProject] = useState<string>("");
  const [selectedOrg, setSelectedOrg] = useState<string>("");
  const [showChannelForm, setShowChannelForm] = useState(false);
  const [showRuleForm, setShowRuleForm] = useState(false);
  const [channelForm, setChannelForm] = useState(EMPTY_CHANNEL);
  const [ruleForm, setRuleForm] = useState(EMPTY_RULE);

  const orgs = useQuery({ queryKey: ["organizations"], queryFn: listOrganizations });

  const projects = useQuery({
    queryKey: ["projects", selectedOrg],
    queryFn: () => apiFetch<Project[]>(`/api/v1/organizations/${selectedOrg}/projects`),
    enabled: !!selectedOrg,
  });

  const channels = useQuery({
    queryKey: ["notif-channels", selectedProject],
    queryFn: () => listChannels(selectedProject),
    enabled: !!selectedProject,
  });

  const rules = useQuery({
    queryKey: ["notif-rules", selectedProject],
    queryFn: () => listRules(selectedProject),
    enabled: !!selectedProject,
  });

  const createCh = useMutation({
    mutationFn: (f: typeof EMPTY_CHANNEL) =>
      createChannel({ ...f, project_id: selectedProject }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notif-channels"] });
      setChannelForm(EMPTY_CHANNEL);
      setShowChannelForm(false);
    },
  });

  const deleteCh = useMutation({
    mutationFn: deleteChannel,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notif-channels"] }),
  });

  const createRl = useMutation({
    mutationFn: (f: typeof EMPTY_RULE) =>
      createRule({ ...f, project_id: selectedProject }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notif-rules"] });
      setRuleForm(EMPTY_RULE);
      setShowRuleForm(false);
    },
  });

  const deleteRl = useMutation({
    mutationFn: deleteRule,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notif-rules"] }),
  });

  const error =
    orgs.error ?? projects.error ?? channels.error ?? rules.error ??
    createCh.error ?? deleteCh.error ?? createRl.error ?? deleteRl.error;

  function handleChannelSubmit(e: React.FormEvent) {
    e.preventDefault();
    createCh.mutate(channelForm);
  }

  function handleRuleSubmit(e: React.FormEvent) {
    e.preventDefault();
    createRl.mutate(ruleForm);
  }

  function setConfigField(k: string, v: string) {
    setChannelForm((f) => ({ ...f, config: { ...f.config, [k]: v } }));
  }

  if (orgs.isLoading) return <LoadingSpinner />;

  const channelMap = Object.fromEntries(
    (channels.data ?? []).map((c) => [c.id, c.name])
  );

  return (
    <div className="p-6 space-y-6 max-w-3xl">
      <div>
        <h1 className="text-lg font-semibold text-white">Notifications</h1>
        <p className="text-xs text-gray-500 mt-0.5">Channels and routing rules per project</p>
      </div>

      <ErrorMessage error={error as Error | null} />

      {/* Org + Project selectors */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div className="flex flex-col gap-1.5">
          <label className="text-xs text-gray-500 uppercase tracking-wider">Organization</label>
          <select
            value={selectedOrg}
            onChange={(e) => { setSelectedOrg(e.target.value); setSelectedProject(""); }}
            className="bg-surface-raised border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
          >
            <option value="">— select —</option>
            {(orgs.data ?? []).map((o) => (
              <option key={o.id} value={o.id}>{o.name}</option>
            ))}
          </select>
        </div>
        <div className="flex flex-col gap-1.5">
          <label className="text-xs text-gray-500 uppercase tracking-wider">Project</label>
          <select
            value={selectedProject}
            onChange={(e) => setSelectedProject(e.target.value)}
            disabled={!selectedOrg}
            className="bg-surface-raised border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand disabled:opacity-40"
          >
            <option value="">— select —</option>
            {(projects.data ?? []).map((p) => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>
        </div>
      </div>

      {!selectedProject ? (
        <EmptyState message="Select an organization and project to manage notifications." />
      ) : (
        <>
          {/* Channels */}
          <Section title="Channels">
            {!showChannelForm && hasMinRole("operator") && (
              <button
                onClick={() => setShowChannelForm(true)}
                className="mb-3 text-xs px-3 py-1.5 rounded border border-brand text-brand hover:bg-brand hover:text-black transition-colors"
              >
                Add Channel
              </button>
            )}

            {showChannelForm && (
              <form
                onSubmit={handleChannelSubmit}
                className="mb-4 bg-surface-raised border border-surface-border rounded-lg p-4 space-y-4"
              >
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="flex flex-col gap-1">
                    <label className="text-xs text-gray-500 uppercase tracking-wider">Name</label>
                    <input
                      required
                      value={channelForm.name}
                      onChange={(e) => setChannelForm((f) => ({ ...f, name: e.target.value }))}
                      placeholder="on-call-slack"
                      className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                    />
                  </div>
                  <div className="flex flex-col gap-1">
                    <label className="text-xs text-gray-500 uppercase tracking-wider">Kind</label>
                    <select
                      value={channelForm.kind}
                      onChange={(e) => {
                        setChannelForm((f) => ({ ...f, kind: e.target.value as ChannelKind, config: {} }));
                      }}
                      className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                    >
                      {CHANNEL_KINDS.map((k) => (
                        <option key={k} value={k}>{kindIcon[k]} {k}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <ConfigFields
                  kind={channelForm.kind}
                  config={channelForm.config}
                  onChange={setConfigField}
                />
                <div className="flex gap-2">
                  <button
                    type="submit"
                    disabled={createCh.isPending}
                    className="text-xs px-3 py-1.5 rounded bg-brand text-black font-medium hover:bg-brand/90 transition-colors disabled:opacity-40"
                  >
                    {createCh.isPending ? "Creating…" : "Create"}
                  </button>
                  <button
                    type="button"
                    onClick={() => { setShowChannelForm(false); setChannelForm(EMPTY_CHANNEL); }}
                    className="text-xs px-3 py-1.5 rounded border border-surface-border text-gray-400 hover:text-white transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            )}

            {channels.isLoading ? (
              <LoadingSpinner />
            ) : (channels.data ?? []).length === 0 ? (
              <EmptyState message="No channels configured." />
            ) : (
              <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-surface-border">
                      <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Name</th>
                      <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Kind</th>
                      <th className="px-4 py-2" />
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-surface-border">
                    {(channels.data ?? []).map((ch: NotificationChannel) => (
                      <tr key={ch.id} className="hover:bg-surface-border/30 transition-colors">
                        <td className="px-4 py-3 text-white font-medium">{ch.name}</td>
                        <td className="px-4 py-3">
                          <span className={`inline-block px-2 py-0.5 rounded text-xs border ${kindStyle[ch.kind]}`}>
                            {kindIcon[ch.kind]} {ch.kind}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          {hasMinRole("admin") && (
                            <button
                              onClick={() => {
                                if (confirm(`Delete channel "${ch.name}"?`)) deleteCh.mutate(ch.id);
                              }}
                              className="text-xs text-gray-500 hover:text-status-fail transition-colors"
                            >
                              delete
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Section>

          {/* Rules */}
          <Section title="Routing Rules">
            {!showRuleForm && hasMinRole("operator") && (channels.data ?? []).length > 0 && (
              <button
                onClick={() => setShowRuleForm(true)}
                className="mb-3 text-xs px-3 py-1.5 rounded border border-brand text-brand hover:bg-brand hover:text-black transition-colors"
              >
                Add Rule
              </button>
            )}

            {showRuleForm && (
              <form
                onSubmit={handleRuleSubmit}
                className="mb-4 bg-surface-raised border border-surface-border rounded-lg p-4 space-y-4"
              >
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="flex flex-col gap-1">
                    <label className="text-xs text-gray-500 uppercase tracking-wider">Event</label>
                    <select
                      value={ruleForm.event_kind}
                      onChange={(e) => setRuleForm((f) => ({ ...f, event_kind: e.target.value }))}
                      className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand font-mono"
                    >
                      {EVENT_KINDS.map((k) => (
                        <option key={k} value={k}>{k}</option>
                      ))}
                    </select>
                  </div>
                  <div className="flex flex-col gap-1">
                    <label className="text-xs text-gray-500 uppercase tracking-wider">Channel</label>
                    <select
                      required
                      value={ruleForm.channel_id}
                      onChange={(e) => setRuleForm((f) => ({ ...f, channel_id: e.target.value }))}
                      className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                    >
                      <option value="">— select channel —</option>
                      {(channels.data ?? []).map((c) => (
                        <option key={c.id} value={c.id}>{c.name}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    type="submit"
                    disabled={createRl.isPending || !ruleForm.channel_id}
                    className="text-xs px-3 py-1.5 rounded bg-brand text-black font-medium hover:bg-brand/90 transition-colors disabled:opacity-40"
                  >
                    {createRl.isPending ? "Creating…" : "Create"}
                  </button>
                  <button
                    type="button"
                    onClick={() => { setShowRuleForm(false); setRuleForm(EMPTY_RULE); }}
                    className="text-xs px-3 py-1.5 rounded border border-surface-border text-gray-400 hover:text-white transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            )}

            {rules.isLoading ? (
              <LoadingSpinner />
            ) : (rules.data ?? []).length === 0 ? (
              <EmptyState message="No routing rules defined." />
            ) : (
              <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-surface-border">
                      <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Event</th>
                      <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Channel</th>
                      <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Tags</th>
                      <th className="px-4 py-2" />
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-surface-border">
                    {(rules.data ?? []).map((r: NotificationRule) => (
                      <tr key={r.id} className="hover:bg-surface-border/30 transition-colors">
                        <td className="px-4 py-3 text-brand font-mono text-xs">{r.event_kind}</td>
                        <td className="px-4 py-3 text-gray-300 text-sm">
                          {channelMap[r.channel_id] ?? <span className="text-gray-600 font-mono">{r.channel_id.slice(0, 8)}…</span>}
                        </td>
                        <td className="px-4 py-3">
                          {(r.tag_filter ?? []).length > 0 ? (
                            <div className="flex flex-wrap gap-1">
                              {r.tag_filter!.map((tag) => (
                                <span key={tag} className="px-1.5 py-0.5 rounded text-xs bg-surface-border text-gray-400">
                                  {tag}
                                </span>
                              ))}
                            </div>
                          ) : (
                            <span className="text-xs text-gray-600">all tags</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-right">
                          {hasMinRole("admin") && (
                            <button
                              onClick={() => {
                                if (confirm("Delete this rule?")) deleteRl.mutate(r.id);
                              }}
                              className="text-xs text-gray-500 hover:text-status-fail transition-colors"
                            >
                              delete
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Section>
        </>
      )}
    </div>
  );
}
