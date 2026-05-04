import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import {
  getService,
  listDependenciesBySource,
  listDependenciesByTarget,
  type ServiceCriticality,
  type DependencyKind,
} from "../api/catalog";
import { listTargets } from "../api/targets";
import { listIncidents } from "../api/incidents";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import StatusBadge from "../components/StatusBadge";
import type { IncidentState } from "../api/client";

const criticalityStyle: Record<ServiceCriticality, string> = {
  critical: "bg-status-fail/15 text-status-fail border-status-fail/30",
  high:     "bg-orange-500/15 text-orange-400 border-orange-500/30",
  medium:   "bg-status-timeout/15 text-status-timeout border-status-timeout/30",
  low:      "bg-surface-border text-gray-400 border-surface-border",
};

const depKindStyle: Record<DependencyKind, string> = {
  http:     "text-brand",
  grpc:     "text-purple-400",
  database: "text-yellow-400",
  queue:    "text-orange-400",
  generic:  "text-gray-400",
};

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h2 className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-3">{title}</h2>
      {children}
    </div>
  );
}

export default function CatalogDetail() {
  const { id } = useParams<{ id: string }>();

  const svc = useQuery({
    queryKey: ["catalog-service", id],
    queryFn: () => getService(id!),
    enabled: !!id,
  });

  const depsOut = useQuery({
    queryKey: ["catalog-deps-out", id],
    queryFn: () => listDependenciesBySource(id!),
    enabled: !!id,
  });

  const depsIn = useQuery({
    queryKey: ["catalog-deps-in", id],
    queryFn: () => listDependenciesByTarget(id!),
    enabled: !!id,
  });

  const targets = useQuery({ queryKey: ["targets"], queryFn: listTargets });

  const incidents = useQuery({
    queryKey: ["incidents", "open"],
    queryFn: () => listIncidents("open"),
  });

  const isLoading = svc.isLoading || targets.isLoading;
  const error = svc.error ?? depsOut.error ?? depsIn.error;

  if (isLoading) return <LoadingSpinner />;
  if (!svc.data) return <div className="p-6 text-gray-500 text-sm">Service not found.</div>;

  const service = svc.data;

  // Targets whose service_id matches — gosre-api stores this FK
  const linkedTargets = (targets.data ?? []).filter(
    (t) => t.metadata?.["service_id"] === service.id
  );

  // Recent open incidents for linked targets
  const linkedTargetIds = new Set(linkedTargets.map((t) => t.id));
  const relatedIncidents = (incidents.data ?? []).filter(
    (i) => i.target_id && linkedTargetIds.has(i.target_id)
  );

  // Build a map of service id → name for dependency display
  return (
    <div className="p-6 space-y-6 max-w-4xl">
      {/* Header */}
      <div>
        <div className="flex items-center gap-2 text-xs text-gray-500 mb-3">
          <Link to="/catalog" className="hover:text-gray-300 transition-colors">Catalog</Link>
          <span>/</span>
          <span className="text-gray-300">{service.name}</span>
        </div>
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-xl font-semibold text-white">{service.name}</h1>
            <p className="text-sm text-gray-400 mt-1">{service.owner || "No owner"}</p>
          </div>
          <span className={`inline-block px-2.5 py-1 rounded text-xs border font-medium ${criticalityStyle[service.criticality]}`}>
            {service.criticality}
          </span>
        </div>

        {(service.runbook_url || service.repo_url) && (
          <div className="flex items-center gap-4 mt-3">
            {service.runbook_url && (
              <a
                href={service.runbook_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-brand hover:text-brand/80 transition-colors font-mono flex items-center gap-1"
              >
                📄 Runbook ↗
              </a>
            )}
            {service.repo_url && (
              <a
                href={service.repo_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-brand hover:text-brand/80 transition-colors font-mono flex items-center gap-1"
              >
                🔗 Repository ↗
              </a>
            )}
          </div>
        )}
      </div>

      <ErrorMessage error={error as Error | null} />

      {/* Linked targets */}
      <Section title={`Targets (${linkedTargets.length})`}>
        {linkedTargets.length === 0 ? (
          <p className="text-xs text-gray-600">
            No targets linked to this service.
            Link them via <code className="font-mono bg-surface-border px-1 rounded">target.metadata.service_id</code>.
          </p>
        ) : (
          <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-surface-border">
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Name</th>
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Type</th>
                  <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Address</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {linkedTargets.map((t) => (
                  <tr key={t.id} className="hover:bg-surface-border/30 transition-colors">
                    <td className="px-4 py-3 text-white font-medium">{t.name}</td>
                    <td className="px-4 py-3 text-brand font-mono text-xs">{t.type}</td>
                    <td className="px-4 py-3 text-gray-400 font-mono text-xs truncate max-w-xs">{t.address}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Section>

      {/* Open incidents */}
      {relatedIncidents.length > 0 && (
        <Section title={`Open Incidents (${relatedIncidents.length})`}>
          <div className="bg-surface-raised border border-surface-border rounded-lg divide-y divide-surface-border">
            {relatedIncidents.map((inc) => {
              const target = linkedTargets.find((t) => t.id === inc.target_id);
              return (
                <div key={inc.id} className="flex items-center justify-between px-4 py-3">
                  <div>
                    <span className="text-sm text-white">{target?.name ?? inc.target_id}</span>
                    {inc.first_seen && (
                      <span className="ml-3 text-xs text-gray-500">
                        since {new Date(inc.first_seen).toLocaleString()}
                      </span>
                    )}
                  </div>
                  <StatusBadge status={(inc.state ?? "open") as IncidentState} />
                </div>
              );
            })}
          </div>
        </Section>
      )}

      {/* Dependencies outbound */}
      <Section title={`Depends On (${depsOut.data?.length ?? 0})`}>
        {(depsOut.data ?? []).length === 0 ? (
          <p className="text-xs text-gray-600">No outbound dependencies.</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {depsOut.data!.map((d) => (
              <div
                key={d.id}
                className="flex items-center gap-2 bg-surface-raised border border-surface-border rounded px-3 py-1.5"
              >
                <span className={`text-xs font-mono font-medium ${depKindStyle[d.kind]}`}>{d.kind}</span>
                <span className="text-xs text-gray-500">→</span>
                <Link
                  to={`/catalog/${d.target_service_id}`}
                  className="text-xs text-gray-300 hover:text-white transition-colors font-mono"
                >
                  {d.target_service_id.slice(0, 8)}…
                </Link>
              </div>
            ))}
          </div>
        )}
      </Section>

      {/* Dependencies inbound */}
      <Section title={`Depended On By (${depsIn.data?.length ?? 0})`}>
        {(depsIn.data ?? []).length === 0 ? (
          <p className="text-xs text-gray-600">No inbound dependencies.</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {depsIn.data!.map((d) => (
              <div
                key={d.id}
                className="flex items-center gap-2 bg-surface-raised border border-surface-border rounded px-3 py-1.5"
              >
                <Link
                  to={`/catalog/${d.source_service_id}`}
                  className="text-xs text-gray-300 hover:text-white transition-colors font-mono"
                >
                  {d.source_service_id.slice(0, 8)}…
                </Link>
                <span className="text-xs text-gray-500">→</span>
                <span className={`text-xs font-mono font-medium ${depKindStyle[d.kind]}`}>{d.kind}</span>
              </div>
            ))}
          </div>
        )}
      </Section>

      {/* Metadata */}
      <Section title="Info">
        <div className="bg-surface-raised border border-surface-border rounded-lg px-4 py-3 space-y-2">
          <div className="flex items-center gap-8">
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">ID</p>
              <p className="text-xs text-gray-300 font-mono">{service.id}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Created</p>
              <p className="text-xs text-gray-300">{new Date(service.created_at).toLocaleDateString()}</p>
            </div>
            {service.project_id && (
              <div>
                <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Project</p>
                <p className="text-xs text-gray-300 font-mono">{service.project_id.slice(0, 8)}…</p>
              </div>
            )}
          </div>
        </div>
      </Section>
    </div>
  );
}
