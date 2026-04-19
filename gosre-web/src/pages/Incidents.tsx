import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listIncidents, patchIncident } from "../api/incidents";
import { listTargets } from "../api/targets";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import EmptyState from "../components/EmptyState";
import StatusBadge from "../components/StatusBadge";
import IncidentDrawer from "../components/IncidentDrawer";
import type { Incident, IncidentState } from "../api/client";

const FILTERS: { label: string; value: IncidentState | undefined }[] = [
  { label: "All", value: undefined },
  { label: "Open", value: "open" },
  { label: "Acknowledged", value: "acknowledged" },
  { label: "Resolved", value: "resolved" },
];

function fmt(iso?: string) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString();
}

export default function Incidents() {
  const qc = useQueryClient();
  const [filter, setFilter] = useState<IncidentState | undefined>(undefined);
  const [selected, setSelected] = useState<Incident | null>(null);

  const targets = useQuery({ queryKey: ["targets"], queryFn: listTargets });
  const { data, isLoading, error } = useQuery({
    queryKey: ["incidents", filter],
    queryFn: () => listIncidents(filter),
  });

  const targetMap = Object.fromEntries(
    (targets.data ?? []).map((t) => [t.id, t.name])
  );

  function resolveTarget(id?: string) {
    if (!id) return "—";
    if (targetMap[id]) return <span className="text-white">{targetMap[id]}</span>;
    return <code className="text-gray-500 font-mono">{id.slice(0, 4)}…{id.slice(-3)}</code>;
  }

  const patch = useMutation({
    mutationFn: ({ id, state }: { id: string; state: IncidentState }) =>
      patchIncident(id, state),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["incidents"] });
      qc.invalidateQueries({ queryKey: ["incidents", "open"] });
    },
  });

  if (isLoading || targets.isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <h1 className="text-lg font-semibold text-white mb-5">Incidents</h1>

      <div className="flex gap-1 mb-4">
        {FILTERS.map(({ label, value }) => (
          <button
            key={label}
            onClick={() => setFilter(value)}
            className={[
              "px-3 py-1 rounded text-xs transition-colors",
              filter === value
                ? "bg-brand text-white"
                : "bg-surface-raised border border-surface-border text-gray-400 hover:text-white",
            ].join(" ")}
          >
            {label}
          </button>
        ))}
      </div>

      <ErrorMessage error={(error ?? targets.error ?? patch.error) as Error | null} />

      {(!data || data.length === 0) ? (
        <EmptyState message="No incidents found." />
      ) : (
        <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Target</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">State</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">First seen</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Last seen</th>
                <th className="px-4 py-2" />
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {data.map((inc) => (
                <tr key={inc.id} onClick={() => setSelected(inc)} className="hover:bg-surface-border/30 transition-colors cursor-pointer">
                  <td className="px-4 py-3 text-sm">{resolveTarget(inc.target_id)}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={inc.state ?? "open"} />
                  </td>
                  <td className="px-4 py-3 text-gray-400 text-xs">{fmt(inc.first_seen)}</td>
                  <td className="px-4 py-3 text-gray-400 text-xs">{fmt(inc.last_seen)}</td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex justify-end gap-2">
                      {inc.state === "open" && (
                        <button
                          onClick={() => patch.mutate({ id: inc.id!, state: "acknowledged" })}
                          disabled={patch.isPending}
                          className="text-xs text-gray-500 hover:text-status-timeout transition-colors disabled:opacity-40"
                        >
                          ack
                        </button>
                      )}
                      {inc.state !== "resolved" && (
                        <button
                          onClick={() => patch.mutate({ id: inc.id!, state: "resolved" })}
                          disabled={patch.isPending}
                          className="text-xs text-gray-500 hover:text-status-ok transition-colors disabled:opacity-40"
                        >
                          resolve
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <IncidentDrawer
        incident={selected}
        targetName={selected ? (targetMap[selected.target_id ?? ""] ?? selected.target_id ?? "—") : ""}
        onClose={() => setSelected(null)}
      />
    </div>
  );
}
