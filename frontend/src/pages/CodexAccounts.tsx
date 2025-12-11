import React, { useState, useEffect, useCallback } from 'react';
import { Plus, Search, RefreshCw } from 'lucide-react';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { Modal } from '@/components/common/Modal';
import { CodexAccountList } from '@/components/features/codex/CodexAccountList';
import { CreateCodexAccountModal } from '@/components/features/codex/CreateCodexAccountModal';
import { EditCodexAccountModal } from '@/components/features/codex/EditCodexAccountModal';
import { CodexOAuthFlow } from '@/components/features/codex/CodexOAuthFlow';
import { codexApi } from '@/api/codex';
import { message } from '@/utils/message';
import type {
  CodexAccount,
  CreateCodexAccountRequest,
  UpdateCodexAccountRequest,
} from '@/types/account';

export const CodexAccounts: React.FC = () => {
  const [accounts, setAccounts] = useState<CodexAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [oauthModalOpen, setOauthModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<CodexAccount | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchAccounts = useCallback(async () => {
    try {
      setLoading(true);
      const response = await codexApi.list({ page: 1, page_size: 100 });
      setAccounts(response.items);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to fetch Codex accounts';
      message.error(errorMessage);
      setAccounts([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAccounts();
  }, [fetchAccounts]);

  const handleCreate = async (data: CreateCodexAccountRequest) => {
    try {
      await codexApi.create(data);
      message.success('Codex account created successfully');
      setCreateModalOpen(false);
      await fetchAccounts();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to create Codex account';
      message.error(errorMessage);
      throw error;
    }
  };

  const handleEdit = (account: CodexAccount) => {
    setSelectedAccount(account);
    setEditModalOpen(true);
  };

  const handleSaveEdit = async (id: number, data: UpdateCodexAccountRequest) => {
    try {
      await codexApi.update(id, data);
      message.success('Codex account updated successfully');
      await fetchAccounts();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to update Codex account';
      message.error(errorMessage);
      throw error;
    }
  };

  const handleDelete = (id: number) => {
    const account = accounts.find((a) => a.id === id);
    setSelectedAccount(account || null);
    setDeleteModalOpen(true);
  };

  const confirmDelete = async () => {
    if (selectedAccount) {
      try {
        await codexApi.delete(selectedAccount.id);
        message.success('Codex account deleted successfully');
        setDeleteModalOpen(false);
        setSelectedAccount(null);
        await fetchAccounts();
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : 'Failed to delete Codex account';
        message.error(errorMessage);
      }
    }
  };

  const handleToggle = async (id: number) => {
    try {
      const result = await codexApi.toggle(id);
      message.success(result.is_active ? 'Account activated' : 'Account deactivated');
      await fetchAccounts();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to toggle account status';
      message.error(errorMessage);
    }
  };

  const handleRefreshToken = async (id: number) => {
    try {
      await codexApi.refreshToken(id);
      message.success('Token refreshed successfully');
      await fetchAccounts();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to refresh token';
      message.error(errorMessage);
    }
  };

  const handleOAuthSuccess = async () => {
    message.success('OAuth account connected successfully');
    setOauthModalOpen(false);
    await fetchAccounts();
  };

  const filteredAccounts = accounts.filter((account) => {
    const searchLower = searchQuery.toLowerCase();
    return (
      account.name?.toLowerCase().includes(searchLower) ||
      account.email?.toLowerCase().includes(searchLower)
    );
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Codex Account Management</h1>
          <p className="text-gray-600 mt-1">Manage Codex AI accounts and API credentials</p>
        </div>
        <div className="flex space-x-2">
          <Button variant="outline" onClick={fetchAccounts}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={() => setCreateModalOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add API Key Account
          </Button>
          <Button variant="outline" onClick={() => setOauthModalOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add OAuth Account
          </Button>
        </div>
      </div>

      {/* Search and Filter */}
      <div className="flex items-center space-x-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <Input
              placeholder="Search accounts by name or email..."
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
      </div>

      {/* Table */}
      <CodexAccountList
        data={filteredAccounts}
        loading={loading}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onToggle={handleToggle}
        onRefreshToken={handleRefreshToken}
      />

      {/* Create Modal */}
      <CreateCodexAccountModal
        isOpen={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onCreate={handleCreate}
      />

      {/* Edit Modal */}
      <EditCodexAccountModal
        isOpen={editModalOpen}
        onClose={() => setEditModalOpen(false)}
        onSave={handleSaveEdit}
        account={selectedAccount}
      />

      {/* OAuth Modal */}
      <Modal
        isOpen={oauthModalOpen}
        onClose={() => setOauthModalOpen(false)}
        title="Add Codex Account via OAuth"
        size="lg"
      >
        <CodexOAuthFlow onSuccess={handleOAuthSuccess} onCancel={() => setOauthModalOpen(false)} />
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
          Are you sure you want to delete the Codex account "
          <span className="font-semibold">{selectedAccount?.name}</span>"? This action cannot be
          undone.
        </p>
      </Modal>
    </div>
  );
};
