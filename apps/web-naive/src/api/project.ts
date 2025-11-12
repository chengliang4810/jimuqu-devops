import { requestClient } from './request';

/**
 * 项目接口类型定义
 */
export interface Project {
  id: number;
  name: string;
  code: string;
  remark?: string;
  git_repo?: string;
  git_username?: string;
  git_password?: string;
  webhook_password?: string;
  created_at?: string;
  updated_at?: string;
  deleted_at?: string;
}

/**
 * 创建项目请求参数
 */
export interface CreateProjectParams {
  name: string;
  code: string;
  remark?: string;
  git_repo?: string;
  git_username?: string;
  git_password?: string;
  webhook_password?: string;
}

/**
 * 更新项目请求参数
 */
export interface UpdateProjectParams extends Partial<CreateProjectParams> {
  id: number;
}

/**
 * 项目列表查询参数
 */
export interface ProjectListParams {
  page?: number;
  pageSize?: number;
  name?: string;
  code?: string;
}

/**
 * 项目列表响应
 */
export interface ProjectListResponse {
  list: Project[];
  total: number;
  page: number;
  pageSize: number;
}

// API方法

/**
 * 获取项目列表
 */
export async function getProjectList(params?: ProjectListParams) {
  return requestClient.get<ProjectListResponse>('/api/project', { params });
}

/**
 * 获取项目详情
 */
export async function getProjectDetail(id: number) {
  return requestClient.get<Project>(`/api/project/${id}`);
}

/**
 * 根据编码获取项目
 */
export async function getProjectByCode(code: string) {
  return requestClient.get<Project>(`/api/project/code/${code}`);
}

/**
 * 创建项目
 */
export async function createProject(params: CreateProjectParams) {
  return requestClient.post<Project>('/api/project', params);
}

/**
 * 更新项目
 */
export async function updateProject(params: UpdateProjectParams) {
  const { id, ...data } = params;
  return requestClient.put<Project>(`/api/project/${id}`, data);
}

/**
 * 删除项目
 */
export async function deleteProject(id: number) {
  return requestClient.delete(`/api/project/${id}`);
}