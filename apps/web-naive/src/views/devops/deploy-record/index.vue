<script setup lang="ts">
import type { VbenFormProps, VxeTableGridOptions } from '#/adapter/vxe-table'
import type { DeployRecord } from '#/api/deploy-record'
import type { DeployRecordQuery } from '#/api/deploy-record'

import { ref, onMounted } from 'vue'
import { Page } from '@vben/common-ui'
import { useVbenVxeGrid } from '#/adapter/vxe-table'
import { getDeployRecords, deleteDeployRecord } from '#/api/deploy-record'
import { useDialog, useMessage, NButton, NEmpty, NFlex } from 'naive-ui'
import { $t } from '@vben/locales'
import { gridColumns, renderStatus, renderAction, formatDuration } from './data'
import DeployRecordDetail from './modules/detail.vue'

// 定义查询参数类型
interface QueryForm extends DeployRecordQuery {
  startTimeRange?: [string, string] | null
}

const showDetailModal = ref(false)
const currentRecord = ref<DeployRecord | null>(null)

// 初始化 Naive UI hooks
const message = useMessage()
const dialog = useDialog()

// 搜索表单配置
const formOptions: VbenFormProps = {
  // 默认展开
  collapsed: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'projectName',
      name: 'projectName',
      label: '项目名称',
    },
    {
      component: 'Input',
      fieldName: 'branch',
      name: 'branch',
      label: '分支',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        placeholder: '请选择状态',
        options: [
          { label: '运行中', value: 'running' },
          { label: '成功', value: 'success' },
          { label: '失败', value: 'failed' },
        ],
      },
      fieldName: 'status',
      name: 'status',
      label: '状态',
    },
    {
      component: 'DatePicker',
      componentProps: {
        type: 'datetimerange',
        allowClear: true,
        format: 'yyyy-MM-dd HH:mm:ss',
        valueFormat: 'yyyy-MM-dd HH:mm:ss',
      },
      fieldName: 'startTimeRange',
      name: 'startTimeRange',
      label: '开始时间',
    },
  ],
  // 控制表单是否显示折叠按钮
  showCollapseButton: true,
  // 按下回车时是否提交表单
  submitOnEnter: true,
}

// VxeGrid配置
const [Grid, gridApi] = useVbenVxeGrid({
  columns: gridColumns,
  formOptions,
  gridOptions: {
    columns: gridColumns,
    height: 'auto',
    keepSource: true,
    rowConfig: {
      keyField: 'id',
    },
    pagerConfig: {
      enabled: true,
      pageSize: 10,
      pageSizes: [10, 20, 50, 100],
    },
    toolbarConfig: {
      refresh: {
        code: 'refresh',
      },
      zoom: true,
      custom: true,
      resizable: true,
    },
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          // 构建查询参数
          const queryParams: DeployRecordQuery = {
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            projectName: formValues.projectName || undefined,
            branch: formValues.branch || undefined,
            status: formValues.status || undefined,
          }

          // 处理时间范围
          if (formValues.startTimeRange && Array.isArray(formValues.startTimeRange) && formValues.startTimeRange.length === 2) {
            queryParams.startTimeStart = formValues.startTimeRange[0]
            queryParams.startTimeEnd = formValues.startTimeRange[1]
          }

          try {
            const response = await getDeployRecords(queryParams)
            return {
              list: response.rows || [],
              total: response.total || 0,
            }
          } catch (error) {
            message.error('获取部署记录列表失败')
            return {
              list: [],
              total: 0,
            }
          }
        },
      },
    },
  } as VxeTableGridOptions,
})

// 刷新数据
function refreshGrid() {
  gridApi?.query()
}

// 查看详情
function handleView(record: DeployRecord) {
  currentRecord.value = record
  showDetailModal.value = true
}

// 查看日志
function handleViewLog(record: DeployRecord) {
  if (record.logPath) {
    // 这里可以实现日志查看功能，比如打开新窗口或弹窗显示日志内容
    message.info(`日志路径: ${record.logPath}`)
  } else {
    message.warning('该部署记录没有日志文件')
  }
}

// 删除单条记录
function handleDelete(record: DeployRecord) {
  dialog.warning({
    title: $t('common.deleteTitle'),
    content: `确定要删除部署记录 "${record.projectName} - ${record.branch}" 吗？此操作不可恢复。`,
    positiveText: $t('common.confirm'),
    negativeText: $t('common.cancel'),
    onPositiveClick: async () => {
      try {
        await deleteDeployRecord(record.id)
        message.success('删除成功')
        refreshGrid()
      } catch (error) {
        message.error('删除失败')
      }
    },
  })
}

// 批量删除
function handleBatchDelete() {
  const selectedRecords = gridApi?.getCheckboxRecords() || []
  if (selectedRecords.length === 0) {
    message.warning('请选择要删除的记录')
    return
  }

  dialog.warning({
    title: $t('common.deleteTitle'),
    content: `确定要删除选中的 ${selectedRecords.length} 条部署记录吗？此操作不可恢复。`,
    positiveText: $t('common.confirm'),
    negativeText: $t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const deletePromises = selectedRecords.map((record: DeployRecord) => deleteDeployRecord(record.id))
        await Promise.all(deletePromises)
        message.success(`成功删除 ${selectedRecords.length} 条记录`)
        refreshGrid()
      } catch (error) {
        message.error('批量删除失败')
      }
    },
  })
}

// 导出数据（预留功能）
function handleExport() {
  message.info('导出功能开发中...')
}

// 页面初始化
onMounted(() => {
  refreshGrid()
})
</script>

<template>
  <Page auto-content-height>
    <Grid>
      <!-- 工具栏左侧操作按钮 -->
      <template #toolbar-left>
        <NButton
          type="error"
          size="small"
          @click="handleBatchDelete"
        >
          批量删除
        </NButton>
        <NButton
          type="default"
          size="small"
          @click="handleExport"
        >
          导出数据
        </NButton>
      </template>

      <!-- 状态列渲染 -->
      <template #status="{ row }">
        <component :is="renderStatus(row.status)" />
      </template>

      <!-- 操作列渲染 -->
      <template #action="{ row }">
        <component :is="renderAction(row, handleView, handleViewLog)" />
      </template>

      <!-- 空状态 -->
      <template #empty>
        <NEmpty description="暂无部署记录" />
      </template>
    </Grid>

    <!-- 详情弹窗 -->
    <DeployRecordDetail
      v-model:show="showDetailModal"
      :record="currentRecord"
      @success="refreshGrid"
    />
  </Page>
</template>
