import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listChecks, runCheck } from "../api/checks";
import { listTargets } from "../api/targets";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import EmptyState from "../components/EmptyState";
import StatusBadge from "../components/StatusBadge";
import type { CheckStatus, Result } from "../api/client";

function fmtDuration(ns?: number) {
  if (ns == null) return "—";
  const ms = ns / 1_000_000;
  if (ms < 1000) return `${ms.toFixed(0)}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

function fmtInterval(ns?: number) {
  if (ns == null) return "—";
  const s = ns / 1_000_000_000;
  if (s < 60) return `${s}s`;
  return `${Math.round(s / 60)}m`;
}

export default function Checks() {
  const qc = useQueryClient();
  const [runResults, setRunResults] = useState<Record<string, Result>>({});

  const targets = useQuery({ queryKey: ["targets"], queryFn: listTargets });
  const checks = useQuery({ queryKey: ["checks"], queryFn: listChecks });
  const { refetch } = checks;

  const targetMap = Object.fromEntries(
    (targets.data ?? []).map((t) => [t.id, t.name]),
  );

  const run = useMutation({
    mutationFn: (id: string) => runCheck(id),
    onSuccess: (result, id) => {
      setRunResults((prev) => ({ ...prev, [id]: result }));
      qc.invalidateQueries({ queryKey: ["results"] });
    },
  });

  const isLoading = targets.isLoading || checks.isLoading;
  const error = targets.error ?? checks.error;

  if (isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold text-white">Checks</h1>
        <button
          onClick={() => refetch()}
          disabled={checks.isFetching}
          className="text-xs px-3 py-1.5 rounded border border-surface-border text-gray-400 hover:text-white hover:border-brand transition-colors disabled:opacity-40"
        >
          {checks.isFetching ? "refreshing…" : "Refresh"}
        </button>
      </div>

      <ErrorMessage error={(error ?? run.error) as Error | null} />

      {(!checks.data || checks.data.length === 0) ? (
        <EmptyState message="No checks found." />
      ) : (
        <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Type</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Target</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Interval</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Timeout</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Last run</th>
                <th className="px-4 py-2" />
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {checks.data.map((c) => {
                const lastResult = runResults[c.id!];
                const isRunning = run.isPending && run.variables === c.id;
                return (
                  <tr key={c.id} className="hover:bg-surface-border/30 transition-colors">
                    <td className="px-4 py-3 text-brand font-mono text-xs">{c.type}</td>
                    <td className="px-4 py-3 text-gray-300 text-xs">
                      {targetMap[c.target_id] ?? (
                        <code className="text-gray-500 font-mono">
                          {c.target_id?.slice(0, 4)}…{c.target_id?.slice(-3)}
                        </code>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-400 font-mono text-xs">{fmtInterval(c.interval)}</td>
                    <td className="px-4 py-3 text-gray-400 font-mono text-xs">{fmtDuration(c.timeout)}</td>
                    <td className="px-4 py-3">
                      {lastResult ? (
                        <div className="flex items-center gap-2">
                          <StatusBadge status={lastResult.status as CheckStatus} />
                          <span className="text-xs text-gray-500 font-mono">
                            {fmtDuration(lastResult.duration_ms)}
                          </span>
                        </div>
                      ) : (
                        <span className="text-xs text-gray-600">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => run.mutate(c.id!)}
                        disabled={isRunning}
                        className="text-xs px-2 py-1 rounded border border-surface-border text-gray-400 hover:text-white hover:border-brand transition-colors disabled:opacity-40"
                      >
                        {isRunning ? "running…" : "run"}
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
