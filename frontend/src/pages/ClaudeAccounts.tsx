import React, { useState, useEffect, useCallback } from 'react';
import { Plus, Search, RefreshCw } from 'lucide-react';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { Modal } from '@/components/common/Modal';
import { ClaudeAccountList } from '@/components/features/claude/ClaudeAccountList';
import { ClaudeOAuthFlow } from '@/components/features/claude/ClaudeOAuthFlow';
import { EditClaudeAccountModal } from '@/components/features/claude/EditClaudeAccountModal';
import { claudeApi } from '@/api/claude';
import { message } from '@/utils/message';
import type { ClaudeAccount, UpdateClaudeAccountRequest } from '@/types/account';

export const ClaudeAccounts: React.FC = () => {
  const [accounts, setAccounts] = useState<ClaudeAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<ClaudeAccount | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchAccounts = useCallback(async () => {
    try {
      setLoading(true);
      const response = await claudeApi.list({ page: 1, page_size: 100 });
      setAccounts(response.items);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to fetch Claude accounts';
      message.error(errorMessage);
      setAccounts([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAccounts();
  }, [fetchAccounts]);

  const handleCreateSuccess = () => {
    setCreateModalOpen(false);
    fetchAccounts();
    message.success('Claude account added successfully');
  };

  const handleEdit = (account: ClaudeAccount) => {
    setSelectedAccount(account);
    setEditModalOpen(true);
  };

  const handleSaveEdit = async (id: number, data: UpdateClaudeAccountRequest) => {
    try {
      await claudeApi.update(id, data);
      message.success('Claude account updated successfully');
      await fetchAccounts();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to update Claude account';
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
        await claudeApi.delete(selectedAccount.id);
        message.success('Claude account deleted successfully');
        setDeleteModalOpen(false);
        setSelectedAccount(null);
        await fetchAccounts();
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : 'Failed to delete Claude account';
        message.error(errorMessage);
      }
    }
  };

  const handleToggle = async (id: number, isActive: boolean) => {
    try {
      await claudeApi.toggle(id, isActive);
      message.success(isActive ? 'Account activated' : 'Account deactivated');
      await fetchAccounts();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to toggle account status';
      message.error(errorMessage);
    }
  };

  const handleRefreshToken = async (id: number) => {
    try {
      await claudeApi.refreshToken(id);
      message.success('Token refreshed successfully');
      await fetchAccounts();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to refresh token';
      message.error(errorMessage);
    }
  };

  const filteredAccounts = accounts.filter((account) =>
    account.email.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Claude Account Management</h1>
          <p className="text-gray-600 mt-1">Manage Claude AI accounts and OAuth authorization</p>
        </div>
        <div className="flex space-x-2">
          <Button variant="outline" onClick={fetchAccounts}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={() => setCreateModalOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add Account
          </Button>
        </div>
      </div>

      {/* Search and Filter */}
      <div className="flex items-center space-x-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <Input
              placeholder="Search accounts by email..."
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
      </div>

      {/* Table */}
      <ClaudeAccountList
        data={filteredAccounts}
        loading={loading}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onToggle={handleToggle}
        onRefreshToken={handleRefreshToken}
      />

      {/* Create OAuth Flow Modal */}
      <ClaudeOAuthFlow
        isOpen={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onSuccess={handleCreateSuccess}
      />

      {/* Edit Modal */}
      <EditClaudeAccountModal
        isOpen={editModalOpen}
        onClose={() => setEditModalOpen(false)}
        onSave={handleSaveEdit}
        account={selectedAccount}
      />

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
          Are you sure you want to delete the Claude account "
          <span className="font-semibold">{selectedAccount?.email}</span>"? This action cannot be
          undone.
        </p>
      </Modal>
    </div>
  );
};
