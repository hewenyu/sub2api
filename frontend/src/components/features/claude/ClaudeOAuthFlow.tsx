import React, { useState } from 'react';
import { Copy, CheckCircle } from 'lucide-react';
import { Modal } from '@/components/common/Modal';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { claudeApi } from '@/api/claude';
import { message } from '@/utils/message';

interface ClaudeOAuthFlowProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const ClaudeOAuthFlow: React.FC<ClaudeOAuthFlowProps> = ({ isOpen, onClose, onSuccess }) => {
  const [step, setStep] = useState<'generate' | 'link' | 'verify' | 'configure'>('generate');
  const [loading, setLoading] = useState(false);
  const [callbackPort, setCallbackPort] = useState(8888);
  const [authUrl, setAuthUrl] = useState('');
  const [callbackUrl, setCallbackUrl] = useState('');
  const [state, setState] = useState('');
  const [authInput, setAuthInput] = useState('');
  const [accountEmail, setAccountEmail] = useState('');
  const [subscriptionLevel, setSubscriptionLevel] = useState<'free' | 'pro' | 'max'>('pro');
  const [priority, setPriority] = useState(1);
  const [schedulable, setSchedulable] = useState(true);
  const [proxyUrl, setProxyUrl] = useState('');
  const [error, setError] = useState('');

  const handleGenerateAuthUrl = async () => {
    try {
      setLoading(true);
      setError('');
      const response = await claudeApi.generateAuthUrl(callbackPort);
      setAuthUrl(response.auth_url);
      setCallbackUrl(response.callback_url);
      setState(response.state);
      setStep('link');
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to generate authorization link';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleCopyAuthUrl = () => {
    navigator.clipboard.writeText(authUrl);
    message.success('Authorization link copied to clipboard');
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
        } catch {
          setError('Invalid callback URL format');
          return;
        }
      } else {
        code = authInput.trim();
      }

      if (!code) {
        setError('Please paste a valid callback URL or authorization code');
        return;
      }

      const account = await claudeApi.verifyAuth(code, state, {
        priority,
        schedulable,
        proxy_url: proxyUrl || undefined,
      });

      setAccountEmail(account.email);
      setSubscriptionLevel(account.subscription_level);
      setStep('configure');
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to verify authorization';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleComplete = () => {
    onSuccess();
    handleClose();
  };

  const handleClose = () => {
    setStep('generate');
    setAuthUrl('');
    setCallbackUrl('');
    setState('');
    setAuthInput('');
    setError('');
    setCallbackPort(8888);
    setPriority(1);
    setSchedulable(true);
    setProxyUrl('');
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Add Claude Account" size="lg">
      {step === 'generate' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">OAuth Authorization Flow (Manual Mode)</h3>
          <p className="text-sm text-gray-600">Step 1: Generate Authorization Link</p>

          <div>
            <label className="block text-sm font-medium mb-2">Callback Port (Optional)</label>
            <Input
              type="number"
              value={callbackPort}
              onChange={(e) => setCallbackPort(parseInt(e.target.value) || 8888)}
              placeholder="8888"
            />
            <p className="text-xs text-gray-500 mt-1">Default: 8888</p>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h4 className="font-semibold text-blue-900 mb-2">Important Notes:</h4>
            <ul className="text-sm text-blue-800 space-y-1">
              <li>- Requires valid Claude Pro/Max subscription</li>
              <li>- You will need to manually copy the callback URL after authorization</li>
              <li>- Token will be automatically refreshed</li>
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
            <Button variant="gradient-orange" onClick={handleGenerateAuthUrl} loading={loading}>
              Generate Authorization Link
            </Button>
          </div>
        </div>
      )}

      {step === 'link' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Step 2: Complete Authorization in Browser</h3>

          <div>
            <label className="block text-sm font-medium mb-2">Authorization Link:</label>
            <textarea
              value={authUrl}
              readOnly
              onClick={(e) => e.currentTarget.select()}
              className="w-full px-4 py-3 border rounded-lg bg-gray-50 font-mono text-sm resize-none cursor-pointer"
              rows={3}
            />
            <Button size="sm" onClick={handleCopyAuthUrl} className="mt-2">
              <Copy className="w-4 h-4 mr-2" />
              Copy Link
            </Button>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Callback Address:</label>
            <div className="px-4 py-2 border rounded-lg bg-gray-50 font-mono text-sm">
              {callbackUrl}
            </div>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <h4 className="font-semibold text-yellow-900 mb-3">Operation Steps:</h4>
            <ul className="text-sm text-yellow-800 space-y-2">
              <li>1. Click "Copy Link"</li>
              <li>2. Paste and open it in a new browser tab</li>
              <li>3. Login and authorize your Claude account</li>
              <li>4. After authorization completes, copy the full URL from browser address bar</li>
              <li>5. Return to this page and paste the callback URL</li>
            </ul>
          </div>

          <div className="flex justify-between">
            <Button variant="ghost" onClick={() => setStep('generate')}>
              Previous
            </Button>
            <Button variant="gradient-orange" onClick={() => setStep('verify')}>
              Next
            </Button>
          </div>
        </div>
      )}

      {step === 'verify' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Step 3: Verify Authorization</h3>
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
                'Paste the full callback URL, for example:\nhttp://localhost:8888/callback?code=xxx&state=yyy'
              }
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Or Just the Authorization Code (code parameter value)
            </label>
            <p className="text-xs text-gray-500">
              Or paste only the value of the code parameter from the URL
            </p>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
            <p className="text-sm text-blue-800">
              Tip: The browser address bar contains the full URL with ?code=xxx
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
            <Button
              variant="gradient-orange"
              onClick={handleVerifyAuth}
              loading={loading}
              disabled={!authInput || loading}
            >
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
            <Input value={accountEmail} readOnly className="bg-gray-50" />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Subscription Level (Auto-detected)
            </label>
            <div className="px-4 py-2 border rounded-lg bg-gray-50">
              <span
                className={`px-3 py-1 text-sm font-semibold rounded ${
                  subscriptionLevel === 'max'
                    ? 'bg-purple-100 text-purple-800'
                    : subscriptionLevel === 'pro'
                      ? 'bg-blue-100 text-blue-800'
                      : 'bg-gray-100 text-gray-800'
                }`}
              >
                {subscriptionLevel.toUpperCase()}
              </span>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Priority</label>
            <Input
              type="number"
              value={priority}
              onChange={(e) => setPriority(parseInt(e.target.value) || 1)}
            />
            <p className="text-xs text-gray-500 mt-1">Higher value means higher priority</p>
          </div>

          <div className="flex items-center">
            <input
              type="checkbox"
              checked={schedulable}
              onChange={(e) => setSchedulable(e.target.checked)}
              className="mr-2"
              id="schedulable"
            />
            <label htmlFor="schedulable" className="text-sm font-medium">
              Schedulable
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Proxy URL (Optional)</label>
            <Input
              value={proxyUrl}
              onChange={(e) => setProxyUrl(e.target.value)}
              placeholder="http://proxy:8080"
            />
          </div>

          <div className="flex justify-end space-x-2">
            <Button variant="ghost" onClick={handleClose}>
              Cancel
            </Button>
            <Button variant="gradient-orange" onClick={handleComplete}>
              Complete
            </Button>
          </div>
        </div>
      )}
    </Modal>
  );
};
