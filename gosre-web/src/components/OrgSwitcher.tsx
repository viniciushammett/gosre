import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listOrganizations, type Organization } from "../api/organizations";

const STORAGE_KEY = "gosre_selected_org";

function getStoredOrg(): Organization | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? (JSON.parse(raw) as Organization) : null;
  } catch {
    return null;
  }
}

function storeOrg(org: Organization) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(org));
}

export default function OrgSwitcher() {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<Organization | null>(getStoredOrg);

  const { data: orgs = [] } = useQuery({
    queryKey: ["organizations"],
    queryFn: listOrganizations,
  });

  useEffect(() => {
    if (!selected && orgs.length > 0) {
      setSelected(orgs[0]);
      storeOrg(orgs[0]);
    }
  }, [orgs, selected]);

  function select(org: Organization) {
    setSelected(org);
    storeOrg(org);
    setOpen(false);
  }

  if (orgs.length === 0) {
    return (
      <Link
        to="/settings/organization"
        className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
      >
        No organization
      </Link>
    );
  }

  return (
    <div className="relative">
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex items-center gap-1.5 text-xs text-gray-300 hover:text-white transition-colors"
      >
        <span className="w-1.5 h-1.5 rounded-full bg-brand" />
        {selected?.name ?? "Select org"}
        <svg className="w-3 h-3 text-gray-500" viewBox="0 0 16 16" fill="none">
          <path d="M4 6l4 4 4-4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute right-0 top-full mt-2 z-20 min-w-40 bg-surface-raised border border-surface-border rounded-lg shadow-lg overflow-hidden">
            {orgs.map((org) => (
              <button
                key={org.id}
                onClick={() => select(org)}
                className={[
                  "w-full text-left px-3 py-2 text-xs transition-colors",
                  selected?.id === org.id
                    ? "text-white bg-surface-border"
                    : "text-gray-400 hover:text-gray-100 hover:bg-surface-border",
                ].join(" ")}
              >
                {org.name}
              </button>
            ))}
            <div className="border-t border-surface-border">
              <Link
                to="/settings/organization"
                onClick={() => setOpen(false)}
                className="block px-3 py-2 text-xs text-gray-500 hover:text-gray-300 transition-colors"
              >
                Manage organizations
              </Link>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
