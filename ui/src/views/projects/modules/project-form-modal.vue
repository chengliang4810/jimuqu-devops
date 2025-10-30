<script setup lang="ts">
import { ref, watch } from 'vue';
import { NButton, NSteps, NStep, NIcon, NInput, NSelect, NCheckbox, NCheckboxGroup, NCard, NSpace, NForm, NFormItem, NTag } from 'naive-ui';
import { createProForm, zhCN } from 'pro-naive-ui';
import { useMessage } from 'naive-ui';
import ConfigProvider from '@/views/pro-naive/ConfigProvider.vue';

interface Props {
  projectData?: any;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  submit: [data: any];
}>();

const message = useMessage();
const step = ref(1);
const submiting = ref(false);

// 创建Pro Naive UI表单
const basicForm = createProForm({
  initialValues: {
    branch: 'main',
    notification_enabled: false,
    notification_types: []
  },
  onReset: () => {
    message.success('重置成功');
  }
});

const notificationForm = createProForm({
  onSubmit: async (values) => {
    submiting.value = true;
    try {
      // 合并所有表单数据
      const submitData = {
        ...basicForm.values.value,
        ...values
      };

      // 模拟提交延迟
      await new Promise(resolve => setTimeout(resolve, 1000));

      emit('submit', submitData);
    } finally {
      submiting.value = false;
    }
  }
});

// 监听props变化，初始化表单数据
watch(() => props.projectData, (newData) => {
  if (newData) {
    Object.assign(basicForm.values.value, {
      name: newData.name || '',
      identifier: newData.identifier || '',
      branch: newData.branch || 'main',
      description: newData.description || '',
      git_url: newData.git_url || '',
      git_username: newData.git_username || '',
      git_password: newData.git_password || '',
      webhook_url: newData.webhook_url || `https://api.jimuqu.com/webhook/${newData.identifier || 'project'}`
    });

    Object.assign(notificationForm.values.value, {
      notification_enabled: newData.notification?.enabled || false,
      notification_types: newData.notification?.types || [],
      dingtalk_webhook: newData.notification?.dingtalk_webhook || ''
    });
  } else {
    // 重置表单
    basicForm.values.value = {
      branch: 'main',
      notification_enabled: false,
      notification_types: []
    };
    notificationForm.values.value = {};
  }
}, { immediate: true });

// 表单验证规则
const basicRules = {
  name: { required: true, message: '请输入项目名称' },
  identifier: { required: true, message: '请输入项目标识符' },
  branch: { required: true, message: '请选择或输入分支名称' },
  git_url: { required: true, message: '请输入Git仓库地址' },
  git_username: { required: true, message: '请输入Git账号' },
  git_password: { required: true, message: '请输入Git密码或Token' }
};

const notificationRules = {
  dingtalk_webhook: {
    required: false,
    validator: (rule: any, value: string) => {
      if (notificationForm.values.value.notification_enabled &&
          notificationForm.values.value.notification_types?.includes('dingtalk') &&
          !value) {
        return new Error('请输入钉钉Webhook地址');
      }
      return true;
    }
  }
};

// 分支选项
const branchOptions = [
  { label: 'main', value: 'main' },
  { label: 'master', value: 'master' },
  { label: 'develop', value: 'develop' },
  { label: 'dev', value: 'dev' },
  { label: 'release', value: 'release' },
  { label: 'feature/*', value: 'feature/*' },
  { label: 'hotfix/*', value: 'hotfix/*' }
];

// 通知类型选项
const notificationOptions = [
  { label: 'Webhook', value: 'webhook' },
  { label: '钉钉', value: 'dingtalk' }
];

function handleClose() {
  emit('close');
}

function toNextStepAfterValidated() {
  // 手动验证必填字段
  const name = basicForm.values.value.name;
  const identifier = basicForm.values.value.identifier;
  const branch = basicForm.values.value.branch;
  const git_url = basicForm.values.value.git_url;
  const git_username = basicForm.values.value.git_username;
  const git_password = basicForm.values.value.git_password;

  if (!name) {
    message.error('请输入项目名称');
    return;
  }
  if (!identifier) {
    message.error('请输入项目标识符');
    return;
  }
  if (!branch) {
    message.error('请选择或输入分支名称');
    return;
  }
  if (!git_url) {
    message.error('请输入Git仓库地址');
    return;
  }
  if (!git_username) {
    message.error('请输入Git账号');
    return;
  }
  if (!git_password) {
    message.error('请输入Git密码或Token');
    return;
  }

  step.value += 1;
}

function toPreviousStep() {
  step.value -= 1;
}

async function handleSubmit() {
  try {
    await notificationForm.validate(notificationRules);

    // 验证通知配置
    const notificationValues = notificationForm.values.value;
    if (notificationValues.notification_enabled && notificationValues.notification_types.length === 0) {
      message.error('请选择至少一种通知方式');
      return;
    }

    await notificationForm.onSubmit(notificationValues);
  } catch (error) {
    message.error('请完善通知配置');
  }
}
</script>

<template>
  <ConfigProvider :locale="zhCN">
    <div class="color-#000">
      <div class="flex flex-col items-center justify-center">
        <div class="flex justify-center w-full mb-8">
          <NSteps :current="step" class="max-w-2xl">
            <NStep title="基本配置" />
            <NStep title="通知配置" />
          </NSteps>
        </div>

        <!-- 第一步：基本配置 -->
        <template v-if="step === 1">
          <div class="w-full max-w-4xl">
            <!-- 基本信息 -->
            <NCard title="基本信息" class="mb-6">

              <NForm :model="basicForm.values.value" label-placement="left" label-width="100px">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <NFormItem label="项目名称" required>
                    <NInput
                      v-model:value="basicForm.values.value.name"
                      placeholder="请输入项目名称，如：Web前端项目"
                      maxlength="100"
                      show-count
                    />
                  </NFormItem>

                  <NFormItem label="项目标识符" required>
                    <NInput
                      v-model:value="basicForm.values.value.identifier"
                      placeholder="请输入项目标识符，如：web-frontend"
                      maxlength="50"
                      show-count
                    />
                  </NFormItem>

                  <NFormItem label="默认分支" required>
                    <NSelect
                      v-model:value="basicForm.values.value.branch"
                      placeholder="请选择默认分支"
                      :options="branchOptions"
                    />
                  </NFormItem>

                  <NFormItem label="描述">
                    <NInput
                      v-model:value="basicForm.values.value.description"
                      type="textarea"
                      placeholder="请输入项目描述信息"
                      :autosize="{ minRows: 2, maxRows: 4 }"
                      maxlength="500"
                      show-count
                    />
                  </NFormItem>
                </div>
              </NForm>
            </NCard>

            <!-- Git配置 -->
            <NCard title="Git配置" class="mb-6">

              <NSpace vertical size="large">
                <NFormItem label="Git仓库地址" required>
                  <NInput
                    v-model:value="basicForm.values.value.git_url"
                    placeholder="请输入Git仓库地址，如：https://github.com/user/project.git"
                  />
                </NFormItem>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <NFormItem label="Git账号" required>
                    <NInput
                      v-model:value="basicForm.values.value.git_username"
                      placeholder="请输入Git账号"
                    />
                  </NFormItem>

                  <NFormItem label="Git密码/Token" required>
                    <NInput
                      v-model:value="basicForm.values.value.git_password"
                      type="password"
                      placeholder="请输入Git密码或访问Token"
                      show-password-on="click"
                    />
                  </NFormItem>
                </div>

              </NSpace>
            </NCard>

            <div class="flex justify-end gap-3 pt-6">
              <NButton @click="handleClose">取消</NButton>
              <NButton type="primary" @click="toNextStepAfterValidated">
                下一步：通知配置
              </NButton>
            </div>
          </div>
        </template>

        <!-- 第二步：通知配置 -->
        <template v-if="step === 2">
          <div class="w-full max-w-4xl">
            <NCard title="消息通知配置">

              <NSpace vertical size="large">
                <!-- 启用消息通知 -->
                <NFormItem>
                  <NSpace align="center">
                    <NCheckbox
                      v-model:checked="notificationForm.values.value.notification_enabled"
                    />
                    <span class="text-base font-medium">启用消息通知</span>
                  </NSpace>
                </NFormItem>

                <!-- 通知方式 -->
                <template v-if="notificationForm.values.value.notification_enabled">
                  <NFormItem label="通知方式" required>
                    <NCheckboxGroup
                      v-model:value="notificationForm.values.value.notification_types"
                    >
                      <NSpace vertical>
                        <NCheckbox
                          v-for="option in notificationOptions"
                          :key="option.value"
                          :value="option.value"
                          :label="option.label"
                        />
                      </NSpace>
                    </NCheckboxGroup>
                  </NFormItem>

                  <!-- 钉钉Webhook地址 -->
                  <template v-if="notificationForm.values.value.notification_types?.includes('dingtalk')">
                    <NFormItem label="钉钉Webhook地址" required>
                      <NInput
                        v-model:value="notificationForm.values.value.dingtalk_webhook"
                        type="textarea"
                        placeholder="请输入钉钉机器人Webhook地址"
                        :autosize="{ minRows: 2, maxRows: 4 }"
                      />
                    </NFormItem>
                  </template>
                </template>
              </NSpace>

              <div class="flex justify-end gap-3 pt-6 border-t">
                <NButton @click="toPreviousStep">上一步</NButton>
                <NButton type="primary" @click="handleSubmit" :loading="submiting">
                  确认添加
                </NButton>
              </div>
            </NCard>
          </div>
        </template>
      </div>
    </div>
  </ConfigProvider>
</template>
