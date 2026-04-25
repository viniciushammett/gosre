import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { useRole } from "../../hooks/useRole";
import {
  listOrganizations,
  createOrganization,
  deleteOrganization,
  type Organization,
} from "../../api/organizations";
import LoadingSpinner from "../../components/LoadingSpinner";
import ErrorMessage from "../../components/ErrorMessage";
import EmptyState from "../../components/EmptyState";

const EMPTY_FORM = { name: "", slug: "" };

export default function OrganizationSettings() {
  const qc = useQueryClient();
  const navigate = useNavigate();
  const { hasMinRole } = useRole();
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);

  const { data, isLoading, error } = useQuery({
    queryKey: ["organizations"],
    queryFn: listOrganizations,
  });

  const del = useMutation({
    mutationFn: deleteOrganization,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["organizations"] }),
  });

  const create = useMutation({
    mutationFn: (body: { name: string; slug?: string }) => createOrganization(body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["organizations"] });
      setForm(EMPTY_FORM);
      setShowForm(false);
    },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const body: { name: string; slug?: string } = { name: form.name.trim() };
    if (form.slug.trim()) body.slug = form.slug.trim();
    create.mutate(body);
  }

  if (isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold text-white">Organizations</h1>
        {!showForm && hasMinRole('admin') && (
          <button
            onClick={() => setShowForm(true)}
            className="text-xs px-3 py-1.5 rounded border border-brand text-brand hover:bg-brand hover:text-black transition-colors"
          >
            New Organization
          </button>
        )}
      </div>

      <ErrorMessage error={(error ?? del.error ?? create.error) as Error | null} />

      {showForm && (
        <form
          onSubmit={handleSubmit}
          className="mb-5 bg-surface-raised border border-surface-border rounded-lg p-4 space-y-4"
        >
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Name</label>
              <input
                required
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                placeholder="Acme Corp"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Slug (optional)</label>
              <input
                value={form.slug}
                onChange={(e) => setForm((f) => ({ ...f, slug: e.target.value }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                placeholder="acme-corp"
              />
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

      {(!data || data.length === 0) ? (
        <EmptyState message="No organizations found." />
      ) : (
        <div className="bg-surface-raised border border-surface-border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border">
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Name</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Slug</th>
                <th className="text-left px-4 py-2 text-xs text-gray-500 uppercase tracking-wider font-medium">Created</th>
                <th className="px-4 py-2" />
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {data.map((org: Organization) => (
                <tr key={org.id} className="hover:bg-surface-border/30 transition-colors">
                  <td className="px-4 py-3 text-white font-medium">{org.name}</td>
                  <td className="px-4 py-3 text-gray-400 font-mono text-xs">{org.slug}</td>
                  <td className="px-4 py-3 text-gray-500 text-xs">
                    {new Date(org.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex justify-end gap-3">
                      <button
                        onClick={() => navigate(`/settings/organization/${org.id}/teams`)}
                        className="text-xs text-gray-500 hover:text-gray-200 transition-colors"
                      >
                        teams
                      </button>
                      {hasMinRole('owner') && <button
                        onClick={() => {
                          if (confirm(`Delete organization "${org.name}"?`)) {
                            del.mutate(org.id);
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
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
