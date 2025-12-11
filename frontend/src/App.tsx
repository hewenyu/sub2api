import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from '@/components/layout/Layout';
import { Dashboard } from '@/pages/Dashboard';
import { APIKeys } from '@/pages/APIKeys';
// import { ClaudeAccounts } from '@/pages/ClaudeAccounts';
import { CodexAccounts } from '@/pages/CodexAccounts';
import { Usage } from '@/pages/Usage';
import { ProxyManagement } from '@/pages/ProxyManagement';
import { Login } from '@/pages/Login';
import { useAuthStore } from '@/store/authStore';
import { ROUTES } from '@/utils/constants';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore();

  if (!isAuthenticated) {
    return <Navigate to={ROUTES.LOGIN} replace />;
  }

  return <>{children}</>;
};

function App() {
  const { initAuth, isAuthenticated } = useAuthStore();

  useEffect(() => {
    initAuth();
  }, [initAuth]);

  return (
    <BrowserRouter>
      <Routes>
        <Route
          path={ROUTES.LOGIN}
          element={isAuthenticated ? <Navigate to={ROUTES.DASHBOARD} replace /> : <Login />}
        />

        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to={ROUTES.DASHBOARD} replace />} />
          <Route path={ROUTES.DASHBOARD} element={<Dashboard />} />
          <Route path={ROUTES.API_KEYS} element={<APIKeys />} />
          {/* <Route path={ROUTES.CLAUDE_ACCOUNTS} element={<ClaudeAccounts />} /> */}
          <Route path={ROUTES.CODEX_ACCOUNTS} element={<CodexAccounts />} />
          <Route path={ROUTES.USAGE_STATS} element={<Usage />} />
          <Route path={ROUTES.PROXIES} element={<ProxyManagement />} />
          <Route
            path={ROUTES.ADMIN}
            element={<div className="p-6">Admin Management - Coming Soon</div>}
          />
        </Route>

        <Route path="*" element={<Navigate to={ROUTES.DASHBOARD} replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
