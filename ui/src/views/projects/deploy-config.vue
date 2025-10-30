<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { NButton, NCard, NSteps, NStep, NIcon, NInput, NSelect, NCheckbox, NSpace, NForm, NFormItem, NTag, NModal, NPopconfirm, NMessage } from 'naive-ui';
import { useMessage } from 'naive-ui';

const route = useRoute();
const router = useRouter();
const message = useMessage();

// 项目ID
const projectId = ref<string>('');
const projectData = ref<any>(null);
const loading = ref(false);

// 部署配置步骤
const currentStep = ref(1);

// 部署步骤配置
const deploymentSteps = ref<Array<{
  id: string;
  name: string;
  type: string;
  command: string;
  timeout: number;
  working_dir: string;
  enabled: boolean;
  retry_count: number;
  target_hosts?: string[];
}>>([]);

// 步骤类型选项
const stepTypes = [
  { label: '代码检出', value: 'checkout' },
  { label: '环境准备', value: 'env_setup' },
  { label: '依赖安装', value: 'install_deps' },
  { label: '构建', value: 'build' },
  { label: '测试', value: 'test' },
  { label: '代码扫描', value: 'scan' },
  { label: 'Docker构建', value: 'docker_build' },
  { label: '应用部署', value: 'deploy' },
  { label: '健康检查', value: 'health_check' },
  { label: '通知', value: 'notify' }
];

// 加载项目数据
async function loadProjectData() {
  loading.value = true;
  try {
    // 模拟从API获取项目数据
    const mockData = {
      id: projectId.value,
      name: 'Web前端项目',
      identifier: 'web-frontend',
      branch: 'main',
      description: '公司官网前端项目，基于Vue3开发',
      webhook_url: `https://api.jimuqu.com/webhook/${projectId.value}`,
      deployment_config: [
        {
          id: '1',
          name: '构建前端',
          type: 'build',
          command: 'npm run build',
          timeout: 300,
          working_dir: '/app',
          enabled: true,
          retry_count: 2
        },
        {
          id: '2',
          name: '部署到服务器',
          type: 'deploy',
          command: 'rsync -av dist/ root@server:/var/www/html/',
          timeout: 600,
          working_dir: '/app',
          enabled: true,
          retry_count: 1,
          target_hosts: ['web-server-01', 'web-server-02']
        }
      ]
    };

    projectData.value = mockData;
    deploymentSteps.value = mockData.deployment_config || [];
  } catch (error) {
    message.error('加载项目数据失败');
  } finally {
    loading.value = false;
  }
}

// 添加新步骤
function addStep() {
  const newStep = {
    id: Date.now().toString(),
    name: '新步骤',
    type: 'build',
    command: '',
    timeout: 300,
    working_dir: '/app',
    enabled: true,
    retry_count: 1
  };
  deploymentSteps.value.push(newStep);
}

// 删除步骤
function deleteStep(stepId: string) {
  const index = deploymentSteps.value.findIndex(step => step.id === stepId);
  if (index > -1) {
    deploymentSteps.value.splice(index, 1);
  }
}

// 上移步骤
function moveUp(index: number) {
  if (index > 0) {
    [deploymentSteps.value[index], deploymentSteps.value[index - 1]] =
    [deploymentSteps.value[index - 1], deploymentSteps.value[index]];
  }
}

// 下移步骤
function moveDown(index: number) {
  if (index < deploymentSteps.value.length - 1) {
    [deploymentSteps.value[index], deploymentSteps.value[index + 1]] =
    [deploymentSteps.value[index + 1], deploymentSteps.value[index]];
  }
}

// 切换步骤启用状态
function toggleStep(stepId: string) {
  const step = deploymentSteps.value.find(s => s.id === stepId);
  if (step) {
    step.enabled = !step.enabled;
  }
}

// 保存配置
function saveConfig() {
  if (projectData.value) {
    projectData.value.deployment_config = deploymentSteps.value;
    projectData.value.updated_at = new Date().toISOString();
    message.success('部署配置已保存');
    router.push('/projects');
  }
}

// 返回项目列表
function goBack() {
  router.push('/projects');
}

// 计算预计执行时间
const estimatedTime = computed(() => {
  const enabledSteps = deploymentSteps.value.filter(step => step.enabled);
  const totalSeconds = enabledSteps.reduce((total, step) => total + step.timeout, 0);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return minutes > 0 ? `${minutes}分${seconds}秒` : `${seconds}秒`;
});

onMounted(() => {
  projectId.value = route.params.id as string;
  if (projectId.value) {
    loadProjectData();
  }
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden">
    <!-- 页面头部 -->
    <NCard>
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <NButton @click="goBack" quaternary>
              <template #icon>
                <NIcon>
                  <svg viewBox="0 0 24 24">
                    <path fill="currentColor" d="M20,11V13H8L13.5,18.5L12.08,19.92L4.16,12L12.08,4.08L13.5,5.5L8,11H20Z"/>
                  </svg>
                </NIcon>
              </template>
              返回项目列表
            </NButton>
            <h2 class="text-xl font-semibold">部署配置</h2>
          </div>
          <div class="flex items-center gap-4">
            <NTag type="info" size="large">
              项目：{{ projectData?.name }}
            </NTag>
            <NTag type="success" size="large">
              Webhook: {{ projectData?.webhook_url }}
            </NTag>
          </div>
        </div>
      </template>
    </NCard>

    <!-- 部署步骤配置 -->
    <NCard title="部署步骤配置">
      <template #header-extra>
        <div class="flex items-center gap-4">
          <NTag type="primary">
            步骤数量：{{ deploymentSteps.length }}
          </NTag>
          <NTag type="warning">
            预计执行时间：{{ estimatedTime }}
          </NTag>
          <NButton type="primary" @click="addStep">
            <template #icon>
              <NIcon>
                <svg viewBox="0 0 24 24">
                  <path fill="currentColor" d="M19,13H13V19H11V13H5V11H11V5H13V11H19V13Z"/>
                </svg>
              </NIcon>
            </template>
            添加步骤
          </NButton>
        </div>
      </template>

      <div v-if="deploymentSteps.length === 0" class="text-center py-12">
        <NIcon size="48" class="text-gray-400 mb-4">
          <svg viewBox="0 0 24 24">
            <path fill="currentColor" d="M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2M12,17A5,5 0 0,1 7,12A5,5 0 0,1 12,7A5,5 0 0,1 17,12A5,5 0 0,1 12,17M12,9A3,3 0 0,0 9,12A3,3 0 0,0 12,15A3,3 0 0,0 15,12A3,3 0 0,0 12,9Z"/>
          </svg>
        </NIcon>
        <p class="text-gray-500">暂无部署步骤，点击"添加步骤"开始配置</p>
      </div>

      <div v-else class="space-y-4">
        <NCard
          v-for="(step, index) in deploymentSteps"
          :key="step.id"
          size="small"
          :class="{ 'opacity-50': !step.enabled }"
        >
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <NCheckbox
                  :checked="step.enabled"
                  @update:checked="toggleStep(step.id)"
                />
                <span class="font-medium">{{ index + 1 }}. {{ step.name }}</span>
                <NTag :type="step.enabled ? 'success' : 'default'" size="small">
                  {{ step.enabled ? '启用' : '禁用' }}
                </NTag>
                <NTag type="info" size="small">
                  {{ stepTypes.find(t => t.value === step.type)?.label || step.type }}
                </NTag>
              </div>
              <div class="flex items-center gap-2">
                <NButton size="small" @click="moveUp(index)" :disabled="index === 0">
                  <template #icon>
                    <NIcon>
                      <svg viewBox="0 0 24 24">
                        <path fill="currentColor" d="M7.41,15.41L12,10.83L16.59,15.41L18,14L12,8L6,14L7.41,15.41Z"/>
                      </svg>
                    </NIcon>
                  </template>
                </NButton>
                <NButton size="small" @click="moveDown(index)" :disabled="index === deploymentSteps.length - 1">
                  <template #icon>
                    <NIcon>
                      <svg viewBox="0 0 24 24">
                        <path fill="currentColor" d="M7.41,8.58L12,13.17L16.59,8.58L18,10L12,16L6,10L7.41,8.58Z"/>
                      </svg>
                    </NIcon>
                  </template>
                </NButton>
                <NPopconfirm onPositiveClick="deleteStep(step.id)">
                  {{
                    default: () => '确认删除此步骤吗？',
                    trigger: () => (
                      <NButton type="error" size="small">
                        <template #icon>
                          <NIcon>
                            <svg viewBox="0 0 24 24">
                              <path fill="currentColor" d="M19,4H15.5L14.5,3H9.5L8.5,4H5V6H19M6,19A2,2 0 0,0 8,21H16A2,2 0 0,0 18,19V7H6V19Z"/>
                            </svg>
                          </NIcon>
                        </template>
                      </NButton>
                    )
                  }}
                </NPopconfirm>
              </div>
            </div>
          </template>

          <NForm :model="step" label-placement="left" label-width="100px">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <NFormItem label="步骤名称">
                <NInput v-model:value="step.name" placeholder="请输入步骤名称" />
              </NFormItem>

              <NFormItem label="步骤类型">
                <NSelect
                  v-model:value="step.type"
                  :options="stepTypes"
                  placeholder="请选择步骤类型"
                />
              </NFormItem>

              <NFormItem label="执行命令">
                <NInput
                  v-model:value="step.command"
                  type="textarea"
                  placeholder="请输入执行命令"
                  :autosize="{ minRows: 2, maxRows: 4 }"
                />
              </NFormItem>

              <NFormItem label="工作目录">
                <NInput v-model:value="step.working_dir" placeholder="请输入工作目录" />
              </NFormItem>

              <NFormItem label="超时时间(秒)">
                <NInput v-model:value="step.timeout" type="number" placeholder="300" />
              </NFormItem>

              <NFormItem label="重试次数">
                <NInput v-model:value="step.retry_count" type="number" placeholder="1" />
              </NFormItem>
            </div>
          </NForm>
        </NCard>
      </div>

      <template #footer>
        <div class="flex justify-end gap-4">
          <NButton @click="goBack">取消</NButton>
          <NButton type="primary" @click="saveConfig" :loading="loading">
            保存配置
          </NButton>
        </div>
      </template>
    </NCard>
  </div>
</template>