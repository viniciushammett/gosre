import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { Incident, Result } from "../api/client";
import { getResult } from "../api/results";
import StatusBadge from "./StatusBadge";

interface Props {
  incident: Incident | null;
  targetName: string;
  onClose: () => void;
}

function fmt(iso?: string) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString();
}

function fmtDuration(ms?: number) {
  if (ms == null) return "—";
  if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`;
  return `${ms}ms`;
}

function useIncidentResults(ids: string[]) {
  return useQuery<Result[]>({
    queryKey: ["incident-results", ...ids],
    queryFn: () =>
      Promise.all(
        ids.map((id) => getResult(id).catch(() => null))
      ).then((r) => r.filter(Boolean) as Result[]),
    enabled: ids.length > 0,
  });
}

export default function IncidentDrawer({ incident, targetName, onClose }: Props) {
  const [showRaw, setShowRaw] = useState(false);

  const ids = incident?.result_ids ?? [];
  const { data: results, isLoading } = useIncidentResults(ids);

  if (!incident) return null;

  return (
    <>
      <div className="fixed inset-0 bg-black/50 z-40" onClick={onClose} />
      <div className="fixed right-0 top-0 h-full w-[480px] bg-surface-raised border-l border-surface-border z-50 flex flex-col shadow-xl">
        <div className="flex items-center justify-between px-5 py-4 border-b border-surface-border">
          <h2 className="text-sm font-semibold text-white">Incident detail</h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-white transition-colors text-lg leading-none"
          >
            ✕
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-5 py-4 space-y-5 text-sm">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Target</p>
              <p className="text-white font-medium">{targetName || "—"}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">State</p>
              <StatusBadge status={incident.state ?? "open"} />
            </div>
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">First seen</p>
              <p className="text-gray-300 text-xs">{fmt(incident.first_seen)}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Last seen</p>
              <p className="text-gray-300 text-xs">{fmt(incident.last_seen)}</p>
            </div>
          </div>

          <div>
            <p className="text-xs text-gray-500 uppercase tracking-wider mb-3">
              Results ({ids.length})
            </p>

            {ids.length === 0 ? (
              <p className="text-gray-600 text-xs">No results associated.</p>
            ) : isLoading ? (
              <p className="text-gray-500 text-xs">Loading results...</p>
            ) : (
              <>
                <div className="rounded border border-surface-border overflow-hidden">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="border-b border-surface-border bg-black/20">
                        <th className="text-left px-3 py-2 text-gray-500 font-medium">Timestamp</th>
                        <th className="text-left px-3 py-2 text-gray-500 font-medium">Status</th>
                        <th className="text-left px-3 py-2 text-gray-500 font-medium">Duration</th>
                        <th className="text-left px-3 py-2 text-gray-500 font-medium">Error</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-surface-border">
                      {(results ?? []).map((r) => (
                        <tr key={r.id} className="hover:bg-surface-border/20">
                          <td className="px-3 py-2 text-gray-400 whitespace-nowrap">
                            {fmt(r.timestamp)}
                          </td>
                          <td className="px-3 py-2">
                            <StatusBadge status={r.status ?? "unknown"} />
                          </td>
                          <td className="px-3 py-2 text-gray-400">
                            {fmtDuration(r.duration_ms as unknown as number)}
                          </td>
                          <td className="px-3 py-2 text-gray-500 truncate max-w-[120px]" title={r.error}>
                            {r.error || "—"}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                <button
                  onClick={() => setShowRaw((v) => !v)}
                  className="mt-3 text-xs text-gray-500 hover:text-gray-300 transition-colors"
                >
                  {showRaw ? "▲ hide raw JSON" : "▼ raw JSON"}
                </button>

                {showRaw && (
                  <pre className="mt-2 text-xs text-gray-400 bg-black/40 border border-surface-border rounded p-3 overflow-x-auto whitespace-pre-wrap break-all">
                    {JSON.stringify(results, null, 2)}
                  </pre>
                )}
              </>
            )}
          </div>
        </div>
      </div>
    </>
  );
}
