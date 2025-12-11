import React from 'react';
import { Edit, Trash2, Power, PowerOff, RefreshCw } from 'lucide-react';
import { Table } from '@/components/common/Table';
import type { Column } from '@/components/common/Table';
import { Button } from '@/components/common/Button';
import { Badge } from '@/components/common/Badge';
import type { CodexAccount } from '@/types/account';

interface CodexAccountListProps {
  data: CodexAccount[];
  loading?: boolean;
  onEdit: (account: CodexAccount) => void;
  onDelete: (id: number) => void;
  onToggle: (id: number) => void;
  onRefreshToken: (id: number) => void;
}

export const CodexAccountList: React.FC<CodexAccountListProps> = ({
  data,
  loading,
  onEdit,
  onDelete,
  onToggle,
  onRefreshToken,
}) => {
  const formatDate = (dateString: string | null): string => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusBadge = (account: CodexAccount) => {
    if (!account.is_active) {
      return <Badge variant="inactive">Inactive</Badge>;
    }
    if (account.rate_limited_until && new Date(account.rate_limited_until) > new Date()) {
      return <Badge variant="rate-limited">Rate Limited</Badge>;
    }
    return <Badge variant="active">Active</Badge>;
  };

  const columns: Column<CodexAccount>[] = [
    {
      key: 'name',
      title: 'Account',
      width: '25%',
      render: (_value, record) => {
        const displayName = record.name || record.email || 'Unknown';
        const subLabel =
          record.email ||
          record.organization_title ||
          record.organization_role ||
          record.organization_id ||
          record.chatgpt_account_id ||
          record.chatgpt_user_id;
        return (
          <div className="flex items-center">
            <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center mr-3">
              <svg
                className="w-4 h-4 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <div className="flex flex-col">
              <span className="font-medium text-gray-900">{displayName}</span>
              {subLabel && subLabel !== displayName && (
                <span className="text-xs text-gray-500 truncate max-w-xs">{subLabel}</span>
              )}
            </div>
          </div>
        );
      },
    },
    {
      key: 'status',
      title: 'Status',
      width: '15%',
      render: (_value, record) => getStatusBadge(record),
    },
    {
      key: 'account_type',
      title: 'Account Type',
      width: '12%',
      render: (value) => {
        const type = String(value);
        return (
          <Badge variant="codex-type">
            {type === 'openai-oauth' ? 'OAuth' : 'OpenAI-Responses'}
          </Badge>
        );
      },
    },
    {
      key: 'proxy_name',
      title: 'Proxy',
      width: '12%',
      render: (value) => (
        <span className="text-gray-700 text-sm">
          {value ? String(value) : <span className="text-gray-400">Default</span>}
        </span>
      ),
    },
    {
      key: 'priority',
      title: 'Priority',
      width: '8%',
      render: (value) => <span className="text-gray-900">{String(value)}</span>,
    },
    {
      key: 'last_used_at',
      title: 'Last Used',
      width: '18%',
      render: (value) => (
        <span className="text-gray-600 text-sm">{formatDate(value as string | null)}</span>
      ),
    },
    {
      key: 'id',
      title: 'Actions',
      width: '15%',
      render: (value, record) => (
        <div className="flex items-center space-x-1">
          {record.account_type === 'openai-oauth' && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onRefreshToken(Number(value))}
              className="p-2"
              title="Refresh Token"
            >
              <RefreshCw className="w-4 h-4 text-blue-600" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onToggle(Number(value))}
            className="p-2"
            title={record.is_active ? 'Deactivate' : 'Activate'}
          >
            {record.is_active ? (
              <PowerOff className="w-4 h-4 text-orange-600" />
            ) : (
              <Power className="w-4 h-4 text-green-600" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(record)}
            className="p-2"
            title="Edit"
          >
            <Edit className="w-4 h-4 text-blue-600" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(Number(value))}
            className="p-2"
            title="Delete"
          >
            <Trash2 className="w-4 h-4 text-red-600" />
          </Button>
        </div>
      ),
    },
  ];

  return <Table<CodexAccount> columns={columns} data={data} loading={loading} />;
};
