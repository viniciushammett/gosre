import { NavLink, Outlet } from "react-router-dom";
import OrgSwitcher from "./OrgSwitcher";

const navItems = [
  { to: "/", label: "Dashboard", end: true },
  { to: "/targets", label: "Targets" },
  { to: "/incidents", label: "Incidents" },
  { to: "/results", label: "Results" },
  { to: "/checks", label: "Checks" },
  { to: "/agents", label: "Agents" },
  { to: "/settings/organization", label: "Settings" },
];

export default function Layout() {
  return (
    <div className="flex h-screen overflow-hidden">
      <aside className="w-52 flex-shrink-0 bg-surface-raised border-r border-surface-border flex flex-col">
        <div className="px-5 py-4 border-b border-surface-border">
          <span className="text-brand font-semibold tracking-wide text-sm uppercase">
            GoSRE
          </span>
        </div>
        <nav className="flex-1 py-3">
          {navItems.map(({ to, label, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                [
                  "flex items-center px-5 py-2 text-sm transition-colors",
                  isActive
                    ? "text-white bg-surface-border"
                    : "text-gray-400 hover:text-gray-100 hover:bg-surface-border",
                ].join(" ")
              }
            >
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="flex-1 flex flex-col min-w-0">
        <header className="h-12 flex items-center justify-between px-6 border-b border-surface-border bg-surface-raised flex-shrink-0">
          <span className="text-xs text-gray-500 font-mono">
            {import.meta.env.VITE_API_URL ?? "http://localhost:8080"}
          </span>
          <OrgSwitcher />
        </header>
        <main className="flex-1 overflow-y-auto bg-surface">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
