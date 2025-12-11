import React, { useState, useEffect } from 'react';
import { Modal } from '@/components/common/Modal';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { proxyApi } from '@/api/proxy';
import type { CodexAccount, UpdateCodexAccountRequest } from '@/types/account';
import type { ProxyNameInfo } from '@/types/proxy';

interface EditCodexAccountModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (id: number, data: UpdateCodexAccountRequest) => Promise<void>;
  account: CodexAccount | null;
}

export const EditCodexAccountModal: React.FC<EditCodexAccountModalProps> = ({
  isOpen,
  onClose,
  onSave,
  account,
}) => {
  const [formData, setFormData] = useState<UpdateCodexAccountRequest>({
    priority: 1,
    schedulable: true,
    proxy_url: '',
  });
  const [loading, setLoading] = useState(false);
  const [proxies, setProxies] = useState<ProxyNameInfo[]>([]);

  useEffect(() => {
    if (isOpen) {
      proxyApi
        .getProxyNames()
        .then((res) => setProxies(res.data.data || []))
        .catch(() => setProxies([]));
    }
  }, [isOpen]);

  useEffect(() => {
    if (account) {
      setFormData({
        priority: account.priority,
        schedulable: account.schedulable,
        proxy_url: account.proxy_url || '',
        proxy_name: account.proxy_name,
      });
    }
  }, [account]);

  const handleSave = async () => {
    if (!account) return;

    try {
      setLoading(true);
      await onSave(account.id, formData);
      onClose();
    } catch (error) {
      console.error('Failed to update account:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Edit Codex Account"
      size="md"
      footer={
        <div className="flex justify-end space-x-3">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSave} loading={loading}>
            Save Changes
          </Button>
        </div>
      }
    >
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Email (Read-only)</label>
          <Input value={account?.email || ''} readOnly className="bg-gray-50" />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Account Type (Read-only)
          </label>
          <Input
            value={account?.account_type === 'openai-oauth' ? 'OAuth' : 'OpenAI-Responses'}
            readOnly
            className="bg-gray-50"
          />
        </div>

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
            id="edit-codex-schedulable"
          />
          <label htmlFor="edit-codex-schedulable" className="text-sm font-medium">
            Schedulable
          </label>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Proxy URL (Optional)
          </label>
          <Input
            value={formData.proxy_url || ''}
            onChange={(e) => setFormData({ ...formData, proxy_url: e.target.value })}
            placeholder="http://proxy:8080"
          />
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
