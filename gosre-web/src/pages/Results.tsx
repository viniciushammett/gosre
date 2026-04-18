import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listResults } from "../api/results";
import { listTargets } from "../api/targets";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import EmptyState from "../components/EmptyState";
import StatusBadge from "../components/StatusBadge";
import type { CheckStatus } from "../api/client";

function fmtDuration(ns?: number) {
  if (ns == null) return "—";
  const ms = ns / 1_000_000;
  if (ms < 1000) return `${ms.toFixed(0)}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

function fmtTime(iso?: string) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString();
}

export default function Results() {
  const [targetId, setTargetId] = useState<string>("");

  const targets = useQuery({ queryKey: ["targets"], queryFn: listTargets });
  const results = useQuery({
    queryKey: ["results", targetId],
    queryFn: () => listResults(targetId || undefined),
  });

  const isLoading = targets.isLoading || results.isLoading;
  const error = targets.error ?? results.error;

  if (isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold text-white">Results</h1>

        <select
          value={targetId}
          onChange={(e) => setTargetId(e.target.value)}
          className="bg-surface-raised border border-surface-border text-sm text-gray-300 rounded px-3 py-1.5 focus:outline-none focus:border-brand"
        >
          <option value="">All targets</option>
          {targets.data?.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
      </div>

      <ErrorMessage error={error as Error | null} />

      {(!results.data || results.data.length === 0) ? (
        <EmptyState message="No results found." />
      ) : (
        <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Status</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Target</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Check</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Duration</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Timestamp</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Error</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {results.data.map((r) => (
                <tr key={r.id} className="hover:bg-surface-border/30 transition-colors">
                  <td className="px-4 py-3">
                    <StatusBadge status={(r.status ?? "unknown") as CheckStatus} />
                  </td>
                  <td className="px-4 py-3 text-gray-300 font-mono text-xs">{r.target_id ?? "—"}</td>
                  <td className="px-4 py-3 text-gray-300 font-mono text-xs">{r.check_id ?? "—"}</td>
                  <td className="px-4 py-3 text-gray-400 font-mono text-xs">{fmtDuration(r.duration_ms)}</td>
                  <td className="px-4 py-3 text-gray-400 text-xs">{fmtTime(r.timestamp)}</td>
                  <td className="px-4 py-3 text-status-fail font-mono text-xs truncate max-w-xs">
                    {r.error ?? ""}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
