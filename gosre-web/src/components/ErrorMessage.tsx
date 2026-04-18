interface Props {
  error: Error | null;
}

export default function ErrorMessage({ error }: Props) {
  if (!error) return null;
  return (
    <div className="mx-6 mt-6 px-4 py-3 rounded bg-status-fail/10 border border-status-fail/30 text-status-fail text-sm font-mono">
      {error.message}
    </div>
  );
}
