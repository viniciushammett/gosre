import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  clearTokens,
  getAccessToken,
  login as apiLogin,
  logout as apiLogout,
  me,
  type AuthUser,
} from "../api/auth";

export type { AuthUser };

export function useAuth() {
  const queryClient = useQueryClient();

  const { data: user, isLoading } = useQuery({
    queryKey: ["auth", "me"],
    queryFn: me,
    retry: false,
    enabled: !!getAccessToken(),
    staleTime: 5 * 60 * 1000,
  });

  const loginMutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      apiLogin(email, password),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["auth", "me"] });
    },
  });

  const logoutMutation = useMutation({
    mutationFn: apiLogout,
    onSuccess: () => {
      clearTokens();
      queryClient.clear();
    },
  });

  return {
    user,
    isLoading: !!getAccessToken() && isLoading,
    isAuthenticated: !!user,
    login: loginMutation.mutateAsync,
    logout: logoutMutation.mutateAsync,
    loginError: loginMutation.error,
    isLoginPending: loginMutation.isPending,
  };
}
