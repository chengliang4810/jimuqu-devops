<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 页面头部 -->
    <div class="flex-y-center justify-between">
      <h1 class="text-18px font-bold">构建管理</h1>
      <div class="flex-y-center gap-12px">
        <NButton type="info" @click="getTableData">
          <Icon icon="material-symbols:refresh" class="mr-4px text-16px" />
          刷新
        </NButton>
      </div>
    </div>

    <!-- 搜索栏 -->
    <div class="card-wrapper">
      <NForm inline label-width="auto" label-placement="left">
        <NFormItem label="应用名称">
          <NSelect
            v-model:value="searchForm.applicationId"
            placeholder="选择应用"
            clearable
            filterable
            :options="applicationOptions"
            class="w-200px"
          />
        </NFormItem>
        <NFormItem label="构建状态">
          <NSelect
            v-model:value="searchForm.status"
            placeholder="选择状态"
            clearable
            :options="statusOptions"
            class="w-150px"
          />
        </NFormItem>
        <NFormItem label="触发方式">
          <NSelect
            v-model:value="searchForm.triggeredBy"
            placeholder="选择触发方式"
            clearable
            :options="triggerOptions"
            class="w-150px"
          />
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
        :scroll-x="1200"
        virtual-scroll
        class="sm:h-full"
      />
    </div>

    <!-- 构建日志弹窗 -->
    <NModal v-model:show="showLogModal" title="构建日志" preset="card" class="w-1000px h-600px">
      <div class="h-full flex flex-col">
        <div class="flex justify-between items-center mb-16px">
          <h3>{{ currentBuild?.applicationName }} - 构建 #{{ currentBuild?.buildNumber }}</h3>
          <div class="flex gap-8px">
            <NTag :type="getStatusType(currentBuild?.status)">
              {{ getStatusText(currentBuild?.status) }}
            </NTag>
            <NButton size="small" @click="refreshLog">
              <Icon icon="material-symbols:refresh" />
            </NButton>
          </div>
        </div>
        <div class="flex-1 overflow-hidden">
          <NScrollbar class="h-full">
            <pre class="bg-gray-900 text-green-400 p-16px rounded text-12px leading-relaxed">{{ buildLog }}</pre>
          </NScrollbar>
        </div>
      </div>
    </NModal>

    <!-- 构建步骤详情弹窗 -->
    <NModal v-model:show="showStepsModal" title="构建步骤" preset="card" class="w-800px">
      <div class="space-y-16px">
        <div v-for="(step, index) in buildSteps" :key="index" class="border rounded p-16px">
          <div class="flex justify-between items-center mb-8px">
            <div class="flex items-center gap-8px">
              <NTag :type="getStatusType(step.status)">
                {{ getStatusText(step.status) }}
              </NTag>
              <span class="font-medium">{{ step.stepName }}</span>
            </div>
            <div class="text-12px text-gray-500">
              耗时: {{ step.durationSeconds ? `${step.durationSeconds}s` : '-' }}
            </div>
          </div>
          <div v-if="step.logContent" class="bg-gray-100 p-8px rounded text-12px max-h-200px overflow-auto">
            <pre>{{ step.logContent }}</pre>
          </div>
          <div v-if="step.errorMessage" class="bg-red-50 text-red-600 p-8px rounded text-12px mt-8px">
            {{ step.errorMessage }}
          </div>
        </div>
      </div>
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
  NSelect,
  NModal, 
  NTag,
  NScrollbar,
  useMessage 
} from 'naive-ui';
import type { DataTableColumns, SelectOption } from 'naive-ui';
import { Icon } from '@iconify/vue';

interface Build {
  id: number;
  applicationId: number;
  applicationName: string;
  buildNumber: number;
  status: 'PENDING' | 'RUNNING' | 'SUCCESS' | 'FAILED' | 'CANCELLED';
  triggeredBy: string;
  triggerUser?: string;
  gitCommit?: string;
  gitBranch?: string;
  startTime?: string;
  endTime?: string;
  durationSeconds?: number;
  createTime: string;
}

interface BuildStep {
  id: number;
  buildId: number;
  stepName: string;
  stepOrder: number;
  status: string;
  startTime?: string;
  endTime?: string;
  durationSeconds?: number;
  logContent?: string;
  errorMessage?: string;
}

const message = useMessage();

// 表格数据
const tableData: Ref<Build[]> = ref([]);
const loading = ref(false);

// 搜索表单
const searchForm = reactive({
  applicationId: null,
  status: null,
  triggeredBy: null
});

// 选项数据
const applicationOptions: Ref<SelectOption[]> = ref([]);
const statusOptions: SelectOption[] = [
  { label: '等待中', value: 'PENDING' },
  { label: '运行中', value: 'RUNNING' },
  { label: '成功', value: 'SUCCESS' },
  { label: '失败', value: 'FAILED' },
  { label: '已取消', value: 'CANCELLED' }
];
const triggerOptions: SelectOption[] = [
  { label: '手动触发', value: 'MANUAL' },
  { label: 'Webhook触发', value: 'WEBHOOK' }
];

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
const columns: DataTableColumns<Build> = [
  {
    title: '构建编号',
    key: 'buildNumber',
    width: 100,
    render: (row) => `#${row.buildNumber}`
  },
  {
    title: '应用名称',
    key: 'applicationName',
    width: 150
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
    title: '触发方式',
    key: 'triggeredBy',
    width: 120,
    render: (row) => {
      const text = row.triggeredBy === 'MANUAL' ? '手动触发' : 'Webhook';
      return h(NTag, { type: 'info' }, () => text);
    }
  },
  {
    title: 'Git分支',
    key: 'gitBranch',
    width: 120
  },
  {
    title: '耗时',
    key: 'durationSeconds',
    width: 100,
    render: (row) => {
      if (row.durationSeconds) {
        return `${row.durationSeconds}s`;
      }
      return '-';
    }
  },
  {
    title: '开始时间',
    key: 'startTime',
    width: 180,
    render: (row) => {
      return row.startTime ? new Date(row.startTime).toLocaleString() : '-';
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
            onClick: () => handleViewLog(row)
          },
          () => '查看日志'
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            onClick: () => handleViewSteps(row)
          },
          () => '查看步骤'
        ),
        row.status === 'RUNNING' ? h(
          NButton,
          {
            size: 'small',
            type: 'error',
            onClick: () => handleCancel(row)
          },
          () => '取消'
        ) : null
      ].filter(Boolean));
    }
  }
];

// 日志相关
const showLogModal = ref(false);
const currentBuild = ref<Build | null>(null);
const buildLog = ref('');

// 步骤相关  
const showStepsModal = ref(false);
const buildSteps = ref<BuildStep[]>([]);

// 状态相关方法
const getStatusType = (status: string) => {
  const typeMap = {
    PENDING: 'default',
    RUNNING: 'info',
    SUCCESS: 'success',
    FAILED: 'error',
    CANCELLED: 'warning'
  };
  return typeMap[status] || 'default';
};

const getStatusText = (status: string) => {
  const textMap = {
    PENDING: '等待中',
    RUNNING: '运行中',
    SUCCESS: '成功',
    FAILED: '失败',
    CANCELLED: '已取消'
  };
  return textMap[status] || '未知';
};

// API调用方法
const getTableData = async () => {
  loading.value = true;
  try {
    // 模拟数据 - 实际应调用后端API
    const mockData: Build[] = [
      {
        id: 1,
        applicationId: 1,
        applicationName: 'demo-spring-boot',
        buildNumber: 15,
        status: 'SUCCESS',
        triggeredBy: 'WEBHOOK',
        triggerUser: 'system',
        gitBranch: 'main',
        gitCommit: 'abc123',
        startTime: new Date(Date.now() - 300000).toISOString(),
        endTime: new Date().toISOString(),
        durationSeconds: 180,
        createTime: new Date().toISOString()
      },
      {
        id: 2,
        applicationId: 1,
        applicationName: 'demo-spring-boot',
        buildNumber: 14,
        status: 'FAILED',
        triggeredBy: 'MANUAL',
        triggerUser: 'admin',
        gitBranch: 'main',
        gitCommit: 'def456',
        startTime: new Date(Date.now() - 600000).toISOString(),
        endTime: new Date(Date.now() - 500000).toISOString(),
        durationSeconds: 120,
        createTime: new Date(Date.now() - 600000).toISOString()
      }
    ];
    
    setTimeout(() => {
      tableData.value = mockData;
      pagination.itemCount = mockData.length;
      loading.value = false;
    }, 500);
  } catch (error) {
    message.error('获取构建列表失败');
    console.error('获取构建列表失败:', error);
    loading.value = false;
  }
};

const getApplicationOptions = async () => {
  try {
    // 模拟数据 - 实际应调用后端API获取应用列表
    applicationOptions.value = [
      { label: 'demo-spring-boot', value: 1 },
      { label: 'demo-vue-app', value: 2 }
    ];
  } catch (error) {
    console.error('获取应用列表失败:', error);
  }
};

// 页面操作方法
const handleViewLog = async (row: Build) => {
  currentBuild.value = row;
  showLogModal.value = true;
  
  // 模拟获取构建日志
  buildLog.value = `构建开始 - ${row.startTime}
正在拉取代码...
Cloning into '/workspace'...
remote: Enumerating objects: 156, done.
remote: Counting objects: 100% (156/156), done.
remote: Compressing objects: 100% (89/89), done.
remote: Total 156 (delta 47), reused 142 (delta 38), pack-reused 0
Receiving objects: 100% (156/156), 45.67 KiB | 1.14 MiB/s, done.
Resolving deltas: 100% (47/47), done.

正在编译项目...
[INFO] Scanning for projects...
[INFO] 
[INFO] ------------------------< com.example:demo >------------------------
[INFO] Building demo 0.0.1-SNAPSHOT
[INFO] --------------------------------[ jar ]---------------------------------
[INFO] 
[INFO] --- maven-clean-plugin:3.2.0:clean (default-clean) @ demo ---
[INFO] Deleting /workspace/target
[INFO] 
[INFO] --- maven-resources-plugin:3.2.0:resources (default-resources) @ demo ---
[INFO] Using 'UTF-8' encoding to copy filtered resources.
[INFO] Using 'UTF-8' encoding to copy filtered properties files.
[INFO] Copying 1 resource
[INFO] Copying 0 resource
[INFO] 
[INFO] --- maven-compiler-plugin:3.10.1:compile (default-compile) @ demo ---
[INFO] Changes detected - recompiling the entire module!
[INFO] Compiling 5 source files to /workspace/target/classes
[INFO] 
[INFO] --- maven-resources-plugin:3.2.0:testResources (default-testResources) @ demo ---
[INFO] Using 'UTF-8' encoding to copy filtered resources.
[INFO] Using 'UTF-8' encoding to copy filtered properties files.
[INFO] Copying 0 resource
[INFO] 
[INFO] --- maven-surefire-plugin:2.22.2:test (default-test) @ demo ---
[INFO] Tests are skipped.
[INFO] 
[INFO] --- maven-jar-plugin:3.2.2:jar (default-jar) @ demo ---
[INFO] Building jar: /workspace/target/demo-0.0.1-SNAPSHOT.jar
[INFO] 
[INFO] --- spring-boot-maven-plugin:2.7.0:repackage (repackage) @ demo ---
[INFO] Replacing main artifact with repackaged archive
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------
[INFO] Total time:  01:23 min
[INFO] Finished at: 2024-01-15T10:32:45Z
[INFO] ------------------------------------------------------------------------

正在构建Docker镜像...
Sending build context to Docker daemon  17.45MB
Step 1/8 : FROM openjdk:11-jre-slim
 ---> 326c7a5e5ede
Step 2/8 : VOLUME /tmp
 ---> Using cache
 ---> 8b4f6b6c8a7d
Step 3/8 : COPY target/*.jar app.jar
 ---> 95a8f3e2b1c4
Step 4/8 : EXPOSE 8080
 ---> Running in d4f5e6a7b8c9
Removing intermediate container d4f5e6a7b8c9
 ---> 2c3d4e5f6a7b
Step 5/8 : ENTRYPOINT ["java","-jar","/app.jar"]
 ---> Running in a1b2c3d4e5f6
Removing intermediate container a1b2c3d4e5f6
 ---> 8c9d0e1f2a3b
Successfully built 8c9d0e1f2a3b
Successfully tagged demo-spring-boot:15

正在启动应用容器...
Container demo-spring-boot started successfully
Application is running on port 8080

构建完成 - ${row.endTime}
总耗时: ${row.durationSeconds}秒`;
};

const handleViewSteps = async (row: Build) => {
  showStepsModal.value = true;
  
  // 模拟获取构建步骤
  buildSteps.value = [
    {
      id: 1,
      buildId: row.id,
      stepName: 'Git代码拉取',
      stepOrder: 1,
      status: 'SUCCESS',
      startTime: row.startTime,
      endTime: new Date(new Date(row.startTime!).getTime() + 30000).toISOString(),
      durationSeconds: 30,
      logContent: 'Cloning repository...\nClone completed successfully.'
    },
    {
      id: 2,
      buildId: row.id,
      stepName: 'Maven编译打包',
      stepOrder: 2,
      status: 'SUCCESS',
      startTime: new Date(new Date(row.startTime!).getTime() + 30000).toISOString(),
      endTime: new Date(new Date(row.startTime!).getTime() + 120000).toISOString(),
      durationSeconds: 90,
      logContent: 'Running maven build...\nBUILD SUCCESS'
    },
    {
      id: 3,
      buildId: row.id,
      stepName: 'Docker镜像构建',
      stepOrder: 3,
      status: 'SUCCESS',
      startTime: new Date(new Date(row.startTime!).getTime() + 120000).toISOString(),
      endTime: new Date(new Date(row.startTime!).getTime() + 180000).toISOString(),
      durationSeconds: 60,
      logContent: 'Building Docker image...\nSuccessfully built 8c9d0e1f2a3b'
    }
  ];
};

const handleCancel = async (row: Build) => {
  try {
    // 调用取消构建API
    message.success('构建已取消');
    getTableData();
  } catch (error) {
    message.error('取消构建失败');
    console.error('取消构建失败:', error);
  }
};

const refreshLog = () => {
  if (currentBuild.value) {
    handleViewLog(currentBuild.value);
  }
};

const reset = () => {
  Object.assign(searchForm, {
    applicationId: null,
    status: null,
    triggeredBy: null
  });
  getTableData();
};

// 生命周期
onMounted(() => {
  getApplicationOptions();
  getTableData();
});
</script>