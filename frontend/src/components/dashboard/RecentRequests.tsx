import React from 'react';
import { Card } from '@/components/common/Card';
import { Badge } from '@/components/common/Badge';
import { formatTokens, formatCost, formatDateTime } from '@/utils/format';
import type { RequestRecord } from '@/types/usage';

interface RecentRequestsProps {
  requests: RequestRecord[];
  isLoading?: boolean;
}

export const RecentRequests: React.FC<RecentRequestsProps> = React.memo(
  ({ requests, isLoading }) => {
    if (isLoading) {
      return (
        <Card title="Recent Activity" description="Latest requests">
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        </Card>
      );
    }

    if (!requests || requests.length === 0) {
      return (
        <Card title="Recent Activity" description="Latest requests">
          <div className="flex items-center justify-center py-12 text-gray-500">
            No recent activity
          </div>
        </Card>
      );
    }

    const getModelBadgeVariant = (model: string): 'default' | 'success' | 'info' | 'warning' => {
      if (model.includes('claude')) return 'info';
      if (model.includes('gpt')) return 'success';
      return 'default';
    };

    return (
      <Card title="Recent Activity" description="Latest requests">
        <div className="divide-y divide-gray-200">
          {requests.map((request) => (
            <div key={request.id} className="px-6 py-4 hover:bg-gray-50 transition-colors">
              <div className="flex items-center justify-between">
                <div className="flex items-center flex-1">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-3"></div>
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <p className="text-sm font-semibold text-gray-900">{request.api_key_name}</p>
                      <Badge variant={getModelBadgeVariant(request.model)}>{request.model}</Badge>
                    </div>
                    <div className="flex items-center gap-4 mt-1 text-xs text-gray-500">
                      <span>Input: {formatTokens(request.input_tokens)}</span>
                      <span>Output: {formatTokens(request.output_tokens)}</span>
                      <span>Cost: {formatCost(request.cost)}</span>
                      <span>{formatDateTime(request.created_at)}</span>
                    </div>
                  </div>
                </div>
                <Badge variant="success">Success</Badge>
              </div>
            </div>
          ))}
        </div>
      </Card>
    );
  }
);

RecentRequests.displayName = 'RecentRequests';
