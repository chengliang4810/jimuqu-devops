import type { ApiResponse, PageResponse, HostGroup, Host, HostGroupForm, HostForm } from '@/types/host';
import request from '@/utils/request';

// 主机分组相关 API
export const hostGroupApi = {
  // 获取所有主机分组
  getAll: () => {
    return request.get<ApiResponse<HostGroup[]>>('/api/host-groups');
  },

  // 创建主机分组
  create: (data: HostGroupForm) => {
    return request.post<ApiResponse<HostGroup>>('/api/host-groups', data);
  },

  // 更新主机分组
  update: (id: string, data: Partial<HostGroupForm>) => {
    return request.put<ApiResponse<HostGroup>>(`/api/host-groups/${id}`, data);
  },

  // 删除主机分组
  delete: (id: string) => {
    return request.delete<ApiResponse<null>>(`/api/host-groups/${id}`);
  },

  // 更新分组排序
  updateSort: (data: { id: string; sort: number }[]) => {
    return request.put<ApiResponse<null>>('/api/host-groups/sort', { items: data });
  },
};

// 主机相关 API
export const hostApi = {
  // 获取主机列表（分页）
  getList: (params: {
    page?: number;
    pageSize?: number;
    groupId?: string;
    keyword?: string;
  }) => {
    return request.get<ApiResponse<PageResponse<Host>>>('/api/hosts', { params });
  },

  // 根据分组获取主机列表
  getByGroup: (groupId: string) => {
    return request.get<ApiResponse<Host[]>>(`/api/host-groups/${groupId}/hosts`);
  },

  // 创建主机
  create: (data: HostForm) => {
    return request.post<ApiResponse<Host>>('/api/hosts', data);
  },

  // 更新主机
  update: (id: string, data: Partial<HostForm>) => {
    return request.put<ApiResponse<Host>>(`/api/hosts/${id}`, data);
  },

  // 删除主机
  delete: (id: string) => {
    return request.delete<ApiResponse<null>>(`/api/hosts/${id}`);
  },

  // 批量删除主机
  batchDelete: (ids: string[]) => {
    return request.post<ApiResponse<null>>('/api/hosts/batch-delete', { ids });
  },

  // 测试主机连接
  testConnection: (id: string) => {
    return request.post<ApiResponse<{ status: 'online' | 'offline'; message: string }>>(`/api/hosts/${id}/test-connection`);
  },

  // 获取主机详情
  getDetail: (id: string) => {
    return request.get<ApiResponse<Host>>(`/api/hosts/${id}`);
  },
};