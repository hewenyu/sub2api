import React, { useState, useEffect, useImperativeHandle, forwardRef } from 'react';
import { Input } from '@/components/common/Input';
import type { CreateProxyRequest, ProxyConfig } from '@/types/proxy';

interface ProxyFormProps {
  initialData?: ProxyConfig;
  onSubmit: (data: CreateProxyRequest) => void;
}

export interface ProxyFormRef {
  submit: () => void;
}

export const ProxyForm = forwardRef<ProxyFormRef, ProxyFormProps>(({ initialData, onSubmit }, ref) => {
  const [formData, setFormData] = useState<CreateProxyRequest>({
    name: '',
    enabled: true,
    protocol: 'http',
    host: '',
    port: 8080,
    username: '',
    password: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (initialData) {
      setFormData({
        name: initialData.name,
        enabled: initialData.enabled,
        protocol: initialData.protocol,
        host: initialData.host,
        port: initialData.port,
        username: initialData.username || '',
        password: '',
      });
    }
  }, [initialData]);

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }
    if (!formData.host.trim()) {
      newErrors.host = 'Host is required';
    }
    if (formData.port < 1 || formData.port > 65535) {
      newErrors.port = 'Port must be between 1 and 65535';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault();
    if (validate()) {
      const submitData = { ...formData };
      if (!submitData.username) delete submitData.username;
      if (!submitData.password) delete submitData.password;
      onSubmit(submitData);
    }
  };

  useImperativeHandle(ref, () => ({
    submit: () => handleSubmit(),
  }));

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Name <span className="text-red-500">*</span>
        </label>
        <Input
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          placeholder="My Proxy"
        />
        {errors.name && <p className="text-red-500 text-xs mt-1">{errors.name}</p>}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Protocol <span className="text-red-500">*</span>
        </label>
        <select
          value={formData.protocol}
          onChange={(e) => setFormData({ ...formData, protocol: e.target.value })}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="http">HTTP</option>
          <option value="https">HTTPS</option>
          <option value="socks5">SOCKS5</option>
        </select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Host <span className="text-red-500">*</span>
          </label>
          <Input
            value={formData.host}
            onChange={(e) => setFormData({ ...formData, host: e.target.value })}
            placeholder="proxy.example.com"
          />
          {errors.host && <p className="text-red-500 text-xs mt-1">{errors.host}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Port <span className="text-red-500">*</span>
          </label>
          <Input
            type="number"
            min="1"
            max="65535"
            value={formData.port}
            onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) || 0 })}
          />
          {errors.port && <p className="text-red-500 text-xs mt-1">{errors.port}</p>}
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">Username</label>
        <Input
          value={formData.username}
          onChange={(e) => setFormData({ ...formData, username: e.target.value })}
          placeholder="Optional"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Password
          {initialData && <span className="text-gray-500 text-xs ml-2">(leave empty to keep current)</span>}
        </label>
        <Input
          type="password"
          value={formData.password}
          onChange={(e) => setFormData({ ...formData, password: e.target.value })}
          placeholder="Optional"
        />
      </div>

      <div className="flex items-center">
        <input
          type="checkbox"
          id="enabled"
          checked={formData.enabled}
          onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
          className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
        />
        <label htmlFor="enabled" className="ml-2 text-sm text-gray-700">
          Enable this proxy
        </label>
      </div>
    </form>
  );
});

ProxyForm.displayName = 'ProxyForm';
