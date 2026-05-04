import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import ProtectedRoute from "./components/ProtectedRoute";
import LoginPage from "./pages/LoginPage";
import Dashboard from "./pages/Dashboard";
import Targets from "./pages/Targets";
import Incidents from "./pages/Incidents";
import Results from "./pages/Results";
import Checks from "./pages/Checks";
import Agents from "./pages/Agents";
import Catalog from "./pages/Catalog";
import CatalogDetail from "./pages/CatalogDetail";
import SLOs from "./pages/SLOs";
import SLODetail from "./pages/SLODetail";
import Organization from "./pages/settings/Organization";
import Teams from "./pages/settings/Teams";
import Notifications from "./pages/settings/Notifications";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Dashboard />} />
            <Route path="targets" element={<Targets />} />
            <Route path="incidents" element={<Incidents />} />
            <Route path="results" element={<Results />} />
            <Route path="checks" element={<Checks />} />
            <Route path="agents" element={<Agents />} />
            <Route path="catalog" element={<Catalog />} />
            <Route path="catalog/:id" element={<CatalogDetail />} />
            <Route path="slos" element={<SLOs />} />
            <Route path="slos/:id" element={<SLODetail />} />
            <Route path="settings/organization" element={<Organization />} />
            <Route path="settings/organization/:orgId/teams" element={<Teams />} />
            <Route path="settings/notifications" element={<Notifications />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
