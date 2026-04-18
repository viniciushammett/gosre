import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listTargets, deleteTarget } from "../api/targets";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import EmptyState from "../components/EmptyState";
import type { TargetType } from "../api/client";

const typeColors: Record<TargetType, string> = {
  http: "text-brand",
  tcp: "text-purple-400",
  dns: "text-yellow-400",
  tls: "text-green-400",
};

export default function Targets() {
  const qc = useQueryClient();
  const { data, isLoading, error } = useQuery({
    queryKey: ["targets"],
    queryFn: listTargets,
  });

  const del = useMutation({
    mutationFn: deleteTarget,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["targets"] }),
  });

  if (isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <h1 className="text-lg font-semibold text-white mb-5">Targets</h1>

      <ErrorMessage error={(error ?? del.error) as Error | null} />

      {(!data || data.length === 0) ? (
        <EmptyState message="No targets found." />
      ) : (
        <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Name</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Type</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Address</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Tags</th>
                <th className="px-4 py-2" />
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {data.map((t) => (
                <tr key={t.id} className="hover:bg-surface-border/30 transition-colors">
                  <td className="px-4 py-3 text-white font-medium">{t.name}</td>
                  <td className="px-4 py-3 font-mono">
                    <span className={typeColors[t.type] ?? "text-gray-400"}>{t.type}</span>
                  </td>
                  <td className="px-4 py-3 text-gray-300 font-mono text-xs">{t.address}</td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-1">
                      {t.tags?.map((tag) => (
                        <span
                          key={tag}
                          className="px-1.5 py-0.5 rounded text-xs bg-surface-border text-gray-400"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => {
                        if (confirm(`Delete target "${t.name}"?`)) {
                          del.mutate(t.id!);
                        }
                      }}
                      disabled={del.isPending}
                      className="text-xs text-gray-500 hover:text-status-fail transition-colors disabled:opacity-40"
                    >
                      delete
                    </button>
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
