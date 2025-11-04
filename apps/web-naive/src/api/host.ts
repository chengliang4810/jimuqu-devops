import { requestClient } from './request';

/**
 * 主机接口类型定义
 */
export interface Host {
  id: number;
  name: string;
  host: string;
  port: number;
  username: string;
  password: string;
  auth_type: 'key' | 'password';
  status: 'inactive' | 'offline' | 'online';
  remark?: string;
  created_at?: string;
  updated_at?: string;
  deleted_at?: string;
}

/**
 * 创建主机请求参数
 */
export interface CreateHostParams {
  name: string;
  host: string;
  port?: number;
  username: string;
  password: string;
  auth_type?: 'key' | 'password';
  status?: 'inactive' | 'offline' | 'online';
  remark?: string;
}

/**
 * 更新主机请求参数
 */
export interface UpdateHostParams extends Partial<CreateHostParams> {
  id: number;
}

/**
 * 主机列表查询参数
 */
export interface HostListParams {
  page?: number;
  pageSize?: number;
  name?: string;
  host?: string;
  status?: 'inactive' | 'offline' | 'online';
}

/**
 * 主机列表响应
 */
export interface HostListResponse {
  items: Host[];
  total: number;
  page: number;
  pageSize: number;
}

/**
 * SSH连接测试请求参数
 */
export interface TestConnectionParams {
  id: number;
}

/**
 * SSH连接测试响应
 */
export interface TestConnectionResponse {
  success: boolean;
  message: string;
  latency?: number;
}

/**
 * 执行SSH命令请求参数
 */
export interface ExecuteCommandParams {
  host_id?: number;
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  command: string;
}

/**
 * 执行SSH命令响应
 */
export interface ExecuteCommandResponse {
  success: boolean;
  output: string;
  error?: string;
  exit_code?: number;
}

/**
 * 批量检查主机状态请求参数
 */
export interface BatchCheckParams {
  host_ids: number[];
}

/**
 * 批量检查主机状态响应
 */
export interface BatchCheckResponse {
  results: Array<{
    host: string;
    id: number;
    latency?: number;
    message: string;
    name: string;
    status: 'offline' | 'online';
  }>;
}

// API方法

/**
 * 获取主机列表
 */
export async function getHostList(params?: HostListParams) {
  return requestClient.get<HostListResponse>('/api/host', { params });
}

/**
 * 获取主机详情
 */
export async function getHostDetail(id: number) {
  return requestClient.get<Host>(`/api/host/${id}`);
}

/**
 * 创建主机
 */
export async function createHost(params: CreateHostParams) {
  return requestClient.post<Host>('/api/host', params);
}

/**
 * 更新主机
 */
export async function updateHost(params: UpdateHostParams) {
  const { id, ...data } = params;
  return requestClient.put<Host>(`/api/host/${id}`, data);
}

/**
 * 删除主机
 */
export async function deleteHost(id: number) {
  return requestClient.delete(`/api/host/${id}`);
}

/**
 * 测试SSH连接
 */
export async function testConnection(params: TestConnectionParams) {
  return requestClient.post<TestConnectionResponse>(
    `/api/host/${params.id}/test`,
  );
}

/**
 * 执行SSH命令
 */
export async function executeCommand(params: ExecuteCommandParams) {
  return requestClient.post<ExecuteCommandResponse>(
    '/api/host/execute',
    params,
  );
}

/**
 * 批量检查主机状态
 */
export async function batchCheckStatus(params: BatchCheckParams) {
  return requestClient.post<BatchCheckResponse>(
    '/api/host/batch-check',
    params,
  );
}

/**
 * 上传单个文件到主机
 */
export async function uploadFile(
  hostId: number,
  file: File,
  remotePath: string,
) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('remote_path', remotePath);
  formData.append('host_id', hostId.toString());

  return requestClient.post<{ message: string; success: boolean }>(
    '/api/host/upload/file',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    },
  );
}

/**
 * 上传目录到主机
 */
export async function uploadDirectory(
  hostId: number,
  files: File[],
  remotePath: string,
) {
  const formData = new FormData();
  files.forEach((file) => formData.append('files', file));
  formData.append('remote_path', remotePath);
  formData.append('host_id', hostId.toString());

  return requestClient.post<{ message: string; success: boolean }>(
    '/api/host/upload/directory',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    },
  );
}

/**
 * 获取Docker信息
 */
export async function getDockerInfo(hostId: number) {
  return requestClient.post('/api/host/docker/info', { host_id: hostId });
}

/**
 * 构建Docker镜像
 */
export async function buildDockerImage(params: {
  context: string;
  dockerfile: string;
  host_id: number;
  tag: string;
}) {
  return requestClient.post('/api/host/docker/build', params);
}

/**
 * 运行Docker容器
 */
export async function runDockerContainer(params: {
  container_name?: string;
  environment?: Record<string, string>;
  host_id: number;
  image: string;
  ports?: string[];
  volumes?: string[];
}) {
  return requestClient.post('/api/host/docker/run', params);
}

/**
 * 执行Docker命令
 */
export async function executeDockerCommand(params: {
  command: string;
  host_id: number;
}) {
  return requestClient.post('/api/host/docker/execute', params);
}
