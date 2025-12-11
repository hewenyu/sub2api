import React, { useMemo } from 'react';
import { Chart as ChartJS, ArcElement, Tooltip, Legend, type TooltipItem } from 'chart.js';
import { Doughnut } from 'react-chartjs-2';
import { Card } from '@/components/common/Card';

ChartJS.register(ArcElement, Tooltip, Legend);

interface ModelDistribution {
  model: string;
  count: number;
  percentage: number;
}

interface ModelDistributionChartProps {
  data?: ModelDistribution[];
  isLoading?: boolean;
}

export const ModelDistributionChart: React.FC<ModelDistributionChartProps> = React.memo(
  ({ data, isLoading }) => {
    const chartData = useMemo(() => {
      if (!data || data.length === 0) {
        return {
          labels: ['Claude Sonnet 4', 'Claude 3.5 Sonnet', 'GPT-4', 'Others'],
          datasets: [
            {
              data: [45, 30, 18, 7],
              backgroundColor: ['#1e40af', '#7c3aed', '#10b981', '#f59e0b'],
              borderWidth: 0,
            },
          ],
        };
      }

      return {
        labels: data.map((item) => item.model),
        datasets: [
          {
            data: data.map((item) => item.count),
            backgroundColor: ['#1e40af', '#7c3aed', '#10b981', '#f59e0b'],
            borderWidth: 0,
          },
        ],
      };
    }, [data]);

    const options = useMemo(
      () => ({
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'bottom' as const,
            labels: {
              padding: 15,
              font: {
                size: 12,
              },
            },
          },
          tooltip: {
            callbacks: {
              label: (context: TooltipItem<'doughnut'>) => {
                const label = context.label || '';
                const value = context.parsed;
                const total = context.dataset.data.reduce(
                  (acc: number, curr) => acc + (curr as number),
                  0
                );
                const percentage = ((value / total) * 100).toFixed(1);
                return `${label}: ${value} (${percentage}%)`;
              },
            },
          },
        },
      }),
      []
    );

    if (isLoading) {
      return (
        <Card title="Model Distribution" description="Usage by model">
          <div className="flex items-center justify-center h-[300px]">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        </Card>
      );
    }

    return (
      <Card title="Model Distribution" description="Usage by model">
        <div className="h-[300px]">
          <Doughnut data={chartData} options={options} />
        </div>
      </Card>
    );
  }
);

ModelDistributionChart.displayName = 'ModelDistributionChart';
