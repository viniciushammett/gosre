import type { CheckStatus, IncidentState } from "../api/client";

type Status = CheckStatus | IncidentState;

const styles: Record<Status, string> = {
  ok: "bg-status-ok/15 text-status-ok",
  fail: "bg-status-fail/15 text-status-fail",
  timeout: "bg-status-timeout/15 text-status-timeout",
  unknown: "bg-status-unknown/15 text-status-unknown",
  open: "bg-status-fail/15 text-status-fail",
  acknowledged: "bg-status-timeout/15 text-status-timeout",
  resolved: "bg-status-ok/15 text-status-ok",
};

interface Props {
  status: Status;
}

export default function StatusBadge({ status }: Props) {
  return (
    <span
      className={[
        "inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium",
        styles[status] ?? styles.unknown,
      ].join(" ")}
    >
      {status}
    </span>
  );
}
