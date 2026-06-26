import { useState, useEffect } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import {
  Code2,
  Menu,
  X,
  Home,
  List,
  Trophy,
  User as UserIcon,
  LogOut,
  ChevronDown,
  Sun,
  Moon,
  Activity,
  PlusCircle,
} from "lucide-react";
import { useAuthStore } from "../store/authStore";
import { useTheme } from "../hooks/useTheme";

const navLinks = [
  { to: "/", label: "Home", icon: Home },
  { to: "/problems", label: "Problems", icon: List },
  { to: "/leaderboard", label: "Leaderboard", icon: Trophy },
];

const mobileNavLinks = [
  { to: "/", label: "Home", icon: Home },
  { to: "/problems", label: "Problems", icon: List },
  { to: "/leaderboard", label: "Rankings", icon: Trophy },
  { to: "/profile", label: "Profile", icon: UserIcon },
];

export function Layout({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const { user, isAuthenticated, logout } = useAuthStore();
  const { theme, toggle: toggleTheme } = useTheme();
  const [isMobile, setIsMobile] = useState(false);

  // Detect mobile
  useEffect(() => {
    const check = () => setIsMobile(window.innerWidth < 768);
    check();
    window.addEventListener("resize", check);
    return () => window.removeEventListener("resize", check);
  }, []);

  // Close sidebar on route change
  useEffect(() => {
    setSidebarOpen(false);
    setUserMenuOpen(false);
  }, [location.pathname]);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="min-h-screen flex flex-col bg-slate-50 dark:bg-neutral-950 transition-colors duration-300">
      {/* Header */}
      <header
        className="bg-white dark:bg-neutral-900 border-b border-slate-200 dark:border-neutral-800 sticky top-0 z-40 transition-colors duration-300"
        role="banner"
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-14 sm:h-16">
            {/* Logo */}
            <Link to="/" className="flex items-center gap-2 text-indigo-600 dark:text-indigo-400 font-bold text-lg sm:text-xl" aria-label="Home">
              <Code2 className="w-6 h-6 sm:w-7 sm:h-7" />
              <span className="hidden sm:inline">CodeChallenge</span>
            </Link>

            {/* Desktop Nav */}
            <nav className="hidden md:flex items-center gap-1" aria-label="Main navigation">
              {navLinks.map((link) => {
                const Icon = link.icon;
                const active = isActive(link.to);
                return (
                  <Link
                    key={link.to}
                    to={link.to}
                    className={`flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                      active
                        ? "bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400"
                        : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-neutral-800 hover:text-slate-900 dark:hover:text-white"
                    }`}
                    aria-current={active ? "page" : undefined}
                  >
                    <Icon className="w-4 h-4" />
                    {link.label}
                  </Link>
                );
              })}
              {/* Admin link */}
              {isAuthenticated && user?.role === "admin" && (
                <Link
                  to="/admin"
                  className={`flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                    isActive("/admin") || location.pathname.startsWith("/admin")
                      ? "bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400"
                      : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-neutral-800"
                  }`}
                >
                  <Activity className="w-4 h-4" />
                  Admin
                </Link>
              )}
            </nav>

            {/* Right section */}
            <div className="flex items-center gap-2 sm:gap-3">
              {/* Theme Toggle */}
              <button
                onClick={toggleTheme}
                className="p-2 rounded-lg text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-neutral-800 transition-colors"
                aria-label={`Switch to ${theme === "light" ? "dark" : "light"} mode`}
                style={{ minWidth: "36px", minHeight: "36px" }}
              >
                {theme === "light" ? (
                  <Moon className="w-4 h-4 sm:w-5 sm:h-5" />
                ) : (
                  <Sun className="w-4 h-4 sm:w-5 sm:h-5" />
                )}
              </button>

              {/* User Menu */}
              {isAuthenticated && user ? (
                <div className="relative">
                  <button
                    onClick={() => setUserMenuOpen(!userMenuOpen)}
                    className="flex items-center gap-2 px-2 sm:px-3 py-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-neutral-800 transition-colors"
                    aria-expanded={userMenuOpen}
                    aria-haspopup="true"
                    aria-label="User menu"
                    style={{ minHeight: "36px" }}
                  >
                    <div className="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-indigo-100 dark:bg-indigo-900/50 flex items-center justify-center">
                      <span className="text-xs sm:text-sm font-semibold text-indigo-700 dark:text-indigo-300">
                        {user.username.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <span className="hidden sm:inline text-sm font-medium text-slate-700 dark:text-slate-300">
                      {user.username}
                    </span>
                    <ChevronDown className="w-4 h-4 text-slate-500" />
                  </button>

                  {userMenuOpen && (
                    <>
                      <div className="fixed inset-0 z-40" onClick={() => setUserMenuOpen(false)} />
                      <div
                        className="absolute right-0 mt-2 w-48 bg-white dark:bg-neutral-900 rounded-xl shadow-xl border border-slate-200 dark:border-neutral-700 py-1 z-50"
                        role="menu"
                      >
                        <Link
                          to="/profile"
                          className="flex items-center gap-2 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-neutral-800"
                          role="menuitem"
                        >
                          <UserIcon className="w-4 h-4" />
                          Profile
                        </Link>
                        {user.role === "admin" && (
                          <Link
                            to="/admin"
                            className="flex items-center gap-2 px-4 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-neutral-800"
                            role="menuitem"
                          >
                            <Activity className="w-4 h-4" />
                            Admin Dashboard
                          </Link>
                        )}
                        <button
                          onClick={handleLogout}
                          className="flex items-center gap-2 w-full px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/30"
                          role="menuitem"
                        >
                          <LogOut className="w-4 h-4" />
                          Logout
                        </button>
                      </div>
                    </>
                  )}
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <Link
                    to="/login"
                    className="px-3 sm:px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white transition-colors"
                    style={{ minHeight: "36px", display: "inline-flex", alignItems: "center" }}
                  >
                    Log in
                  </Link>
                  <Link
                    to="/register"
                    className="px-3 sm:px-4 py-2 text-sm font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
                    style={{ minHeight: "36px", display: "inline-flex", alignItems: "center" }}
                  >
                    Sign up
                  </Link>
                </div>
              )}

              {/* Mobile hamburger */}
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="md:hidden p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-neutral-800"
                aria-expanded={sidebarOpen}
                aria-label="Toggle navigation menu"
                style={{ minWidth: "36px", minHeight: "36px" }}
              >
                {sidebarOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div className="md:hidden fixed inset-0 z-30">
          <div className="fixed inset-0 bg-black/40 dark:bg-black/60" onClick={() => setSidebarOpen(false)} />
          <nav
            className="fixed left-0 top-14 bottom-16 w-72 bg-white dark:bg-neutral-900 border-r border-slate-200 dark:border-neutral-700 p-4 overflow-y-auto shadow-xl"
            aria-label="Mobile navigation"
          >
            <ul className="space-y-1">
              {navLinks.map((link) => {
                const Icon = link.icon;
                const active = isActive(link.to);
                return (
                  <li key={link.to}>
                    <Link
                      to={link.to}
                      onClick={() => setSidebarOpen(false)}
                      className={`flex items-center gap-3 px-3 py-3 rounded-lg text-sm font-medium transition-colors ${
                        active
                          ? "bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400"
                          : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-neutral-800"
                      }`}
                      aria-current={active ? "page" : undefined}
                    >
                      <Icon className="w-5 h-5" />
                      {link.label}
                    </Link>
                  </li>
                );
              })}
              {/* Admin link in mobile */}
              {isAuthenticated && user?.role === "admin" && (
                <li>
                  <Link
                    to="/admin"
                    onClick={() => setSidebarOpen(false)}
                    className={`flex items-center gap-3 px-3 py-3 rounded-lg text-sm font-medium transition-colors ${
                      isActive("/admin") || location.pathname.startsWith("/admin")
                        ? "bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400"
                        : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-neutral-800"
                    }`}
                  >
                    <Activity className="w-5 h-5" />
                    Admin
                  </Link>
                </li>
              )}
            </ul>
          </nav>
        </div>
      )}

      {/* Main content — add bottom padding for mobile nav bar */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-4 sm:py-6 pb-20 md:pb-6" role="main">
        {children}
      </main>

      {/* Mobile Bottom Navigation Bar */}
      <nav
        className="md:hidden fixed bottom-0 left-0 right-0 z-40 bg-white dark:bg-neutral-900 border-t border-slate-200 dark:border-neutral-800 safe-area-bottom transition-colors duration-300"
        aria-label="Mobile bottom navigation"
      >
        <div className="flex items-center justify-around h-14">
          {mobileNavLinks.map((link) => {
            const Icon = link.icon;
            const active = isActive(link.to) || 
              (link.to !== "/" && link.to !== "/profile" && location.pathname.startsWith(link.to));
            return (
              <Link
                key={link.to}
                to={link.to}
                className={`flex flex-col items-center justify-center gap-0.5 px-3 py-1 rounded-lg transition-colors ${
                  active
                    ? "text-indigo-600 dark:text-indigo-400"
                    : "text-slate-500 dark:text-slate-500 hover:text-slate-700 dark:hover:text-slate-300"
                }`}
                aria-current={active ? "page" : undefined}
                style={{ minWidth: "44px", minHeight: "44px" }}
              >
                <Icon className="w-5 h-5" />
                <span className="text-[10px] font-medium">{link.label}</span>
              </Link>
            );
          })}
        </div>
      </nav>

      {/* Footer — hidden on mobile since bottom nav replaces it */}
      <footer
        className="hidden md:block bg-white dark:bg-neutral-900 border-t border-slate-200 dark:border-neutral-800 py-6 transition-colors duration-300"
        role="contentinfo"
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2 text-slate-600 dark:text-slate-400">
              <Code2 className="w-5 h-5" />
              <span className="text-sm font-medium">CodeChallenge Platform</span>
            </div>
            <p className="text-sm text-slate-500 dark:text-slate-500">
              &copy; {new Date().getFullYear()} CodeChallenge. All rights reserved.
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}
