import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useRole } from "../hooks/useRole";
import { listTargets, deleteTarget, createTarget, updateTarget } from "../api/targets";
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

const TAG_SUGGESTIONS = [
  "production", "staging", "development",
  "http", "tcp", "dns", "tls",
  "critical", "monitoring",
];

const EMPTY_FORM = { name: "", type: "http" as TargetType, address: "", tags: [] as string[] };

type FormState = { name: string; type: TargetType; address: string; tags: string[] };

function TagChips({ tags, onToggle }: { tags: string[]; onToggle: (tag: string) => void }) {
  return (
    <div className="flex flex-wrap gap-2">
      {TAG_SUGGESTIONS.map((tag) => (
        <button
          key={tag}
          type="button"
          onClick={() => onToggle(tag)}
          className={`px-2 py-0.5 rounded text-xs border transition-colors ${
            tags.includes(tag)
              ? "border-brand text-brand bg-brand/10"
              : "border-surface-border text-gray-500 hover:border-gray-400 hover:text-gray-300"
          }`}
        >
          {tag}
        </button>
      ))}
    </div>
  );
}

export default function Targets() {
  const qc = useQueryClient();
  const { hasMinRole } = useRole();
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<FormState>(EMPTY_FORM);

  const { data, isLoading, error } = useQuery({
    queryKey: ["targets"],
    queryFn: listTargets,
  });

  const del = useMutation({
    mutationFn: deleteTarget,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["targets"] });
      qc.invalidateQueries({ queryKey: ["checks"] });
      qc.invalidateQueries({ queryKey: ["results"] });
    },
  });

  const create = useMutation({
    mutationFn: createTarget,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["targets"] });
      qc.invalidateQueries({ queryKey: ["checks"] });
      setForm(EMPTY_FORM);
      setShowForm(false);
    },
  });

  const update = useMutation({
    mutationFn: ({ id, body }: { id: string; body: FormState }) =>
      updateTarget(id, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["targets"] });
      setEditingId(null);
    },
  });

  function toggleTag(tag: string) {
    setForm((f) => ({
      ...f,
      tags: f.tags.includes(tag) ? f.tags.filter((t) => t !== tag) : [...f.tags, tag],
    }));
  }

  function toggleEditTag(tag: string) {
    setEditForm((f) => ({
      ...f,
      tags: f.tags.includes(tag) ? f.tags.filter((t) => t !== tag) : [...f.tags, tag],
    }));
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    let address = form.address.trim();
    if (form.type === "http" && !/^https?:\/\//i.test(address)) {
      address = "https://" + address;
    }
    create.mutate({ name: form.name, type: form.type, address, tags: form.tags });
  }

  function handleEditSubmit(e: React.FormEvent, id: string) {
    e.preventDefault();
    let address = editForm.address.trim();
    if (
      editForm.type === "http" &&
      !address.startsWith("http://") &&
      !address.startsWith("https://")
    ) {
      address = "https://" + address;
    }
    update.mutate({ id, body: { ...editForm, address } });
  }

  function startEdit(t: { id?: string; name?: string; type?: TargetType; address?: string; tags?: string[] }) {
    setEditingId(t.id!);
    setEditForm({
      name: t.name ?? "",
      type: t.type ?? "http",
      address: t.address ?? "",
      tags: t.tags ?? [],
    });
  }

  if (isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold text-white">Targets</h1>
        {!showForm && hasMinRole('operator') && (
          <button
            onClick={() => setShowForm(true)}
            className="text-xs px-3 py-1.5 rounded border border-brand text-brand hover:bg-brand hover:text-black transition-colors"
          >
            Add Target
          </button>
        )}
      </div>

      <ErrorMessage error={(error ?? del.error ?? create.error ?? update.error) as Error | null} />

      {showForm && (
        <form
          onSubmit={handleSubmit}
          className="mb-5 bg-surface-raised border border-surface-border rounded-lg p-4 space-y-4"
        >
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Name</label>
              <input
                required
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                placeholder="my-api"
              />
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Type</label>
              <select
                value={form.type}
                onChange={(e) => setForm((f) => ({ ...f, type: e.target.value as TargetType }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
              >
                <option value="http">http</option>
                <option value="tcp">tcp</option>
                <option value="dns">dns</option>
                <option value="tls">tls</option>
              </select>
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Address</label>
              <input
                required
                value={form.address}
                onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                placeholder="https://example.com"
              />
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-xs text-gray-500 uppercase tracking-wider">Tags</label>
            <TagChips tags={form.tags} onToggle={toggleTag} />
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
                editingId === t.id ? (
                  <tr key={t.id}>
                    <td colSpan={5} className="px-4 py-3">
                      <form onSubmit={(e) => handleEditSubmit(e, t.id!)} className="space-y-3">
                        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                          <div className="flex flex-col gap-1">
                            <label className="text-xs text-gray-500 uppercase tracking-wider">Name</label>
                            <input
                              required
                              value={editForm.name}
                              onChange={(e) => setEditForm((f) => ({ ...f, name: e.target.value }))}
                              className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                            />
                          </div>
                          <div className="flex flex-col gap-1">
                            <label className="text-xs text-gray-500 uppercase tracking-wider">Type</label>
                            <select
                              value={editForm.type}
                              onChange={(e) => setEditForm((f) => ({ ...f, type: e.target.value as TargetType }))}
                              className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                            >
                              <option value="http">http</option>
                              <option value="tcp">tcp</option>
                              <option value="dns">dns</option>
                              <option value="tls">tls</option>
                            </select>
                          </div>
                          <div className="flex flex-col gap-1">
                            <label className="text-xs text-gray-500 uppercase tracking-wider">Address</label>
                            <input
                              required
                              value={editForm.address}
                              onChange={(e) => setEditForm((f) => ({ ...f, address: e.target.value }))}
                              className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                            />
                          </div>
                        </div>
                        <div className="flex flex-col gap-2">
                          <label className="text-xs text-gray-500 uppercase tracking-wider">Tags</label>
                          <TagChips tags={editForm.tags} onToggle={toggleEditTag} />
                        </div>
                        <div className="flex gap-2">
                          <button
                            type="submit"
                            disabled={update.isPending}
                            className="text-xs px-3 py-1.5 rounded bg-brand text-black font-medium hover:bg-brand/90 transition-colors disabled:opacity-40"
                          >
                            {update.isPending ? "Saving…" : "Save"}
                          </button>
                          <button
                            type="button"
                            onClick={() => setEditingId(null)}
                            className="text-xs px-3 py-1.5 rounded border border-surface-border text-gray-400 hover:text-white transition-colors"
                          >
                            Cancel
                          </button>
                        </div>
                      </form>
                    </td>
                  </tr>
                ) : (
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
                      <div className="flex justify-end gap-3">
                        {hasMinRole('operator') && <button
                          onClick={() => startEdit(t)}
                          className="text-xs text-gray-500 hover:text-gray-200 transition-colors"
                        >
                          edit
                        </button>}
                        {hasMinRole('admin') && <button
                          onClick={() => {
                            if (confirm(`Delete target "${t.name}"?`)) {
                              del.mutate(t.id!);
                            }
                          }}
                          disabled={del.isPending}
                          className="text-xs text-gray-500 hover:text-status-fail transition-colors disabled:opacity-40"
                        >
                          delete
                        </button>}
                      </div>
                    </td>
                  </tr>
                )
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
