<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import { NButton, NIcon, NForm, NFormItem, NInput, NInputNumber, NDivider, NAlert } from 'naive-ui';
import { useMessage } from 'naive-ui';

interface Props {
  hostData?: any;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  submit: [data: any];
}>();

const message = useMessage();
const formRef = ref();
const loading = ref(false);
const isEdit = ref(!!props.hostData);

const formData = reactive({
  name: '',
  host: '',
  port: 22,
  username: '',
  ssh_key_path: '',
  ssh_password: '',
  use_password: true,
  description: ''
});

const rules = {
  name: { required: true, message: '请输入主机名称', trigger: 'blur' },
  host: { required: true, message: '请输入主机地址', trigger: 'blur' },
  port: { required: true, type: 'number', message: '请输入端口号', trigger: 'blur' },
  username: { required: true, message: '请输入用户名', trigger: 'blur' }
};

// 监听props变化，初始化表单数据
watch(() => props.hostData, (newData) => {
  if (newData) {
    isEdit.value = true;
    Object.assign(formData, {
      name: newData.name || '',
      host: newData.host || '',
      port: newData.port || 22,
      username: newData.username || '',
      ssh_key_path: newData.ssh_key_path || '',
      ssh_password: newData.ssh_password || '',
      use_password: !!newData.ssh_password,
      description: newData.description || ''
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

    if (!formData.use_password && !formData.ssh_key_path) {
      message.error('请选择SSH密钥路径或设置密码');
      return;
    }

    if (formData.use_password && !formData.ssh_password) {
      message.error('请输入SSH密码');
      return;
    }

    loading.value = true;

    // 第一步：验证连接
    message.info('正在验证主机连接...');

    try {
      // 模拟连接验证
      await new Promise((resolve, reject) => {
        setTimeout(() => {
          const success = Math.random() > 0.2; // 80% 成功率
          if (success) {
            resolve(true);
          } else {
            reject(new Error('连接失败：无法连接到主机'));
          }
        }, 2000);
      });

      // 第二步：连接成功后添加到数据库
      const submitData = {
        name: formData.name,
        description: formData.description || '',
        host: formData.host,
        port: formData.port,
        username: formData.username,
        ssh_key_path: formData.use_password ? null : formData.ssh_key_path,
        ssh_password: formData.use_password ? formData.ssh_password : null
      };

      emit('submit', submitData);
    } catch (error) {
      message.error(error instanceof Error ? error.message : '主机连接验证失败');
    }
  } catch (error) {
    console.error('表单验证失败:', error);
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  Object.assign(formData, {
    name: '',
    host: '',
    port: 22,
    username: '',
    ssh_key_path: '',
    ssh_password: '',
    use_password: true,
    description: ''
  });
  formRef.value?.restoreValidation();
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
              <path fill="currentColor" d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10 10-4.5 10-10S17.5 2 12 2m0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8m0-14c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6m0 10c-2.21 0-4-1.79-4-4s1.79-4 4-4 4 1.79 4 4-1.79 4-4 4Z"/>
            </svg>
          </NIcon>
          基本信息
        </h4>

        <div class="grid grid-cols-1 gap-4">
          <NFormItem label="主机名称" path="name">
            <NInput
              v-model:value="formData.name"
              placeholder="请输入主机名称，如：生产服务器-01"
              maxlength="100"
              show-count
            />
          </NFormItem>

          <NFormItem label="描述">
            <NInput
              v-model:value="formData.description"
              type="textarea"
              placeholder="请输入主机描述信息（可选）"
              :autosize="{ minRows: 2, maxRows: 4 }"
              maxlength="500"
              show-count
            />
          </NFormItem>
        </div>
      </div>

      <NDivider class="my-6" />

      <!-- 连接信息 -->
      <div class="mb-6">
        <h4 class="text-base font-medium mb-4 flex items-center gap-2">
          <NIcon size="18" class="text-green-500">
            <svg viewBox="0 0 24 24">
              <path fill="currentColor" d="M12 1L3 5V11C3 16.55 6.84 21.74 12 23C17.16 21.74 21 16.55 21 11V5L12 1M12 5A3 3 0 0 0 9 8A3 3 0 0 0 12 11A3 3 0 0 0 15 8A3 3 0 0 0 12 5M17.13 17C15.92 18.85 14.11 20.24 12 20.92C9.89 20.24 8.08 18.85 6.87 17C6.53 16.5 6.24 16 6 15.47C6 13.82 8.71 12.47 12 12.47C15.29 12.47 18 13.79 18 15.47C17.76 16 17.47 16.5 17.13 17Z"/>
            </svg>
          </NIcon>
          连接信息
        </h4>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <NFormItem label="主机地址" path="host">
            <NInput
              v-model:value="formData.host"
              placeholder="请输入IP地址或域名，如：192.168.1.100"
            />
          </NFormItem>

          <NFormItem label="SSH端口" path="port">
            <NInputNumber
              v-model:value="formData.port"
              :min="1"
              :max="65535"
              placeholder="默认22"
              class="w-full"
            />
          </NFormItem>

          <NFormItem label="用户名" path="username">
            <NInput
              v-model:value="formData.username"
              placeholder="请输入SSH用户名，如：root、deploy"
            />
          </NFormItem>
        </div>
      </div>

      <NDivider class="my-6" />

      <!-- SSH认证 -->
      <div class="mb-6">
        <h4 class="text-base font-medium mb-4 flex items-center gap-2">
          <NIcon size="18" class="text-orange-500">
            <svg viewBox="0 0 24 24">
              <path fill="currentColor" d="M12 17C10.89 17 10 16.11 10 15C10 13.89 10.89 13 12 13C13.11 13 14 13.89 14 15C14 16.11 13.11 17 12 17M8 9H16C16 5.69 13.31 3 10 3H8V9M6 10V3C6 1.89 6.89 1 8 1H10C13.31 1 16 3.69 16 7V9C16 10.11 16.89 11 18 11V13C18 13.74 17.81 14.39 17.44 14.83C16.78 15.54 16.33 16.44 16.33 17.44C16.33 18.31 16.67 19.1 17.21 19.64C17.76 20.19 18.55 20.53 19.41 20.53C21.39 20.53 23 18.92 23 16.94C23 14.9 21.53 13.26 19.54 13.05C19.84 12.45 20 11.75 20 11H19V10H6M20 17.44C20 18.28 19.28 19 18.44 19C17.6 19 16.89 18.28 16.89 17.44C16.89 16.6 17.6 15.89 18.44 15.89C19.28 15.89 20 16.6 20 17.44Z"/>
            </svg>
          </NIcon>
          SSH认证配置
        </h4>

        <div class="mb-4">
          <div class="flex gap-4 mb-4">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                v-model="formData.use_password"
                :value="true"
                class="text-blue-500"
              />
              <span>密码认证</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                v-model="formData.use_password"
                :value="false"
                class="text-blue-500"
              />
              <span>SSH密钥认证（推荐）</span>
            </label>
          </div>
        </div>

        <!-- 密码认证 -->
        <div v-if="formData.use_password" class="space-y-4">
          <NFormItem label="SSH密码">
            <NInput
              v-model:value="formData.ssh_password"
              type="password"
              show-password-on="click"
              placeholder="请输入SSH密码"
            />
          </NFormItem>

          <NAlert type="warning" show-icon>
            <template #icon>
              <NIcon>
                <svg viewBox="0 0 24 24">
                  <path fill="currentColor" d="M12 2L10.11 8.26L3 9L8 14.14L6.89 21L12 17.77L17.11 21L16 14.14L21 9L13.89 8.26L12 2M17.5 8A1.5 1.5 0 0 1 16 6.5A1.5 1.5 0 0 1 17.5 5A1.5 1.5 0 0 1 19 6.5A1.5 1.5 0 0 1 17.5 8M14.5 11A1.5 1.5 0 0 1 13 9.5A1.5 1.5 0 0 1 14.5 8A1.5 1.5 0 0 1 16 9.5A1.5 1.5 0 0 1 14.5 11M20.5 14A1.5 1.5 0 0 1 19 12.5A1.5 1.5 0 0 1 20.5 11A1.5 1.5 0 0 1 22 12.5A1.5 1.5 0 0 1 20.5 14Z"/>
                </svg>
              </NIcon>
            </template>
            密码认证方式相对不够安全，建议仅在测试环境或无法使用密钥认证时使用。
          </NAlert>
        </div>

        <!-- SSH密钥认证 -->
        <div v-else class="space-y-4">
          <NFormItem label="SSH私钥内容">
            <NInput
              v-model:value="formData.ssh_key_path"
              type="textarea"
              placeholder="请粘贴SSH私钥内容，如：-----BEGIN RSA PRIVATE KEY-----"
              :autosize="{ minRows: 4, maxRows: 8 }"
              show-count
            />
          </NFormItem>

          <NAlert type="info" show-icon>
            <template #icon>
              <NIcon>
                <svg viewBox="0 0 24 24">
                  <path fill="currentColor" d="M13 9H11V7H13M13 17H11V11H13M12 2C17.53 2 22 6.47 22 12C22 17.53 17.53 22 12 22C6.47 22 2 17.53 2 12C2 6.47 6.47 2 12 2Z"/>
                </svg>
              </NIcon>
            </template>
            使用SSH密钥认证更安全，建议优先选择此方式。请直接粘贴私钥内容，系统会安全保存。
          </NAlert>
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
input[type="radio"]:checked {
  background-color: #3b82f6;
  border-color: #3b82f6;
}
</style>