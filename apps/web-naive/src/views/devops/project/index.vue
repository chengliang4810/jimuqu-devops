<script lang="ts" setup>
import type { VbenFormProps, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { Project } from '#/api/project';

import { Page, useVbenModal } from '@vben/common-ui';

import { NButton, NSpace, useDialog, useMessage, NPopconfirm, NTooltip } from 'naive-ui';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { deleteProject, getProjectList } from '#/api/project';

import { useColumns } from './data';
import Form from './modules/form.vue';

const [FormModal, formModalApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

// 初始化消息提示
const message = useMessage();

// 搜索表单配置
const formOptions: VbenFormProps = {
  // 默认展开
  collapsed: false,
  schema: [
    {
      component: 'Input',
      fieldName: 'name',
      label: '项目名称',
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '项目编码',
    },
  ],
  // 控制表单是否显示折叠按钮
  showCollapseButton: true,
  // 按下回车时是否提交表单
  submitOnEnter: true,
};

/**
 * 编辑项目
 */
function onEdit(row: Project) {
  formModalApi.setData(row).open();
}

/**
 * 创建新项目
 */
function onCreate() {
  formModalApi.setData(null).open();
}

/**
 * 删除项目
 */
async function onDelete(row: Project) {
  try {
    await deleteProject(row.id);
    message.success(`项目 "${row.name}" 删除成功`);
    refreshGrid();
  } catch (error) {
    console.error(`项目 "${row.name}" 删除失败:`, error);
    message.error(`项目 "${row.name}" 删除失败`);
  }
}

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions,
  gridEvents: {},
  gridOptions: {
    columns: useColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getProjectList({
            page: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    toolbarConfig: {
      custom: true,
      refresh: true,
      resizable: true,
      zoom: true,
    },
  } as VxeTableGridOptions,
});

/**
 * 复制webhook URL
 */
async function onCopyWebhookUrl(row: Project) {
  try {
    const baseUrl = window.location.origin;
    const webhookUrl = `${baseUrl}/api/webhook/${row.code}`;

    await navigator.clipboard.writeText(webhookUrl);
    message.success(`项目 "${row.name}" 的Webhook URL已复制到剪贴板`);
  } catch (error) {
    console.error('复制失败:', error);
    message.error('复制失败，请手动复制');
  }
}

/**
 * 刷新表格
 */
function refreshGrid() {
  gridApi.query();
}
</script>

<template>
  <Page auto-content-height>
    <FormModal @success="refreshGrid" />
    <Grid table-title="项目列表">
      <template #toolbar-tools>
        <NButton type="primary" @click="onCreate"> 添加项目 </NButton>
      </template>
      <template #action="{ row }">
        <NSpace :wrap="false">
          <NTooltip trigger="hover">
            <template #trigger>
              <NButton type="info" size="small" @click="onCopyWebhookUrl(row)">
                复制URL
              </NButton>
            </template>
            复制Webhook URL
          </NTooltip>
          <NButton type="warning" size="small" @click="onEdit(row)">
            编辑
          </NButton>
          <NPopconfirm
            :show-arrow="true"
            :show-icon="true"
            @positive-click="() => onDelete(row)"
          >
            <template #trigger>
              <NButton type="error" size="small">
                删除
              </NButton>
            </template>
            <template #default>
              确定要删除项目 "{{ row.name }}" 吗？此操作不可撤销。
            </template>
          </NPopconfirm>
        </NSpace>
      </template>
    </Grid>
  </Page>
</template>
