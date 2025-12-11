import React, { useState } from 'react';
import { Plus, Search, Copy } from 'lucide-react';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { Modal } from '@/components/common/Modal';
import { APIKeyList } from '@/components/features/apikey/APIKeyList';
import { CreateAPIKeyModal } from '@/components/features/apikey/CreateAPIKeyModal';
import { useAPIKeys } from '@/hooks/useAPIKeys';
import type { APIKey, CreateAPIKeyRequest } from '@/types/apikey';

export const APIKeys: React.FC = () => {
  const {
    apiKeys,
    loading,
    createAPIKey,
    updateAPIKey,
    deleteAPIKey,
    toggleAPIKey,
    refetch,
  } = useAPIKeys();

  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [keyDisplayModalOpen, setKeyDisplayModalOpen] = useState(false);
  const [selectedAPIKey, setSelectedAPIKey] = useState<APIKey | null>(null);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [editFormData, setEditFormData] = useState<Partial<CreateAPIKeyRequest>>({});
  const [searchQuery, setSearchQuery] = useState('');

  const handleCreateSuccess = (result: { api_key: string; api_key_object: APIKey }) => {
    setCreatedKey(result.api_key);
    setKeyDisplayModalOpen(true);
    refetch();
  };

  const handleEdit = (apiKey: APIKey) => {
    setSelectedAPIKey(apiKey);
    setEditFormData({
      name: apiKey.name,
      max_concurrent_requests: apiKey.max_concurrent_requests,
      rate_limit_per_minute: apiKey.rate_limit_per_minute,
      rate_limit_per_hour: apiKey.rate_limit_per_hour,
      rate_limit_per_day: apiKey.rate_limit_per_day,
      daily_cost_limit: apiKey.daily_cost_limit,
      weekly_cost_limit: apiKey.weekly_cost_limit,
      monthly_cost_limit: apiKey.monthly_cost_limit,
      total_cost_limit: apiKey.total_cost_limit,
      enable_model_restriction: apiKey.enable_model_restriction,
      restricted_models: apiKey.restricted_models,
      enable_client_restriction: apiKey.enable_client_restriction,
      allowed_clients: apiKey.allowed_clients,
    });
    setEditModalOpen(true);
  };

  const handleSaveEdit = async () => {
    if (selectedAPIKey) {
      try {
        await updateAPIKey(selectedAPIKey.id, editFormData);
        setEditModalOpen(false);
        setSelectedAPIKey(null);
      } catch (error) {
        // Error handled in hook
      }
    }
  };

  const handleDelete = (id: number) => {
    const apiKey = apiKeys.find((k) => k.id === id);
    setSelectedAPIKey(apiKey || null);
    setDeleteModalOpen(true);
  };

  const confirmDelete = async () => {
    if (selectedAPIKey) {
      try {
        await deleteAPIKey(selectedAPIKey.id);
        setDeleteModalOpen(false);
        setSelectedAPIKey(null);
      } catch (error) {
        // Error handled in hook
      }
    }
  };

  const handleCopyKey = async () => {
    if (createdKey) {
      try {
        await navigator.clipboard.writeText(createdKey);
        alert('API Key copied to clipboard!');
      } catch (error) {
        console.error('Failed to copy:', error);
      }
    }
  };

  const filteredAPIKeys = apiKeys.filter((key) =>
    key.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">API Key Management</h1>
          <p className="text-gray-600 mt-1">Manage and configure API Keys</p>
        </div>
        <Button onClick={() => setCreateModalOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create API Key
        </Button>
      </div>

      {/* Search and Filter */}
      <div className="flex items-center space-x-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <Input
              placeholder="Search API Keys..."
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
      </div>

      {/* Table */}
      <APIKeyList
        data={filteredAPIKeys}
        loading={loading}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onToggle={toggleAPIKey}
      />

      {/* Create Modal */}
      <CreateAPIKeyModal
        isOpen={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onSuccess={handleCreateSuccess}
        onCreate={createAPIKey}
      />

      {/* Show Created Key Modal */}
      <Modal
        isOpen={keyDisplayModalOpen}
        onClose={() => {
          setKeyDisplayModalOpen(false);
          setCreatedKey(null);
        }}
        title="API Key Created Successfully"
        size="md"
        footer={
          <div className="flex justify-end space-x-3">
            <Button
              onClick={() => {
                setKeyDisplayModalOpen(false);
                setCreatedKey(null);
              }}
            >
              Close
            </Button>
          </div>
        }
      >
        <div className="space-y-4">
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <p className="text-sm text-yellow-800 font-medium">
              Important: Copy and save this API Key now!
            </p>
            <p className="text-xs text-yellow-700 mt-1">
              For security reasons, you won't be able to see the full key again after
              closing this window.
            </p>
          </div>

          <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <label className="block text-xs font-medium text-gray-700 mb-2">
              Your API Key
            </label>
            <div className="bg-white p-3 rounded border border-gray-300 font-mono text-sm break-all">
              {createdKey}
            </div>
          </div>

          <Button onClick={handleCopyKey} className="w-full" variant="outline">
            <Copy className="w-4 h-4 mr-2" />
            Copy to Clipboard
          </Button>
        </div>
      </Modal>

      {/* Edit Modal */}
      <Modal
        isOpen={editModalOpen}
        onClose={() => setEditModalOpen(false)}
        title="Edit API Key"
        size="lg"
        footer={
          <div className="flex justify-end space-x-3">
            <Button variant="ghost" onClick={() => setEditModalOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSaveEdit}>Save Changes</Button>
          </div>
        }
      >
        <div className="space-y-6">
          {/* Name */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Name
            </label>
            <Input
              value={editFormData.name || ''}
              onChange={(e) =>
                setEditFormData({ ...editFormData, name: e.target.value })
              }
            />
          </div>

          {/* Max Concurrent Requests */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Max Concurrent Requests
            </label>
            <Input
              type="number"
              min="1"
              value={editFormData.max_concurrent_requests || 0}
              onChange={(e) =>
                setEditFormData({
                  ...editFormData,
                  max_concurrent_requests: parseInt(e.target.value) || 1,
                })
              }
            />
          </div>

          {/* Rate Limits */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Rate Limits
            </label>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="block text-xs text-gray-600 mb-1">Per Minute</label>
                <Input
                  type="number"
                  min="0"
                  value={editFormData.rate_limit_per_minute || 0}
                  onChange={(e) =>
                    setEditFormData({
                      ...editFormData,
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
                  value={editFormData.rate_limit_per_hour || 0}
                  onChange={(e) =>
                    setEditFormData({
                      ...editFormData,
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
                  value={editFormData.rate_limit_per_day || 0}
                  onChange={(e) =>
                    setEditFormData({
                      ...editFormData,
                      rate_limit_per_day: parseInt(e.target.value) || 0,
                    })
                  }
                />
              </div>
            </div>
          </div>

          {/* Cost Limits */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Cost Limits (USD)
            </label>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs text-gray-600 mb-1">Daily</label>
                <Input
                  type="number"
                  min="0"
                  step="0.01"
                  value={editFormData.daily_cost_limit || 0}
                  onChange={(e) =>
                    setEditFormData({
                      ...editFormData,
                      daily_cost_limit: parseFloat(e.target.value) || 0,
                    })
                  }
                />
              </div>
              <div>
                <label className="block text-xs text-gray-600 mb-1">Weekly</label>
                <Input
                  type="number"
                  min="0"
                  step="0.01"
                  value={editFormData.weekly_cost_limit || 0}
                  onChange={(e) =>
                    setEditFormData({
                      ...editFormData,
                      weekly_cost_limit: parseFloat(e.target.value) || 0,
                    })
                  }
                />
              </div>
              <div>
                <label className="block text-xs text-gray-600 mb-1">Monthly</label>
                <Input
                  type="number"
                  min="0"
                  step="0.01"
                  value={editFormData.monthly_cost_limit || 0}
                  onChange={(e) =>
                    setEditFormData({
                      ...editFormData,
                      monthly_cost_limit: parseFloat(e.target.value) || 0,
                    })
                  }
                />
              </div>
              <div>
                <label className="block text-xs text-gray-600 mb-1">Total</label>
                <Input
                  type="number"
                  min="0"
                  step="0.01"
                  value={editFormData.total_cost_limit || 0}
                  onChange={(e) =>
                    setEditFormData({
                      ...editFormData,
                      total_cost_limit: parseFloat(e.target.value) || 0,
                    })
                  }
                />
              </div>
            </div>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Confirm Delete"
        size="sm"
        footer={
          <div className="flex justify-end space-x-3">
            <Button variant="ghost" onClick={() => setDeleteModalOpen(false)}>
              Cancel
            </Button>
            <Button variant="danger" onClick={confirmDelete}>
              Delete
            </Button>
          </div>
        }
      >
        <p className="text-gray-700">
          Are you sure you want to delete the API Key "
          <span className="font-semibold">{selectedAPIKey?.name}</span>"? This action
          cannot be undone.
        </p>
      </Modal>
    </div>
  );
};
