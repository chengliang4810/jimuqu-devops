import { requestClient } from './request';

/**
 * 部署配置接口类型定义
 */
export interface DeployConfig {
  id: number;
  project_id: number;
  branch: string;
  config: DeployConfigItem[];
  created_at?: string;
  updated_at?: string;
}

/**
 * 部署配置项
 */
export interface DeployConfigItem {
  key: string;
  value: any;
  desc: string;
}

/**
 * 创建部署配置请求参数
 */
export interface CreateDeployConfigParams {
  project_id: number;
  branch: string;
  config: DeployConfigItem[];
}

/**
 * 更新部署配置请求参数
 */
export interface UpdateDeployConfigParams {
  branch: string;
  config: DeployConfigItem[];
}

/**
 * 部署配置列表查询参数
 */
export interface DeployConfigListParams {
  page?: number;
  pageSize?: number;
  project_id?: number;
  branch?: string;
}

/**
 * 部署配置列表响应
 */
export interface DeployConfigListResponse {
  items: DeployConfig[];
  total: number;
  page: number;
  pageSize: number;
}

// API方法

/**
 * 获取部署配置列表
 */
export async function getDeployConfigList(params?: DeployConfigListParams) {
  return requestClient.get<DeployConfigListResponse>('/api/deploy-config', { params });
}

/**
 * 获取部署配置详情
 */
export async function getDeployConfigDetail(id: number) {
  return requestClient.get<DeployConfig>(`/api/deploy-config/${id}`);
}

/**
 * 根据项目ID获取部署配置列表
 */
export async function getDeployConfigByProjectId(projectId: number) {
  return requestClient.get<DeployConfig[]>(`/api/deploy-config/project/${projectId}`);
}

/**
 * 根据项目ID和分支获取部署配置
 */
export async function getDeployConfigByProjectAndBranch(projectId: number, branch: string) {
  return requestClient.get<DeployConfig>(`/api/deploy-config/project/${projectId}/branch/${branch}`);
}

/**
 * 创建部署配置
 */
export async function createDeployConfig(params: CreateDeployConfigParams) {
  return requestClient.post<DeployConfig>('/api/deploy-config', params);
}

/**
 * 更新部署配置
 */
export async function updateDeployConfig(id: number, params: UpdateDeployConfigParams) {
  return requestClient.put<DeployConfig>(`/api/deploy-config/${id}`, params);
}

/**
 * 删除部署配置
 */
export async function deleteDeployConfig(id: number) {
  return requestClient.delete(`/api/deploy-config/${id}`);
}