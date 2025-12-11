import React, { useState } from 'react';
import { Modal } from '@/components/common/Modal';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import type { CreateAPIKeyRequest, CreateAPIKeyResponse } from '@/types/apikey';

interface CreateAPIKeyModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (result: CreateAPIKeyResponse) => void;
  onCreate: (data: CreateAPIKeyRequest) => Promise<CreateAPIKeyResponse>;
}

export const CreateAPIKeyModal: React.FC<CreateAPIKeyModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
  onCreate,
}) => {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<CreateAPIKeyRequest>({
    name: '',
    permissions: ['claude', 'codex'],
    max_concurrent_requests: 10,
    rate_limit_per_minute: 60,
    rate_limit_per_hour: 3600,
    rate_limit_per_day: 86400,
    daily_cost_limit: 100,
    weekly_cost_limit: 500,
    monthly_cost_limit: 2000,
    total_cost_limit: 0,
    enable_model_restriction: false,
    restricted_models: [],
    enable_client_restriction: false,
    allowed_clients: [],
  });

  const handleSubmit = async () => {
    if (!formData.name.trim()) {
      alert('Please enter a name for the API key');
      return;
    }

    try {
      setLoading(true);
      const result = await onCreate(formData);
      onSuccess(result);
      onClose();
      // Reset form
      setFormData({
        name: '',
        permissions: ['claude', 'codex'],
        max_concurrent_requests: 10,
        rate_limit_per_minute: 60,
        rate_limit_per_hour: 3600,
        rate_limit_per_day: 86400,
        daily_cost_limit: 100,
        weekly_cost_limit: 500,
        monthly_cost_limit: 2000,
        total_cost_limit: 0,
        enable_model_restriction: false,
        restricted_models: [],
        enable_client_restriction: false,
        allowed_clients: [],
      });
    } catch {
      // Error already handled in hook
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      onClose();
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Create API Key"
      size="lg"
      footer={
        <div className="flex justify-end space-x-3">
          <Button variant="ghost" onClick={handleClose} disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} loading={loading}>
            Create
          </Button>
        </div>
      }
    >
      <div className="space-y-6">
        {/* Name */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Name <span className="text-red-500">*</span>
          </label>
          <Input
            placeholder="Enter API Key name"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          />
        </div>

        {/* Concurrency Limits */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Max Concurrent Requests
          </label>
          <Input
            type="number"
            min="1"
            value={formData.max_concurrent_requests}
            onChange={(e) =>
              setFormData({
                ...formData,
                max_concurrent_requests: parseInt(e.target.value) || 1,
              })
            }
          />
          <p className="mt-1 text-xs text-gray-500">
            Maximum number of concurrent requests allowed
          </p>
        </div>

        {/* Rate Limits */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Rate Limits</label>
          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="block text-xs text-gray-600 mb-1">Per Minute</label>
              <Input
                type="number"
                min="0"
                placeholder="Per minute"
                value={formData.rate_limit_per_minute}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    rate_limit_per_minute: parseInt(e.target.value) || 0,
                  })
                }
              />
            </div>
            <div>
              <label className="block text-xs text-gray-600 mb-1">Per Hour</label>
              <Input
                type="number"
                min="0"
                placeholder="Per hour"
                value={formData.rate_limit_per_hour}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    rate_limit_per_hour: parseInt(e.target.value) || 0,
                  })
                }
              />
            </div>
            <div>
              <label className="block text-xs text-gray-600 mb-1">Per Day</label>
              <Input
                type="number"
                min="0"
                placeholder="Per day"
                value={formData.rate_limit_per_day}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    rate_limit_per_day: parseInt(e.target.value) || 0,
                  })
                }
              />
            </div>
          </div>
        </div>

        {/* Cost Limits */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Cost Limits (USD)</label>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-gray-600 mb-1">Daily Limit</label>
              <Input
                type="number"
                min="0"
                step="0.01"
                placeholder="Daily"
                value={formData.daily_cost_limit}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    daily_cost_limit: parseFloat(e.target.value) || 0,
                  })
                }
              />
            </div>
            <div>
              <label className="block text-xs text-gray-600 mb-1">Weekly Limit</label>
              <Input
                type="number"
                min="0"
                step="0.01"
                placeholder="Weekly"
                value={formData.weekly_cost_limit}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    weekly_cost_limit: parseFloat(e.target.value) || 0,
                  })
                }
              />
            </div>
            <div>
              <label className="block text-xs text-gray-600 mb-1">Monthly Limit</label>
              <Input
                type="number"
                min="0"
                step="0.01"
                placeholder="Monthly"
                value={formData.monthly_cost_limit}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    monthly_cost_limit: parseFloat(e.target.value) || 0,
                  })
                }
              />
            </div>
            <div>
              <label className="block text-xs text-gray-600 mb-1">Total Limit</label>
              <Input
                type="number"
                min="0"
                step="0.01"
                placeholder="Total (0 = unlimited)"
                value={formData.total_cost_limit}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    total_cost_limit: parseFloat(e.target.value) || 0,
                  })
                }
              />
            </div>
          </div>
          <p className="mt-1 text-xs text-gray-500">
            Set to 0 for unlimited. Requests will be blocked when limit is reached.
          </p>
        </div>

        {/* Model Restrictions */}
        <div>
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={formData.enable_model_restriction}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  enable_model_restriction: e.target.checked,
                })
              }
              className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            <span className="text-sm font-medium text-gray-700">Enable Model Restrictions</span>
          </label>
          {formData.enable_model_restriction && (
            <div className="mt-2">
              <Input
                placeholder="Comma-separated model names to restrict"
                value={formData.restricted_models?.join(', ') || ''}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    restricted_models: e.target.value
                      .split(',')
                      .map((m) => m.trim())
                      .filter((m) => m),
                  })
                }
              />
              <p className="mt-1 text-xs text-gray-500">
                Blacklist mode: These models will be blocked
              </p>
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
};
