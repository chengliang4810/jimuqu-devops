import { request } from '../request';

/**
 * 获取主机列表
 */
export function fetchGetHostList(params?: Api.Host.HostSearchParams) {
  return request<Api.Host.HostListResponse>({
    url: '/hosts',
    method: 'get',
    params
  });
}

/**
 * 获取单个主机详情
 */
export function fetchGetHost(id: number) {
  return request<Api.Host.Host>({
    url: `/hosts/${id}`,
    method: 'get'
  });
}

/**
 * 创建主机
 */
export function fetchCreateHost(data: Api.Host.HostCreate) {
  return request<Api.Host.Host>({
    url: '/hosts',
    method: 'post',
    data
  });
}

/**
 * 更新主机
 */
export function fetchUpdateHost(id: number, data: Api.Host.HostUpdate) {
  return request<Api.Host.Host>({
    url: `/hosts/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除主机
 */
export function deleteHost(id: number) {
  return request<void>({
    url: `/hosts/${id}`,
    method: 'delete'
  });
}

/**
 * 测试主机连接
 */
export function testHostConnection(id: number) {
  return request<Api.Host.HostTestConnection>({
    url: `/hosts/${id}/test-connection`,
    method: 'post'
  });
}

/**
 * 切换主机状态
 */
export function toggleHostStatus(id: number) {
  return request<{ id: number; is_active: boolean; message: string }>({
    url: `/hosts/${id}/toggle-status`,
    method: 'post'
  });
}

/**
 * 获取主机分组列表
 */
export function fetchGetHostGroups() {
  return request<{ groups: string[] }>({
    url: '/hosts/groups',
    method: 'get'
  });
}