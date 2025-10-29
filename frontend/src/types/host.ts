// 主机分组类型定义
export interface HostGroup {
  id: string;
  name: string;
  description?: string;
  sort: number;
  hostCount: number;
  createdAt: string;
  updatedAt: string;
}

// 主机信息类型定义
export interface Host {
  id: string;
  groupId: string;
  groupName: string;
  name: string;
  ip: string;
  port: number;
  username: string;
  password: string;
  authType: 'password' | 'key';
  privateKey?: string;
  description?: string;
  status: 'online' | 'offline' | 'unknown';
  lastConnected?: string;
  createdAt: string;
  updatedAt: string;
}

// 主机分组表单类型
export interface HostGroupForm {
  name: string;
  description?: string;
}

// 主机信息表单类型
export interface HostForm {
  groupId: string;
  name: string;
  ip: string;
  port: number;
  username: string;
  password: string;
  authType: 'password' | 'key';
  privateKey?: string;
  description?: string;
}

// API 响应类型
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  success: boolean;
}

// 分页响应类型
export interface PageResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}