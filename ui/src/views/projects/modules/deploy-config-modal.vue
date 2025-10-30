<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue';
import { NButton, NIcon, NCard, NFormItem, NInput, NInputNumber, NSelect, NTag, NAlert, NDivider } from 'naive-ui';
import { useMessage } from 'naive-ui';

interface Props {
  projectData?: any;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  submit: [data: any];
}>();

const message = useMessage();
const loading = ref(false);

// 部署步骤配置
const deploymentConfig = ref<any[]>([]);

// 监听props变化，初始化数据
watch(() => props.projectData, (newData) => {
  if (newData) {
    deploymentConfig.value = newData.deployment_config ? [...newData.deployment_config] : [];
  } else {
    deploymentConfig.value = [];
  }
}, { immediate: true });

// 部署步骤类型选项
const stepTypeOptions = [
  { label: '代码检出', value: 'checkout' },
  { label: '环境准备', value: 'setup' },
  { label: '依赖安装', value: 'install' },
  { label: '代码构建', value: 'build' },
  { label: '单元测试', value: 'test' },
  { label: '代码扫描', value: 'scan' },
  { label: 'Docker构建', value: 'docker' },
  { label: '应用部署', value: 'deploy' },
  { label: '健康检查', value: 'health' },
  { label: '通知', value: 'notify' }
];

// 获取默认步骤名称
function getDefaultStepName(type: string): string {
  const nameMap: Record<string, string> = {
    checkout: '代码检出',
    setup: '环境准备',
    install: '依赖安装',
    build: '代码构建',
    test: '单元测试',
    scan: '代码扫描',
    docker: 'Docker构建',
    deploy: '应用部署',
    health: '健康检查',
    notify: '通知'
  };
  return nameMap[type] || '未知步骤';
}

// 获取默认步骤命令
function getDefaultStepCommand(type: string): string {
  const commandMap: Record<string, string> = {
    checkout: 'git clone $GIT_URL . && git checkout $GIT_BRANCH',
    setup: 'echo "准备部署环境"',
    install: 'npm install',
    build: 'npm run build',
    test: 'npm run test',
    scan: 'npm run lint',
    docker: 'docker build -t $PROJECT_NAME:$BUILD_NUMBER .',
    deploy: 'kubectl apply -f deployment.yaml',
    health: 'curl -f http://localhost:3000/health || exit 1',
    notify: 'echo "部署完成通知"'
  };
  return commandMap[type] || '';
}

// 获取默认步骤超时时间
function getDefaultStepTimeout(type: string): number {
  const timeoutMap: Record<string, number> = {
    checkout: 120,
    setup: 60,
    install: 300,
    build: 600,
    test: 300,
    scan: 180,
    docker: 600,
    deploy: 120,
    health: 30,
    notify: 30
  };
  return timeoutMap[type] || 300;
}

// 获取步骤类型颜色
function getStepTypeColor(type: string): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' {
  const colorMap: Record<string, 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'> = {
    checkout: 'info',
    setup: 'default',
    install: 'primary',
    build: 'success',
    test: 'warning',
    scan: 'error',
    docker: 'info',
    deploy: 'primary',
    health: 'success',
    notify: 'default'
  };
  return colorMap[type] || 'default';
}

// 添加部署步骤
function addDeploymentStep() {
  const stepCount = deploymentConfig.value.length;
  const stepType = stepCount === 0 ? 'checkout' : 'build';

  const newStep = {
    id: Date.now(),
    name: getDefaultStepName(stepType),
    type: stepType,
    command: getDefaultStepCommand(stepType),
    timeout: getDefaultStepTimeout(stepType),
    working_dir: '/app',
    target_hosts: [],
    enabled: true,
    continue_on_error: false
  };
  deploymentConfig.value.push(newStep);
}

// 删除部署步骤
function removeDeploymentStep(index: number) {
  deploymentConfig.value.splice(index, 1);
}

// 移动部署步骤
function moveStep(index: number, direction: 'up' | 'down') {
  const steps = deploymentConfig.value;
  if (direction === 'up' && index > 0) {
    [steps[index], steps[index - 1]] = [steps[index - 1], steps[index]];
  } else if (direction === 'down' && index < steps.length - 1) {
    [steps[index], steps[index + 1]] = [steps[index + 1], steps[index]];
  }
}

// 更新步骤属性
function updateStep(index: number, field: string, value: any) {
  deploymentConfig.value[index][field] = value;
}

// 计算总执行时间
const totalExecutionTime = computed(() => {
  return deploymentConfig.value.reduce((total, step) => total + (step.timeout || 0), 0);
});

function handleClose() {
  emit('close');
}

async function handleSubmit() {
  try {
    // 验证部署配置
    if (deploymentConfig.value.length === 0) {
      message.error('请至少添加一个部署步骤');
      return;
    }

    // 验证部署步骤的必填字段
    for (const step of deploymentConfig.value) {
      if (!step.name || !step.command) {
        message.error('所有部署步骤必须填写名称和执行命令');
        return;
      }
    }

    loading.value = true;

    // 模拟提交延迟
    await new Promise(resolve => setTimeout(resolve, 1000));

    const submitData = {
      deployment_config: deploymentConfig.value
    };

    emit('submit', submitData);
    message.success('部署配置已保存');
  } catch (error) {
    console.error('表单提交失败:', error);
    message.error('保存失败，请重试');
  } finally {
    loading.value = false;
  }
}

// 格式化时间显示
function formatTime(seconds: number): string {
  if (seconds < 60) {
    return `${seconds}秒`;
  } else if (seconds < 3600) {
    return `${Math.floor(seconds / 60)}分${seconds % 60}秒`;
  } else {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const remainingSeconds = seconds % 60;
    return `${hours}小时${minutes}分${remainingSeconds}秒`;
  }
}
</script>

<template>
  <div class="p-6">
    <!-- 项目信息展示 -->
    <div v-if="props.projectData" class="mb-6 p-4 bg-gray-50 rounded-lg">
      <h3 class="text-lg font-medium mb-2">{{ props.projectData.name }}</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-gray-600">
        <div>
          <span class="font-medium">标识符：</span>{{ props.projectData.identifier }}
        </div>
        <div>
          <span class="font-medium">分支：</span>{{ props.projectData.branch }}
        </div>
        <div>
          <span class="font-medium">Git仓库：</span>
          <span class="font-mono text-xs">{{ props.projectData.git_url }}</span>
        </div>
      </div>
    </div>

    <!-- 部署步骤配置 -->
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <span class="text-lg font-medium">部署步骤配置</span>
        <NButton size="small" @click="addDeploymentStep">
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

      <div v-if="deploymentConfig.length === 0" class="text-center py-8 text-gray-400 border-2 border-dashed border-gray-300 rounded-lg">
        <NIcon size="48" class="mb-4">
          <svg viewBox="0 0 24 24">
            <path fill="currentColor" d="M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2M12,4A8,8 0 0,1 20,12A8,8 0 0,1 12,20A8,8 0 0,1 4,12A8,8 0 0,1 12,4M11,7H13V13H11V7M11,15H13V17H11V15Z"/>
          </svg>
        </NIcon>
        <p class="mb-2">暂无部署步骤</p>
        <p class="text-sm">点击上方按钮添加第一个部署步骤</p>
      </div>

      <div v-else class="space-y-3">
        <div v-for="(step, index) in deploymentConfig" :key="step.id" class="border border-gray-200 rounded-lg p-4">
          <div class="flex justify-between items-start mb-3">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-3">
                <NTag :type="getStepTypeColor(step.type)" size="small">
                  {{ stepTypeOptions.find(opt => opt.value === step.type)?.label || step.type }}
                </NTag>
                <span v-if="!step.enabled" class="text-gray-400 text-xs line-through">已禁用</span>
                <span class="font-medium text-sm">步骤 {{ index + 1 }}</span>
              </div>

              <!-- 步骤配置表单 -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">步骤名称</label>
                  <NInput
                    :value="step.name"
                    @update:value="(value) => updateStep(index, 'name', value)"
                    placeholder="请输入步骤名称"
                    size="small"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">步骤类型</label>
                  <NSelect
                    :value="step.type"
                    @update:value="(value) => {
                      updateStep(index, 'type', value);
                      updateStep(index, 'name', getDefaultStepName(value));
                      updateStep(index, 'command', getDefaultStepCommand(value));
                      updateStep(index, 'timeout', getDefaultStepTimeout(value));
                    }"
                    :options="stepTypeOptions"
                    size="small"
                  />
                </div>
                <div class="md:col-span-2">
                  <label class="block text-sm font-medium text-gray-700 mb-1">执行命令</label>
                  <NInput
                    :value="step.command"
                    @update:value="(value) => updateStep(index, 'command', value)"
                    type="textarea"
                    placeholder="请输入执行命令"
                    :autosize="{ minRows: 2, maxRows: 4 }"
                    size="small"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">超时时间(秒)</label>
                  <NInputNumber
                    :value="step.timeout"
                    @update:value="(value) => updateStep(index, 'timeout', value)"
                    :min="1"
                    :max="3600"
                    size="small"
                    class="w-full"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">工作目录</label>
                  <NInput
                    :value="step.working_dir"
                    @update:value="(value) => updateStep(index, 'working_dir', value)"
                    placeholder="/app"
                    size="small"
                  />
                </div>
              </div>

              <!-- 高级选项 -->
              <div class="mt-3 pt-3 border-t border-gray-100">
                <div class="flex items-center gap-4">
                  <div class="flex items-center gap-2">
                    <input
                      type="checkbox"
                      :checked="step.enabled"
                      @change="(e) => updateStep(index, 'enabled', (e.target as HTMLInputElement).checked)"
                      class="text-blue-500"
                    />
                    <label class="text-sm text-gray-700">启用此步骤</label>
                  </div>
                  <div class="flex items-center gap-2">
                    <input
                      type="checkbox"
                      :checked="step.continue_on_error"
                      @change="(e) => updateStep(index, 'continue_on_error', (e.target as HTMLInputElement).checked)"
                      class="text-blue-500"
                    />
                    <label class="text-sm text-gray-700">出错时继续执行</label>
                  </div>
                </div>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="flex gap-2 ml-4">
              <NButton size="small" @click="moveStep(index, 'up')" :disabled="index === 0">
                上移
              </NButton>
              <NButton size="small" @click="moveStep(index, 'down')" :disabled="index === deploymentConfig.length - 1">
                下移
              </NButton>
              <NButton size="small" type="error" @click="removeDeploymentStep(index)">
                删除
              </NButton>
            </div>
          </div>
        </div>

        <!-- 统计信息 -->
        <div class="text-center py-2 bg-gray-50 rounded-lg">
          <span class="text-sm text-gray-600">
            共 {{ deploymentConfig.length }} 个步骤，预计执行时间: {{ formatTime(totalExecutionTime) }}
          </span>
        </div>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="flex justify-end gap-3 pt-6 border-t">
      <NButton @click="handleClose">取消</NButton>
      <NButton type="primary" @click="handleSubmit" :loading="loading" :disabled="deploymentConfig.length === 0">
        保存配置
      </NButton>
    </div>
  </div>
</template>

<style scoped>
input[type="checkbox"]:checked {
  background-color: #3b82f6;
  border-color: #3b82f6;
}
</style>