import { create } from "zustand";
import { persist } from "zustand/middleware";
import { authApi } from "../services/api";
import type { LoginRequest, RegisterRequest, User } from "../types";

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  login: (credentials: LoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => void;
  refreshTokenAction: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (credentials: LoginRequest) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authApi.login(credentials);
          const { user, access_token, refresh_token } = response.data;
          set({
            user,
            accessToken: access_token,
            refreshToken: refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });
          localStorage.setItem("access_token", access_token);
          localStorage.setItem("refresh_token", refresh_token);
        } catch (error: unknown) {
          const message =
            error instanceof Error ? error.message : "Login failed. Please try again.";
          set({ isLoading: false, error: message });
          throw error;
        }
      },

      register: async (data: RegisterRequest) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authApi.register(data);
          const { user, access_token, refresh_token } = response.data;
          set({
            user,
            accessToken: access_token,
            refreshToken: refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });
          localStorage.setItem("access_token", access_token);
          localStorage.setItem("refresh_token", refresh_token);
        } catch (error: unknown) {
          const message =
            error instanceof Error ? error.message : "Registration failed. Please try again.";
          set({ isLoading: false, error: message });
          throw error;
        }
      },

      logout: () => {
        const { refreshToken } = get();
        if (refreshToken) {
          authApi.logout().catch(() => {
            /* ignore errors on logout */
          });
        }
        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
          error: null,
        });
      },

      refreshTokenAction: async () => {
        const { refreshToken } = get();
        if (!refreshToken) {
          get().logout();
          return;
        }
        try {
          const response = await authApi.refresh(refreshToken);
          const { access_token, refresh_token: newRefresh } = response.data;
          set({
            accessToken: access_token,
            refreshToken: newRefresh,
          });
          localStorage.setItem("access_token", access_token);
          localStorage.setItem("refresh_token", newRefresh);
        } catch {
          get().logout();
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
