import { createContext, ReactNode, useContext, useState } from 'react';
import { authApi } from '../api/auth';
import { User } from '../types';

interface AuthContextValue {
  user: User | null;
  isLoggedIn: boolean;
  isCitizen: boolean;
  isOfficer: boolean;
  isAdmin: boolean;
  login: (email: string, password: string) => Promise<User>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function loadUser(): User | null {
  const raw = localStorage.getItem('user');
  return raw ? (JSON.parse(raw) as User) : null;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(loadUser());

  async function login(email: string, password: string): Promise<User> {
    const res = await authApi.login(email, password);
    localStorage.setItem('token', res.token);
    localStorage.setItem('user', JSON.stringify(res.user));
    setUser(res.user);
    return res.user;
  }

  function logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setUser(null);
  }

  const value: AuthContextValue = {
    user,
    isLoggedIn: user !== null,
    isCitizen: user?.role === 'citizen',
    isOfficer: user?.role === 'officer',
    isAdmin: user?.role === 'admin',
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth mora biti pozvan unutar AuthProvider-a');
  return ctx;
}
