<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 页面头部 -->
    <div class="flex-y-center justify-between">
      <h1 class="text-18px font-bold">应用管理</h1>
      <NButton type="primary" @click="handleAdd">
        <Icon icon="material-symbols:add" class="mr-4px text-16px" />
        添加应用
      </NButton>
    </div>

    <!-- 搜索栏 -->
    <div class="card-wrapper">
      <NForm inline label-width="auto" label-placement="left">
        <NFormItem label="搜索">
          <NInput
            v-model:value="searchKeyword"
            placeholder="请输入应用名称或Git仓库地址"
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
        :scroll-x="1000"
        virtual-scroll
        class="sm:h-full"
      />
    </div>

    <!-- 添加/编辑应用弹窗 -->
    <NModal v-model:show="showModal" title="应用信息" preset="card" class="w-800px">
      <NForm ref="formRef" :model="formModel" :rules="rules" label-width="auto" label-placement="left">
        <div class="grid grid-cols-1 gap-16px sm:grid-cols-2">
          <NFormItem label="应用名称" path="name">
            <NInput v-model:value="formModel.name" placeholder="请输入应用名称" />
          </NFormItem>
          <NFormItem label="Git仓库地址" path="gitRepoUrl">
            <NInput v-model:value="formModel.gitRepoUrl" placeholder="https://gitee.com/user/repo.git" />
          </NFormItem>
          <NFormItem label="Git分支" path="gitBranch">
            <NInput v-model:value="formModel.gitBranch" placeholder="main" />
          </NFormItem>
          <NFormItem label="Git用户名">
            <NInput v-model:value="formModel.gitUsername" placeholder="Git用户名（可选）" />
          </NFormItem>
          <NFormItem label="Git密码">
            <NInput
              v-model:value="formModel.gitPassword"
              type="password"
              placeholder="Git密码（可选）"
              show-password-on="click"
            />
          </NFormItem>
          <NFormItem label="自动触发">
            <NSwitch v-model:value="formModel.autoTrigger" />
          </NFormItem>
        </div>
        <NFormItem label="通知地址">
          <NInput v-model:value="formModel.notificationUrl" placeholder="构建完成后的通知回调地址" />
        </NFormItem>
        <NFormItem label="通知令牌">
          <NInput v-model:value="formModel.notificationToken" placeholder="通知验证令牌" />
        </NFormItem>
        <NFormItem label="Webhook密钥">
          <NInput v-model:value="formModel.webhookSecret" placeholder="用于验证Webhook请求的密钥" />
        </NFormItem>
        <NFormItem label="Git私钥">
          <NInput
            v-model:value="formModel.gitPrivateKey"
            type="textarea"
            placeholder="Git私钥内容（如果使用SSH方式）"
            :rows="4"
          />
        </NFormItem>
        <NFormItem label="环境变量">
          <div class="w-full">
            <div v-for="(item, index) in envVariables" :key="index" class="flex gap-8px mb-8px">
              <NInput v-model:value="item.key" placeholder="变量名" class="flex-1" />
              <NInput v-model:value="item.value" placeholder="变量值" class="flex-1" />
              <NButton type="error" size="small" @click="removeEnvVariable(index)">
                <Icon icon="material-symbols:delete" />
              </NButton>
            </div>
            <NButton type="dashed" @click="addEnvVariable" class="w-full">
              <Icon icon="material-symbols:add" class="mr-4px" />
              添加环境变量
            </NButton>
          </div>
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-12px">
          <NButton @click="closeModal">取消</NButton>
          <NButton type="primary" @click="handleSave">保存</NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue';
import type { Ref } from 'vue';
import { 
  NButton, 
  NDataTable, 
  NForm, 
  NFormItem, 
  NInput, 
  NModal, 
  NTag, 
  NSwitch,
  useDialog, 
  useMessage 
} from 'naive-ui';
import type { DataTableColumns, FormInst } from 'naive-ui';
import { Icon } from '@iconify/vue';

interface Application {
  id?: number;
  name: string;
  gitRepoUrl: string;
  gitBranch: string;
  gitUsername?: string;
  gitPassword?: string;
  gitPrivateKey?: string;
  notificationUrl?: string;
  notificationToken?: string;
  autoTrigger: boolean;
  webhookSecret?: string;
  status: string;
  createTime?: string;
  updateTime?: string;
  variables?: Record<string, string>;
}

interface EnvVariable {
  key: string;
  value: string;
}

const dialog = useDialog();
const message = useMessage();

// 表格数据
const tableData: Ref<Application[]> = ref([]);
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
const columns: DataTableColumns<Application> = [
  {
    title: '应用名称',
    key: 'name',
    width: 150
  },
  {
    title: 'Git仓库',
    key: 'gitRepoUrl',
    width: 250,
    ellipsis: true
  },
  {
    title: '分支',
    key: 'gitBranch',
    width: 100
  },
  {
    title: '自动触发',
    key: 'autoTrigger',
    width: 100,
    render: (row) => {
      return h(NTag, { type: row.autoTrigger ? 'success' : 'default' }, () => 
        row.autoTrigger ? '已启用' : '已禁用'
      );
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const type = row.status === 'ACTIVE' ? 'success' : 'default';
      const text = row.status === 'ACTIVE' ? '激活' : '非激活';
      return h(NTag, { type }, () => text);
    }
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
    width: 250,
    render: (row) => {
      return h('div', { class: 'flex gap-8px flex-wrap' }, [
        h(
          NButton,
          {
            size: 'small',
            type: 'success',
            onClick: () => handleBuild(row)
          },
          () => '构建'
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'info',
            onClick: () => handleConfig(row)
          },
          () => '配置'
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
const formModel = reactive<Application>({
  name: '',
  gitRepoUrl: '',
  gitBranch: 'main',
  gitUsername: '',
  gitPassword: '',
  gitPrivateKey: '',
  notificationUrl: '',
  notificationToken: '',
  autoTrigger: false,
  webhookSecret: '',
  status: 'ACTIVE'
});

// 环境变量
const envVariables = ref<EnvVariable[]>([]);

// 表单验证规则
const rules = {
  name: { required: true, message: '请输入应用名称', trigger: 'blur' },
  gitRepoUrl: { required: true, message: '请输入Git仓库地址', trigger: 'blur' },
  gitBranch: { required: true, message: '请输入Git分支', trigger: 'blur' }
};

// API调用方法
const getTableData = async () => {
  loading.value = true;
  try {
    // 模拟数据 - 实际应调用后端API
    const mockData = [
      {
        id: 1,
        name: 'demo-spring-boot',
        gitRepoUrl: 'https://gitee.com/user/demo-spring-boot.git',
        gitBranch: 'main',
        autoTrigger: true,
        status: 'ACTIVE',
        createTime: new Date().toISOString()
      }
    ];
    
    setTimeout(() => {
      tableData.value = mockData;
      pagination.itemCount = mockData.length;
      loading.value = false;
    }, 500);
  } catch (error) {
    message.error('获取应用列表失败');
    console.error('获取应用列表失败:', error);
    loading.value = false;
  }
};

// 页面操作方法
const handleAdd = () => {
  Object.assign(formModel, {
    id: undefined,
    name: '',
    gitRepoUrl: '',
    gitBranch: 'main',
    gitUsername: '',
    gitPassword: '',
    gitPrivateKey: '',
    notificationUrl: '',
    notificationToken: '',
    autoTrigger: false,
    webhookSecret: '',
    status: 'ACTIVE'
  });
  envVariables.value = [];
  showModal.value = true;
};

const handleEdit = (row: Application) => {
  Object.assign(formModel, { ...row });
  
  // 转换环境变量格式
  envVariables.value = row.variables 
    ? Object.entries(row.variables).map(([key, value]) => ({ key, value }))
    : [];
    
  showModal.value = true;
};

const handleDelete = (row: Application) => {
  dialog.warning({
    title: '删除确认',
    content: `确定要删除应用 "${row.name}" 吗？此操作不可恢复。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        // 调用删除API
        message.success('删除成功');
        getTableData();
      } catch (error) {
        message.error('删除失败');
        console.error('删除应用失败:', error);
      }
    }
  });
};

const handleSave = async () => {
  try {
    await formRef.value?.validate();
    
    // 转换环境变量格式
    const variables = envVariables.value.reduce((acc, item) => {
      if (item.key && item.value) {
        acc[item.key] = item.value;
      }
      return acc;
    }, {} as Record<string, string>);
    
    const saveData = {
      ...formModel,
      variables
    };
    
    // 调用保存API
    message.success(formModel.id ? '更新成功' : '创建成功');
    closeModal();
    getTableData();
  } catch (error) {
    message.error('保存失败');
    console.error('保存应用失败:', error);
  }
};

const handleBuild = async (row: Application) => {
  const loadingMessage = message.loading('正在触发构建...', { duration: 0 });
  try {
    const response = await fetch(`/api/applications/${row.id}/build`, {
      method: 'POST'
    });
    const result = await response.json();
    
    loadingMessage.destroy();
    
    if (result.code === 200) {
      message.success('构建已开始，请查看构建管理页面获取详细信息');
    } else {
      message.error(result.message || '触发构建失败');
    }
  } catch (error) {
    loadingMessage.destroy();
    message.error('触发构建失败');
    console.error('触发构建失败:', error);
  }
};

const handleConfig = (row: Application) => {
  // 跳转到流水线配置页面
  message.info('即将跳转到流水线配置页面');
};

// 环境变量操作
const addEnvVariable = () => {
  envVariables.value.push({ key: '', value: '' });
};

const removeEnvVariable = (index: number) => {
  envVariables.value.splice(index, 1);
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