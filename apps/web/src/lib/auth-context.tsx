"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiException, apiFetch } from "./api";

interface AuthState {
  isInitialized: boolean;
  accessToken: string | null;
  expiresAt: number;
}

interface AuthContextValue extends AuthState {
  login: (accessToken: string, expiresAt: string) => void;
  getAccessToken: () => Promise<string | null>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [state, setState] = useState<AuthState>({
    isInitialized: false,
    accessToken: null,
    expiresAt: 0,
  });

  const clearAuth = useCallback(() => {
    setState({ isInitialized: true, accessToken: null, expiresAt: 0 });
  }, []);

  const login = useCallback((accessToken: string, expiresAt: string) => {
    const expiry = new Date(expiresAt).getTime();
    setState({
      isInitialized: true,
      accessToken,
      expiresAt: Number.isFinite(expiry) ? expiry : 0,
    });
  }, []);

  const refresh = useCallback(async (): Promise<string | null> => {
    try {
      const result = await apiFetch<{ access_token: string; expires_at: string }>("auth/refresh", {
        method: "POST",
      });
      login(result.access_token, result.expires_at);
      return result.access_token;
    } catch (error) {
      if (!(error instanceof ApiException) || error.status === 401 || error.status === 403) {
        clearAuth();
      }
      return null;
    }
  }, [clearAuth, login]);

  const getAccessToken = useCallback(async (): Promise<string | null> => {
    if (!state.accessToken || Date.now() >= state.expiresAt - 15_000) {
      return refresh();
    }
    return state.accessToken;
  }, [refresh, state.accessToken, state.expiresAt]);

  const logout = useCallback(async () => {
    const accessToken = await getAccessToken();
    if (accessToken) {
      try {
        await apiFetch("auth/logout", {
          method: "POST",
          headers: { Authorization: `Bearer ${accessToken}` },
        });
      } catch {
        // Server-side session may already be expired; always clear browser state.
      }
    }

    clearAuth();
    router.replace("/login");
  }, [clearAuth, getAccessToken, router]);

  useEffect(() => {
    if (!state.isInitialized) {
      // Background initialization from the HttpOnly refresh-token cookie.
      // eslint-disable-next-line react-hooks/set-state-in-effect
      void refresh();
    }
  }, [refresh, state.isInitialized]);

  return (
    <AuthContext.Provider value={{ ...state, login, getAccessToken, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}