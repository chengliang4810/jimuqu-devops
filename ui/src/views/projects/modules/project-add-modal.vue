<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import { NButton, NIcon, NForm, NFormItem, NInput, NInputNumber, NDivider, NAlert, NCheckboxGroup, NCheckbox, NSelect, NCard, NSpace, NTag } from 'naive-ui';
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
const formRef = ref();
const loading = ref(false);
const isEdit = ref(!!props.projectData);

const formData = reactive({
  name: '',
  identifier: '',
  branch: 'main',
  description: '',
  git_url: '',
  git_username: '',
  git_password: '',
  webhook_url: '',
  notification_enabled: false,
  notification_types: [],
  dingtalk_webhook: '',
  deployment_config: []
});

const rules = {
  name: { required: true, message: '请输入项目名称', trigger: 'blur' },
  identifier: { required: true, message: '请输入项目标识符', trigger: 'blur' },
  branch: { required: true, message: '请输入分支名称', trigger: 'blur' },
  git_url: { required: true, message: '请输入Git仓库地址', trigger: 'blur' },
  git_username: { required: true, message: '请输入Git账号', trigger: 'blur' },
  git_password: { required: true, message: '请输入Git密码或Token', trigger: 'blur' }
};

// 通知类型选项
const notificationOptions = [
  { label: 'Webhook', value: 'webhook' },
  { label: '钉钉', value: 'dingtalk' }
];

// 分支选项（常见分支）
const branchOptions = [
  { label: 'main', value: 'main' },
  { label: 'master', value: 'master' },
  { label: 'develop', value: 'develop' },
  { label: 'dev', value: 'dev' },
  { label: 'release', value: 'release' },
  { label: 'feature/*', value: 'feature/*' },
  { label: 'hotfix/*', value: 'hotfix/*' }
];

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

// 监听props变化，初始化表单数据
watch(() => props.projectData, (newData) => {
  if (newData) {
    isEdit.value = true;
    Object.assign(formData, {
      name: newData.name || '',
      identifier: newData.identifier || '',
      branch: newData.branch || 'main',
      description: newData.description || '',
      git_url: newData.git_url || '',
      git_username: newData.git_username || '',
      git_password: newData.git_password || '',
      webhook_url: newData.webhook_url || '',
      notification_enabled: newData.notification?.enabled || false,
      notification_types: newData.notification?.types || [],
      dingtalk_webhook: newData.notification?.dingtalk_webhook || '',
      deployment_config: newData.deployment_config || []
    });
  } else {
    isEdit.value = false;
    resetForm();
  }
}, { immediate: true });

function handleClose() {
  emit('close');
}

async function handleSubmit() {
  try {
    await formRef.value?.validate();

    // 验证通知配置
    if (formData.notification_enabled && formData.notification_types.length === 0) {
      message.error('请选择至少一种通知方式');
      return;
    }

    if (formData.notification_enabled && formData.notification_types.includes('dingtalk') && !formData.dingtalk_webhook) {
      message.error('请输入钉钉Webhook地址');
      return;
    }

    // 验证部署配置
    if (formData.deployment_config.length === 0) {
      message.error('请至少添加一个部署步骤');
      return;
    }

    // 验证部署步骤的必填字段
    for (const step of formData.deployment_config) {
      if (!step.name || !step.command) {
        message.error('所有部署步骤必须填写名称和执行命令');
        return;
      }
    }

    loading.value = true;

    // 模拟提交延迟
    await new Promise(resolve => setTimeout(resolve, 1000));

    const submitData = {
      name: formData.name,
      identifier: formData.identifier,
      branch: formData.branch,
      description: formData.description,
      git_url: formData.git_url,
      git_username: formData.git_username,
      git_password: formData.git_password,
      webhook_url: formData.webhook_url,
      notification_enabled: formData.notification_enabled,
      notification_types: formData.notification_types,
      dingtalk_webhook: formData.dingtalk_webhook,
      deployment_config: formData.deployment_config
    };

    emit('submit', submitData);
  } catch (error) {
    console.error('表单验证失败:', error);
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  Object.assign(formData, {
    name: '',
    identifier: '',
    branch: 'main',
    description: '',
    git_url: '',
    git_username: '',
    git_password: '',
    webhook_url: '',
    notification_enabled: false,
    notification_types: [],
    dingtalk_webhook: '',
    deployment_config: []
  });
  formRef.value?.restoreValidation();
}

// 添加部署步骤
function addDeploymentStep() {
  const stepCount = formData.deployment_config.length;
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
  formData.deployment_config.push(newStep);
}

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

// 删除部署步骤
function removeDeploymentStep(index: number) {
  formData.deployment_config.splice(index, 1);
}

// 移动部署步骤
function moveStep(index: number, direction: 'up' | 'down') {
  const steps = formData.deployment_config;
  if (direction === 'up' && index > 0) {
    [steps[index], steps[index - 1]] = [steps[index - 1], steps[index]];
  } else if (direction === 'down' && index < steps.length - 1) {
    [steps[index], steps[index + 1]] = [steps[index + 1], steps[index]];
  }
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

// 计算总执行时间
function getTotalExecutionTime(): number {
  return formData.deployment_config.reduce((total, step) => total + (step.timeout || 0), 0);
}
</script>

<template>
  <div class="p-6">
    <NForm ref="formRef" :model="formData" :rules="rules" label-placement="top" label-width="100">
      <!-- 基本信息 -->
      <div class="mb-6">
        <h4 class="text-base font-medium mb-4 flex items-center gap-2">
          <NIcon size="18" class="text-blue-500">
            <svg viewBox="0 0 24 24">
              <path fill="currentColor" d="M17,8H8V6L5,9L8,12V10H17M12,20A8,8 0 0,0 20,12A8,8 0 0,0 12,4A8,8 0 0,0 4,12A8,8 0 0,0 12,20M12,2A10,10 0 0,1 22,12A10,10 0 0,1 12,22C6.47,22 2,17.5 2,12A10,10 0 0,1 12,2Z"/>
            </svg>
          </NIcon>
          基本信息
        </h4>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <NFormItem label="项目名称" path="name">
            <NInput
              v-model:value="formData.name"
              placeholder="请输入项目名称，如：Web前端项目"
              maxlength="100"
              show-count
            />
          </NFormItem>

          <NFormItem label="项目标识符" path="identifier">
            <NInput
              v-model:value="formData.identifier"
              placeholder="请输入项目标识符，如：web-frontend"
              maxlength="50"
              show-count
            />
          </NFormItem>

          <NFormItem label="默认分支" path="branch">
            <NSelect
              v-model:value="formData.branch"
              placeholder="请选择或输入默认分支"
              :options="branchOptions"
              filterable
              tag
              clearable
            />
          </NFormItem>

          <NFormItem label="描述">
            <NInput
              v-model:value="formData.description"
              type="textarea"
              placeholder="请输入项目描述信息"
              :autosize="{ minRows: 2, maxRows: 4 }"
              maxlength="500"
              show-count
            />
          </NFormItem>
        </div>
      </div>

      <NDivider class="my-6" />

      <!-- Git配置 -->
      <div class="mb-6">
        <h4 class="text-base font-medium mb-4 flex items-center gap-2">
          <NIcon size="18" class="text-green-500">
            <svg viewBox="0 0 24 24">
              <path fill="currentColor" d="M12,2A10,10 0 0,0 2,12C2,16.42 4.87,20.17 8.84,21.5C9.34,21.58 9.79,21.95 10,21.82C10,21.82 10.03,18.22 10.03,15.5C10.03,15.5 9.5,13.56 9.5,13.56C8.75,13.31 8.5,13.18 8.25,13C8.03,12.82 7.5,12.41 7.5,12.41C6.15,12.41 6.15,13.66 6.15,13.66C6.15,13.66 6.03,15.5 6.03,15.5C6.03,17.68 7.58,19.61 9.5,20.82C9.66,20.95 9.82,21.08 10,21.17V16.5C10,16.5 9.5,14.5 9.5,14.5C8.83,14.5 8.33,14.17 8,13.83C8.33,13.5 8.66,13.17 9.33,13.17C10.67,13.17 10.83,14.17 10.83,14.17C11.33,14.17 11.17,13.33 11.17,13.33C11.5,13.67 11.83,13.5 12.33,13.5C13,13.5 13.17,14.17 13.17,14.17C13.17,14.17 13.5,15.5 13.5,15.5C13.5,15.5 13.83,13.17 14.17,13.17C14.5,13.17 14.67,13.5 15.17,13.5C15.67,13.5 15.5,14.17 15.5,14.17C15.5,14.17 16,13.5 16,13.5C16,13.5 16.17,13.67 16.5,14C16.17,14.33 16.5,13.67 16.83,14C17.33,14.17 17.5,13.17 18.17,13.17C19.5,13.17 19.83,13.5 20.17,13.83C19.83,13.17 19.67,13.5 20,13.83C19.67,14.17 20,13.5 20.33,13.83C20.67,13.5 21,13.17 21.33,13.17C21.67,13.17 22,14.5 22,14.5C22,14.5 21.83,13.17 22.17,13.17C22.17,13.17 22.33,13.5 22.33,13.5C22.33,13.5 21.83,13.17 21.5,13.17C21.17,13.17 20.83,13.5 20.5,13.83C20.83,13.5 20.5,13.17 20.17,13.17C19.83,13.17 19.67,13.5 19.33,13.17C19,13.5 18.67,13.17 18.33,13.17C17.67,13.17 17.5,13.5 17.17,13.83C17.5,13.5 17.17,13.17 16.83,13.5C16.5,13.17 16.17,13.5 15.83,13.17C15.5,13.5 15.17,13.17 14.83,13.5C14.5,13.17 14.17,13.5 13.83,13.17C13.5,13.5 13.17,13.17 12.83,13.5C12.5,13.17 12.17,13.5 11.83,13.17C11.5,13.5 11.17,13.17 10.83,13.5C10.5,13.17 10.17,13.5 9.83,13.17C9.5,13.5 9.17,13.17 8.83,13.5C8.5,13.17 8.17,13.5 7.83,13.17C7.5,13.5 7.17,13.17 6.83,13.5C6.5,13.17 6.17,13.5 5.83,13.17C5.5,13.5 5.17,13.17 4.83,13.5C4.5,13.17 4.17,13.5 3.83,13.17C3.5,13.5 3.17,13.17 2.83,13.5C2.5,13.17 2.17,13.5 1.83,13.17C1.5,13.5 1.17,13.17 0.83,13.5C0.5,13.17 0.17,13.5 -0.17,13.17C-0.5,13.5 -0.83,13.17 -1.17,13.5C-1.5,13.17 -1.83,13.5 -2.17,13.17C-2.5,13.5 -2.83,13.17 -3.17,13.5C-3.5,13.5 -3.83,13.17 -4.17,13.5C-4.5,13.5 -4.83,13.17 -5.17,13.5C-5.5,13.5 -5.83,13.17 -6.17,13.5C-6.5,13.5 -6.83,13.17 -7.17,13.5C-7.5,13.5 -7.83,13.17 -8.17,13.5C-8.5,13.5 -8.83,13.17 -9.17,13.5C-9.5,13.5 -9.83,13.17 -10.17,13.5C-10.5,13.5 -10.83,13.17 -11.17,13.5C-11.5,13.5 -11.83,13.17 -12.17,13.5C-12.5,13.5 -12.83,13.17 -13.17,13.5C-13.5,13.5 -13.83,13.17 -14.17,13.5C-14.5,13.5 -14.83,13.17 -15.17,13.5C-15.5,13.5 -15.83,13.17 -16.17,13.5C-16.5,13.5 -16.83,13.17 -17.17,13.5C-17.5,13.5 -17.83,13.17 -18.17,13.5C-18.5,13.5 -18.83,13.17 -19.17,13.5C-19.5,13.5 -19.83,13.17 -20.17,13.5C-20.5,13.5 -20.83,13.17 -21.17,13.5C-21.5,13.5 -21.83,13.17 -22.17,13.5Z" />
            </svg>
          </NIcon>
          Git配置
        </h4>

        <div class="grid grid-cols-1 gap-4">
          <NFormItem label="Git仓库地址" path="git_url">
            <NInput
              v-model:value="formData.git_url"
              placeholder="请输入Git仓库地址，如：https://github.com/user/project.git"
            />
          </NFormItem>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <NFormItem label="Git账号" path="git_username">
              <NInput
                v-model:value="formData.git_username"
                placeholder="请输入Git账号"
              />
            </NFormItem>

            <NFormItem label="Git密码/Token" path="git_password">
              <NInput
                v-model:value="formData.git_password"
                type="password"
                show-password-on="click"
                placeholder="请输入Git密码或访问Token"
              />
            </NFormItem>
          </div>

          <NFormItem label="Webhook URL">
            <NInput
              v-model:value="formData.webhook_url"
              placeholder="请输入Webhook接收地址"
            />
          </NFormItem>
        </div>
      </div>

      <NDivider class="my-6" />

      <!-- 消息通知配置 -->
      <div class="mb-6">
        <h4 class="text-base font-medium mb-4 flex items-center gap-2">
          <NIcon size="18" class="text-orange-500">
            <svg viewBox="0 0 24 24">
              <path fill="currentColor" d="M21,10.12H14.22L13.5,10.06L13.44,10.5H9.06L8.94,10.5L8.22,10.12H3V10.12M17.5,1.5C17.5,1.5 16.97,1.03 16.97,1.03C16.97,1.03 16.97,1.03 16.97,1.03C15.5,1.03 14.5,3.5 14.5,5.5C14.5,6.5 14.97,7.5 15.5,8.5C16.03,9.5 16.97,10.5 16.97,10.5C16.97,10.5 16.97,10.5 16.97,10.5C18.47,10.5 19.5,8.5 19.5,8.5C19.5,8.5 19.03,7.5 18.5,6.5C17.97,5.5 16.97,3.5 16.97,3.5C16.97,3.5 16.97,3.5 16.97,3.5C16.97,1.03 17.5,1.5 17.5,1.5M17.5,3.5C18.03,3.5 18.47,3.97 18.47,3.5C18.47,3.03 18.03,2.5 17.5,2.5C16.97,2.5 16.53,3.03 16.53,3.5C16.53,3.97 16.97,4.5 17.5,4.5C18.03,4.5 18.47,3.97 18.47,3.5C18.47,3.03 18.03,2.5 17.5,2.5M17.5,6.5C18.03,6.5 18.47,6.97 18.47,6.5C18.47,6.03 18.03,5.5 17.5,5.5C16.97,5.5 16.53,6.03 16.53,6.5C16.53,6.97 16.97,7.5 17.5,7.5C18.03,7.5 18.47,6.97 18.47,6.5C18.47,6.03 18.03,5.5 17.5,5.5M18,11H21V15H20V17H18V19H16V17H14V15H13V11H16V9H18V11Z"/>
            </svg>
          </NIcon>
          消息通知配置
        </h4>

        <div class="space-y-4">
          <NFormItem>
            <div class="flex items-center gap-2">
              <input
                type="checkbox"
                v-model="formData.notification_enabled"
                class="text-blue-500"
              />
              <span>启用消息通知</span>
            </div>
          </NFormItem>

          <div v-if="formData.notification_enabled">
            <NFormItem label="通知方式">
              <NCheckboxGroup v-model:value="formData.notification_types">
                <NCheckbox value="webhook">Webhook</NCheckbox>
                <NCheckbox value="dingtalk">钉钉</NCheckbox>
              </NCheckboxGroup>
            </NFormItem>

            <NFormItem v-if="formData.notification_types.includes('dingtalk')" label="钉钉Webhook地址">
              <NInput
                v-model:value="formData.dingtalk_webhook"
                placeholder="请输入钉钉机器人Webhook地址"
              />
            </NFormItem>
          </div>
        </div>
      </div>

      <NDivider class="my-6" />

      <!-- 部署配置 -->
      <div class="mb-6">
        <h4 class="text-base font-medium mb-4 flex items-center gap-2">
          <NIcon size="18" class="text-purple-500">
            <svg viewBox="0 0 24 24">
              <path fill="currentColor" d="M4,10V14H13L9.5,17.5L6,14H4M10,4H12L14.5,6.5L11,10H20L16.5,6.5L13,4H10Z"/>
            </svg>
          </NIcon>
          部署配置
        </h4>

        <div class="space-y-4">
          <div class="flex justify-between items-center">
            <span class="text-sm text-gray-600">部署步骤列表</span>
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

          <div v-if="formData.deployment_config.length === 0" class="text-center py-8 text-gray-400">
            暂无部署步骤，点击上方按钮添加
          </div>

          <div v-else class="space-y-3">
            <div v-for="(step, index) in formData.deployment_config" :key="step.id" class="border border-gray-200 rounded-lg p-4">
              <div class="flex justify-between items-start mb-3">
                <div class="flex-1">
                  <div class="flex items-center gap-2 mb-2">
                    <NTag :type="getStepTypeColor(step.type)" size="small">
                      {{ stepTypeOptions.find(opt => opt.value === step.type)?.label || step.type }}
                    </NTag>
                    <span v-if="!step.enabled" class="text-gray-400 text-xs line-through">已禁用</span>
                    <span class="font-medium text-sm">{{ step.name || `步骤 ${index + 1}` }}</span>
                    <NTag v-if="step.continue_on_error" type="warning" size="small">容错</NTag>
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <div>
                      <label class="text-xs text-gray-500">执行命令</label>
                      <div class="text-sm font-mono bg-gray-50 p-2 rounded mt-1 max-h-16 overflow-y-auto">
                        {{ step.command || '未设置命令' }}
                      </div>
                    </div>
                    <div>
                      <label class="text-xs text-gray-500">工作目录</label>
                      <div class="text-sm bg-gray-50 p-2 rounded mt-1">
                        {{ step.working_dir || '/app' }}
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center gap-4 mt-2">
                    <div class="flex items-center gap-1">
                      <span class="text-xs text-gray-500">超时:</span>
                      <span class="text-sm">{{ step.timeout }}秒</span>
                    </div>
                    <div v-if="step.target_hosts && step.target_hosts.length > 0" class="flex items-center gap-1">
                      <span class="text-xs text-gray-500">目标主机:</span>
                      <div class="flex gap-1">
                        <NTag v-for="host in step.target_hosts" :key="host" size="small">
                          {{ host }}
                        </NTag>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="flex gap-2">
                  <NButton size="small" @click="moveStep(index, 'up')" :disabled="index === 0">
                    上移
                  </NButton>
                  <NButton size="small" @click="moveStep(index, 'down')" :disabled="index === formData.deployment_config.length - 1">
                    下移
                  </NButton>
                  <NButton size="small" type="error" @click="removeDeploymentStep(index)">
                    删除
                  </NButton>
                </div>
              </div>
            </div>
            <div class="text-center py-2">
              <span class="text-xs text-gray-500">
                共 {{ formData.deployment_config.length }} 个步骤，预计执行时间: {{ getTotalExecutionTime() }}秒
              </span>
            </div>
          </div>
        </div>
      </div>
    </NForm>

    <!-- 操作按钮 -->
    <div class="flex justify-end gap-3 pt-6 border-t">
      <NButton @click="handleClose">取消</NButton>
      <NButton type="primary" @click="handleSubmit" :loading="loading">
        {{ isEdit ? '确认更新' : '确认添加' }}
      </NButton>
    </div>
  </div>
</template>

<style scoped>

</style>
