import { Navigate } from "react-router-dom";
import { getAccessToken } from "../api/auth";

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  if (!getAccessToken()) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}
