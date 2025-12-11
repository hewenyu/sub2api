import { useState, useEffect, useCallback } from 'react';
// import { claudeApi } from '@/api/claude';
import { codexApi } from '@/api/codex';
import type { ClaudeAccount, CodexAccount, AccountStatus } from '@/types/account';

interface UseAccountsResult {
  status: AccountStatus | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export const useAccounts = (
  autoRefresh: boolean = false,
  refreshInterval: number = 30000
): UseAccountsResult => {
  const [status, setStatus] = useState<AccountStatus | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      // Claude accounts API call disabled - feature not yet developed
      // const [claudeResponse, codexResponse] = await Promise.all([
      //   claudeApi.list({ page: 1, page_size: 100 }),
      //   codexApi.list({ page: 1, page_size: 100 }),
      // ]);
      const codexResponse = await codexApi.list({ page: 1, page_size: 100 });

      // const claudeAccounts: ClaudeAccount[] = claudeResponse.items || [];
      const claudeAccounts: ClaudeAccount[] = [];
      const codexAccounts: CodexAccount[] = codexResponse.items || [];

      const claudeHealthy = claudeAccounts.filter(
        (acc) => acc.is_active && acc.schedulable && !acc.rate_limited_until
      ).length;

      const codexHealthy = codexAccounts.filter(
        (acc) => acc.is_active && acc.schedulable && !acc.rate_limited_until
      ).length;

      setStatus({
        claude_accounts: claudeAccounts,
        codex_accounts: codexAccounts,
        claude_healthy: claudeHealthy,
        codex_healthy: codexHealthy,
        claude_total: claudeAccounts.length,
        codex_total: codexAccounts.length,
      });
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch account data'));
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (autoRefresh && refreshInterval > 0) {
      const timer = setInterval(fetchData, refreshInterval);
      return () => clearInterval(timer);
    }
  }, [autoRefresh, refreshInterval, fetchData]);

  return {
    status,
    isLoading,
    error,
    refetch: fetchData,
  };
};
