import React, { useState } from 'react';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import type { DateRange } from '@/types/usage';

interface DateRangePickerProps {
  value: DateRange;
  onChange: (range: DateRange) => void;
}

type PresetOption = {
  label: string;
  getDates: () => DateRange;
};

export const DateRangePicker: React.FC<DateRangePickerProps> = ({ value, onChange }) => {
  const [customMode, setCustomMode] = useState(false);
  const [tempRange, setTempRange] = useState<DateRange>(value);

  const formatDate = (date: Date): string => {
    return date.toISOString().split('T')[0];
  };

  const presets: PresetOption[] = [
    {
      label: 'Today',
      getDates: () => {
        const today = new Date();
        return {
          start_date: formatDate(today),
          end_date: formatDate(today),
        };
      },
    },
    {
      label: 'Last 7 Days',
      getDates: () => {
        const end = new Date();
        const start = new Date();
        start.setDate(start.getDate() - 7);
        return {
          start_date: formatDate(start),
          end_date: formatDate(end),
        };
      },
    },
    {
      label: 'Last 30 Days',
      getDates: () => {
        const end = new Date();
        const start = new Date();
        start.setDate(start.getDate() - 30);
        return {
          start_date: formatDate(start),
          end_date: formatDate(end),
        };
      },
    },
    {
      label: 'This Month',
      getDates: () => {
        const now = new Date();
        const start = new Date(now.getFullYear(), now.getMonth(), 1);
        const end = new Date(now.getFullYear(), now.getMonth() + 1, 0);
        return {
          start_date: formatDate(start),
          end_date: formatDate(end),
        };
      },
    },
    {
      label: 'Last Month',
      getDates: () => {
        const now = new Date();
        const start = new Date(now.getFullYear(), now.getMonth() - 1, 1);
        const end = new Date(now.getFullYear(), now.getMonth(), 0);
        return {
          start_date: formatDate(start),
          end_date: formatDate(end),
        };
      },
    },
  ];

  const handlePresetClick = (preset: PresetOption) => {
    const range = preset.getDates();
    onChange(range);
    setCustomMode(false);
  };

  const handleCustomApply = () => {
    onChange(tempRange);
    setCustomMode(false);
  };

  const handleCustomCancel = () => {
    setTempRange(value);
    setCustomMode(false);
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2">
        {presets.map((preset) => (
          <Button
            key={preset.label}
            variant="outline"
            size="sm"
            onClick={() => handlePresetClick(preset)}
          >
            {preset.label}
          </Button>
        ))}
        <Button
          variant={customMode ? 'primary' : 'outline'}
          size="sm"
          onClick={() => setCustomMode(!customMode)}
        >
          Custom Range
        </Button>
      </div>

      {customMode && (
        <div className="border border-gray-300 rounded-lg p-4 bg-white space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Start Date
              </label>
              <Input
                type="date"
                value={tempRange.start_date}
                onChange={(e) =>
                  setTempRange({ ...tempRange, start_date: e.target.value })
                }
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                End Date
              </label>
              <Input
                type="date"
                value={tempRange.end_date}
                onChange={(e) =>
                  setTempRange({ ...tempRange, end_date: e.target.value })
                }
              />
            </div>
          </div>
          <div className="flex gap-2">
            <Button variant="primary" size="sm" onClick={handleCustomApply}>
              Apply
            </Button>
            <Button variant="outline" size="sm" onClick={handleCustomCancel}>
              Cancel
            </Button>
          </div>
        </div>
      )}

      <div className="text-sm text-gray-600">
        Selected: {value.start_date} to {value.end_date}
      </div>
    </div>
  );
};
