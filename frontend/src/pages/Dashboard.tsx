import React from 'react';
import { StatsCard } from '@/components/common/StatsCard';
import { Button } from '@/components/common/Button';
import { Alert } from '@/components/common/Alert';
import { UsageChart } from '@/components/dashboard/UsageChart';
import { ModelDistributionChart } from '@/components/dashboard/ModelDistributionChart';
import { AccountStatus } from '@/components/dashboard/AccountStatus';
import { RecentRequests } from '@/components/dashboard/RecentRequests';
import { useUsage } from '@/hooks/useUsage';
import { useAccounts } from '@/hooks/useAccounts';
import { formatNumber, formatTokens, formatCost } from '@/utils/format';

export const Dashboard: React.FC = () => {
  const { summary, recentRequests, isLoading: usageLoading, error: usageError, refetch: refetchUsage } = useUsage(undefined, true, 30000);
  const { status: accountStatus, isLoading: accountsLoading, error: accountsError, refetch: refetchAccounts } = useAccounts(true, 30000);

  const handleRefresh = async () => {
    await Promise.all([refetchUsage(), refetchAccounts()]);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-600">
            Welcome to Claude Relay Admin Dashboard
          </p>
        </div>
        <Button
          onClick={handleRefresh}
          variant="outline"
          disabled={usageLoading || accountsLoading}
        >
          {usageLoading || accountsLoading ? 'Refreshing...' : 'Refresh'}
        </Button>
      </div>

      {(usageError || accountsError) && (
        <Alert variant="error">
          {usageError?.message || accountsError?.message || 'An error occurred while loading data'}
        </Alert>
      )}

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard
          label="Total Requests"
          value={formatNumber(summary?.total_requests)}
          icon={
            <svg className="w-full h-full" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          }
          accentColor="blue"
        />

        <StatsCard
          label="Total Cost"
          value={formatCost(summary?.total_cost)}
          icon={
            <svg className="w-full h-full" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          }
          accentColor="purple"
        />

        <StatsCard
          label="Total Tokens"
          value={formatTokens((summary?.total_input_tokens || 0) + (summary?.total_output_tokens || 0))}
          icon={
            <svg className="w-full h-full" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          }
          accentColor="green"
        />

        <StatsCard
          label="API Keys"
          value={formatNumber(summary?.api_key_count)}
          icon={
            <svg className="w-full h-full" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
          }
          accentColor="orange"
        />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <UsageChart
          data={summary?.daily_usage || []}
          isLoading={usageLoading}
        />

        <ModelDistributionChart
          isLoading={usageLoading}
        />
      </div>

      <AccountStatus
        status={accountStatus}
        isLoading={accountsLoading}
      />

      <RecentRequests
        requests={recentRequests}
        isLoading={usageLoading}
      />
    </div>
  );
};
