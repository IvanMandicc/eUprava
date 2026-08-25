import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

export function Layout() {
  const auth = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    auth.logout();
    navigate('/login');
  }

  return (
    <>
      <header className="topbar">
        <span className="brand">eUprava — MUP Vozila</span>
        <nav>
          <NavLink to="/verify" className={({ isActive }) => (isActive ? 'active' : undefined)}>
            Provera izveštaja
          </NavLink>
          {auth.isCitizen && (
            <>
              <NavLink to="/citizen/vehicles" className={({ isActive }) => (isActive ? 'active' : undefined)}>
                Moja vozila
              </NavLink>
              <NavLink to="/citizen/plates" className={({ isActive }) => (isActive ? 'active' : undefined)}>
                Personalizovane tablice
              </NavLink>
            </>
          )}
          {auth.isOfficer && (
            <>
              <NavLink to="/officer/vehicles" className={({ isActive }) => (isActive ? 'active' : undefined)}>
                Registar vozila
              </NavLink>
              <NavLink to="/officer/plate-requests" className={({ isActive }) => (isActive ? 'active' : undefined)}>
                Zahtevi za tablice
              </NavLink>
            </>
          )}
        </nav>
        {auth.isLoggedIn ? (
          <div className="user-box">
            <span>
              {auth.user?.firstName} {auth.user?.lastName}
            </span>
            <button className="secondary" onClick={handleLogout}>
              Odjava
            </button>
          </div>
        ) : (
          <div className="user-box">
            <NavLink to="/login">Prijava</NavLink>
            <NavLink to="/register">Registracija</NavLink>
          </div>
        )}
      </header>
      <main className="container">
        <Outlet />
      </main>
    </>
  );
}
