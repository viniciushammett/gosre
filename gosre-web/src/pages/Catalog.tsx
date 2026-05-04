import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listServices, type CatalogService, type ServiceCriticality } from "../api/catalog";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import EmptyState from "../components/EmptyState";

const criticalityRank: Record<ServiceCriticality, number> = {
  critical: 4,
  high: 3,
  medium: 2,
  low: 1,
};

const criticalityStyle: Record<ServiceCriticality, string> = {
  critical: "bg-status-fail/15 text-status-fail border-status-fail/30",
  high:     "bg-orange-500/15 text-orange-400 border-orange-500/30",
  medium:   "bg-status-timeout/15 text-status-timeout border-status-timeout/30",
  low:      "bg-surface-border text-gray-400 border-surface-border",
};

function CriticalityBadge({ level }: { level: ServiceCriticality }) {
  return (
    <span className={`inline-block px-2 py-0.5 rounded text-xs border font-medium ${criticalityStyle[level]}`}>
      {level}
    </span>
  );
}

function HealthScore({ service }: { service: CatalogService }) {
  const score = criticalityRank[service.criticality] ?? 0;
  const bars = [1, 2, 3, 4];
  return (
    <div className="flex items-center gap-1" title={`Criticality rank ${score}/4`}>
      {bars.map((b) => (
        <span
          key={b}
          className={`inline-block w-1.5 rounded-full transition-colors ${
            b <= score
              ? score >= 3 ? "bg-status-fail h-3" : score === 2 ? "bg-status-timeout h-2.5" : "bg-status-ok h-2"
              : "bg-surface-border h-2"
          }`}
        />
      ))}
    </div>
  );
}

export default function Catalog() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["catalog-services"],
    queryFn: () => listServices(),
  });

  if (isLoading) return <LoadingSpinner />;

  const sorted = [...(data ?? [])].sort(
    (a, b) => criticalityRank[b.criticality] - criticalityRank[a.criticality]
  );

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-white">Service Catalog</h1>
          <p className="text-xs text-gray-500 mt-0.5">
            {data?.length ?? 0} service{data?.length !== 1 ? "s" : ""} registered
          </p>
        </div>
      </div>

      <ErrorMessage error={error as Error | null} />

      {sorted.length === 0 ? (
        <EmptyState message="No services in catalog." />
      ) : (
        <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Service</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Owner</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Criticality</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Impact</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Links</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {sorted.map((svc) => (
                <tr key={svc.id} className="hover:bg-surface-border/30 transition-colors group">
                  <td className="px-4 py-3">
                    <Link
                      to={`/catalog/${svc.id}`}
                      className="font-medium text-white group-hover:text-brand transition-colors"
                    >
                      {svc.name}
                    </Link>
                    {svc.project_id && (
                      <p className="text-xs text-gray-600 font-mono mt-0.5 truncate max-w-xs">
                        {svc.project_id.slice(0, 8)}…
                      </p>
                    )}
                  </td>
                  <td className="px-4 py-3 text-gray-300 text-sm">{svc.owner || "—"}</td>
                  <td className="px-4 py-3">
                    <CriticalityBadge level={svc.criticality} />
                  </td>
                  <td className="px-4 py-3">
                    <HealthScore service={svc} />
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      {svc.runbook_url && (
                        <a
                          href={svc.runbook_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-xs text-gray-500 hover:text-brand transition-colors font-mono"
                        >
                          runbook ↗
                        </a>
                      )}
                      {svc.repo_url && (
                        <a
                          href={svc.repo_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-xs text-gray-500 hover:text-brand transition-colors font-mono"
                        >
                          repo ↗
                        </a>
                      )}
                      {!svc.runbook_url && !svc.repo_url && (
                        <span className="text-xs text-gray-600">—</span>
                      )}
                    </div>
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
