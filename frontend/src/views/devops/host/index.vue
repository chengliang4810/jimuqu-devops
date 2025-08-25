<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 页面头部 -->
    <div class="flex-y-center justify-between">
      <h1 class="text-18px font-bold">主机管理</h1>
      <div class="flex-y-center gap-12px">
        <NButton type="info" @click="refreshStatus">
          <Icon icon="material-symbols:refresh" class="mr-4px text-16px" />
          刷新状态
        </NButton>
        <NButton type="primary" @click="handleAdd">
          <Icon icon="material-symbols:add" class="mr-4px text-16px" />
          添加主机
        </NButton>
      </div>
    </div>

    <!-- 搜索栏 -->
    <div class="card-wrapper">
      <NForm inline label-width="auto" label-placement="left">
        <NFormItem label="搜索">
          <NInput
            v-model:value="searchKeyword"
            placeholder="请输入主机名称或IP地址"
            clearable
            @keyup.enter="getTableData"
          >
            <template #suffix>
              <Icon icon="material-symbols:search" class="text-16px cursor-pointer" @click="getTableData" />
            </template>
          </NInput>
        </NFormItem>
        <NFormItem>
          <NButton type="primary" @click="getTableData">查询</NButton>
          <NButton class="ml-8px" @click="reset">重置</NButton>
        </NFormItem>
      </NForm>
    </div>

    <!-- 数据表格 -->
    <div class="card-wrapper flex-1-hidden">
      <NDataTable
        :columns="columns"
        :data="tableData"
        :loading="loading"
        :pagination="mobilePagination"
        :scroll-x="800"
        virtual-scroll
        class="sm:h-full"
      />
    </div>

    <!-- 添加/编辑主机弹窗 -->
    <NModal v-model:show="showModal" title="主机信息" preset="card" class="w-700px">
      <NForm ref="formRef" :model="formModel" :rules="rules" label-width="auto" label-placement="left">
        <div class="grid grid-cols-1 gap-16px sm:grid-cols-2">
          <NFormItem label="主机名称" path="name">
            <NInput v-model:value="formModel.name" placeholder="请输入主机名称" />
          </NFormItem>
          <NFormItem label="主机IP" path="hostIp">
            <NInput v-model:value="formModel.hostIp" placeholder="请输入主机IP地址" />
          </NFormItem>
          <NFormItem label="SSH端口" path="port">
            <NInputNumber v-model:value="formModel.port" placeholder="SSH端口，默认22" :min="1" :max="65535" />
          </NFormItem>
          <NFormItem label="用户名" path="username">
            <NInput v-model:value="formModel.username" placeholder="请输入SSH用户名" />
          </NFormItem>
          <NFormItem label="密码" path="password">
            <NInput
              v-model:value="formModel.password"
              type="password"
              placeholder="请输入SSH密码"
              show-password-on="click"
            />
          </NFormItem>
          <NFormItem label="状态">
            <NTag :type="getStatusType(formModel.status)">
              {{ getStatusText(formModel.status) }}
            </NTag>
          </NFormItem>
        </div>
        <NFormItem label="描述" path="description">
          <NInput
            v-model:value="formModel.description"
            type="textarea"
            placeholder="请输入主机描述信息"
            :rows="3"
          />
        </NFormItem>
        <NFormItem label="SSH私钥" path="privateKey">
          <NInput
            v-model:value="formModel.privateKey"
            type="textarea"
            placeholder="可选：SSH私钥内容（如果使用密钥认证）"
            :rows="4"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-12px">
          <NButton @click="closeModal">取消</NButton>
          <NButton v-if="formModel.id" type="warning" @click="testConnection">测试连接</NButton>
          <NButton type="primary" @click="handleSave">保存</NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue';
import type { Ref } from 'vue';
import { NButton, NDataTable, NForm, NFormItem, NInput, NInputNumber, NModal, NTag, useDialog, useMessage } from 'naive-ui';
import type { DataTableColumns, FormInst } from 'naive-ui';
import { Icon } from '@iconify/vue';

interface Host {
  id?: number;
  name: string;
  hostIp: string;
  port: number;
  username: string;
  password: string;
  privateKey?: string;
  status: 'ONLINE' | 'OFFLINE' | 'ERROR';
  description?: string;
  createTime?: string;
  updateTime?: string;
}

const dialog = useDialog();
const message = useMessage();

// 表格数据
const tableData: Ref<Host[]> = ref([]);
const loading = ref(false);
const searchKeyword = ref('');

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  onChange: (page: number) => {
    pagination.page = page;
    getTableData();
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.page = 1;
    getTableData();
  }
});

const mobilePagination = computed(() => ({
  ...pagination,
  showQuickJumper: false,
  showSizePicker: false
}));

// 表格列定义
const columns: DataTableColumns<Host> = [
  {
    title: '主机名称',
    key: 'name',
    width: 150
  },
  {
    title: '主机IP',
    key: 'hostIp',
    width: 140
  },
  {
    title: '端口',
    key: 'port',
    width: 80
  },
  {
    title: '用户名',
    key: 'username',
    width: 100
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      return h(NTag, { type: getStatusType(row.status) }, () => getStatusText(row.status));
    }
  },
  {
    title: '描述',
    key: 'description',
    ellipsis: true
  },
  {
    title: '创建时间',
    key: 'createTime',
    width: 180,
    render: (row) => {
      return row.createTime ? new Date(row.createTime).toLocaleString() : '-';
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render: (row) => {
      return h('div', { class: 'flex gap-8px' }, [
        h(
          NButton,
          {
            size: 'small',
            type: 'info',
            onClick: () => handleTestConnection(row)
          },
          () => '测试连接'
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            onClick: () => handleEdit(row)
          },
          () => '编辑'
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'error',
            onClick: () => handleDelete(row)
          },
          () => '删除'
        )
      ]);
    }
  }
];

// 表单相关
const showModal = ref(false);
const formRef = ref<FormInst | null>(null);
const formModel = reactive<Host>({
  name: '',
  hostIp: '',
  port: 22,
  username: '',
  password: '',
  privateKey: '',
  status: 'OFFLINE',
  description: ''
});

// 表单验证规则
const rules = {
  name: { required: true, message: '请输入主机名称', trigger: 'blur' },
  hostIp: { required: true, message: '请输入主机IP地址', trigger: 'blur' },
  port: { required: true, type: 'number', message: '请输入有效的端口号', trigger: 'blur' },
  username: { required: true, message: '请输入SSH用户名', trigger: 'blur' },
  password: { required: true, message: '请输入SSH密码', trigger: 'blur' }
};

// 状态相关方法
const getStatusType = (status: string) => {
  const typeMap = {
    ONLINE: 'success',
    OFFLINE: 'default',
    ERROR: 'error'
  };
  return typeMap[status] || 'default';
};

const getStatusText = (status: string) => {
  const textMap = {
    ONLINE: '在线',
    OFFLINE: '离线',
    ERROR: '错误'
  };
  return textMap[status] || '未知';
};

// API调用方法
const getTableData = async () => {
  loading.value = true;
  try {
    // 调用后端API获取主机列表
    const params = new URLSearchParams({
      page: pagination.page.toString(),
      size: pagination.pageSize.toString()
    });
    if (searchKeyword.value) {
      params.append('keyword', searchKeyword.value);
    }

    const response = await fetch(`/api/hosts?${params}`);
    const result = await response.json();

    if (result.code === 200) {
      tableData.value = result.data.records;
      pagination.itemCount = result.data.total;
    } else {
      message.error(result.message || '获取主机列表失败');
    }
  } catch (error) {
    message.error('网络请求失败');
    console.error('获取主机列表失败:', error);
  } finally {
    loading.value = false;
  }
};

// 页面操作方法
const handleAdd = () => {
  Object.assign(formModel, {
    id: undefined,
    name: '',
    hostIp: '',
    port: 22,
    username: '',
    password: '',
    privateKey: '',
    status: 'OFFLINE',
    description: ''
  });
  showModal.value = true;
};

const handleEdit = (row: Host) => {
  Object.assign(formModel, { ...row });
  showModal.value = true;
};

const handleDelete = (row: Host) => {
  dialog.warning({
    title: '删除确认',
    content: `确定要删除主机 "${row.name}" 吗？此操作不可恢复。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const response = await fetch(`/api/hosts/${row.id}`, {
          method: 'DELETE'
        });
        const result = await response.json();

        if (result.code === 200) {
          message.success('删除成功');
          getTableData();
        } else {
          message.error(result.message || '删除失败');
        }
      } catch (error) {
        message.error('删除失败');
        console.error('删除主机失败:', error);
      }
    }
  });
};

const handleSave = async () => {
  try {
    await formRef.value?.validate();
    
    const method = formModel.id ? 'PUT' : 'POST';
    const url = formModel.id ? `/api/hosts/${formModel.id}` : '/api/hosts';
    
    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(formModel)
    });
    
    const result = await response.json();
    
    if (result.code === 200) {
      message.success(formModel.id ? '更新成功' : '创建成功');
      closeModal();
      getTableData();
    } else {
      message.error(result.message || '保存失败');
    }
  } catch (error) {
    message.error('保存失败');
    console.error('保存主机失败:', error);
  }
};

const handleTestConnection = async (row: Host) => {
  const loadingMessage = message.loading('正在测试连接...', { duration: 0 });
  try {
    const response = await fetch(`/api/hosts/${row.id}/test`, {
      method: 'POST'
    });
    const result = await response.json();
    
    loadingMessage.destroy();
    
    if (result.code === 200) {
      if (result.data) {
        message.success('连接测试成功');
      } else {
        message.error('连接测试失败');
      }
      getTableData(); // 刷新状态
    } else {
      message.error(result.message || '连接测试失败');
    }
  } catch (error) {
    loadingMessage.destroy();
    message.error('连接测试失败');
    console.error('测试连接失败:', error);
  }
};

const testConnection = () => {
  if (formModel.id) {
    handleTestConnection(formModel);
  } else {
    message.warning('请先保存主机信息后再测试连接');
  }
};

const refreshStatus = async () => {
  const loadingMessage = message.loading('正在刷新主机状态...', { duration: 0 });
  try {
    const response = await fetch('/api/hosts/status/update', {
      method: 'POST'
    });
    const result = await response.json();
    
    loadingMessage.destroy();
    
    if (result.code === 200) {
      message.success('状态刷新已开始，请稍后查看结果');
      setTimeout(getTableData, 3000); // 3秒后刷新列表
    } else {
      message.error(result.message || '刷新状态失败');
    }
  } catch (error) {
    loadingMessage.destroy();
    message.error('刷新状态失败');
    console.error('刷新状态失败:', error);
  }
};

const closeModal = () => {
  showModal.value = false;
};

const reset = () => {
  searchKeyword.value = '';
  getTableData();
};

// 生命周期
onMounted(() => {
  getTableData();
});
</script>