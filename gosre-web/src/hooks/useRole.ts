import { useAuth } from "./useAuth";

export type Role = 'viewer' | 'operator' | 'admin' | 'owner';

const ROLE_HIERARCHY: Record<Role, number> = {
  viewer: 0,
  operator: 1,
  admin: 2,
  owner: 3,
};

export function hasMinRole(userRole: Role, minRole: Role): boolean {
  return (ROLE_HIERARCHY[userRole] ?? -1) >= ROLE_HIERARCHY[minRole];
}

export function useRole() {
  const { user } = useAuth();
  const role = (user?.role ?? 'viewer') as Role;
  return {
    role,
    hasMinRole: (minRole: Role) => hasMinRole(role, minRole),
  };
}
