import { request } from '@/utils/request';
import type {
  ProxyConfig,
  ProxyListResponse,
  CreateProxyRequest,
  UpdateProxyRequest,
  TestProxyResult,
  ProxyNameInfo,
} from '@/types/proxy';

export const proxyApi = {
  list: (page = 1, pageSize = 20) =>
    request.get<ProxyListResponse>('/admin/proxies', { page, page_size: pageSize }),

  create: (data: CreateProxyRequest) => request.post<ProxyConfig>('/admin/proxies', data),

  get: (id: number) => request.get<ProxyConfig>(`/admin/proxies/${id}`),

  update: (id: number, data: UpdateProxyRequest) => request.put<void>(`/admin/proxies/${id}`, data),

  delete: (id: number) => request.delete<void>(`/admin/proxies/${id}`),

  setDefault: (id: number) => request.post<void>(`/admin/proxies/${id}/set-default`),

  test: (id: number) => request.post<TestProxyResult>(`/admin/proxies/${id}/test`),

  getProxyNames: () => request.get<ProxyNameInfo[]>('/admin/proxies/names'),
};
