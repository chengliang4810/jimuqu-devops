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
  useMessage,
  useDialog,
  NSpin,
  NEmpty
} from 'naive-ui';
import { Plus, Copy } from '@vben/icons';
import type { DeployConfigItem } from '#/api/deploy-config';
import {
  getDeployConfigByProjectId,
  createDeployConfig,
  deleteDeployConfig,
  updateDeployConfig
} from '#/api/deploy-config';

const route = useRoute();
const message = useMessage();
const dialog = useDialog();

const projectInfo = ref({
  id: '',
  name: '',
  code: '',
});

// 分支相关
const branches = ref<Array<{ name: string; config: DeployConfigItem[]; id?: number }>>([]);
const activeTab = ref('main');
const showAddBranchModal = ref(false);
const newBranchName = ref('');
const showCopyBranchModal = ref(false);
const copySourceBranch = ref('');
const copyTargetBranch = ref('');

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
  }
});

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
      branches.value = response.map(config => ({
        name: config.branch,
        config: config.config || [],
        id: config.id
      }));

      // 设置默认选中的分支
      if (branches.value.length > 0) {
        activeTab.value = branches.value[0]?.name || 'main';
      }
    } else {
      // 如果没有配置，初始化默认分支
      branches.value = [
        { name: 'main', config: [] },
        { name: 'develop', config: [] }
      ];
      activeTab.value = 'main';
    }
  } catch (error) {
    console.error('加载部署配置失败:', error);
    message.error('加载部署配置失败');
    // 初始化默认分支
    branches.value = [
      { name: 'main', config: [] },
      { name: 'develop', config: [] }
    ];
    activeTab.value = 'main';
  } finally {
    loading.value = false;
  }
}

// 保存配置到后端
async function saveConfig(branchName: string, config: DeployConfigItem[]) {
  if (!projectInfo.value.id) return;

  try {
    const projectId = parseInt(projectInfo.value.id);
    const branchData = branches.value.find(b => b.name === branchName);

    if (branchData?.id) {
      // 更新现有配置
      await updateDeployConfig(branchData.id, {
        branch: branchName,
        config: config
      });
      message.success('配置保存成功');
    } else {
      // 创建新配置
      await createDeployConfig({
        project_id: projectId,
        branch: branchName,
        config: config
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
  branches.value.push({ name: newBranchName.value, config: [] });

  // 保存到后端
  await saveConfig(newBranchName.value, []);

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

  const sourceConfig = branches.value.find(b => b.name === copySourceBranch.value)?.config || [];

  // 添加到本地列表
  branches.value.push({ name: copyTargetBranch.value, config: [...sourceConfig] });

  // 保存到后端
  await saveConfig(copyTargetBranch.value, sourceConfig);

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
              <div class="py-32 text-center text-gray-400">
                <div class="text-6xl mb-4">📝</div>
                <div class="text-xl">分支 "{{ branch.name }}" 的配置内容正在开发中...</div>
                <div class="text-sm mt-2 text-gray-500">
                  配置项数量: {{ branch.config?.length || 0 }}
                </div>
              </div>
            </div>
          </NTabPane>
        </NTabs>
      </NCard>
    </div>

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
</style>
