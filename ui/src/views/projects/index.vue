<script setup lang="tsx">
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { NButton, NPopconfirm, NTag, NIcon, NModal } from 'naive-ui';
import { useAppStore } from '@/store/modules/app';
import { useMessage } from 'naive-ui';
import ProjectAddModal from './modules/project-form-modal.vue';

const appStore = useAppStore();
const message = useMessage();
const router = useRouter();

// 静态模拟数据
const mockProjects = ref([
  {
    id: 1,
    name: 'Web前端项目',
    identifier: 'web-frontend',
    branch: 'main',
    description: '公司官网前端项目，基于Vue3开发',
    git_url: 'https://github.com/company/web-frontend.git',
    git_username: 'deploy-user',
    git_password: 'encrypted_token',
    webhook_url: 'https://api.jimuqu.com/webhook/web-frontend',
    notification: {
      enabled: true,
      types: ['webhook', 'dingtalk'],
      dingtalk_webhook: 'https://oapi.dingtalk.com/robot/send?access_token=xxx'
    },
    deployment_config: [
      {
        id: 1,
        name: '构建前端',
        type: 'build',
        command: 'npm run build',
        timeout: 300,
        working_dir: '/app'
      },
      {
        id: 2,
        name: '部署到服务器',
        type: 'deploy',
        command: 'rsync -av dist/ root@server:/var/www/html/',
        timeout: 600,
        target_hosts: ['web-server-01', 'web-server-02']
      }
    ],
    created_at: '2025-10-01T00:00:00Z',
    updated_at: '2025-10-30T11:30:00Z'
  },
  {
    id: 2,
    name: 'API后端服务',
    identifier: 'api-backend',
    branch: 'develop',
    description: 'Spring Boot后端API服务',
    git_url: 'https://github.com/company/api-backend.git',
    git_username: 'deploy-user',
    git_password: 'encrypted_token',
    webhook_url: 'https://api.jimuqu.com/webhook/api-backend',
    notification: {
      enabled: true,
      types: ['webhook'],
      dingtalk_webhook: ''
    },
    deployment_config: [
      {
        id: 1,
        name: '打包Java应用',
        type: 'build',
        command: './mvnw clean package -DskipTests',
        timeout: 600,
        working_dir: '/app'
      },
      {
        id: 2,
        name: 'Docker构建',
        type: 'docker',
        command: 'docker build -t api-backend:latest .',
        timeout: 300,
        working_dir: '/app'
      }
    ],
    created_at: '2025-10-05T00:00:00Z',
    updated_at: '2025-10-30T10:20:00Z'
  },
  {
    id: 3,
    name: '移动端APP',
    identifier: 'mobile-app',
    branch: 'release',
    description: 'React Native移动应用项目',
    git_url: 'https://github.com/company/mobile-app.git',
    git_username: 'deploy-user',
    git_password: 'encrypted_token',
    webhook_url: 'https://api.jimuqu.com/webhook/mobile-app',
    notification: {
      enabled: false,
      types: [],
      dingtalk_webhook: ''
    },
    deployment_config: [
      {
        id: 1,
        name: '构建Android APK',
        type: 'build',
        command: 'npx react-native build-android --release',
        timeout: 1200,
        working_dir: '/app'
      }
    ],
    created_at: '2025-10-10T00:00:00Z',
    updated_at: '2025-10-29T18:30:00Z'
  }
]);

const searchParams: any = reactive({
  current: 1,
  size: 10,
  name: '',
  identifier: '',
  branch: ''
});

// 弹窗状态
const showModal = ref(false);
const editingProject = ref<any>(null);
const loading = ref(false);

// 过滤后的数据
const filteredProjects = ref(mockProjects.value);

// 过滤函数
function filterProjects() {
  let filtered = mockProjects.value;

  // 项目名称搜索过滤
  if (searchParams.name) {
    const searchLower = searchParams.name.toLowerCase();
    filtered = filtered.filter(project =>
      project.name.toLowerCase().includes(searchLower)
    );
  }

  // 标识符搜索过滤
  if (searchParams.identifier) {
    const searchLower = searchParams.identifier.toLowerCase();
    filtered = filtered.filter(project =>
      project.identifier.toLowerCase().includes(searchLower)
    );
  }

  // 分支搜索过滤
  if (searchParams.branch) {
    const searchLower = searchParams.branch.toLowerCase();
    filtered = filtered.filter(project =>
      project.branch.toLowerCase().includes(searchLower)
    );
  }

  return filtered;
}

// 获取分页数据
function getPaginatedData() {
  const filtered = filterProjects();
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
    title: '项目名称',
    align: 'center',
    minWidth: 150,
    render: row => (
      <div class="flex items-center justify-center">
        <span class="font-medium">{row.name}</span>
      </div>
    )
  },
  {
    key: 'identifier',
    title: '标识符',
    align: 'center',
    width: 120,
    render: row => (
      <NTag type="info" size="small">
        {row.identifier}
      </NTag>
    )
  },
  {
    key: 'branch',
    title: '分支',
    align: 'center',
    width: 100,
    render: row => (
      <NTag type="success" size="small">
        {row.branch}
      </NTag>
    )
  },
  {
    key: 'webhook_url',
    title: 'Webhook URL',
    align: 'center',
    minWidth: 200,
    render: row => (
      <div class="flex items-center justify-center gap-2">
        <span class="text-xs font-mono text-gray-600 truncate max-w-32" title={row.webhook_url}>
          {row.webhook_url || '-'}
        </span>
        {row.webhook_url && (
          <NButton
            size="tiny"
            quaternary
            onClick={() => {
              navigator.clipboard.writeText(row.webhook_url);
              message.success('Webhook URL已复制到剪贴板');
            }}
            class="ml-1"
          >
            复制
          </NButton>
        )}
      </div>
    )
  },
  {
    key: 'notification',
    title: '消息通知',
    align: 'center',
    width: 120,
    render: row => {
      if (!row.notification.enabled) {
        return <NTag type="default" size="small">关闭</NTag>;
      }

      const types = row.notification.types || [];
      const typeLabels = {
        webhook: 'Webhook',
        dingtalk: '钉钉'
      };

      return (
        <div class="flex justify-center gap-1">
          {types.map((type: string, index: number) => (
            <NTag key={index} type="primary" size="small">
              {typeLabels[type] || type}
            </NTag>
          ))}
        </div>
      );
    }
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
    width: 180,
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
        <NButton
          type="info"
          ghost
          size="small"
          onClick={() => goToDeployConfig(row.id)}
        >
          部署配置
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

// 新增项目
function handleAdd() {
  editingProject.value = null;
  showModal.value = true;
}

// 编辑项目
function edit(id: number) {
  const project = mockProjects.value.find(item => item.id === id);
  if (project) {
    editingProject.value = project;
    showModal.value = true;
  }
}

// 跳转到部署配置页面
function goToDeployConfig(id: number) {
  router.push(`/projects/deploy-config/${id}`);
}

// 提交表单
function handleSubmit(formData: any) {
  if (editingProject.value) {
    // 编辑模式
    Object.assign(editingProject.value, formData);
    message.success('项目信息已更新');
  } else {
    // 新增模式
    const identifier = formData.identifier || 'project';
    const newProject = {
      id: Date.now(),
      ...formData,
      webhook_url: `https://api.jimuqu.com/webhook/${identifier}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      notification: {
        enabled: formData.notification_enabled || false,
        types: formData.notification_types || [],
        dingtalk_webhook: formData.dingtalk_webhook || ''
      },
      deployment_config: []
    };
    mockProjects.value.push(newProject);
    message.success('项目添加成功');

    // 提示用户是否配置部署
    setTimeout(() => {
      if (confirm('项目创建成功！是否立即配置部署信息？')) {
        router.push(`/projects/deploy-config/${newProject.id}`);
      }
    }, 500);
  }
  showModal.value = false;
  editingProject.value = null;
  getDataByPage();
}

// 删除项目
function handleDelete(id: number) {
  const index = mockProjects.value.findIndex(project => project.id === id);
  if (index > -1) {
    mockProjects.value.splice(index, 1);
    getDataByPage();
    message.success('项目已删除');
  }
}

// 刷新数据
function getData() {
  getDataByPage();
}

// 清空搜索条件
function clearSearch() {
  Object.assign(searchParams, {
    name: '',
    identifier: '',
    branch: ''
  });
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
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">项目名称</label>
            <NInput
              v-model:value="searchParams.name"
              placeholder="请输入项目名称"
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
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">标识符</label>
            <NInput
              v-model:value="searchParams.identifier"
              placeholder="请输入项目标识符"
              clearable
              @input="getDataByPage"
            >
              <template #prefix>
                <NIcon>
                  <svg viewBox="0 0 24 24">
                    <path fill="currentColor" d="M12,2A10,10 0 0,1 22,12A10,10 0 0,1 12,22A10,10 0 0,1 2,12A10,10 0 0,1 12,2M12,4A8,8 0 0,0 4,12A8,8 0 0,0 12,20A8,8 0 0,0 20,12A8,8 0 0,0 12,4M11,7H13V13H11V7M11,15H13V17H11V15Z"/>
                  </svg>
                </NIcon>
              </template>
            </NInput>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">分支</label>
            <NInput
              v-model:value="searchParams.branch"
              placeholder="请输入分支名称"
              clearable
              @input="getDataByPage"
            >
              <template #prefix>
                <NIcon>
                  <svg viewBox="0 0 24 24">
                    <path fill="currentColor" d="M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2M11,7H13V13H11V7M11,15H13V17H11V15Z"/>
                  </svg>
                </NIcon>
              </template>
            </NInput>
          </div>
          <div class="flex gap-2">
            <NButton @click="getDataByPage">
              <template #icon>
                <NIcon>
                  <svg viewBox="0 0 24 24">
                    <path fill="currentColor" d="M9.5,3A6.5,6.5 0 0,1 16,9.5C16,11.11 15.41,12.59 14.44,13.73L14.71,14H15.5L20.5,19L19,20.5L14,15.5V14.71L13.73,14.44C12.59,15.41 11.11,16 9.5,16A6.5,6.5 0 0,1 3,9.5A6.5,6.5 0 0,1 9.5,3M9.5,5C7,5 5,7 5,9.5C5,12 7,14 9.5,14C12,14 14,12 14,9.5C14,7 12,5 9.5,5Z"/>
                  </svg>
                </NIcon>
              </template>
              搜索
            </NButton>
            <NButton @click="clearSearch">
              <template #icon>
                <NIcon>
                  <svg viewBox="0 0 24 24">
                    <path fill="currentColor" d="M19,6.41L17.59,5L12,10.59L6.41,5L5,6.41L10.59,12L5,17.59L6.41,19L12,13.41L17.59,19L19,17.59L13.41,12L19,6.41Z"/>
                  </svg>
                </NIcon>
              </template>
              清空
            </NButton>
          </div>
        </div>
      </div>
    </NCard>

    <!-- 项目管理卡片 -->
    <NCard :bordered="false" size="small" class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex justify-between items-center w-full">
          <span class="text-lg font-medium">项目管理</span>
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
              新增项目
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
        :scroll-x="1200"
        :loading="loading"
        remote
        :row-key="row => row.id"
        :pagination="mobilePagination"
        class="sm:h-full"
      />
    </NCard>

    <!-- 新增/编辑项目弹窗 -->
    <NModal v-model:show="showModal" :mask-closable="false" preset="card" style="width: 1000px;" :title="editingProject ? '编辑项目' : '新增项目'">
      <ProjectAddModal
        v-if="showModal"
        :project-data="editingProject"
        @close="showModal = false"
        @submit="handleSubmit"
      />
    </NModal>
  </div>
</template>

<style scoped></style>