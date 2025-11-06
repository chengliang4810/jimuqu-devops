<script setup lang="ts">
import { Page } from '@vben/common-ui';
import { useRoute } from 'vue-router';
import { onMounted, ref, watch } from 'vue';
import {
  NCard,
  NTabs,
  NTabPane,
  NButton,
  NIcon,
  NSpace,
  NInput,
  NModal,
  NForm,
  NFormItem,
  NSteps,
  NStep,
  useMessage,
  useDialog,
  NSpin,
  NEmpty,
  NSelect
} from 'naive-ui';
import { Plus, Copy } from '@vben/icons';
import type { DeployConfigContent } from '#/api/deploy-config';
import type { Host } from '#/api/host';
import {
  getDeployConfigByProjectId,
  createDeployConfig,
  deleteDeployConfig,
  updateDeployConfig
} from '#/api/deploy-config';
import { getHostList } from '#/api/host';

const route = useRoute();
const message = useMessage();
const dialog = useDialog();

const projectInfo = ref({
  id: '',
  name: '',
  code: '',
});

// 分支相关
const branches = ref<Array<{ name: string; config?: DeployConfigContent; id?: number }>>([]);
const activeTab = ref('main');
const showAddBranchModal = ref(false);
const newBranchName = ref('');
const showCopyBranchModal = ref(false);
const copySourceBranch = ref('');
const copyTargetBranch = ref('');

// 配置表单相关
const currentStep = ref(1);
const showConfigForm = ref(false);
const configForm = ref<DeployConfigContent>({
  compile: {
    docker_image: '',
    build_commands: []
  },
  deploy: {
    target_hosts: [],
    deploy_directory: '',
    pre_deploy_commands: [],
    post_deploy_commands: []
  }
});

// 主机列表
const hostList = ref<Host[]>([]);
const loadingHosts = ref(false);

// 加载状态
const loading = ref(false);

onMounted(() => {
  // 从路由参数中获取项目信息
  const { projectId, projectName, projectCode } = route.query;

  if (projectId && projectName && projectCode) {
    projectInfo.value = {
      id: projectId as string,
      name: projectName as string,
      code: projectCode as string,
    };
    // 加载部署配置
    loadDeployConfigs();
    // 加载主机列表
    loadHostList();
  }
});

// 加载主机列表
async function loadHostList() {
  loadingHosts.value = true;
  try {
    const response = await getHostList({ pageSize: 1000 });
    if (response && response.items) {
      hostList.value = response.items.filter(host => host.status === 'online' && !host.deleted_at);
    }
  } catch (error) {
    console.error('加载主机列表失败:', error);
    message.error('加载主机列表失败');
  } finally {
    loadingHosts.value = false;
  }
}

// 监听项目变化，重新加载配置
watch(() => projectInfo.value.id, (newId) => {
  if (newId) {
    loadDeployConfigs();
  }
});

// 加载部署配置
async function loadDeployConfigs() {
  if (!projectInfo.value.id) return;

  loading.value = true;
  try {
    const projectId = parseInt(projectInfo.value.id);
    const response = await getDeployConfigByProjectId(projectId);

    if (response && response.length > 0) {
      // 转换数据格式
      branches.value = response.map(config => {
        let deployContent: DeployConfigContent | undefined;

        // 如果config存在且是字符串格式，尝试解析YAML
        if (config.config && typeof config.config === 'string') {
          try {
            deployContent = JSON.parse(config.config);
          } catch (error) {
            console.error('解析配置失败:', error);
            deployContent = undefined;
          }
        } else if (config.config && Array.isArray(config.config) && config.config.length > 0) {
          // 兼容旧格式，从config数组中提取
          const configObj = config.config.reduce((acc, item) => {
            acc[item.key] = item.value;
            return acc;
          }, {} as any);

          deployContent = {
            compile: configObj.compile || { docker_image: '', build_commands: [] },
            deploy: configObj.deploy || {
              target_hosts: [],
              deploy_directory: '',
              pre_deploy_commands: [],
              post_deploy_commands: []
            }
          };
        }

        return {
          name: config.branch,
          config: deployContent,
          id: config.id
        };
      });

      // 设置默认选中的分支
      if (branches.value.length > 0) {
        activeTab.value = branches.value[0]?.name || 'main';
      }
    } else {
      // 如果没有配置，初始化默认分支
      branches.value = [
        { name: 'main', config: undefined },
        { name: 'develop', config: undefined }
      ];
      activeTab.value = 'main';
    }
  } catch (error) {
    console.error('加载部署配置失败:', error);
    message.error('加载部署配置失败');
    // 初始化默认分支
    branches.value = [
      { name: 'main', config: undefined },
      { name: 'develop', config: undefined }
    ];
    activeTab.value = 'main';
  } finally {
    loading.value = false;
  }
}

// 保存配置到后端
async function saveConfig(branchName: string, config: DeployConfigContent) {
  if (!projectInfo.value.id) return;

  try {
    const projectId = parseInt(projectInfo.value.id);
    const branchData = branches.value.find(b => b.name === branchName);

    // 将配置转换为YAML格式存储
    const configJson = JSON.stringify(config);

    if (branchData?.id) {
      // 更新现有配置 - 使用旧格式兼容
      await updateDeployConfig(branchData.id, {
        branch: branchName,
        config: [
          { key: 'content', value: configJson, desc: '部署配置内容' },
          { key: 'compile', value: config.compile, desc: '编译配置' },
          { key: 'deploy', value: config.deploy, desc: '部署配置' }
        ]
      });
      message.success('配置保存成功');
    } else {
      // 创建新配置
      await createDeployConfig({
        project_id: projectId,
        branch: branchName,
        config: [
          { key: 'content', value: configJson, desc: '部署配置内容' },
          { key: 'compile', value: config.compile, desc: '编译配置' },
          { key: 'deploy', value: config.deploy, desc: '部署配置' }
        ]
      });
      message.success('配置创建成功');
      // 重新加载配置以获取ID
      await loadDeployConfigs();
    }
  } catch (error) {
    console.error('保存配置失败:', error);
    message.error('保存配置失败');
  }
}

// 添加分支
async function handleAddBranch() {
  if (!newBranchName.value.trim()) {
    message.warning('请输入分支名称');
    return;
  }

  if (branches.value.some(b => b.name === newBranchName.value)) {
    message.warning('分支已存在');
    return;
  }

  // 添加到本地列表
  branches.value.push({ name: newBranchName.value, config: undefined });

  showAddBranchModal.value = false;
  newBranchName.value = '';
  activeTab.value = newBranchName.value;
}

// 复制分支
async function handleCopyBranch() {
  if (!copyTargetBranch.value.trim()) {
    message.warning('请输入目标分支名称');
    return;
  }

  if (branches.value.some(b => b.name === copyTargetBranch.value)) {
    message.warning('目标分支已存在');
    return;
  }

  const sourceConfig = branches.value.find(b => b.name === copySourceBranch.value)?.config;

  // 添加到本地列表
  branches.value.push({
    name: copyTargetBranch.value,
    config: sourceConfig ? JSON.parse(JSON.stringify(sourceConfig)) : undefined
  });

  showCopyBranchModal.value = false;
  copyTargetBranch.value = '';
  activeTab.value = copyTargetBranch.value;
}

// 删除分支
async function handleDeleteBranch(branchName: string) {
  if (branches.value.length <= 1) {
    message.warning('至少需要保留一个分支');
    return;
  }

  const branchData = branches.value.find(b => b.name === branchName);

  dialog.warning({
    title: '确认删除',
    content: `确定要删除分支 "${branchName}" 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        // 如果有后端ID，删除后端数据
        if (branchData?.id) {
          await deleteDeployConfig(branchData.id);
        }

        // 从本地列表移除
        branches.value = branches.value.filter(b => b.name !== branchName);

        // 如果删除的是当前激活的分支，切换到第一个分支
        if (activeTab.value === branchName) {
          activeTab.value = branches.value[0]?.name || '';
        }

        message.success('分支删除成功');
      } catch (error) {
        console.error('删除分支失败:', error);
        message.error('删除分支失败');
      }
    }
  });
}

// 配置表单相关函数
function openConfigForm(branchName: string) {
  const branch = branches.value.find(b => b.name === branchName);
  if (branch?.config) {
    configForm.value = JSON.parse(JSON.stringify(branch.config));
  } else {
    configForm.value = {
      compile: {
        docker_image: '',
        build_commands: []
      },
      deploy: {
        target_hosts: [],
        deploy_directory: '',
        pre_deploy_commands: [],
        post_deploy_commands: []
      }
    };
  }
  currentStep.value = 1;
  showConfigForm.value = true;
}

function closeConfigForm() {
  showConfigForm.value = false;
  currentStep.value = 1;
}

function nextStep() {
  if (currentStep.value < 2) {
    currentStep.value++;
  }
}

function prevStep() {
  if (currentStep.value > 1) {
    currentStep.value--;
  }
}

async function saveConfigForm() {
  const branchName = activeTab.value;
  await saveConfig(branchName, configForm.value);
  closeConfigForm();
}

// 添加构建命令
function addBuildCommand() {
  configForm.value.compile.build_commands.push('');
}

function removeBuildCommand(index: number) {
  configForm.value.compile.build_commands.splice(index, 1);
}

// 添加部署前命令
function addPreDeployCommand() {
  configForm.value.deploy.pre_deploy_commands.push('');
}

function removePreDeployCommand(index: number) {
  configForm.value.deploy.pre_deploy_commands.splice(index, 1);
}

// 添加部署后命令
function addPostDeployCommand() {
  configForm.value.deploy.post_deploy_commands.push('');
}

function removePostDeployCommand(index: number) {
  configForm.value.deploy.post_deploy_commands.splice(index, 1);
}
</script>

<template>
  <Page auto-content-height>
    <div class="deploy-config">
      <!-- 部署配置标题和操作按钮 -->
      <NCard
        :title="`部署配置 - ${projectInfo.name} (${projectInfo.code})`"
        class="mb-4"
      >
        <template #header-extra>
          <NSpace>
            <NButton type="primary" size="small" @click="showAddBranchModal = true" :disabled="loading">
              <template #icon>
                <NIcon :component="Plus" />
              </template>
              添加分支
            </NButton>
          </NSpace>
        </template>

        <!-- 加载状态 -->
        <div v-if="loading" class="flex justify-center items-center py-32">
          <NSpin size="large" />
        </div>

        <!-- 无配置状态 -->
        <div v-else-if="branches.length === 0" class="py-32">
          <NEmpty description="暂无部署配置">
            <template #extra>
              <NButton type="primary" @click="showAddBranchModal = true">
                创建第一个配置
              </NButton>
            </template>
          </NEmpty>
        </div>

        <!-- 分支配置Tab -->
        <NTabs
          v-else
          v-model:value="activeTab"
          type="card"
          placement="left"
          tab-style="min-width: 140px; max-width: 160px;"
        >
          <NTabPane
            v-for="branch in branches"
            :key="branch.name"
            :name="branch.name"
            :tab="branch.name"
          >
            <template #tab>
              <div class="branch-tab">
                <div class="branch-name" :title="branch.name">
                  {{ branch.name }}
                </div>
                <div class="branch-actions">
                  <NButton
                    text
                    type="primary"
                    size="tiny"
                    @click.stop="copySourceBranch = branch.name; copyTargetBranch = ''; showCopyBranchModal = true"
                  >
                    <template #icon>
                      <NIcon :component="Copy" />
                    </template>
                  </NButton>
                  <NButton
                    v-if="branches.length > 1"
                    text
                    type="error"
                    size="tiny"
                    @click.stop="handleDeleteBranch(branch.name)"
                  >
                    删除
                  </NButton>
                </div>
              </div>
            </template>

            <!-- 分支配置内容 -->
            <div class="branch-config-content">
              <!-- 配置表单 -->
              <div v-if="branch.config" class="config-display">
                <NCard title="当前配置" class="mb-4">
                  <div class="config-summary">
                    <div class="config-section">
                      <h4>编译配置</h4>
                      <p><strong>Docker镜像:</strong> {{ branch.config.compile.docker_image || '未设置' }}</p>
                      <p><strong>构建命令:</strong> {{ branch.config.compile.build_commands.length }} 条</p>
                    </div>
                    <div class="config-section">
                      <h4>部署配置</h4>
                      <p><strong>目标主机:</strong> {{ branch.config.deploy.target_hosts.length }} 台</p>
                      <p><strong>部署目录:</strong> {{ branch.config.deploy.deploy_directory || '未设置' }}</p>
                      <p><strong>部署前命令:</strong> {{ branch.config.deploy.pre_deploy_commands.length }} 条</p>
                      <p><strong>部署后命令:</strong> {{ branch.config.deploy.post_deploy_commands.length }} 条</p>
                    </div>
                  </div>
                  <div class="mt-4">
                    <NButton type="primary" @click="openConfigForm(branch.name)">编辑配置</NButton>
                  </div>
                </NCard>
              </div>

              <!-- 无配置状态 -->
              <div v-else class="py-32 text-center text-gray-400">
                <div class="text-6xl mb-4">⚙️</div>
                <div class="text-xl">分支 "{{ branch.name }}" 尚未配置部署参数</div>
                <div class="text-sm mt-2 text-gray-500">
                  点击下方按钮开始配置部署流程
                </div>
                <div class="mt-6">
                  <NButton type="primary" size="large" @click="openConfigForm(branch.name)">
                    开始配置
                  </NButton>
                </div>
              </div>
            </div>
          </NTabPane>
        </NTabs>
      </NCard>
    </div>

    <!-- 配置表单弹窗 -->
    <NModal
      v-model:show="showConfigForm"
      preset="card"
      title="配置部署流程"
      style="width: 800px"
      :closable="false"
      :mask-closable="false"
    >
      <NSteps :current="currentStep" class="mb-6">
        <NStep title="编译配置" description="设置Docker镜像和构建命令" />
        <NStep title="部署配置" description="选择目标主机和部署参数" />
      </NSteps>

      <!-- 步骤1: 编译配置 -->
      <div v-if="currentStep === 1">
        <NForm>
          <NFormItem label="Docker镜像名称" required>
            <NInput
              v-model:value="configForm.compile.docker_image"
              placeholder="例如: node:18-alpine, golang:1.21-alpine"
            />
          </NFormItem>

          <NFormItem label="构建命令">
            <div class="w-full">
              <div
                v-for="(_, index) in configForm.compile.build_commands"
                :key="index"
                class="mb-2"
              >
                <div class="flex gap-2">
                  <NInput
                    v-model:value="configForm.compile.build_commands[index]"
                    placeholder="输入构建命令，例如: npm run build"
                    class="flex-1"
                  />
                  <NButton
                    type="error"
                    size="small"
                    @click="removeBuildCommand(index)"
                    :disabled="configForm.compile.build_commands.length <= 1"
                  >
                    删除
                  </NButton>
                </div>
              </div>
              <NButton dashed @click="addBuildCommand" class="w-full">
                添加构建命令
              </NButton>
            </div>
          </NFormItem>
        </NForm>
      </div>

      <!-- 步骤2: 部署配置 -->
      <div v-if="currentStep === 2">
        <NForm>
          <NFormItem label="目标主机" required>
            <NSelect
              v-model:value="configForm.deploy.target_hosts"
              :options="hostList.map(host => ({
                label: `${host.name} (${host.host})`,
                value: host.id
              }))"
              multiple
              placeholder="选择要部署到的主机"
              :loading="loadingHosts"
            />
          </NFormItem>

          <NFormItem label="部署目录" required>
            <NInput
              v-model:value="configForm.deploy.deploy_directory"
              placeholder="例如: /var/www/app, /opt/myapp"
            />
          </NFormItem>

          <NFormItem label="部署前执行的命令">
            <div class="w-full">
              <div
                v-for="(_, index) in configForm.deploy.pre_deploy_commands"
                :key="index"
                class="mb-2"
              >
                <div class="flex gap-2">
                  <NInput
                    v-model:value="configForm.deploy.pre_deploy_commands[index]"
                    placeholder="输入部署前执行的命令，例如: systemctl stop nginx"
                    class="flex-1"
                  />
                  <NButton
                    type="error"
                    size="small"
                    @click="removePreDeployCommand(index)"
                  >
                    删除
                  </NButton>
                </div>
              </div>
              <NButton dashed @click="addPreDeployCommand" class="w-full">
                添加部署前命令
              </NButton>
            </div>
          </NFormItem>

          <NFormItem label="部署后执行的命令">
            <div class="w-full">
              <div
                v-for="(_, index) in configForm.deploy.post_deploy_commands"
                :key="index"
                class="mb-2"
              >
                <div class="flex gap-2">
                  <NInput
                    v-model:value="configForm.deploy.post_deploy_commands[index]"
                    placeholder="输入部署后执行的命令，例如: systemctl start nginx"
                    class="flex-1"
                  />
                  <NButton
                    type="error"
                    size="small"
                    @click="removePostDeployCommand(index)"
                  >
                    删除
                  </NButton>
                </div>
              </div>
              <NButton dashed @click="addPostDeployCommand" class="w-full">
                添加部署后命令
              </NButton>
            </div>
          </NFormItem>
        </NForm>
      </div>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="closeConfigForm">取消</NButton>
          <NButton v-if="currentStep > 1" @click="prevStep">上一步</NButton>
          <NButton v-if="currentStep < 2" type="primary" @click="nextStep">下一步</NButton>
          <NButton v-if="currentStep === 2" type="primary" @click="saveConfigForm">保存配置</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 添加分支弹窗 -->
    <NModal
      v-model:show="showAddBranchModal"
      preset="card"
      title="添加分支"
      style="width: 400px"
    >
      <NForm>
        <NFormItem label="分支名称">
          <NInput
            v-model:value="newBranchName"
            placeholder="请输入分支名称"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showAddBranchModal = false">取消</NButton>
          <NButton type="primary" @click="handleAddBranch">确定</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 复制分支弹窗 -->
    <NModal
      v-model:show="showCopyBranchModal"
      preset="card"
      title="复制分支"
      style="width: 400px"
    >
      <NForm>
        <NFormItem label="源分支">
          <NInput
            v-model:value="copySourceBranch"
            readonly
          />
        </NFormItem>
        <NFormItem label="目标分支名称">
          <NInput
            v-model:value="copyTargetBranch"
            placeholder="请输入目标分支名称"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showCopyBranchModal = false">取消</NButton>
          <NButton type="primary" @click="handleCopyBranch">确定</NButton>
        </NSpace>
      </template>
    </NModal>
  </Page>
</template>

<style scoped>
.branch-tab {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  padding: 4px 0;
}

.branch-name {
  font-size: 12px;
  font-weight: 500;
  text-align: center;
  word-break: break-all;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  max-width: 100%;
  line-height: 1.2;
  margin-bottom: 4px;
  min-height: 14px;
}

.branch-actions {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
  opacity: 0.8;
  transition: opacity 0.2s;
}

.branch-actions:hover {
  opacity: 1;
}

/* 当Tab激活时，操作按钮更明显 */
:deep(.n-tabs-tab--active .branch-actions) {
  opacity: 1;
}

/* 确保Tab内容不会被挤压 */
:deep(.n-tabs-tab) {
  padding: 8px 12px !important;
}

/* 调整Tab卡片样式 */
:deep(.n-tabs--left .n-tabs-tab) {
  justify-content: center;
}

/* 分支配置内容区域 */
.branch-config-content {
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 配置显示样式 */
.config-display {
  width: 100%;
}

.config-summary {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.config-section {
  padding: 16px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background-color: #fafafa;
}

.config-section h4 {
  margin: 0 0 12px 0;
  color: #333;
  font-size: 16px;
  font-weight: 600;
  border-bottom: 2px solid #18a058;
  padding-bottom: 4px;
}

.config-section p {
  margin: 6px 0;
  color: #666;
  font-size: 14px;
}

.config-section strong {
  color: #333;
  font-weight: 600;
}
</style>
