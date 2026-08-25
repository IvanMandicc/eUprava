import { Navigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import { Role } from '../types';

// RequireRole je ekvivalent Angular-ovim guard-ovima (authGuard/citizenGuard/
// officerGuard) — ruta je dostupna samo ulogovanom korisniku sa dozvoljenom ulogom.
export function RequireRole({ roles, children }: { roles?: Role[]; children: React.ReactNode }) {
  const auth = useAuth();
  if (!auth.isLoggedIn) return <Navigate to="/login" replace />;
  if (roles && auth.user && !roles.includes(auth.user.role)) return <Navigate to="/login" replace />;
  return <>{children}</>;
}
