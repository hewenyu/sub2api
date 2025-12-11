import React, { useState, useEffect, useCallback } from 'react';
import { FilterPanel } from '@/components/features/usage/FilterPanel';
import { UsageSummary } from '@/components/features/usage/UsageSummary';
import { CostBreakdown } from '@/components/features/usage/CostBreakdown';
import { UsageTable } from '@/components/features/usage/UsageTable';
import { ExportButton } from '@/components/features/usage/ExportButton';
import { UsageChart } from '@/components/dashboard/UsageChart';
import { Alert } from '@/components/common/Alert';
import { usageApi } from '@/api/usage';
import type {
  UsageFilters,
  UsageStats,
  UsageByModel,
  UsageByKey,
  UsageRecord,
  DailyUsage,
} from '@/types/usage';

export const Usage: React.FC = () => {
  const [filters, setFilters] = useState<UsageFilters>(() => {
    const today = new Date();
    const lastWeek = new Date();
    lastWeek.setDate(lastWeek.getDate() - 7);

    return {
      start_date: lastWeek.toISOString().split('T')[0],
      end_date: today.toISOString().split('T')[0],
    };
  });

  const [stats, setStats] = useState<UsageStats | null>(null);
  const [byModel, setByModel] = useState<UsageByModel[]>([]);
  const [byKey, setByKey] = useState<UsageByKey[]>([]);
  const [records, setRecords] = useState<UsageRecord[]>([]);
  const [dailyUsage, setDailyUsage] = useState<DailyUsage[]>([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(50);

  const [isLoadingStats, setIsLoadingStats] = useState(false);
  const [isLoadingRecords, setIsLoadingRecords] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = useCallback(async () => {
    try {
      setIsLoadingStats(true);
      setError(null);

      const [statsData, byModelData, byKeyData, summaryData] = await Promise.all([
        usageApi.getStats(filters),
        usageApi.getByModel(filters),
        usageApi.getByKey(filters),
        usageApi.getSummary(filters),
      ]);

      setStats(statsData);
      setByModel(byModelData);
      setByKey(byKeyData);
      setDailyUsage(summaryData.daily_usage || []);
    } catch (err) {
      console.error('Failed to fetch usage stats:', err);
      setError('Failed to load usage statistics. Please try again.');

      setStats({
        total_requests: 0,
        total_cost: 0,
        total_input_tokens: 0,
        total_output_tokens: 0,
        avg_cost_per_request: 0,
        most_used_model: 'N/A',
      });
      setByModel([]);
      setByKey([]);
      setDailyUsage([]);
    } finally {
      setIsLoadingStats(false);
    }
  }, [filters]);

  const fetchRecords = useCallback(async () => {
    try {
      setIsLoadingRecords(true);
      setError(null);

      const data = await usageApi.getRecords(filters, currentPage, pageSize);
      const transformedRecords: UsageRecord[] = (data.items || []).map((item) => ({
        id: item.id,
        timestamp: item.created_at || item.timestamp,
        api_key_name: item.api_key_name,
        account_email: item.account_email || 'N/A',
        model: item.model,
        input_tokens: item.input_tokens,
        output_tokens: item.output_tokens,
        cost: item.cost,
      }));
      setRecords(transformedRecords);
      setTotal(data.total || 0);
    } catch (err) {
      console.error('Failed to fetch usage records:', err);
      setError('Failed to load usage records. Please try again.');
      setRecords([]);
      setTotal(0);
    } finally {
      setIsLoadingRecords(false);
    }
  }, [filters, currentPage, pageSize]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  useEffect(() => {
    fetchRecords();
  }, [fetchRecords]);

  const handleFiltersChange = (newFilters: UsageFilters) => {
    setFilters(newFilters);
    setCurrentPage(1);
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Usage Statistics</h1>
          <p className="mt-1 text-sm text-gray-600">Comprehensive usage analytics and reports</p>
        </div>
        <ExportButton filters={filters} records={records} disabled={isLoadingRecords} />
      </div>

      {error && (
        <Alert variant="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-1">
          <FilterPanel filters={filters} onFiltersChange={handleFiltersChange} />
        </div>

        <div className="lg:col-span-3 space-y-6">
          <UsageSummary stats={stats} isLoading={isLoadingStats} />

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <CostBreakdown byModel={byModel} type="model" isLoading={isLoadingStats} />
            <CostBreakdown byKey={byKey} type="key" isLoading={isLoadingStats} />
          </div>

          <UsageChart data={dailyUsage} isLoading={isLoadingStats} />

          <UsageTable
            records={records}
            total={total}
            currentPage={currentPage}
            pageSize={pageSize}
            isLoading={isLoadingRecords}
            onPageChange={handlePageChange}
          />
        </div>
      </div>
    </div>
  );
};
