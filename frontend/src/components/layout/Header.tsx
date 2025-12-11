import React from 'react';
import { LogOut, User } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/common/Button';

export const Header: React.FC = () => {
  const { admin, logout } = useAuth();

  return (
    <header className="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-6">
      <div className="flex items-center gap-4">
        <h2 className="text-lg font-semibold text-gray-900">Claude Relay Admin</h2>
      </div>

      <div className="flex items-center gap-4">
        {admin && (
          <div className="flex items-center gap-2 rounded-md bg-gray-100 px-3 py-2">
            <User className="h-4 w-4 text-gray-600" />
            <span className="text-sm font-medium text-gray-700">{admin.username}</span>
          </div>
        )}

        <Button
          variant="outline"
          size="sm"
          onClick={logout}
          className="flex items-center gap-2"
        >
          <LogOut className="h-4 w-4" />
          Logout
        </Button>
      </div>
    </header>
  );
};
