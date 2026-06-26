import { useEffect, lazy, Suspense } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Toaster } from "react-hot-toast";
import { Layout } from "./components/Layout";
import { LoadingShell } from "./components/ui/LoadingSpinner";
import { useAuthStore } from "./store/authStore";
import { useSubmissionStore } from "./store/submissionStore";
import { QueryClientProvider } from "./providers/QueryProvider";
import { ThemeProvider, useTheme } from "./hooks/useTheme";

// Lazy-loaded page components for code splitting
const HomePage = lazy(() => import("./pages/HomePage").then((m) => ({ default: m.HomePage })));
const LoginPage = lazy(() => import("./pages/LoginPage").then((m) => ({ default: m.LoginPage })));
const ProblemsPage = lazy(() => import("./pages/ProblemsPage").then((m) => ({ default: m.ProblemsPage })));
const ProblemDetailPage = lazy(() => import("./pages/ProblemDetailPage").then((m) => ({ default: m.ProblemDetailPage })));
const LeaderboardPage = lazy(() => import("./pages/LeaderboardPage").then((m) => ({ default: m.LeaderboardPage })));
const ProfilePage = lazy(() => import("./pages/ProfilePage").then((m) => ({ default: m.ProfilePage })));
const AdminDashboard = lazy(() => import("./pages/admin/Dashboard").then((m) => ({ default: m.AdminDashboard })));
const AdminProblemForm = lazy(() => import("./pages/admin/ProblemForm").then((m) => ({ default: m.AdminProblemForm })));

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore();
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const { accessToken } = useAuthStore();
  const { connectWebSocket, disconnectWebSocket } = useSubmissionStore();

  useEffect(() => {
    if (accessToken) {
      connectWebSocket(accessToken);
    }
    return () => {
      disconnectWebSocket();
    };
  }, [accessToken, connectWebSocket, disconnectWebSocket]);

  return <>{children}</>;
}

function AppRoutes() {
  return (
    <Suspense fallback={<LoadingShell />}>
      <Routes>
        {/* Public routes */}
        <Route
          path="/"
          element={
            <Layout>
              <HomePage />
            </Layout>
          }
        />
        <Route
          path="/login"
          element={
            <Layout>
              <LoginPage />
            </Layout>
          }
        />
        <Route
          path="/register"
          element={
            <Layout>
              <LoginPage />
            </Layout>
          }
        />

        {/* Protected routes */}
        <Route
          path="/problems"
          element={
            <ProtectedRoute>
              <Layout>
                <ProblemsPage />
              </Layout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/problems/:slug"
          element={
            <ProtectedRoute>
              <Layout>
                <ProblemDetailPage />
              </Layout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/leaderboard"
          element={
            <ProtectedRoute>
              <Layout>
                <LeaderboardPage />
              </Layout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/profile"
          element={
            <ProtectedRoute>
              <Layout>
                <ProfilePage />
              </Layout>
            </ProtectedRoute>
          }
        />

        {/* Admin routes */}
        <Route
          path="/admin"
          element={
            <ProtectedRoute>
              <Layout>
                <AdminDashboard />
              </Layout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/problems/new"
          element={
            <ProtectedRoute>
              <Layout>
                <AdminProblemForm />
              </Layout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/problems/:slug/edit"
          element={
            <ProtectedRoute>
              <Layout>
                <AdminProblemForm />
              </Layout>
            </ProtectedRoute>
          }
        />

        {/* Catch all */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  );
}

function AppContent() {
  const { theme } = useTheme();

  return (
    <>
      <Toaster
        position="top-right"
        toastOptions={{
          duration: 3000,
          style: {
            background: theme === "dark" ? "#1e293b" : "#ffffff",
            color: theme === "dark" ? "#f1f5f9" : "#0f172a",
            borderRadius: "0.75rem",
            border: theme === "dark" ? "1px solid #334155" : "1px solid #e2e8f0",
          },
        }}
      />
      <AppRoutes />
    </>
  );
}

export default function App() {
  return (
    <QueryClientProvider>
      <BrowserRouter>
        <ThemeProvider>
          <WebSocketProvider>
            <AppContent />
          </WebSocketProvider>
        </ThemeProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
