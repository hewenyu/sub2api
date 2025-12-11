import React, { useState, useEffect } from 'react';
import { Modal } from '@/components/common/Modal';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import type { ClaudeAccount, UpdateClaudeAccountRequest } from '@/types/account';

interface EditClaudeAccountModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (id: number, data: UpdateClaudeAccountRequest) => Promise<void>;
  account: ClaudeAccount | null;
}

export const EditClaudeAccountModal: React.FC<EditClaudeAccountModalProps> = ({
  isOpen,
  onClose,
  onSave,
  account,
}) => {
  const [formData, setFormData] = useState<UpdateClaudeAccountRequest>({
    priority: 1,
    schedulable: true,
    proxy_url: '',
  });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (account) {
      setFormData({
        priority: account.priority,
        schedulable: account.schedulable,
        proxy_url: account.proxy_url || '',
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
      title="Edit Claude Account"
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
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Email (Read-only)
          </label>
          <Input value={account?.email || ''} readOnly className="bg-gray-50" />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Priority
          </label>
          <Input
            type="number"
            min="1"
            value={formData.priority}
            onChange={(e) =>
              setFormData({ ...formData, priority: parseInt(e.target.value) || 1 })
            }
          />
          <p className="text-xs text-gray-500 mt-1">Higher value means higher priority</p>
        </div>

        <div className="flex items-center">
          <input
            type="checkbox"
            checked={formData.schedulable}
            onChange={(e) =>
              setFormData({ ...formData, schedulable: e.target.checked })
            }
            className="mr-2"
            id="edit-schedulable"
          />
          <label htmlFor="edit-schedulable" className="text-sm font-medium">
            Schedulable
          </label>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Proxy URL (Optional)
          </label>
          <Input
            value={formData.proxy_url || ''}
            onChange={(e) =>
              setFormData({ ...formData, proxy_url: e.target.value })
            }
            placeholder="http://proxy:8080"
          />
        </div>
      </div>
    </Modal>
  );
};
