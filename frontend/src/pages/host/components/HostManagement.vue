<template>
  <div class="host-management">
    <div class="host-header">
      <div class="header-left">
        <h2>主机列表</h2>
        <span class="group-info" v-if="currentGroup">
          {{ currentGroup.name }} ({{ hostList.length }} 台主机)
        </span>
      </div>
      <div class="header-right">
        <t-button @click="showCreateDialog = true">
          <template #icon>
            <add-icon />
          </template>
          新建主机
        </t-button>
        <t-button variant="outline" @click="refreshList">
          <template #icon>
            <refresh-icon />
          </template>
          刷新
        </t-button>
      </div>
    </div>

    <div class="host-content">
      <t-table
        :data="hostList"
        :columns="columns"
        :loading="loading"
        row-key="id"
        :selected-row-keys="selectedHostIds"
        @select-change="handleSelectChange"
        stripe
        hover
        size="medium"
      >
        <template #status="{ row }">
          <t-tag
            :theme="row.status === 'online' ? 'success' : row.status === 'offline' ? 'danger' : 'warning'"
            variant="light"
          >
            {{ getStatusText(row.status) }}
          </t-tag>
        </template>

        <template #authType="{ row }">
          <t-tag variant="light">
            {{ row.authType === 'password' ? '密码' : '密钥' }}
          </t-tag>
        </template>

        <template #operation="{ row }">
          <t-space>
            <t-button variant="text" size="small" @click="testConnection(row)">
              <template #icon>
                <desktop-icon />
              </template>
              连接测试
            </t-button>
            <t-button variant="text" size="small" @click="editHost(row)">
              <template #icon>
                <edit-icon />
              </template>
              编辑
            </t-button>
            <t-popconfirm
              content="确定要删除该主机吗？"
              @confirm="deleteHost(row.id)"
            >
              <t-button variant="text" size="small" theme="danger">
                <template #icon>
                  <delete-icon />
                </template>
                删除
              </t-button>
            </t-popconfirm>
          </t-space>
        </template>

        <template #empty>
          <t-empty description="暂无主机数据">
            <t-button theme="primary" @click="showCreateDialog = true">
              新建主机
            </t-button>
          </t-empty>
        </template>
      </t-table>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="total > 0">
        <t-pagination
          v-model="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-size-options="[10, 20, 50, 100]"
          @change="handlePageChange"
          show-jumper
          show-sizer
        />
      </div>
    </div>

    <!-- 新建/编辑主机对话框 -->
    <t-dialog
      v-model:visible="showCreateDialog"
      :header="editingHost ? '编辑主机' : '新建主机'"
      width="600px"
      @confirm="handleHostSubmit"
    >
      <t-form
        ref="hostFormRef"
        :data="hostForm"
        :rules="hostRules"
        label-width="100px"
      >
        <t-form-item label="所属分组" name="groupId">
          <t-select
            v-model="hostForm.groupId"
            placeholder="请选择分组"
            :disabled="!!editingHost"
          >
            <t-option
              v-for="group in groupList"
              :key="group.id"
              :value="group.id"
              :label="group.name"
            />
          </t-select>
        </t-form-item>

        <t-form-item label="主机名称" name="name">
          <t-input
            v-model="hostForm.name"
            placeholder="请输入主机名称"
            clearable
          />
        </t-form-item>

        <t-form-item label="IP地址" name="ip">
          <t-input
            v-model="hostForm.ip"
            placeholder="请输入IP地址"
            clearable
          />
        </t-form-item>

        <t-form-item label="端口" name="port">
          <t-input-number
            v-model="hostForm.port"
            :min="1"
            :max="65535"
            placeholder="请输入端口号"
          />
        </t-form-item>

        <t-form-item label="用户名" name="username">
          <t-input
            v-model="hostForm.username"
            placeholder="请输入用户名"
            clearable
          />
        </t-form-item>

        <t-form-item label="认证方式" name="authType">
          <t-radio-group v-model="hostForm.authType">
            <t-radio value="password">密码认证</t-radio>
            <t-radio value="key">密钥认证</t-radio>
          </t-radio-group>
        </t-form-item>

        <t-form-item
          label="密码"
          name="password"
          v-if="hostForm.authType === 'password'"
        >
          <t-input
            v-model="hostForm.password"
            type="password"
            placeholder="请输入密码"
            clearable
          />
        </t-form-item>

        <t-form-item
          label="私钥"
          name="privateKey"
          v-if="hostForm.authType === 'key'"
        >
          <t-textarea
            v-model="hostForm.privateKey"
            placeholder="请输入私钥内容"
            :autosize="{ minRows: 4, maxRows: 8 }"
          />
        </t-form-item>

        <t-form-item label="备注" name="description">
          <t-textarea
            v-model="hostForm.description"
            placeholder="请输入备注信息（可选）"
            :autosize="{ minRows: 2, maxRows: 4 }"
            :maxlength="200"
          />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 连接测试结果对话框 -->
    <t-dialog
      v-model:visible="showTestDialog"
      header="连接测试结果"
      width="400px"
      :footer="false"
    >
      <div class="test-result">
        <t-result
          :theme="testResult.status === 'online' ? 'success' : 'error'"
          :title="testResult.status === 'online' ? '连接成功' : '连接失败'"
          :description="testResult.message"
        />
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import {
  AddIcon,
  EditIcon,
  DeleteIcon,
  RefreshIcon,
  DesktopIcon,
} from 'tdesign-icons-vue-next';

import { hostApi } from '@/api/host';
import type { Host, HostForm, HostGroup } from '@/types/host';

const props = defineProps<{
  selectedGroupId: string;
  groupList: HostGroup[];
}>();

// 响应式数据
const hostList = ref<Host[]>([]);
const loading = ref(false);
const selectedHostIds = ref<string[]>([]);
const showCreateDialog = ref(false);
const showTestDialog = ref(false);
const editingHost = ref<Host | null>(null);
const hostFormRef = ref();

// 分页数据
const pagination = reactive({
  page: 1,
  pageSize: 20,
});
const total = ref(0);

// 连接测试结果
const testResult = reactive({
  status: 'offline' as 'online' | 'offline',
  message: '',
});

// 当前选中的分组
const currentGroup = computed(() => {
  return props.groupList.find(group => group.id === props.selectedGroupId);
});

// 表单数据
const hostForm = reactive<HostForm>({
  groupId: props.selectedGroupId,
  name: '',
  ip: '',
  port: 22,
  username: '',
  password: '',
  authType: 'password',
  privateKey: '',
  description: '',
});

// 表单验证规则
const hostRules = {
  groupId: [{ required: true, message: '请选择分组' }],
  name: [
    { required: true, message: '请输入主机名称' },
    { min: 1, max: 50, message: '主机名称长度为1-50个字符' },
  ],
  ip: [
    { required: true, message: '请输入IP地址' },
    {
      pattern: /^(\d{1,3}\.){3}\d{1,3}$/,
      message: '请输入有效的IP地址',
    },
  ],
  port: [
    { required: true, message: '请输入端口号' },
    { type: 'number', min: 1, max: 65535, message: '端口范围为1-65535' },
  ],
  username: [
    { required: true, message: '请输入用户名' },
    { min: 1, max: 50, message: '用户名长度为1-50个字符' },
  ],
  password: [
    {
      required: true,
      message: '请输入密码',
      validator: (val: string) => {
        if (hostForm.authType === 'password' && !val) {
          return false;
        }
        return true;
      },
    },
  ],
  privateKey: [
    {
      required: true,
      message: '请输入私钥',
      validator: (val: string) => {
        if (hostForm.authType === 'key' && !val) {
          return false;
        }
        return true;
      },
    },
  ],
};

// 表格列定义
const columns = [
  {
    colKey: 'select',
    type: 'multiple',
    width: 50,
  },
  {
    colKey: 'name',
    title: '主机名称',
    ellipsis: true,
  },
  {
    colKey: 'ip',
    title: 'IP地址',
    width: 140,
  },
  {
    colKey: 'port',
    title: '端口',
    width: 80,
  },
  {
    colKey: 'username',
    title: '用户名',
    width: 100,
  },
  {
    colKey: 'authType',
    title: '认证方式',
    width: 100,
    cell: { slot: 'authType' },
  },
  {
    colKey: 'status',
    title: '状态',
    width: 100,
    cell: { slot: 'status' },
  },
  {
    colKey: 'description',
    title: '备注',
    ellipsis: true,
  },
  {
    colKey: 'operation',
    title: '操作',
    width: 200,
    cell: { slot: 'operation' },
  },
];

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap = {
    online: '在线',
    offline: '离线',
    unknown: '未知',
  };
  return statusMap[status] || '未知';
};

// 获取主机列表
const getHostList = async () => {
  loading.value = true;
  try {
    const res = await hostApi.getList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      groupId: props.selectedGroupId,
    });
    if (res.success) {
      hostList.value = res.data.items;
      total.value = res.data.total;
    }
  } catch (error) {
    console.error('获取主机列表失败:', error);
    MessagePlugin.error('获取主机列表失败');
  } finally {
    loading.value = false;
  }
};

// 刷新列表
const refreshList = () => {
  getHostList();
};

// 选择变化处理
const handleSelectChange = (selectedRowKeys: string[]) => {
  selectedHostIds.value = selectedRowKeys;
};

// 分页变化处理
const handlePageChange = (pageInfo: any) => {
  pagination.page = pageInfo.page;
  pagination.pageSize = pageInfo.pageSize;
  getHostList();
};

// 编辑主机
const editHost = (host: Host) => {
  editingHost.value = host;
  Object.assign(hostForm, {
    groupId: host.groupId,
    name: host.name,
    ip: host.ip,
    port: host.port,
    username: host.username,
    password: host.password,
    authType: host.authType,
    privateKey: host.privateKey,
    description: host.description,
  });
  showCreateDialog.value = true;
};

// 删除主机
const deleteHost = async (id: string) => {
  try {
    const res = await hostApi.delete(id);
    if (res.success) {
      MessagePlugin.success('删除成功');
      await getHostList();
    }
  } catch (error) {
    console.error('删除主机失败:', error);
    MessagePlugin.error('删除失败');
  }
};

// 测试连接
const testConnection = async (host: Host) => {
  try {
    const res = await hostApi.testConnection(host.id);
    if (res.success) {
      testResult.status = res.data.status;
      testResult.message = res.data.message;
      showTestDialog.value = true;
    }
  } catch (error) {
    console.error('连接测试失败:', error);
    testResult.status = 'offline';
    testResult.message = '连接测试失败';
    showTestDialog.value = true;
  }
};

// 处理主机提交
const handleHostSubmit = async () => {
  const valid = await hostFormRef.value?.validate();
  if (!valid) return;

  try {
    if (editingHost.value) {
      // 更新主机
      const res = await hostApi.update(editingHost.value.id, hostForm);
      if (res.success) {
        MessagePlugin.success('更新成功');
      }
    } else {
      // 创建主机
      const res = await hostApi.create(hostForm);
      if (res.success) {
        MessagePlugin.success('创建成功');
      }
    }

    await getHostList();
    showCreateDialog.value = false;
    resetForm();
  } catch (error) {
    console.error('操作失败:', error);
    MessagePlugin.error('操作失败');
  }
};

// 重置表单
const resetForm = () => {
  hostForm.groupId = props.selectedGroupId;
  hostForm.name = '';
  hostForm.ip = '';
  hostForm.port = 22;
  hostForm.username = '';
  hostForm.password = '';
  hostForm.authType = 'password';
  hostForm.privateKey = '';
  hostForm.description = '';
  editingHost.value = null;
};

// 监听分组选择变化
watch(
  () => props.selectedGroupId,
  (newGroupId) => {
    hostForm.groupId = newGroupId;
    pagination.page = 1;
    getHostList();
  },
  { immediate: true }
);

// 组件挂载时获取数据
onMounted(() => {
  getHostList();
});
</script>

<style lang="less" scoped>
.host-management {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--td-bg-color-container);
}

.host-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--td-component-border);

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;

    h2 {
      margin: 0;
      font: var(--td-font-title-medium);
      color: var(--td-text-color-primary);
    }

    .group-info {
      font: var(--td-font-body-small);
      color: var(--td-text-color-secondary);
    }
  }

  .header-right {
    display: flex;
    gap: 12px;
  }
}

.host-content {
  flex: 1;
  padding: 24px;
  overflow: auto;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

.test-result {
  text-align: center;
  padding: 20px 0;
}
</style>