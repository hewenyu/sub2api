import { useState, useEffect, useCallback } from 'react';
import { usageApi } from '@/api/usage';
import type { UsageSummary, UsageFilters, RequestRecord } from '@/types/usage';

interface UseUsageResult {
  summary: UsageSummary | null;
  recentRequests: RequestRecord[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export const useUsage = (
  filters?: UsageFilters,
  autoRefresh: boolean = false,
  refreshInterval: number = 30000
): UseUsageResult => {
  const [summary, setSummary] = useState<UsageSummary | null>(null);
  const [recentRequests, setRecentRequests] = useState<RequestRecord[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      const [summaryData, recentData] = await Promise.all([
        usageApi.getSummary(filters),
        usageApi.getRecent(10),
      ]);

      setSummary(summaryData);
      setRecentRequests(recentData);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch usage data'));
    } finally {
      setIsLoading(false);
    }
  }, [filters]);

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
    summary,
    recentRequests,
    isLoading,
    error,
    refetch: fetchData,
  };
};
