"use client";

import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { getMe } from "./api";

interface AuthState {
  email: string | null;
  userId: string | null;
  loading: boolean;
}

const AuthContext = createContext<AuthState>({
  email: null,
  userId: null,
  loading: true,
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState>({
    email: null,
    userId: null,
    loading: true,
  });

  useEffect(() => {
    getMe()
      .then((data) => setAuth({ email: data.email, userId: data.user_id, loading: false }))
      .catch(() => setAuth({ email: null, userId: null, loading: false }));
  }, []);

  return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}
