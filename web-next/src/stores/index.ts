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
  setActiveView: (view: string) => void;
}

export const useNavStore = create<NavState>()((set) => ({
  activeView: "home",
  setActiveView: (view) => set({ activeView: view }),
}));
