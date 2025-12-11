import React, { useState, useEffect, useRef } from 'react';
import { Plus, Search, Edit, Trash2, Star, TestTube } from 'lucide-react';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { Modal } from '@/components/common/Modal';
import { ProxyForm, type ProxyFormRef } from '@/components/ProxyForm';
import { proxyApi } from '@/api/proxy';
import type { ProxyConfig, CreateProxyRequest, TestProxyResult } from '@/types/proxy';

export const ProxyManagement: React.FC = () => {
  const [proxies, setProxies] = useState<ProxyConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [testModalOpen, setTestModalOpen] = useState(false);
  const [selectedProxy, setSelectedProxy] = useState<ProxyConfig | null>(null);
  const [testResult, setTestResult] = useState<TestProxyResult | null>(null);
  const createFormRef = useRef<ProxyFormRef>(null);
  const editFormRef = useRef<ProxyFormRef>(null);

  const fetchProxies = async () => {
    setLoading(true);
    try {
      const response = await proxyApi.list(1, 100);
      setProxies(response.data.data?.items || []);
    } catch (error) {
      console.error('Failed to fetch proxies:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProxies();
  }, []);

  const handleCreate = async (data: CreateProxyRequest) => {
    try {
      await proxyApi.create(data);
      setCreateModalOpen(false);
      fetchProxies();
    } catch (error) {
      console.error('Failed to create proxy:', error);
    }
  };

  const handleEdit = (proxy: ProxyConfig) => {
    setSelectedProxy(proxy);
    setEditModalOpen(true);
  };

  const handleUpdate = async (data: CreateProxyRequest) => {
    if (!selectedProxy) return;
    try {
      const updateData: any = { ...data };
      if (!updateData.password) delete updateData.password;
      await proxyApi.update(selectedProxy.id, updateData);
      setEditModalOpen(false);
      setSelectedProxy(null);
      fetchProxies();
    } catch (error) {
      console.error('Failed to update proxy:', error);
    }
  };

  const handleDelete = (proxy: ProxyConfig) => {
    setSelectedProxy(proxy);
    setDeleteModalOpen(true);
  };

  const confirmDelete = async () => {
    if (!selectedProxy) return;
    try {
      await proxyApi.delete(selectedProxy.id);
      setDeleteModalOpen(false);
      setSelectedProxy(null);
      fetchProxies();
    } catch (error) {
      console.error('Failed to delete proxy:', error);
    }
  };

  const handleSetDefault = async (id: number) => {
    try {
      await proxyApi.setDefault(id);
      fetchProxies();
    } catch (error) {
      console.error('Failed to set default proxy:', error);
    }
  };

  const handleTest = async (proxy: ProxyConfig) => {
    setSelectedProxy(proxy);
    setTestResult(null);
    setTestModalOpen(true);
    try {
      const response = await proxyApi.test(proxy.id);
      setTestResult(response.data.data || null);
    } catch (error: any) {
      setTestResult({
        success: false,
        message: 'Test failed',
        error: error.message || 'Unknown error',
      });
    }
  };

  const handleToggle = async (proxy: ProxyConfig) => {
    try {
      await proxyApi.update(proxy.id, { enabled: !proxy.enabled });
      fetchProxies();
    } catch (error) {
      console.error('Failed to toggle proxy:', error);
    }
  };

  const filteredProxies = proxies.filter((proxy) =>
    proxy.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Proxy Management</h1>
          <p className="text-gray-600 mt-1">Manage proxy configurations</p>
        </div>
        <Button onClick={() => setCreateModalOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create Proxy
        </Button>
      </div>

      <div className="flex items-center space-x-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <Input
              placeholder="Search proxies..."
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Protocol</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Address</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Default</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {loading ? (
              <tr>
                <td colSpan={6} className="px-6 py-4 text-center text-gray-500">Loading...</td>
              </tr>
            ) : filteredProxies.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-4 text-center text-gray-500">No proxies found</td>
              </tr>
            ) : (
              filteredProxies.map((proxy) => (
                <tr key={proxy.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">{proxy.name}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-sm text-gray-900 uppercase">{proxy.protocol}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-sm text-gray-900">{proxy.host}:{proxy.port}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <button
                      onClick={() => handleToggle(proxy)}
                      className={`px-2 py-1 text-xs rounded-full ${
                        proxy.enabled
                          ? 'bg-green-100 text-green-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}
                    >
                      {proxy.enabled ? 'Enabled' : 'Disabled'}
                    </button>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {proxy.is_default ? (
                      <Star className="w-5 h-5 text-yellow-500 fill-yellow-500" />
                    ) : (
                      <button
                        onClick={() => handleSetDefault(proxy.id)}
                        className="text-gray-400 hover:text-yellow-500"
                      >
                        <Star className="w-5 h-5" />
                      </button>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end space-x-2">
                      <button
                        onClick={() => handleTest(proxy)}
                        className="text-blue-600 hover:text-blue-900"
                      >
                        <TestTube className="w-5 h-5" />
                      </button>
                      <button
                        onClick={() => handleEdit(proxy)}
                        className="text-blue-600 hover:text-blue-900"
                      >
                        <Edit className="w-5 h-5" />
                      </button>
                      <button
                        onClick={() => handleDelete(proxy)}
                        className="text-red-600 hover:text-red-900"
                      >
                        <Trash2 className="w-5 h-5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <Modal
        isOpen={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        title="Create Proxy"
        size="md"
        footer={
          <div className="flex justify-end space-x-3">
            <Button variant="ghost" onClick={() => setCreateModalOpen(false)}>Cancel</Button>
            <Button onClick={() => createFormRef.current?.submit()}>Create</Button>
          </div>
        }
      >
        <ProxyForm
          ref={createFormRef}
          onSubmit={handleCreate}
        />
      </Modal>

      <Modal
        isOpen={editModalOpen}
        onClose={() => setEditModalOpen(false)}
        title="Edit Proxy"
        size="md"
        footer={
          <div className="flex justify-end space-x-3">
            <Button variant="ghost" onClick={() => setEditModalOpen(false)}>Cancel</Button>
            <Button onClick={() => editFormRef.current?.submit()}>Save</Button>
          </div>
        }
      >
        {selectedProxy && (
          <ProxyForm
            ref={editFormRef}
            initialData={selectedProxy}
            onSubmit={handleUpdate}
          />
        )}
      </Modal>

      <Modal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Confirm Delete"
        size="sm"
        footer={
          <div className="flex justify-end space-x-3">
            <Button variant="ghost" onClick={() => setDeleteModalOpen(false)}>Cancel</Button>
            <Button variant="danger" onClick={confirmDelete}>Delete</Button>
          </div>
        }
      >
        <p className="text-gray-700">
          Are you sure you want to delete the proxy "<span className="font-semibold">{selectedProxy?.name}</span>"? This action cannot be undone.
        </p>
      </Modal>

      <Modal
        isOpen={testModalOpen}
        onClose={() => setTestModalOpen(false)}
        title="Proxy Test Result"
        size="md"
        footer={
          <div className="flex justify-end">
            <Button onClick={() => setTestModalOpen(false)}>Close</Button>
          </div>
        }
      >
        {!testResult ? (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">Testing proxy connection...</p>
          </div>
        ) : (
          <div className="space-y-4">
            <div className={`p-4 rounded-lg ${testResult.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
              <p className={`font-medium ${testResult.success ? 'text-green-800' : 'text-red-800'}`}>
                {testResult.message}
              </p>
            </div>
            {testResult.success && testResult.ip && (
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs text-gray-500">IP Address</label>
                    <p className="text-sm font-medium text-gray-900">{testResult.ip}</p>
                  </div>
                  {testResult.country && (
                    <div>
                      <label className="text-xs text-gray-500">Country</label>
                      <p className="text-sm font-medium text-gray-900">{testResult.country}</p>
                    </div>
                  )}
                  {testResult.region && (
                    <div>
                      <label className="text-xs text-gray-500">Region</label>
                      <p className="text-sm font-medium text-gray-900">{testResult.region}</p>
                    </div>
                  )}
                  {testResult.city && (
                    <div>
                      <label className="text-xs text-gray-500">City</label>
                      <p className="text-sm font-medium text-gray-900">{testResult.city}</p>
                    </div>
                  )}
                  {testResult.isp && (
                    <div>
                      <label className="text-xs text-gray-500">ISP</label>
                      <p className="text-sm font-medium text-gray-900">{testResult.isp}</p>
                    </div>
                  )}
                  {testResult.response_ms && (
                    <div>
                      <label className="text-xs text-gray-500">Response Time</label>
                      <p className="text-sm font-medium text-gray-900">{testResult.response_ms}ms</p>
                    </div>
                  )}
                </div>
              </div>
            )}
            {testResult.error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                <p className="text-sm text-red-800">{testResult.error}</p>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};
