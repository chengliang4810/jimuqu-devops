import type { VxeGridProps } from '#/adapter/vxe-table'
import type { DeployRecord } from '#/api/deploy-record'
import { h } from 'vue'
import { NTag, NButton, NFlex } from 'naive-ui'

// 状态类型和配置
export type DeployStatus = 'running' | 'success' | 'failed'

export const statusOptions = [
  { label: '运行中', value: 'running' as const },
  { label: '成功', value: 'success' as const },
  { label: '失败', value: 'failed' as const },
]

export const statusConfig = {
  running: { label: '运行中', type: 'info' as const },
  success: { label: '成功', type: 'success' as const },
  failed: { label: '失败', type: 'error' as const },
}

// 格式化耗时显示（秒转为可读格式）
export function formatDuration(duration: number): string {
  if (duration < 60) {
    return `${duration}秒`
  } else if (duration < 3600) {
    const minutes = Math.floor(duration / 60)
    const seconds = duration % 60
    return `${minutes}分${seconds}秒`
  } else {
    const hours = Math.floor(duration / 3600)
    const minutes = Math.floor((duration % 3600) / 60)
    const seconds = duration % 60
    return `${hours}小时${minutes}分${seconds}秒`
  }
}

// 状态渲染函数
export function renderStatus(status: DeployStatus) {
  const config = statusConfig[status]
  return h(NTag, { type: config.type }, { default: () => config.label })
}

// 操作列渲染函数
export function renderAction(row: DeployRecord, onView: (record: DeployRecord) => void, onViewLog: (record: DeployRecord) => void) {
  return h(NFlex, { size: 'small' }, () => [
    h(NButton, {
      type: 'info',
      size: 'small',
      onClick: () => onView(row),
    }, { default: () => '查看详情' }),

    row.logPath && h(NButton, {
      type: 'primary',
      size: 'small',
      onClick: () => onViewLog(row),
    }, { default: () => '查看日志' }),
  ])
}

// VxeGrid列配置
export const gridColumns: VxeGridProps['columns'] = [
  {
    type: 'checkbox',
    width: 60,
    align: 'center',
  },
  {
    field: 'id',
    title: 'ID',
    width: 80,
    align: 'center',
  },
  {
    field: 'projectName',
    title: '项目名称',
    minWidth: 200,
    showOverflow: 'tooltip',
  },
  {
    field: 'branch',
    title: '分支',
    width: 150,
    showOverflow: 'tooltip',
  },
  {
    field: 'status',
    title: '部署状态',
    width: 120,
    align: 'center',
    slots: { default: 'status' },
  },
  {
    field: 'startTime',
    title: '开始时间',
    width: 180,
    align: 'center',
    sortable: true,
  },
  {
    field: 'duration',
    title: '耗时',
    width: 120,
    align: 'center',
    sortable: true,
    formatter: ({ cellValue }) => formatDuration(cellValue),
  },
  {
    field: 'logPath',
    title: '日志路径',
    minWidth: 250,
    showOverflow: 'tooltip',
  },
  {
    field: 'createdAt',
    title: '创建时间',
    width: 180,
    align: 'center',
    sortable: true,
  },
  {
    title: '操作',
    width: 240,
    fixed: 'right',
    align: 'center',
    showOverflow: false,
    slots: { default: 'action' },
  },
]