"use client";

import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { getMe } from "./api";

interface AuthState {
  email: string | null;
  userId: string | null;
  loading: boolean;
  waking: boolean;
}

const AuthContext = createContext<AuthState>({
  email: null,
  userId: null,
  loading: true,
  waking: false,
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState>({
    email: null,
    userId: null,
    loading: true,
    waking: false,
  });

  useEffect(() => {
    // If the request takes >3s, show "waking up" message (container cold start)
    const wakingTimer = setTimeout(() => {
      setAuth((prev) => (prev.loading ? { ...prev, waking: true } : prev));
    }, 3000);

    getMe()
      .then((data) => setAuth({ email: data.email, userId: data.user_id, loading: false, waking: false }))
      .catch(() => setAuth({ email: null, userId: null, loading: false, waking: false }))
      .finally(() => clearTimeout(wakingTimer));
  }, []);

  return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}
