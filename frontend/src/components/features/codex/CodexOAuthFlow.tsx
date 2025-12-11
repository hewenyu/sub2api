import React, { useState, useEffect } from 'react';
import { Copy, CheckCircle } from 'lucide-react';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { codexApi } from '@/api/codex';
import { proxyApi } from '@/api/proxy';
import { message } from '@/utils/message';
import type { CodexAccount } from '@/types/account';
import type { ProxyNameInfo } from '@/types/proxy';

interface CodexOAuthFlowProps {
  onSuccess: (account: CodexAccount) => void;
  onCancel: () => void;
}

type OAuthStep = 'generate' | 'link' | 'verify' | 'configure';

export const CodexOAuthFlow: React.FC<CodexOAuthFlowProps> = ({ onSuccess, onCancel }) => {
  const [step, setStep] = useState<OAuthStep>('generate');
  const [loading, setLoading] = useState(false);
  const [callbackPort, setCallbackPort] = useState(1455);
  const [authURL, setAuthURL] = useState('');
  const [callbackURL, setCallbackURL] = useState('');
  const [state, setState] = useState('');
  const [authInput, setAuthInput] = useState('');
  const [accountInfo, setAccountInfo] = useState<{
    email: string;
    subscription_level?: string;
  }>({ email: '' });
  const [accountConfig, setAccountConfig] = useState({
    name: '',
    priority: 100,
    schedulable: true,
    remarks: '',
    proxy_name: undefined as string | undefined,
  });
  const [verifiedAccount, setVerifiedAccount] = useState<CodexAccount | null>(null);
  const [error, setError] = useState('');
  const [proxies, setProxies] = useState<ProxyNameInfo[]>([]);

  useEffect(() => {
    proxyApi
      .getProxyNames()
      .then((res) => setProxies(res.data.data || []))
      .catch(() => setProxies([]));
  }, []);

  const handleGenerateAuthURL = async () => {
    try {
      setLoading(true);
      setError('');
      const response = await codexApi.generateAuthURL(callbackPort);
      setAuthURL(response.auth_url);
      setCallbackURL(response.callback_url);
      setState(response.state);
      setStep('link');
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to generate authorization URL';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleCopyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(authURL);
      message.success('Authorization link copied to clipboard');
    } catch {
      setError('Failed to copy to clipboard');
    }
  };

  const handleVerifyAuth = async () => {
    try {
      setLoading(true);
      setError('');

      let code = '';
      if (authInput.includes('code=')) {
        try {
          const url = new URL(authInput);
          code = url.searchParams.get('code') || '';
          const returnedState = url.searchParams.get('state') || '';

          if (returnedState !== state) {
            throw new Error('State mismatch - authorization may have been tampered with');
          }
        } catch (urlError) {
          if (urlError instanceof Error && urlError.message.includes('State mismatch')) {
            throw urlError;
          }
          setError('Invalid callback URL format');
          return;
        }
      } else {
        code = authInput.trim();
      }

      if (!code) {
        throw new Error('Invalid authorization code');
      }

      const account = await codexApi.verifyAuth(code, state, {
        name: 'Temporary',
        account_type: 'openai-oauth',
        priority: 100,
        schedulable: true,
        proxy_name: accountConfig.proxy_name,
      });

      setVerifiedAccount(account);

      setAccountInfo({
        email: account.email || '',
        subscription_level: account.account_type,
      });

      setAccountConfig((prev) => ({
        ...prev,
        name: account.name || account.email || prev.name,
      }));

      setStep('configure');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to verify authorization';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveAccount = async () => {
    try {
      setLoading(true);
      setError('');

      if (!verifiedAccount) {
        throw new Error('Please complete authorization verification first');
      }

      // Update configurable fields (name / priority / schedulable / proxy_name) on the created account
      await codexApi.update(verifiedAccount.id, {
        name: accountConfig.name || verifiedAccount.name,
        priority: accountConfig.priority,
        schedulable: accountConfig.schedulable,
        proxy_name: accountConfig.proxy_name,
      });

      const updatedAccount: CodexAccount = {
        ...verifiedAccount,
        name: accountConfig.name || verifiedAccount.name,
        priority: accountConfig.priority,
        schedulable: accountConfig.schedulable,
      };

      message.success('Codex account created successfully!');
      onSuccess(updatedAccount);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to save account';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setStep('generate');
    setAuthURL('');
    setCallbackURL('');
    setState('');
    setAuthInput('');
    setAccountInfo({ email: '' });
    setAccountConfig({
      name: '',
      priority: 100,
      schedulable: true,
      remarks: '',
      proxy_name: undefined,
    });
    setVerifiedAccount(null);
    setError('');
    setCallbackPort(1455);
    onCancel();
  };

  return (
    <div className="space-y-6">
      {step === 'generate' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">OAuth Authorization Flow (Manual Mode)</h3>
          <p className="text-sm text-gray-600">Step 1/4: Generate Authorization Link</p>

          <div>
            <label className="block text-sm font-medium mb-2">Callback Port</label>
            <Input
              type="number"
              value={callbackPort}
              onChange={(e) => setCallbackPort(parseInt(e.target.value) || 1455)}
              placeholder="1455"
            />
            <p className="text-xs text-gray-500 mt-1">
              Must be 1455 to match official Codex CLI redirect URI.
            </p>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h4 className="font-semibold text-blue-900 mb-2">Important Notes:</h4>
            <ul className="text-sm text-blue-800 space-y-1">
              <li>- Requires valid OpenAI account with API access</li>
              <li>- You will need to manually copy the callback URL after authorization</li>
              <li>- Token will be securely encrypted and stored</li>
            </ul>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <div className="flex justify-end space-x-2">
            <Button variant="ghost" onClick={handleClose}>
              Cancel
            </Button>
            <Button onClick={handleGenerateAuthURL} loading={loading}>
              Generate Authorization Link
            </Button>
          </div>
        </div>
      )}

      {step === 'link' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Step 2/4: Complete Authorization in Browser</h3>

          <div>
            <label className="block text-sm font-medium mb-2">Authorization Link:</label>
            <textarea
              value={authURL}
              readOnly
              onClick={(e) => e.currentTarget.select()}
              className="w-full px-4 py-3 border rounded-lg bg-gray-50 font-mono text-sm resize-none cursor-pointer"
              rows={3}
            />
            <Button size="sm" onClick={handleCopyToClipboard} className="mt-2">
              <Copy className="w-4 h-4 mr-2" />
              Copy Link
            </Button>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Callback Address:</label>
            <div className="px-4 py-2 border rounded-lg bg-gray-50 font-mono text-sm">
              {callbackURL}
            </div>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <h4 className="font-semibold text-yellow-900 mb-3">Operation Steps:</h4>
            <ul className="text-sm text-yellow-800 space-y-2">
              <li>1. Click "Copy Link" button above</li>
              <li>2. Paste and open it in a new browser tab</li>
              <li>3. Login and authorize your OpenAI account</li>
              <li>4. After authorization completes, copy the full URL from browser address bar</li>
              <li>5. Return to this page and paste the callback URL in next step</li>
            </ul>
          </div>

          <div className="flex justify-between">
            <Button variant="ghost" onClick={() => setStep('generate')}>
              Previous
            </Button>
            <Button onClick={() => setStep('verify')}>Next</Button>
          </div>
        </div>
      )}

      {step === 'verify' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Step 3/4: Verify Authorization</h3>
          <p className="text-sm text-gray-600">Please paste one of the following:</p>

          <div>
            <label className="block text-sm font-medium mb-2">
              Full Callback URL (Recommended)
            </label>
            <textarea
              value={authInput}
              onChange={(e) => setAuthInput(e.target.value)}
              className="w-full px-4 py-3 border rounded-lg font-mono text-sm"
              rows={3}
              placeholder={
                'Paste the full callback URL, for example:\nhttp://localhost:1455/auth/callback?code=xxx&state=yyy'
              }
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Or Just the Authorization Code</label>
            <p className="text-xs text-gray-500">
              You can paste only the value of the code parameter from the URL
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Proxy Configuration (Optional)</label>
            <select
              value={accountConfig.proxy_name || ''}
              onChange={(e) =>
                setAccountConfig({ ...accountConfig, proxy_name: e.target.value || undefined })
              }
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">No Proxy (Direct Connection)</option>
              {proxies
                .filter((proxy) => proxy.enabled)
                .map((proxy) => (
                  <option key={proxy.name} value={proxy.name}>
                    {proxy.name} ({proxy.protocol.toUpperCase()})
                  </option>
                ))}
            </select>
            <p className="text-xs text-gray-500 mt-1">
              Select a proxy if you need to route OAuth requests through a proxy server. Most users
              can leave this as "No Proxy".
            </p>
            {proxies.filter((p) => p.enabled).length === 0 && (
              <p className="text-xs text-yellow-600 mt-1">
                No proxies configured. You can add proxies in the Proxy Management page.
              </p>
            )}
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
            <p className="text-sm text-blue-800">
              Tip: The browser address bar contains the full URL with ?code=xxx&state=yyy
            </p>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <div className="flex justify-between">
            <Button variant="ghost" onClick={() => setStep('link')}>
              Previous
            </Button>
            <Button onClick={handleVerifyAuth} loading={loading} disabled={!authInput || loading}>
              {loading ? 'Verifying...' : 'Verify Authorization'}
            </Button>
          </div>
        </div>
      )}

      {step === 'configure' && (
        <div className="space-y-4">
          <div className="bg-green-50 border border-green-200 rounded-lg p-4">
            <div className="flex items-center">
              <CheckCircle className="w-5 h-5 text-green-600 mr-2" />
              <span className="text-green-800 font-semibold">Authorization Successful!</span>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Email (Read-only)</label>
            <Input value={accountInfo.email} readOnly className="bg-gray-50" />
          </div>

          {accountInfo.subscription_level && (
            <div>
              <label className="block text-sm font-medium mb-2">Account Type (Auto-detected)</label>
              <div className="px-4 py-2 border rounded-lg bg-gray-50">
                <span className="px-3 py-1 text-sm font-semibold rounded bg-blue-100 text-blue-800">
                  {accountInfo.subscription_level}
                </span>
              </div>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium mb-2">Account Name</label>
            <Input
              value={accountConfig.name}
              onChange={(e) => setAccountConfig({ ...accountConfig, name: e.target.value })}
              placeholder="My Codex Account"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Priority</label>
            <Input
              type="number"
              value={accountConfig.priority}
              onChange={(e) =>
                setAccountConfig({ ...accountConfig, priority: parseInt(e.target.value) || 100 })
              }
              placeholder="100"
            />
            <p className="text-xs text-gray-500 mt-1">Higher value means higher priority</p>
          </div>

          <div className="flex items-center">
            <input
              type="checkbox"
              checked={accountConfig.schedulable}
              onChange={(e) =>
                setAccountConfig({ ...accountConfig, schedulable: e.target.checked })
              }
              className="mr-2"
              id="schedulable"
            />
            <label htmlFor="schedulable" className="text-sm font-medium">
              Schedulable
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Proxy Configuration (Optional)</label>
            <select
              value={accountConfig.proxy_name || ''}
              onChange={(e) =>
                setAccountConfig({ ...accountConfig, proxy_name: e.target.value || undefined })
              }
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">No Proxy (Direct Connection)</option>
              {proxies
                .filter((proxy) => proxy.enabled)
                .map((proxy) => (
                  <option key={proxy.name} value={proxy.name}>
                    {proxy.name} ({proxy.protocol.toUpperCase()})
                  </option>
                ))}
            </select>
            <p className="text-xs text-gray-500 mt-1">
              Select a proxy if you need to route OAuth requests through a proxy server. Most users
              can leave this as "No Proxy".
            </p>
            {proxies.filter((p) => p.enabled).length === 0 && (
              <p className="text-xs text-yellow-600 mt-1">
                No proxies configured. You can add proxies in the Proxy Management page.
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Remarks (Optional)</label>
            <textarea
              value={accountConfig.remarks}
              onChange={(e) => setAccountConfig({ ...accountConfig, remarks: e.target.value })}
              className="w-full px-4 py-3 border rounded-lg text-sm"
              rows={2}
              placeholder="Optional notes about this account"
            />
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <div className="flex justify-end space-x-2">
            <Button variant="ghost" onClick={handleClose}>
              Cancel
            </Button>
            <Button onClick={handleSaveAccount} loading={loading}>
              Save Account
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};
