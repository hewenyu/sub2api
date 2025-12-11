import React, { useMemo } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Tooltip,
  Legend,
  type TooltipItem,
} from 'chart.js';
import { Bar } from 'react-chartjs-2';
import { Card } from '@/components/common/Card';
import type { UsageByModel, UsageByKey } from '@/types/usage';

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip, Legend);

interface CostBreakdownProps {
  byModel?: UsageByModel[];
  byKey?: UsageByKey[];
  isLoading?: boolean;
  type: 'model' | 'key';
}

export const CostBreakdown: React.FC<CostBreakdownProps> = ({
  byModel,
  byKey,
  isLoading,
  type,
}) => {
  const chartData = useMemo(() => {
    if (type === 'model' && byModel && byModel.length > 0) {
      return {
        labels: byModel.map((item) => item.model),
        datasets: [
          {
            label: 'Cost ($)',
            data: byModel.map((item) => item.cost),
            backgroundColor: '#1e40af',
            borderWidth: 0,
          },
        ],
      };
    }

    if (type === 'key' && byKey && byKey.length > 0) {
      return {
        labels: byKey.map((item) => item.api_key_name),
        datasets: [
          {
            label: 'Cost ($)',
            data: byKey.map((item) => item.cost),
            backgroundColor: '#1e40af',
            borderWidth: 0,
          },
        ],
      };
    }

    return {
      labels: ['No Data'],
      datasets: [
        {
          label: 'Cost ($)',
          data: [1],
          backgroundColor: '#e5e7eb',
          borderWidth: 0,
        },
      ],
    };
  }, [byModel, byKey, type]);

  const options = useMemo(
    () => ({
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false,
        },
        tooltip: {
          callbacks: {
            label: (context: TooltipItem<'bar'>) => {
              const value = context.parsed.y ?? 0;
              const formattedValue = new Intl.NumberFormat('en-US', {
                style: 'currency',
                currency: 'USD',
                minimumFractionDigits: 4,
                maximumFractionDigits: 4,
              }).format(value);
              return `Cost: ${formattedValue}`;
            },
          },
        },
      },
      scales: {
        y: {
          beginAtZero: true,
          ticks: {
            callback: (value: number | string) => `$${value}`,
          },
        },
      },
    }),
    []
  );

  const title = type === 'model' ? 'Cost Breakdown by Model' : 'Cost Breakdown by API Key';
  const description =
    type === 'model' ? 'Usage cost distribution by model' : 'Usage cost distribution by API key';

  if (isLoading) {
    return (
      <Card title={title} description={description}>
        <div className="flex items-center justify-center h-[300px]">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      </Card>
    );
  }

  const hasData =
    (type === 'model' && byModel && byModel.length > 0) ||
    (type === 'key' && byKey && byKey.length > 0);

  if (!hasData) {
    return (
      <Card title={title} description={description}>
        <div className="flex items-center justify-center h-[300px] text-gray-500">
          No data available for the selected period
        </div>
      </Card>
    );
  }

  return (
    <Card title={title} description={description}>
      <div className="h-[300px]">
        <Bar data={chartData} options={options} />
      </div>
    </Card>
  );
};
