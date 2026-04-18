import { useQuery } from "@tanstack/react-query";
import { listTargets } from "../api/targets";
import { listChecks } from "../api/checks";
import { listIncidents } from "../api/incidents";
import { listResults } from "../api/results";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import StatusBadge from "../components/StatusBadge";
import type { CheckStatus } from "../api/client";

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

export default function Dashboard() {
  const targets = useQuery({ queryKey: ["targets"], queryFn: listTargets });
  const checks = useQuery({ queryKey: ["checks"], queryFn: listChecks });
  const incidents = useQuery({
    queryKey: ["incidents", "open"],
    queryFn: () => listIncidents("open"),
  });
  const results = useQuery({ queryKey: ["results"], queryFn: () => listResults() });

  const isLoading =
    targets.isLoading || checks.isLoading || incidents.isLoading || results.isLoading;
  const error =
    targets.error ?? checks.error ?? incidents.error ?? results.error;

  const latestResult = results.data?.[0];

  const targetMap = Object.fromEntries(
    (targets.data ?? []).map((t) => [t.id, t.name])
  );

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

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <SummaryCard
          label="Targets"
          value={targets.data?.length ?? 0}
        />
        <SummaryCard
          label="Checks"
          value={checks.data?.length ?? 0}
        />
        <SummaryCard
          label="Open incidents"
          value={incidents.data?.length ?? 0}
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
    </div>
  );
}
