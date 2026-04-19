import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listTargets } from "../api/targets";
import { listChecks } from "../api/checks";
import { listIncidents } from "../api/incidents";
import { listResults } from "../api/results";
import { listAgents, type Agent } from "../api/agents";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import StatusBadge from "../components/StatusBadge";
import type { CheckStatus, Target } from "../api/client";

interface CardProps {
  label: string;
  value: string | number;
  sub?: string;
}

function SummaryCard({ label, value, sub }: CardProps) {
  return (
    <div className="bg-surface-raised border border-surface-border rounded-lg px-5 py-4">
      <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">{label}</p>
      <p className="text-2xl font-semibold text-white">{value}</p>
      {sub && <p className="text-xs text-gray-500 mt-1">{sub}</p>}
    </div>
  );
}

function SectionToggle({
  label,
  open,
  onToggle,
}: {
  label: string;
  open: boolean;
  onToggle: () => void;
}) {
  return (
    <button
      onClick={onToggle}
      className="flex items-center gap-2 text-sm font-medium text-gray-300 hover:text-white transition-colors mb-3"
    >
      <span className="text-gray-500 text-xs">{open ? "▼" : "▶"}</span>
      {label}
    </button>
  );
}

function isOnline(lastSeen: string): boolean {
  return Date.now() - new Date(lastSeen).getTime() < 60_000;
}

function fmtLastSeen(lastSeen: string): string {
  const diff = Math.floor((Date.now() - new Date(lastSeen).getTime()) / 1000);
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  return `${Math.floor(diff / 3600)}h ago`;
}

export default function Dashboard() {
  const [showTargets, setShowTargets] = useState(false);
  const [showAgents, setShowAgents] = useState(false);

  const targets = useQuery({ queryKey: ["targets"], queryFn: listTargets });
  const checks = useQuery({ queryKey: ["checks"], queryFn: listChecks });
  const incidents = useQuery({
    queryKey: ["incidents", "open"],
    queryFn: () => listIncidents("open"),
  });
  const results = useQuery({ queryKey: ["results"], queryFn: () => listResults() });
  const agents = useQuery({
    queryKey: ["agents"],
    queryFn: listAgents,
    refetchInterval: 30_000,
  });

  const isLoading =
    targets.isLoading ||
    checks.isLoading ||
    incidents.isLoading ||
    results.isLoading ||
    agents.isLoading;
  const error =
    targets.error ?? checks.error ?? incidents.error ?? results.error ?? agents.error;

  const latestResult = results.data?.[0];

  const targetMap = Object.fromEntries(
    (targets.data ?? []).map((t) => [t.id, t.name])
  );

  // last result status per target_id (results are newest-first)
  const lastResultByTarget: Record<string, CheckStatus> = {};
  for (const r of results.data ?? []) {
    if (r.target_id && !(r.target_id in lastResultByTarget)) {
      lastResultByTarget[r.target_id] = r.status as CheckStatus;
    }
  }

  const agentsOnline = (agents.data ?? []).filter((a) => isOnline(a.last_seen)).length;

  function resolveTarget(id?: string) {
    if (!id) return "—";
    if (targetMap[id]) return <span>{targetMap[id]}</span>;
    return <code className="text-gray-500 font-mono">{id.slice(0, 4)}…{id.slice(-3)}</code>;
  }

  if (isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <h1 className="text-lg font-semibold text-white mb-5">Dashboard</h1>

      <ErrorMessage error={error as Error | null} />

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-5">
        <SummaryCard label="Targets" value={targets.data?.length ?? 0} />
        <SummaryCard label="Checks" value={checks.data?.length ?? 0} />
        <SummaryCard label="Open incidents" value={incidents.data?.length ?? 0} />
        <SummaryCard
          label="Agents online"
          value={agentsOnline}
          sub={`of ${agents.data?.length ?? 0} registered`}
        />
        <div className="bg-surface-raised border border-surface-border rounded-lg px-5 py-4">
          <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Last result</p>
          {latestResult ? (
            <>
              <div className="mt-1">
                <StatusBadge status={latestResult.status as CheckStatus} />
              </div>
              <p className="text-xs text-gray-500 mt-2 font-mono">
                {latestResult.timestamp
                  ? new Date(latestResult.timestamp).toLocaleTimeString()
                  : "—"}
              </p>
            </>
          ) : (
            <p className="text-2xl font-semibold text-white">—</p>
          )}
        </div>
      </div>

      {(incidents.data?.length ?? 0) > 0 && (
        <div className="mt-6">
          <h2 className="text-sm font-medium text-gray-300 mb-3">Open incidents</h2>
          <div className="bg-surface-raised border border-surface-border rounded-lg divide-y divide-surface-border">
            {incidents.data!.map((inc) => (
              <div key={inc.id} className="flex items-center justify-between px-4 py-3">
                <div>
                  <span className="text-sm text-white">{resolveTarget(inc.target_id)}</span>
                  {inc.first_seen && (
                    <span className="ml-3 text-xs text-gray-500">
                      since {new Date(inc.first_seen).toLocaleString()}
                    </span>
                  )}
                </div>
                <StatusBadge status={inc.state ?? "open"} />
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="mt-6">
        <SectionToggle
          label={`Targets (${targets.data?.length ?? 0})`}
          open={showTargets}
          onToggle={() => setShowTargets((v) => !v)}
        />
        {showTargets && (
          <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-surface-border">
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Name</th>
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Type</th>
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Address</th>
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Last status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {(targets.data ?? []).map((t: Target) => (
                  <tr key={t.id} className="hover:bg-surface-border/30 transition-colors">
                    <td className="px-4 py-3 text-gray-200">{t.name}</td>
                    <td className="px-4 py-3 text-brand font-mono text-xs">{t.type}</td>
                    <td className="px-4 py-3 text-gray-400 font-mono text-xs truncate max-w-xs">{t.address}</td>
                    <td className="px-4 py-3">
                      {t.id && lastResultByTarget[t.id] ? (
                        <StatusBadge status={lastResultByTarget[t.id]} />
                      ) : (
                        <span className="text-xs text-gray-600">no data</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="mt-4">
        <SectionToggle
          label={`Agents (${agents.data?.length ?? 0})`}
          open={showAgents}
          onToggle={() => setShowAgents((v) => !v)}
        />
        {showAgents && (
          <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-surface-border">
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Status</th>
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Hostname</th>
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Last Seen</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {(agents.data ?? []).map((a: Agent) => (
                  <tr key={a.id} className="hover:bg-surface-border/30 transition-colors">
                    <td className="px-4 py-3">
                      {isOnline(a.last_seen) ? (
                        <span className="inline-flex items-center gap-1.5 text-xs text-emerald-400">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 inline-block" />
                          online
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 text-xs text-gray-500">
                          <span className="w-1.5 h-1.5 rounded-full bg-gray-600 inline-block" />
                          offline
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-200">{a.hostname}</td>
                    <td className="px-4 py-3 text-gray-400 text-xs">{fmtLastSeen(a.last_seen)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
