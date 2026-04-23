import { useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  listOrganizations,
  listTeams,
  createTeam,
  deleteTeam,
  type Team,
} from "../../api/organizations";
import LoadingSpinner from "../../components/LoadingSpinner";
import ErrorMessage from "../../components/ErrorMessage";
import EmptyState from "../../components/EmptyState";

const EMPTY_FORM = { name: "", slug: "" };

export default function TeamsSettings() {
  const { orgId } = useParams<{ orgId: string }>();
  const qc = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(EMPTY_FORM);

  const { data: orgs } = useQuery({
    queryKey: ["organizations"],
    queryFn: listOrganizations,
  });

  const { data, isLoading, error } = useQuery({
    queryKey: ["teams", orgId],
    queryFn: () => listTeams(orgId!),
    enabled: !!orgId,
  });

  const del = useMutation({
    mutationFn: (teamId: string) => deleteTeam(orgId!, teamId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["teams", orgId] }),
  });

  const create = useMutation({
    mutationFn: (body: { name: string; slug?: string }) => createTeam(orgId!, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["teams", orgId] });
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

  const orgName = orgs?.find((o) => o.id === orgId)?.name ?? orgId;

  if (isLoading) return <LoadingSpinner />;

  return (
    <div className="p-6">
      <div className="mb-5">
        <div className="text-xs text-gray-500 mb-2 flex items-center gap-1.5">
          <Link to="/settings/organization" className="hover:text-gray-300 transition-colors">
            Organizations
          </Link>
          <span>/</span>
          <span className="text-gray-400">{orgName}</span>
          <span>/</span>
          <span className="text-gray-300">Teams</span>
        </div>
        <div className="flex items-center justify-between">
          <h1 className="text-lg font-semibold text-white">Teams</h1>
          {!showForm && (
            <button
              onClick={() => setShowForm(true)}
              className="text-xs px-3 py-1.5 rounded border border-brand text-brand hover:bg-brand hover:text-black transition-colors"
            >
              New Team
            </button>
          )}
        </div>
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
                placeholder="Platform Engineering"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs text-gray-500 uppercase tracking-wider">Slug (optional)</label>
              <input
                value={form.slug}
                onChange={(e) => setForm((f) => ({ ...f, slug: e.target.value }))}
                className="bg-surface border border-surface-border rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-brand"
                placeholder="platform-eng"
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
        <EmptyState message="No teams found." />
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
              {data.map((team: Team) => (
                <tr key={team.id} className="hover:bg-surface-border/30 transition-colors">
                  <td className="px-4 py-3 text-white font-medium">{team.name}</td>
                  <td className="px-4 py-3 text-gray-400 font-mono text-xs">{team.slug}</td>
                  <td className="px-4 py-3 text-gray-500 text-xs">
                    {new Date(team.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => {
                        if (confirm(`Delete team "${team.name}"?`)) {
                          del.mutate(team.id);
                        }
                      }}
                      disabled={del.isPending}
                      className="text-xs text-gray-500 hover:text-status-fail transition-colors disabled:opacity-40"
                    >
                      delete
                    </button>
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
