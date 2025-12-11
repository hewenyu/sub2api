import React, { useMemo } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
  type TooltipItem,
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import { Card } from '@/components/common/Card';
import type { DailyUsage } from '@/types/usage';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

interface UsageChartProps {
  data: DailyUsage[];
  isLoading?: boolean;
}

export const UsageChart: React.FC<UsageChartProps> = React.memo(({ data, isLoading }) => {
  const chartData = useMemo(() => {
    const sortedData = [...data].sort((a, b) =>
      new Date(a.date).getTime() - new Date(b.date).getTime()
    );

    const limitedData = sortedData.slice(-30);

    return {
      labels: limitedData.map(item => {
        const date = new Date(item.date);
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
      }),
      datasets: [
        {
          label: 'Requests',
          data: limitedData.map(item => item.requests),
          borderColor: '#1e40af',
          backgroundColor: 'rgba(30, 64, 175, 0.1)',
          fill: true,
          tension: 0.4,
        },
        {
          label: 'Cost ($)',
          data: limitedData.map(item => item.cost),
          borderColor: '#10b981',
          backgroundColor: 'rgba(16, 185, 129, 0.1)',
          fill: true,
          tension: 0.4,
          yAxisID: 'y1',
        },
      ],
    };
  }, [data]);

  const options = useMemo(() => ({
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      mode: 'index' as const,
      intersect: false,
    },
    plugins: {
      legend: {
        position: 'top' as const,
      },
      title: {
        display: false,
      },
      tooltip: {
        callbacks: {
          label: (context: TooltipItem<'line'>) => {
            const label = context.dataset.label || '';
            const value = context.parsed.y;
            if (label === 'Cost ($)') {
              return `${label}: $${value?.toFixed(4) || '0'}`;
            }
            return `${label}: ${value || 0}`;
          },
        },
      },
    },
    scales: {
      y: {
        type: 'linear' as const,
        display: true,
        position: 'left' as const,
        title: {
          display: true,
          text: 'Requests',
        },
      },
      y1: {
        type: 'linear' as const,
        display: true,
        position: 'right' as const,
        title: {
          display: true,
          text: 'Cost ($)',
        },
        grid: {
          drawOnChartArea: false,
        },
      },
    },
  }), []);

  if (isLoading) {
    return (
      <Card title="Request Trend" description="Last 7 days">
        <div className="flex items-center justify-center h-[300px]">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      </Card>
    );
  }

  if (!data || data.length === 0) {
    return (
      <Card title="Request Trend" description="Last 7 days">
        <div className="flex items-center justify-center h-[300px] text-gray-500">
          No usage data available
        </div>
      </Card>
    );
  }

  return (
    <Card title="Request Trend" description="Last 7 days">
      <div className="h-[300px]">
        <Line data={chartData} options={options} />
      </div>
    </Card>
  );
});

UsageChart.displayName = 'UsageChart';
