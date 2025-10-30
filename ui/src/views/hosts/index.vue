<script setup lang="tsx">
import { reactive, ref } from 'vue';
import { NButton, NPopconfirm, NTag, NIcon, NCard, NDataTable, NModal, NSpace } from 'naive-ui';
import { useAppStore } from '@/store/modules/app';
import { useMessage } from 'naive-ui';
import HostAddModal from './modules/host-add-modal.vue';

const appStore = useAppStore();
const message = useMessage();

// 静态模拟数据
const mockHosts = ref([
  {
    id: 1,
    name: '生产服务器-01',
    description: '主要生产环境服务器',
    host: '192.168.1.100',
    port: 22,
    username: 'deploy',
    ssh_key_path: '-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----',
    ssh_password: null,
    created_at: '2025-10-01T00:00:00Z',
    updated_at: '2025-10-30T11:30:00Z'
  },
  {
    id: 2,
    name: '测试服务器-01',
    description: '测试环境专用服务器',
    host: '192.168.1.101',
    port: 22,
    username: 'test',
    ssh_key_path: null,
    ssh_password: 'encrypted_password',
    created_at: '2025-10-05T00:00:00Z',
    updated_at: '2025-10-30T10:20:00Z'
  },
  {
    id: 3,
    name: '备份服务器-01',
    description: '数据备份和存储服务器',
    host: '192.168.1.200',
    port: 22,
    username: 'backup',
    ssh_key_path: '-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----',
    ssh_password: null,
    created_at: '2025-10-10T00:00:00Z',
    updated_at: '2025-10-29T18:30:00Z'
  },
  {
    id: 4,
    name: '开发服务器-01',
    description: '开发环境服务器',
    host: '192.168.1.50',
    port: 2222,
    username: 'dev',
    ssh_key_path: null,
    ssh_password: 'encrypted_dev_password',
    created_at: '2025-10-15T00:00:00Z',
    updated_at: '2025-10-28T16:45:00Z'
  },
  {
    id: 5,
    name: '监控服务器-01',
    description: '系统监控和日志收集',
    host: '192.168.1.150',
    port: 22,
    username: 'monitor',
    ssh_key_path: '-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----',
    ssh_password: null,
    created_at: '2025-10-12T00:00:00Z',
    updated_at: '2025-10-30T12:00:00Z'
  }
]);

const searchParams: any = reactive({
  current: 1,
  size: 10,
  name: ''
});

// 弹窗状态
const showModal = ref(false);
const editingHost = ref<any>(null);
const loading = ref(false);

// 过滤后的数据
const filteredHosts = ref(mockHosts.value);

// 过滤函数
function filterHosts() {
  let filtered = mockHosts.value;

  // 主机名称搜索过滤
  if (searchParams.name) {
    const searchLower = searchParams.name.toLowerCase();
    filtered = filtered.filter(host =>
      host.name.toLowerCase().includes(searchLower)
    );
  }

  return filtered;
}

// 获取分页数据
function getPaginatedData() {
  const filtered = filterHosts();
  const start = (searchParams.current - 1) * searchParams.size;
  const end = start + searchParams.size;
  return filtered.slice(start, end);
}

const data = ref(getPaginatedData());

const columns = [
  {
    key: 'index',
    title: '序号',
    align: 'center',
    width: 64,
    render: (_, index) => (searchParams.current - 1) * searchParams.size + index + 1
  },
  {
    key: 'name',
    title: '主机名称',
    align: 'center',
    minWidth: 150,
    render: row => (
      <div class="flex items-center justify-center gap-2">
        <NIcon size="16">
          <svg viewBox="0 0 24 24" class="text-primary">
            <path fill="currentColor" d="M4 1c-1.11 0-2 .89-2 2v4c0 1.11.89 2 2 2h1v1a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1V9h2v1a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1V9h2v1a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1V9h1c1.11 0 2-.89 2-2V3c0-1.11-.89-2-2-2H4m0 2h16v4H4V3m0 6h1v1H5V9m3 0h1v1H8V9m3 0h1v1h-1V9m3 0h1v1h-1V9m3 0h1v1h-1V9M5 12h14a1 1 0 0 1 1 1v2a1 1 0 0 1-1 1h-1v1a1 1 0 0 1-1 1h-2a1 1 0 0 1-1-1v-1H9v1a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1v-1H4a1 1 0 0 1-1-1v-2a1 1 0 0 1 1-1m0 1v2h14v-2H5z"/>
          </svg>
        </NIcon>
        <span class="font-medium">{row.name}</span>
      </div>
    )
  },
  {
    key: 'host',
    title: '连接信息',
    align: 'center',
    minWidth: 200,
    render: row => (
      <div class="flex flex-col">
        <span class="font-mono text-sm">{row.host}:{row.port}</span>
        <span class="text-xs text-gray-500">@{row.username}</span>
      </div>
    )
  },
  {
    key: 'ssh_auth',
    title: '认证方式',
    align: 'center',
    width: 120,
    render: row => (
      <NTag type={row.ssh_key_path ? 'success' : 'warning'} size="small">
        {row.ssh_key_path ? 'SSH密钥' : '密码'}
      </NTag>
    )
  },
  {
    key: 'description',
    title: '描述',
    align: 'left',
    minWidth: 200,
    render: row => row.description || '-'
  },
  {
    key: 'operate',
    title: '操作',
    align: 'center',
    width: 130,
    render: row => (
      <div class="flex-center gap-8px">
        <NButton
          type="primary"
          ghost
          size="small"
          onClick={() => edit(row.id)}
        >
          编辑
        </NButton>
        <NPopconfirm onPositiveClick={() => handleDelete(row.id)}>
          {{
            default: () => '确认删除吗？',
            trigger: () => (
              <NButton type="error" ghost size="small">
                删除
              </NButton>
            )
          }}
        </NPopconfirm>
      </div>
    )
  }
];

// 搜索功能
function getDataByPage() {
  loading.value = true;
  setTimeout(() => {
    data.value = getPaginatedData();
    loading.value = false;
  }, 300);
}

// 新增主机
function handleAdd() {
  editingHost.value = null;
  showModal.value = true;
}

// 编辑主机
function edit(id: number) {
  const host = mockHosts.value.find(item => item.id === id);
  if (host) {
    editingHost.value = host;
    showModal.value = true;
  }
}

// 提交表单
function handleSubmit(formData: any) {
  if (editingHost.value) {
    // 编辑模式
    Object.assign(editingHost.value, formData);
    message.success('主机信息已更新');
  } else {
    // 新增模式
    const newHost = {
      id: Date.now(),
      ...formData,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };
    mockHosts.value.push(newHost);
    message.success('主机添加成功');
  }
  getDataByPage();
}

// 删除主机
function handleDelete(id: number) {
  const index = mockHosts.value.findIndex(host => host.id === id);
  if (index > -1) {
    mockHosts.value.splice(index, 1);
    getDataByPage();
    message.success('主机已删除');
  }
}

// 刷新数据
function getData() {
  getDataByPage();
}

// 移动端分页
const mobilePagination = {
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 20, 50]
};
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 搜索区域 -->
    <NCard :bordered="false" size="small" class="card-wrapper">
      <div class="p-4">
        <div class="flex items-center gap-4">
          <div class="flex-1 max-w-md">
            <NInput
              v-model:value="searchParams.name"
              placeholder="请输入主机名称"
              clearable
              @input="getDataByPage"
            >
              <template #prefix>
                <NIcon>
                  <svg viewBox="0 0 24 24">
                    <path fill="currentColor" d="M9.5,3A6.5,6.5 0 0,1 16,9.5C16,11.11 15.41,12.59 14.44,13.73L14.71,14H15.5L20.5,19L19,20.5L14,15.5V14.71L13.73,14.44C12.59,15.41 11.11,16 9.5,16A6.5,6.5 0 0,1 3,9.5A6.5,6.5 0 0,1 9.5,3M9.5,5C7,5 5,7 5,9.5C5,12 7,14 9.5,14C12,14 14,12 14,9.5C14,7 12,5 9.5,5Z"/>
                  </svg>
                </NIcon>
              </template>
            </NInput>
          </div>
          <NButton @click="getDataByPage">
            <template #icon>
              <icon-ic-round-search class="text-icon" />
            </template>
            搜索
          </NButton>
        </div>
      </div>
    </NCard>

    <!-- 主机管理卡片 -->
    <NCard :bordered="false" size="small" class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex justify-between items-center w-full">
          <span class="text-lg font-medium">主机管理</span>
          <div class="flex items-center gap-3">
            <NButton @click="getData">
              <template #icon>
                <icon-ic-round-refresh class="text-icon" />
              </template>
              刷新
            </NButton>
            <NButton type="primary" @click="handleAdd">
              <template #icon>
                <icon-ic-round-add class="text-icon" />
              </template>
              新增主机
            </NButton>
          </div>
        </div>
      </template>

      <!-- 数据表格 -->
      <NDataTable
        :columns="columns"
        :data="data"
        size="small"
        :flex-height="!appStore.isMobile"
        :scroll-x="1000"
        :loading="loading"
        remote
        :row-key="row => row.id"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
    </NCard>

    <!-- 新增/编辑主机弹窗 -->
    <NModal v-model:show="showModal" :mask-closable="false" preset="card" style="width: 600px;" :title="editingHost ? '编辑主机' : '新增主机'">
      <HostAddModal
        v-if="showModal"
        :host-data="editingHost"
        @close="showModal = false"
        @submit="handleSubmit"
      />
    </NModal>
  </div>
</template>

<style scoped></style>