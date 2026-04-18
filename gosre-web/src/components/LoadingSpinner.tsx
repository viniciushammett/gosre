interface Props {
  className?: string;
}

export default function LoadingSpinner({ className = "" }: Props) {
  return (
    <div
      className={["flex items-center justify-center p-8", className].join(" ")}
    >
      <div className="w-6 h-6 border-2 border-surface-border border-t-brand rounded-full animate-spin" />
    </div>
  );
}
