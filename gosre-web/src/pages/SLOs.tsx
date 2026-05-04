import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listSLOs, createSLO, deleteSLO, getSLOBudget, type SLO } from "../api/slos";
import { listTargets } from "../api/targets";
import { useRole } from "../hooks/useRole";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";
import EmptyState from "../components/EmptyState";

function ComplianceBadge({ value, insufficient }: { value: number; insufficient: boolean }) {
  if (insufficient) {
    return (
      <span className="inline-block px-2 py-0.5 rounded text-xs bg-surface-border text-gray-500 border border-surface-border font-mono">
        no data
      </span>
    );
  }
  const pct = (value * 100).toFixed(2);
  const ok = value >= 0.99;
  const warn = value >= 0.95;
  return (
    <span
      className={`inline-block px-2 py-0.5 rounded text-xs border font-mono font-medium ${
        ok
          ? "bg-status-ok/15 text-status-ok border-status-ok/30"
          : warn
          ? "bg-status-timeout/15 text-status-timeout border-status-timeout/30"
          : "bg-status-fail/15 text-status-fail border-status-fail/30"
      }`}
    >
      {pct}%
    </span>
  );
}

function fmtWindow(secs: number): string {
  if (secs >= 86400 * 7) return `${secs / (86400 * 7)}w`;
  if (secs >= 86400) return `${secs / 86400}d`;
  if (secs >= 3600) return `${secs / 3600}h`;
  return `${secs}s`;
}

const METRICS = ["availability", "latency", "error_rate"];
const WINDOWS_SECS: { label: string; value: number }[] = [
  { label: "1h", value: 3600 },
  { label: "6h", value: 21600 },
  { label: "24h", value: 86400 },
  { label: "7d", value: 604800 },
  { label: "30d", value: 2592000 },
];

interface CreateForm {
  target_id: string;
  name: string;
  metric: string;
  threshold: string;
  window_seconds: number;
}

const EMPTY_FORM: CreateForm = {
  target_id: "",
  name: "",
  metric: "availability",
  threshold: "0.99",
  window_seconds: 86400,
};

export default function SLOs() {
  const qc = useQueryClient();
  const { hasMinRole } = useRole();
  const [selectedTarget, setSelectedTarget] = useState<string>("");
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<CreateForm>(EMPTY_FORM);

  const targets = useQuery({ queryKey: ["targets"], queryFn: listTargets });

  const slos = useQuery({
    queryKey: ["slos", selectedTarget],
    queryFn: () => listSLOs(selectedTarget),
    enabled: !!selectedTarget,
  });

  const create = useMutation({
    mutationFn: (f: CreateForm) =>
      createSLO({
        target_id: f.target_id,
        name: f.name,
        metric: f.metric,
        threshold: parseFloat(f.threshold),
        window_seconds: f.window_seconds,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["slos"] });
      setForm(EMPTY_FORM);
      setShowForm(false);
    },
  });

  const del = useMutation({
    mutationFn: deleteSLO,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["slos"] }),
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    create.mutate(form);
  }

  const isLoading = targets.isLoading;
  const error = targets.error ?? slos.error ?? create.error ?? del.error;

  if (isLoading) return <LoadingSpinner />;

  const targetMap = Object.fromEntries((targets.data ?? []).map((t) => [t.id, t.name]));

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-lg font-semibold text-white">SLOs</h1>
          <p className="text-xs text-gray-500 mt-0.5">Service Level Objectives per target</p>
        </div>
        {!showForm && selectedTarget && hasMinRole("operator") && (
          <button
            onClick={() => { setShowForm(true); setForm((f) => ({ ...f, target_id: selectedTarget })); }}
            className="text-xs px-3 py-1.5 rounded border border-brand text-brand hover:bg-brand hover:text-black transition-colors"
          >
            Add SLO
          </button>
        )}
      </div>

      <ErrorMessage error={error as Error | null} />

      {/* Target selector */}
      <div className="mb-5">
        <label className="text-xs text-gray-500 uppercase tracking-wider block mb-1.5">Target</label>
        <select
          value={selectedTarget}
          onChange={(e) => setSelectedTarget(e.target.value)}
          className="bg-surface-raised border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand w-64"
        >
          <option value="">— select a target —</option>
          {(targets.data ?? []).map((t) => (
            <option key={t.id} value={t.id}>{t.name}</option>
          ))}
        </select>
      </div>

      {/* Create form */}
      {showForm && (
        <form
          onSubmit={handleSubmit}
          className="mb-5 bg-surface-raised border border-surface-border rounded-lg p-4 space-y-4"
        >
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Name</label>
              <input
                required
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                placeholder="API availability"
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Metric</label>
              <select
                value={form.metric}
                onChange={(e) => setForm((f) => ({ ...f, metric: e.target.value }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
              >
                {METRICS.map((m) => <option key={m} value={m}>{m}</option>)}
              </select>
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Threshold</label>
              <input
                required
                type="number"
                step="0.001"
                min="0.001"
                max="0.999"
                value={form.threshold}
                onChange={(e) => setForm((f) => ({ ...f, threshold: e.target.value }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand font-mono"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Window</label>
              <select
                value={form.window_seconds}
                onChange={(e) => setForm((f) => ({ ...f, window_seconds: parseInt(e.target.value) }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
              >
                {WINDOWS_SECS.map((w) => (
                  <option key={w.value} value={w.value}>{w.label}</option>
                ))}
              </select>
            </div>
          </div>
          <div className="flex gap-2">
            <button
              type="submit"
              disabled={create.isPending}
              className="text-xs px-3 py-1.5 rounded bg-brand text-black font-medium hover:bg-brand/90 transition-colors disabled:opacity-40"
            >
              {create.isPending ? "Creating…" : "Create"}
            </button>
            <button
              type="button"
              onClick={() => { setShowForm(false); setForm(EMPTY_FORM); }}
              className="text-xs px-3 py-1.5 rounded border border-surface-border text-gray-400 hover:text-white transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      {/* SLO list */}
      {!selectedTarget ? (
        <EmptyState message="Select a target to view its SLOs." />
      ) : slos.isLoading ? (
        <LoadingSpinner />
      ) : (slos.data ?? []).length === 0 ? (
        <EmptyState message="No SLOs defined for this target." />
      ) : (
        <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Name</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Target</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Metric</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Threshold</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Window</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Compliance</th>
                <th className="px-4 py-2" />
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {(slos.data ?? []).map((slo: SLO) => (
                <SLORow
                  key={slo.id}
                  slo={slo}
                  targetName={targetMap[slo.target_id] ?? slo.target_id}
                  canDelete={hasMinRole("admin")}
                  onDelete={() => {
                    if (confirm(`Delete SLO "${slo.name}"?`)) del.mutate(slo.id);
                  }}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function SLORow({
  slo,
  targetName,
  canDelete,
  onDelete,
}: {
  slo: SLO;
  targetName: string;
  canDelete: boolean;
  onDelete: () => void;
}) {
  const budget = useQuery({
    queryKey: ["slo-budget", slo.id],
    queryFn: () => getSLOBudget(slo.id),
    staleTime: 60_000,
  });

  return (
    <tr className="hover:bg-surface-border/30 transition-colors">
      <td className="px-4 py-3">
        <Link to={`/slos/${slo.id}`} className="text-white font-medium hover:text-brand transition-colors">
          {slo.name}
        </Link>
      </td>
      <td className="px-4 py-3 text-gray-300 text-sm">{targetName}</td>
      <td className="px-4 py-3 text-brand font-mono text-xs">{slo.metric}</td>
      <td className="px-4 py-3 text-gray-300 font-mono text-xs">
        {(slo.threshold * 100).toFixed(1)}%
      </td>
      <td className="px-4 py-3 text-gray-400 font-mono text-xs">{fmtWindow(slo.window_seconds)}</td>
      <td className="px-4 py-3">
        {budget.isLoading ? (
          <span className="text-xs text-gray-600 font-mono">…</span>
        ) : budget.data ? (
          <ComplianceBadge
            value={budget.data.compliance}
            insufficient={budget.data.insufficient_data}
          />
        ) : (
          <span className="text-xs text-gray-600">—</span>
        )}
      </td>
      <td className="px-4 py-3 text-right">
        <div className="flex justify-end gap-3">
          <Link
            to={`/slos/${slo.id}`}
            className="text-xs text-gray-500 hover:text-gray-200 transition-colors"
          >
            detail
          </Link>
          {canDelete && (
            <button
              onClick={onDelete}
              className="text-xs text-gray-500 hover:text-status-fail transition-colors"
            >
              delete
            </button>
          )}
        </div>
      </td>
    </tr>
  );
}
