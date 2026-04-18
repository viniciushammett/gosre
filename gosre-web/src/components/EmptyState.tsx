interface Props {
  message: string;
}

export default function EmptyState({ message }: Props) {
  return (
    <div className="flex items-center justify-center p-12 text-gray-500 text-sm">
      {message}
    </div>
  );
}
