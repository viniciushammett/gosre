import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import Dashboard from "./pages/Dashboard";
import Targets from "./pages/Targets";
import Incidents from "./pages/Incidents";
import Results from "./pages/Results";
import Checks from "./pages/Checks";
import Agents from "./pages/Agents";
import Organization from "./pages/settings/Organization";
import Teams from "./pages/settings/Teams";

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
          <Route path="/" element={<Layout />}>
            <Route index element={<Dashboard />} />
            <Route path="targets" element={<Targets />} />
            <Route path="incidents" element={<Incidents />} />
            <Route path="results" element={<Results />} />
            <Route path="checks" element={<Checks />} />
            <Route path="agents" element={<Agents />} />
            <Route path="settings/organization" element={<Organization />} />
            <Route path="settings/organization/:orgId/teams" element={<Teams />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
