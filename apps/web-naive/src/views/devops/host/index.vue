<script lang="ts" setup>
import type { VbenFormProps, VxeTableGridOptions } from '#/adapter/vxe-table';
import type { Host } from '#/api/host';

import { Page, useVbenModal } from '@vben/common-ui';

import { NButton, NSpace, useDialog, useMessage, NPopconfirm } from 'naive-ui';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { deleteHost, getHostList } from '#/api/host';

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
      label: '主机名称',
    },
    {
      component: 'Input',
      fieldName: 'host',
      label: 'IP地址',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: '在线', value: 'online' },
          { label: '离线', value: 'offline' },
          { label: '未激活', value: 'inactive' },
        ],
        placeholder: '请选择状态',
      },
      fieldName: 'status',
      label: '状态',
    },
  ],
  // 控制表单是否显示折叠按钮
  showCollapseButton: true,
  // 是否在字段值改变时提交表单
  submitOnChange: true,
  // 按下回车时是否提交表单
  submitOnEnter: false,
};

/**
 * 编辑主机
 */
function onEdit(row: Host) {
  formModalApi.setData(row).open();
}

/**
 * 创建新主机
 */
function onCreate() {
  formModalApi.setData(null).open();
}

/**
 * 删除主机
 */
async function onDelete(row: Host) {
  try {
    await deleteHost(row.id);
    message.success(`主机 "${row.name}" 删除成功`);
    refreshGrid();
  } catch (error) {
    console.error(`主机 "${row.name}" 删除失败:`, error);
    message.error(`主机 "${row.name}" 删除失败`);
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
          return await getHostList({
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
 * 刷新表格
 */
function refreshGrid() {
  gridApi.query();
}
</script>

<template>
  <Page auto-content-height>
    <FormModal @success="refreshGrid" />
    <Grid table-title="主机列表">
      <template #toolbar-tools>
        <NButton type="primary" @click="onCreate"> 添加主机 </NButton>
      </template>
      <template #action="{ row }">
        <NSpace :wrap="false">
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
              确定要删除主机 "{{ row.name }}" 吗？此操作不可撤销。
            </template>
          </NPopconfirm>
        </NSpace>
      </template>
    </Grid>
  </Page>
</template>
