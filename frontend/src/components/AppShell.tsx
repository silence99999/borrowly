import { NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { api } from "../lib/api";
import { classNames } from "../lib/utils";

export function AppShell({ children }: { children: React.ReactNode }) {
  const { user, refresh } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    await api.logout();
    await refresh();
    navigate("/login");
  }

  return (
    <div className="app-shell">
      <div className="page-backdrop" />
      <header className="site-header">
        <NavLink to="/" className="brand">
          <span className="brand-mark">B</span>
          <span>
            <strong>Borrowly</strong>
            <small>Rent smarter, list faster.</small>
          </span>
        </NavLink>

        <nav className="site-nav">
          <NavItem to="/">Home</NavItem>
          <NavItem to="/items">Browse</NavItem>
          {user ? <NavItem to="/dashboard">Dashboard</NavItem> : null}
          {user ? <NavItem to="/chat">Chat</NavItem> : null}
          {!user ? <NavItem to="/login">Login</NavItem> : null}
        </nav>

        {user ? (
          <div className="header-actions">
            <button className="button secondary" onClick={() => void handleLogout()} type="button">
              Logout
            </button>
          </div>
        ) : null}
      </header>

      <main className="page-container">{children}</main>
    </div>
  );
}

function NavItem({ to, children }: { to: string; children: React.ReactNode }) {
  return (
    <NavLink to={to} className={({ isActive }) => classNames("nav-link", isActive && "active")}>
      {children}
    </NavLink>
  );
}
