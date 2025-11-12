import { requestClient } from './request'

// 部署记录接口类型定义
export interface DeployRecord {
  id: number
  projectId: number
  projectName: string
  branch: string
  startTime: string
  duration: number
  logPath: string
  status: 'running' | 'success' | 'failed'
  createdAt: string
  updatedAt: string
}

// 部署记录查询参数
export interface DeployRecordQuery {
  pageNum?: number
  pageSize?: number
  projectId?: number
  projectName?: string
  branch?: string
  status?: 'running' | 'success' | 'failed'
  startTimeStart?: string
  startTimeEnd?: string
}

// 创建部署记录请求参数
export interface CreateDeployRecordParams {
  projectId: number
  projectName: string
  branch: string
  startTime?: string
  duration?: number
  logPath?: string
  status?: 'running' | 'success' | 'failed'
}

// 分页响应数据
export interface PageData<T> {
  rows: T[]
  total: number
}

// 部署记录统计信息
export interface DeployRecordStats {
  total: number
  success: number
  failed: number
  running: number
}

// 获取部署记录列表（分页）
export function getDeployRecords(params: DeployRecordQuery = {}) {
  return requestClient.get<PageData<DeployRecord>>('/api/deploy-record', {
    params,
  })
}

// 根据ID获取部署记录详情
export function getDeployRecord(id: number) {
  return requestClient.get<DeployRecord>(`/api/deploy-record/${id}`)
}

// 创建部署记录
export function createDeployRecord(data: CreateDeployRecordParams) {
  return requestClient.post<DeployRecord>('/api/deploy-record', data)
}

// 更新部署记录
export function updateDeployRecord(id: number, data: Partial<DeployRecord>) {
  return requestClient.put<DeployRecord>(`/api/deploy-record/${id}`, data)
}

// 删除部署记录
export function deleteDeployRecord(id: number) {
  return requestClient.delete<void>(`/api/deploy-record/${id}`)
}

// 根据项目ID获取部署记录
export function getDeployRecordsByProject(projectId: number) {
  return requestClient.get<DeployRecord[]>(`/api/deploy-record/project/${projectId}`)
}

// 根据分支获取部署记录
export function getDeployRecordsByBranch(branch: string) {
  return requestClient.get<DeployRecord[]>(`/api/deploy-record/branch/${branch}`)
}

// 根据状态获取部署记录
export function getDeployRecordsByStatus(status: 'running' | 'success' | 'failed') {
  return requestClient.get<DeployRecord[]>(`/api/deploy-record/status/${status}`)
}

// 获取指定项目和分支的最新部署记录
export function getLatestDeployRecord(projectId: number, branch: string) {
  return requestClient.get<DeployRecord>(`/api/deploy-record/project/${projectId}/branch/${branch}/latest`)
}

// 获取部署记录统计信息
export function getDeployRecordStats(projectId?: number) {
  const url = projectId ? `/api/deploy-record/stats?project_id=${projectId}` : '/api/deploy-record/stats'
  return requestClient.get<DeployRecordStats>(url)
}