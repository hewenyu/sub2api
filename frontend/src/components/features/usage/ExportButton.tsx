import React, { useState } from 'react';
import { Button } from '@/components/common/Button';
import { usageApi } from '@/api/usage';
import type { UsageFilters, UsageRecord } from '@/types/usage';

interface ExportButtonProps {
  filters: UsageFilters;
  records: UsageRecord[];
  disabled?: boolean;
}

export const ExportButton: React.FC<ExportButtonProps> = ({ filters, records, disabled }) => {
  const [isExporting, setIsExporting] = useState(false);
  const [showMenu, setShowMenu] = useState(false);

  const exportToCSV = () => {
    if (records.length === 0) {
      alert('No data to export');
      return;
    }

    setIsExporting(true);
    try {
      const headers = [
        'Time',
        'API Key',
        'Account',
        'Model',
        'Input Tokens',
        'Output Tokens',
        'Cost',
      ];

      const csvRows = [
        headers.join(','),
        ...records.map((record) => {
          const row = [
            record.timestamp,
            `"${record.api_key_name}"`,
            `"${record.account_email}"`,
            `"${record.model}"`,
            record.input_tokens,
            record.output_tokens,
            record.cost,
          ];
          return row.join(',');
        }),
      ];

      const csvContent = csvRows.join('\n');
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      const url = URL.createObjectURL(blob);

      const filename = `usage_${filters.start_date || 'all'}_${filters.end_date || 'all'}.csv`;
      link.setAttribute('href', url);
      link.setAttribute('download', filename);
      link.style.visibility = 'hidden';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export CSV:', error);
      alert('Failed to export CSV. Please try again.');
    } finally {
      setIsExporting(false);
      setShowMenu(false);
    }
  };

  const exportToExcel = async () => {
    setIsExporting(true);
    try {
      const blob = await usageApi.exportExcel(filters);
      const link = document.createElement('a');
      const url = URL.createObjectURL(blob);

      const filename = `usage_${filters.start_date || 'all'}_${filters.end_date || 'all'}.xlsx`;
      link.setAttribute('href', url);
      link.setAttribute('download', filename);
      link.style.visibility = 'hidden';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export Excel:', error);
      alert('Excel export is not available at the moment. Please use CSV export instead.');
    } finally {
      setIsExporting(false);
      setShowMenu(false);
    }
  };

  return (
    <div className="relative inline-block">
      <Button
        variant="success"
        onClick={() => setShowMenu(!showMenu)}
        disabled={disabled || isExporting}
      >
        {isExporting ? (
          <>
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
            Exporting...
          </>
        ) : (
          <>
            <svg
              className="w-4 h-4 mr-2"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
              />
            </svg>
            Export CSV
          </>
        )}
      </Button>

      {showMenu && !isExporting && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setShowMenu(false)}
          ></div>
          <div className="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 z-20">
            <div className="py-1" role="menu">
              <button
                className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                role="menuitem"
                onClick={exportToCSV}
              >
                <div className="flex items-center">
                  <svg
                    className="w-4 h-4 mr-2"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                  Export as CSV
                </div>
              </button>
              <button
                className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                role="menuitem"
                onClick={exportToExcel}
              >
                <div className="flex items-center">
                  <svg
                    className="w-4 h-4 mr-2"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                  Export as Excel
                </div>
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
