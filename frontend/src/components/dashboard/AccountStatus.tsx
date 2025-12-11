import React from 'react';
import { Card } from '@/components/common/Card';
import { Badge } from '@/components/common/Badge';
import { formatPercentage } from '@/utils/format';
import type { AccountStatus as AccountStatusType } from '@/types/account';

interface AccountStatusProps {
  status: AccountStatusType | null;
  isLoading?: boolean;
}

export const AccountStatus: React.FC<AccountStatusProps> = React.memo(({ status, isLoading }) => {
  if (isLoading) {
    return (
      <Card title="Account Status" description="Account health overview">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      </Card>
    );
  }

  if (!status) {
    return (
      <Card title="Account Status" description="Account health overview">
        <div className="flex items-center justify-center py-12 text-gray-500">
          No account data available
        </div>
      </Card>
    );
  }

  // const claudeHealthPercentage = status.claude_total > 0
  //   ? (status.claude_healthy / status.claude_total) * 100
  //   : 0;

  const codexHealthPercentage = status.codex_total > 0
    ? (status.codex_healthy / status.codex_total) * 100
    : 0;

  const getHealthBadge = (percentage: number) => {
    if (percentage >= 80) return <Badge variant="success">Healthy</Badge>;
    if (percentage >= 50) return <Badge variant="warning">Warning</Badge>;
    return <Badge variant="error">Critical</Badge>;
  };

  return (
    <Card title="Account Status" description="Account health overview">
      <div className="space-y-6">
        {/* Claude Accounts section hidden - feature not yet developed
        <div className="border-b pb-4">
          <div className="flex items-center justify-between mb-2">
            <div>
              <h3 className="text-sm font-semibold text-gray-700">Claude Accounts</h3>
              <p className="text-xs text-gray-500 mt-1">
                {status.claude_healthy} healthy out of {status.claude_total} total
              </p>
            </div>
            {getHealthBadge(claudeHealthPercentage)}
          </div>
          <div className="mt-3">
            <div className="flex items-center justify-between text-xs text-gray-600 mb-1">
              <span>Health Status</span>
              <span>{formatPercentage(status.claude_healthy, status.claude_total)}</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className={`h-2 rounded-full transition-all ${
                  claudeHealthPercentage >= 80 ? 'bg-green-500' :
                  claudeHealthPercentage >= 50 ? 'bg-yellow-500' : 'bg-red-500'
                }`}
                style={{ width: `${claudeHealthPercentage}%` }}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4 mt-4">
            <div className="bg-gray-50 p-3 rounded-lg">
              <p className="text-xs text-gray-600">Pro Accounts</p>
              <p className="text-lg font-semibold text-gray-900 mt-1">
                {status.claude_accounts.filter(acc => acc.subscription_level === 'pro').length}
              </p>
            </div>
            <div className="bg-gray-50 p-3 rounded-lg">
              <p className="text-xs text-gray-600">Max Accounts</p>
              <p className="text-lg font-semibold text-gray-900 mt-1">
                {status.claude_accounts.filter(acc => acc.subscription_level === 'max').length}
              </p>
            </div>
          </div>
        </div>
        */}

        <div>
          <div className="flex items-center justify-between mb-2">
            <div>
              <h3 className="text-sm font-semibold text-gray-700">Codex Accounts</h3>
              <p className="text-xs text-gray-500 mt-1">
                {status.codex_healthy} healthy out of {status.codex_total} total
              </p>
            </div>
            {getHealthBadge(codexHealthPercentage)}
          </div>
          <div className="mt-3">
            <div className="flex items-center justify-between text-xs text-gray-600 mb-1">
              <span>Health Status</span>
              <span>{formatPercentage(status.codex_healthy, status.codex_total)}</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className={`h-2 rounded-full transition-all ${
                  codexHealthPercentage >= 80 ? 'bg-green-500' :
                  codexHealthPercentage >= 50 ? 'bg-yellow-500' : 'bg-red-500'
                }`}
                style={{ width: `${codexHealthPercentage}%` }}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4 mt-4">
            <div className="bg-gray-50 p-3 rounded-lg">
              <p className="text-xs text-gray-600">OAuth Accounts</p>
              <p className="text-lg font-semibold text-gray-900 mt-1">
                {status.codex_accounts.filter(acc => acc.account_type === 'openai-oauth').length}
              </p>
            </div>
            <div className="bg-gray-50 p-3 rounded-lg">
              <p className="text-xs text-gray-600">OpenAI Responses</p>
              <p className="text-lg font-semibold text-gray-900 mt-1">
                {status.codex_accounts.filter(acc => acc.account_type === 'openai-responses').length}
              </p>
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
});

AccountStatus.displayName = 'AccountStatus';
