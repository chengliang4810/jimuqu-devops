import { create } from "zustand";
import { persist } from "zustand/middleware";

// 认证状态
interface AuthState {
  token: string | null;
  isAuthenticated: boolean;
  setToken: (token: string) => void;
  clearToken: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      isAuthenticated: false,
      setToken: (token) =>
        set({ token, isAuthenticated: true }),
      clearToken: () =>
        set({ token: null, isAuthenticated: false }),
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({ token: state.token, isAuthenticated: state.isAuthenticated }),
    }
  )
);

// 导航状态
interface NavState {
  activeView: string;
  pendingRunId: number | null;
  setActiveView: (view: string) => void;
  setPendingRunId: (runId: number) => void;
  clearPendingRunId: () => void;
}

export const useNavStore = create<NavState>()(
  persist(
    (set) => ({
      activeView: "home",
      pendingRunId: null,
      setActiveView: (view) => set({ activeView: view }),
      setPendingRunId: (runId) => set({ pendingRunId: runId }),
      clearPendingRunId: () => set({ pendingRunId: null }),
    }),
    {
      name: "nav-storage",
      partialize: (state) => ({ activeView: state.activeView }),
    }
  )
);
