import React from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Key,
  Users,
  // UserCircle,
  BarChart3,
  Network,
  Settings,
} from 'lucide-react';
import { useUIStore } from '@/store/uiStore';
import { useAuthStore } from '@/store/authStore';
import { ROUTES } from '@/utils/constants';

interface NavItem {
  path: string;
  label: string;
  icon: React.ElementType;
}

const navItems: NavItem[] = [
  { path: ROUTES.DASHBOARD, label: 'Dashboard', icon: LayoutDashboard },
  { path: ROUTES.API_KEYS, label: 'API Keys', icon: Key },
  // { path: ROUTES.CLAUDE_ACCOUNTS, label: 'Claude Accounts', icon: UserCircle },
  { path: ROUTES.CODEX_ACCOUNTS, label: 'Codex Accounts', icon: Users },
  { path: ROUTES.USAGE_STATS, label: 'Usage Statistics', icon: BarChart3 },
  { path: ROUTES.PROXIES, label: 'Proxy Management', icon: Network },
  { path: ROUTES.ADMIN, label: 'Settings', icon: Settings },
];

export const Sidebar: React.FC = () => {
  const { sidebarCollapsed } = useUIStore();
  const { admin } = useAuthStore();

  if (sidebarCollapsed) return null;

  return (
    <aside className="w-64 bg-white border-r border-gray-200 flex flex-col fixed h-screen">
      {/* Logo Section */}
      <div className="p-6">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-blue-900 rounded-lg flex items-center justify-center">
            <svg
              className="w-6 h-6 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
          </div>
          <div>
            <h1 className="font-bold text-gray-900">Relay Go</h1>
            <p className="text-xs text-gray-500">v1.0.0</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4">
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
          >
            <item.icon className="w-5 h-5 mr-3" />
            <span>{item.label}</span>
          </NavLink>
        ))}
      </nav>

      {/* User Profile Section */}
      <div className="border-t border-gray-200 p-4 bg-white">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center">
            <span className="text-gray-700 font-semibold text-sm">
              {admin?.username?.charAt(0).toUpperCase() || 'A'}
            </span>
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold text-gray-900 truncate">
              {admin?.username || 'Admin'}
            </p>
            <p className="text-xs text-gray-500">Administrator</p>
          </div>
        </div>
      </div>
    </aside>
  );
};
