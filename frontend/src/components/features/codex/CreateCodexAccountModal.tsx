import React, { useState, useEffect } from 'react';
import { Modal } from '@/components/common/Modal';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { proxyApi } from '@/api/proxy';
import type { CreateCodexAccountRequest } from '@/types/account';
import type { ProxyNameInfo } from '@/types/proxy';

interface CreateCodexAccountModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (data: CreateCodexAccountRequest) => Promise<void>;
}

export const CreateCodexAccountModal: React.FC<CreateCodexAccountModalProps> = ({
  isOpen,
  onClose,
  onCreate,
}) => {
  const [formData, setFormData] = useState<CreateCodexAccountRequest>({
    name: '',
    account_type: 'openai-responses',
    email: '',
    api_key: '',
    base_api: 'https://api.openai.com/v1',
    custom_user_agent: '',
    daily_quota: 0,
    quota_reset_time: '00:00',
    priority: 1,
    schedulable: true,
  });
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [proxies, setProxies] = useState<ProxyNameInfo[]>([]);

  useEffect(() => {
    if (isOpen) {
      proxyApi
        .getProxyNames()
        .then((res) => setProxies(res.data.data || []))
        .catch(() => setProxies([]));
    }
  }, [isOpen]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name) {
      newErrors.name = 'Name is required';
    }

    if (!formData.api_key) {
      newErrors.api_key = 'API Key is required';
    }

    if (!formData.base_api) {
      newErrors.base_api = 'Base API URL is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleCreate = async () => {
    if (!validateForm()) return;

    try {
      setLoading(true);
      const data: CreateCodexAccountRequest = {
        ...formData,
        account_type: 'openai-responses',
      };
      await onCreate(data);
      handleClose();
    } catch (error) {
      console.error('Failed to create account:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setFormData({
      name: '',
      account_type: 'openai-responses',
      email: '',
      api_key: '',
      base_api: 'https://api.openai.com/v1',
      custom_user_agent: '',
      daily_quota: 0,
      quota_reset_time: '00:00',
      priority: 1,
      schedulable: true,
    });
    setErrors({});
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Add Codex Account (API Key)"
      size="lg"
      footer={
        <div className="flex justify-end space-x-3">
          <Button variant="ghost" onClick={handleClose}>
            Cancel
          </Button>
          <Button onClick={handleCreate} loading={loading}>
            Create Account
          </Button>
        </div>
      }
    >
      <div className="space-y-4">
        {/* Informational Message */}
        <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
          <p className="text-sm text-blue-800">
            <strong>Note:</strong> Codex accounts use OpenAI API Keys for authentication. You can
            generate API Keys from your{' '}
            <a
              href="https://platform.openai.com/api-keys"
              target="_blank"
              rel="noopener noreferrer"
              className="underline hover:text-blue-900"
            >
              OpenAI Dashboard
            </a>
            .
          </p>
        </div>
        {/* Name */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Name <span className="text-red-500">*</span>
          </label>
          <Input
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="Enter account name"
          />
          {errors.name && <p className="text-xs text-red-600 mt-1">{errors.name}</p>}
        </div>

        {/* API Key */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            API Key <span className="text-red-500">*</span>
          </label>
          <Input
            value={formData.api_key || ''}
            onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
            placeholder="sk-..."
          />
          {errors.api_key && <p className="text-xs text-red-600 mt-1">{errors.api_key}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Base API URL <span className="text-red-500">*</span>
          </label>
          <Input
            value={formData.base_api || ''}
            onChange={(e) => setFormData({ ...formData, base_api: e.target.value })}
            placeholder="https://api.openai.com/v1"
          />
          {errors.base_api && <p className="text-xs text-red-600 mt-1">{errors.base_api}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Custom User-Agent (Optional)
          </label>
          <Input
            value={formData.custom_user_agent || ''}
            onChange={(e) => setFormData({ ...formData, custom_user_agent: e.target.value })}
            placeholder="Enter custom user-agent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Daily Quota (Optional)
          </label>
          <Input
            type="number"
            min="0"
            value={formData.daily_quota || 0}
            onChange={(e) =>
              setFormData({ ...formData, daily_quota: parseInt(e.target.value) || 0 })
            }
            placeholder="0 for unlimited"
          />
          <p className="text-xs text-gray-500 mt-1">Daily request quota (0 for unlimited)</p>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Quota Reset Time (Optional)
          </label>
          <Input
            type="time"
            value={formData.quota_reset_time || '00:00'}
            onChange={(e) => setFormData({ ...formData, quota_reset_time: e.target.value })}
          />
          <p className="text-xs text-gray-500 mt-1">Time when daily quota resets</p>
        </div>

        {/* Common Fields */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Priority</label>
          <Input
            type="number"
            min="1"
            value={formData.priority}
            onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 1 })}
          />
          <p className="text-xs text-gray-500 mt-1">Higher value means higher priority</p>
        </div>

        <div className="flex items-center">
          <input
            type="checkbox"
            checked={formData.schedulable}
            onChange={(e) => setFormData({ ...formData, schedulable: e.target.checked })}
            className="mr-2"
            id="create-schedulable"
          />
          <label htmlFor="create-schedulable" className="text-sm font-medium">
            Schedulable
          </label>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Proxy (Optional)</label>
          <select
            value={formData.proxy_name || ''}
            onChange={(e) => setFormData({ ...formData, proxy_name: e.target.value || undefined })}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Default Proxy</option>
            {proxies.map((proxy) => (
              <option key={proxy.name} value={proxy.name}>
                {proxy.name} ({proxy.protocol})
              </option>
            ))}
          </select>
          <p className="text-xs text-gray-500 mt-1">Select proxy configuration for this account</p>
        </div>
      </div>
    </Modal>
  );
};
