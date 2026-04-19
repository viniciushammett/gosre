import { useQuery } from "@tanstack/react-query";
import { listAgents, type Agent } from "../api/agents";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import EmptyState from "../components/EmptyState";

function isOnline(lastSeen: string): boolean {
  return Date.now() - new Date(lastSeen).getTime() < 60_000;
}

function fmtLastSeen(lastSeen: string): string {
  const diff = Math.floor((Date.now() - new Date(lastSeen).getTime()) / 1000);
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  return `${Math.floor(diff / 3600)}h ago`;
}

export default function Agents() {
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ["agents"],
    queryFn: listAgents,
    refetchInterval: 30_000,
  });

  if (isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold text-white">Agents</h1>
        <button
          onClick={() => refetch()}
          disabled={isFetching}
          className="text-xs px-3 py-1.5 rounded border border-surface-border text-gray-400 hover:text-white hover:border-brand transition-colors disabled:opacity-40"
        >
          {isFetching ? "refreshing…" : "Refresh"}
        </button>
      </div>

      <ErrorMessage error={error as Error | null} />

      {!data || data.length === 0 ? (
        <EmptyState message="No agents registered." />
      ) : (
        <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">
                  Status
                </th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">
                  Hostname
                </th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">
                  Version
                </th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">
                  Last Seen
                </th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">
                  ID
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {data.map((a: Agent) => (
                <tr
                  key={a.id}
                  className="hover:bg-surface-border/30 transition-colors"
                >
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
                  <td className="px-4 py-3 text-gray-400 font-mono text-xs">
                    {a.version}
                  </td>
                  <td className="px-4 py-3 text-gray-400 text-xs">
                    {fmtLastSeen(a.last_seen)}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-gray-600">
                    {a.id.slice(0, 8)}…
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
