import React, { useEffect, useState } from 'react';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { DateRangePicker } from './DateRangePicker';
import { apikeyApi } from '@/api/apikey';
import type { UsageFilters, DateRange } from '@/types/usage';
import type { APIKey } from '@/types/apikey';

interface FilterPanelProps {
  filters: UsageFilters;
  onFiltersChange: (filters: UsageFilters) => void;
}

export const FilterPanel: React.FC<FilterPanelProps> = ({ filters, onFiltersChange }) => {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadApiKeys();
  }, []);

  const loadApiKeys = async () => {
    try {
      setLoading(true);
      const response = await apikeyApi.list();
      setApiKeys(response.items || []);
    } catch (error) {
      console.error('Failed to load API keys:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDateRangeChange = (range: DateRange) => {
    onFiltersChange({
      ...filters,
      start_date: range.start_date,
      end_date: range.end_date,
    });
  };

  const handleApiKeyChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    onFiltersChange({
      ...filters,
      api_key_id: value ? parseInt(value) : undefined,
    });
  };

  const handleModelChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    onFiltersChange({
      ...filters,
      model: value || undefined,
    });
  };

  const handleAccountTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value as 'claude' | 'codex' | '';
    onFiltersChange({
      ...filters,
      account_type: value || undefined,
    });
  };

  const handleResetFilters = () => {
    const today = new Date();
    const lastWeek = new Date();
    lastWeek.setDate(lastWeek.getDate() - 7);

    onFiltersChange({
      start_date: lastWeek.toISOString().split('T')[0],
      end_date: today.toISOString().split('T')[0],
    });
  };

  const models = [
    'claude-3-5-sonnet-20241022',
    'claude-3-5-sonnet-20240620',
    'claude-3-opus-20240229',
    'claude-3-sonnet-20240229',
    'claude-3-haiku-20240307',
    'gpt-4-turbo',
    'gpt-4',
    'gpt-3.5-turbo',
  ];

  return (
    <Card title="Filters" className="mb-6">
      <div className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Time Range
          </label>
          <DateRangePicker
            value={{
              start_date: filters.start_date || '',
              end_date: filters.end_date || '',
            }}
            onChange={handleDateRangeChange}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            API Key
          </label>
          <select
            value={filters.api_key_id || ''}
            onChange={handleApiKeyChange}
            disabled={loading}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All API Keys</option>
            {apiKeys.map((key) => (
              <option key={key.id} value={key.id}>
                {key.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Account Type
          </label>
          <select
            value={filters.account_type || ''}
            onChange={handleAccountTypeChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Account Types</option>
            <option value="claude">Claude</option>
            <option value="codex">Codex</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Model
          </label>
          <select
            value={filters.model || ''}
            onChange={handleModelChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Models</option>
            {models.map((model) => (
              <option key={model} value={model}>
                {model}
              </option>
            ))}
          </select>
        </div>

        <Button
          variant="outline"
          className="w-full"
          onClick={handleResetFilters}
        >
          Reset Filters
        </Button>
      </div>
    </Card>
  );
};
