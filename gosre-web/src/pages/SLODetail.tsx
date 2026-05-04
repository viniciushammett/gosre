import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
  Cell,
} from "recharts";
import { getSLO, getSLOBudget } from "../api/slos";
import { getTarget } from "../api/targets";
import LoadingSpinner from "../components/LoadingSpinner";
import ErrorMessage from "../components/ErrorMessage";

function fmtWindow(secs: number): string {
  if (secs >= 86400 * 7) return `${secs / (86400 * 7)}w`;
  if (secs >= 86400) return `${secs / 86400}d`;
  if (secs >= 3600) return `${secs / 3600}h`;
  return `${secs}s`;
}

function ComplianceGauge({ value, threshold }: { value: number; threshold: number }) {
  const pct = value * 100;
  const ok = value >= threshold;
  const color = ok ? "#22c55e" : value >= threshold * 0.99 ? "#f59e0b" : "#ef4444";
  return (
    <div className="flex flex-col items-center gap-2">
      <div
        className="text-4xl font-semibold font-mono tabular-nums"
        style={{ color }}
      >
        {pct.toFixed(3)}%
      </div>
      <div className="text-xs text-gray-500">
        target: {(threshold * 100).toFixed(1)}%
      </div>
    </div>
  );
}

interface BurnBarProps {
  label: string;
  value: number;
}

function BurnBar({ label, value }: BurnBarProps) {
  const capped = Math.min(value, 5);
  const color = value > 3 ? "#ef4444" : value > 1 ? "#f59e0b" : "#22c55e";
  return (
    <div className="flex items-center gap-3">
      <span className="text-xs text-gray-500 font-mono w-6 text-right">{label}</span>
      <div className="flex-1 bg-surface-border rounded-full h-2 overflow-hidden">
        <div
          className="h-2 rounded-full transition-all"
          style={{ width: `${Math.min((capped / 5) * 100, 100)}%`, backgroundColor: color }}
        />
      </div>
      <span className="text-xs font-mono tabular-nums w-10 text-right" style={{ color }}>
        {value.toFixed(2)}x
      </span>
    </div>
  );
}

const CHART_DATA_LABELS = ["1h", "6h", "24h"];

export default function SLODetail() {
  const { id } = useParams<{ id: string }>();

  const slo = useQuery({
    queryKey: ["slo", id],
    queryFn: () => getSLO(id!),
    enabled: !!id,
  });

  const budget = useQuery({
    queryKey: ["slo-budget", id],
    queryFn: () => getSLOBudget(id!),
    enabled: !!id,
    refetchInterval: 60_000,
  });

  const target = useQuery({
    queryKey: ["target", slo.data?.target_id],
    queryFn: () => getTarget(slo.data!.target_id),
    enabled: !!slo.data?.target_id,
  });

  const isLoading = slo.isLoading;
  const error = slo.error ?? budget.error;

  if (isLoading) return <LoadingSpinner />;
  if (!slo.data) return <div className="p-6 text-gray-500 text-sm">SLO not found.</div>;

  const s = slo.data;
  const b = budget.data;

  const chartData = b && !b.insufficient_data
    ? [
        { name: "1h", value: b.burn_rate_1h },
        { name: "6h", value: b.burn_rate_6h },
        { name: "24h", value: b.burn_rate_24h },
      ]
    : CHART_DATA_LABELS.map((name) => ({ name, value: 0 }));

  return (
    <div className="p-6 max-w-3xl space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-xs text-gray-500">
        <Link to="/slos" className="hover:text-gray-300 transition-colors">SLOs</Link>
        <span>/</span>
        <span className="text-gray-300">{s.name}</span>
      </div>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold text-white">{s.name}</h1>
          <div className="flex items-center gap-3 mt-1.5">
            <span className="text-xs text-gray-500 font-mono">{s.metric}</span>
            <span className="text-xs text-gray-600">·</span>
            <span className="text-xs text-gray-500 font-mono">{fmtWindow(s.window_seconds)} window</span>
            {target.data && (
              <>
                <span className="text-xs text-gray-600">·</span>
                <span className="text-xs text-brand font-mono">{target.data.name}</span>
              </>
            )}
          </div>
        </div>
      </div>

      <ErrorMessage error={error as Error | null} />

      {/* Compliance + burn rates */}
      {b?.insufficient_data ? (
        <div className="bg-surface-raised border border-surface-border rounded-lg px-5 py-6 text-center">
          <p className="text-sm text-gray-400">Insufficient data</p>
          <p className="text-xs text-gray-600 mt-1">
            Compliance requires at least 1,000 results in the window.
            Currently {b.total_results} result{b.total_results !== 1 ? "s" : ""}.
          </p>
        </div>
      ) : b ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {/* Compliance */}
          <div className="bg-surface-raised border border-surface-border rounded-lg px-5 py-5">
            <p className="text-xs text-gray-500 uppercase tracking-wider mb-4">Compliance</p>
            <ComplianceGauge value={b.compliance} threshold={s.threshold} />
            <p className="text-xs text-gray-600 text-center mt-3">
              {b.total_results.toLocaleString()} results in window
            </p>
          </div>

          {/* Burn rates */}
          <div className="bg-surface-raised border border-surface-border rounded-lg px-5 py-5">
            <p className="text-xs text-gray-500 uppercase tracking-wider mb-4">Burn Rate</p>
            <div className="space-y-3">
              <BurnBar label="1h" value={b.burn_rate_1h} />
              <BurnBar label="6h" value={b.burn_rate_6h} />
              <BurnBar label="24h" value={b.burn_rate_24h} />
            </div>
            <p className="text-xs text-gray-600 mt-3">
              &gt;1× = consuming budget faster than sustainable
            </p>
          </div>
        </div>
      ) : budget.isLoading ? (
        <LoadingSpinner />
      ) : null}

      {/* Burn rate chart */}
      {b && !b.insufficient_data && (
        <div className="bg-surface-raised border border-surface-border rounded-lg px-5 py-5">
          <p className="text-xs text-gray-500 uppercase tracking-wider mb-4">Burn Rate by Window</p>
          <ResponsiveContainer width="100%" height={160}>
            <BarChart data={chartData} barSize={32} margin={{ top: 4, right: 4, bottom: 0, left: -20 }}>
              <XAxis
                dataKey="name"
                tick={{ fontSize: 11, fill: "#6b7280", fontFamily: "IBM Plex Mono" }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis
                tick={{ fontSize: 10, fill: "#4b5563", fontFamily: "IBM Plex Mono" }}
                axisLine={false}
                tickLine={false}
                domain={[0, "auto"]}
              />
              <Tooltip
                contentStyle={{
                  background: "#171b26",
                  border: "1px solid #1e2233",
                  borderRadius: "6px",
                  fontSize: "11px",
                  fontFamily: "IBM Plex Mono",
                  color: "#e5e7eb",
                }}
                formatter={(v) => [`${(v as number).toFixed(3)}×`, "burn rate"]}
                cursor={{ fill: "rgba(30,34,51,0.5)" }}
              />
              <ReferenceLine
                y={1}
                stroke="#f59e0b"
                strokeDasharray="3 3"
                label={{ value: "1×", fill: "#f59e0b", fontSize: 10, fontFamily: "IBM Plex Mono" }}
              />
              <Bar dataKey="value" radius={[3, 3, 0, 0]}>
                {chartData.map((entry) => (
                  <Cell
                    key={entry.name}
                    fill={entry.value > 3 ? "#ef4444" : entry.value > 1 ? "#f59e0b" : "#22c55e"}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* SLO info */}
      <div className="bg-surface-raised border border-surface-border rounded-lg px-4 py-3">
        <div className="flex items-center gap-8 flex-wrap">
          <div>
            <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">ID</p>
            <p className="text-xs text-gray-300 font-mono">{s.id}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Threshold</p>
            <p className="text-xs text-gray-300 font-mono">{(s.threshold * 100).toFixed(2)}%</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Window</p>
            <p className="text-xs text-gray-300 font-mono">{fmtWindow(s.window_seconds)}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">Target ID</p>
            <p className="text-xs text-gray-300 font-mono">{s.target_id}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
