<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { getProjectList } from '#/api/project'
import { getDeployConfigByProjectId } from '#/api/deploy-config'
import { createDeployRecord } from '#/api/deploy-record'
import type { Project } from '#/api/project'
import type { CreateDeployRecordParams } from '#/api/deploy-record'
import {
  NModal,
  NForm,
  NFormItem,
  NSelect,
  NInput,
  NButton,
  NSpin,
  useMessage,
  useDialog
} from 'naive-ui'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
})

const emit = defineEmits<Emits>()
const message = useMessage()
const dialog = useDialog()

// 表单引用
const formRef = ref<any>(null)

// 表单数据
const formData = ref<Partial<CreateDeployRecordParams>>({
  projectId: undefined,
  projectName: '',
  branch: '',
  startTime: new Date().toLocaleString('zh-CN', {
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).replace(/\//g, '-'),
  status: 'running'
})

// 项目列表
const projectList = ref<Project[]>([])
const projectOptions = computed(() =>
  projectList.value.map(item => ({
    label: item.name,
    value: item.id,
    project: item
  }))
)

// 分支列表
const branchList = ref<string[]>([])
const branchOptions = computed(() =>
  branchList.value.map(branch => ({
    label: branch,
    value: branch
  }))
)

// 加载状态
const projectLoading = ref(false)
const configLoading = ref(false)

// 表单验证规则
const formRules = {
  projectId: {
    required: true,
    message: '请选择项目',
    trigger: ['change', 'blur']
  },
  branch: {
    required: true,
    message: '请选择分支',
    trigger: ['change', 'blur']
  }
}

// 监听显示状态变化
watch(() => props.show, (newVal) => {
  if (newVal) {
    loadProjects()
    resetForm()
  }
})

// 加载项目列表
async function loadProjects() {
  try {
    projectLoading.value = true
    const response = await getProjectList({ page: 1, pageSize: 1000 })
    projectList.value = response.list || []
  } catch (error) {
    console.error('加载项目列表失败:', error)
    message.error('加载项目列表失败')
  } finally {
    projectLoading.value = false
  }
}

// 项目变化处理
async function handleProjectChange(projectId: number) {
  if (!projectId) {
    formData.value.projectName = ''
    branchList.value = []
    return
  }

  try {
    // 查找项目名称（用于提交时使用）
    const project = projectList.value.find(item => item.id === projectId)
    if (project) {
      formData.value.projectName = project.name
    }

    // 加载部署配置获取分支列表
    configLoading.value = true
    const configs = await getDeployConfigByProjectId(projectId)

    // 提取所有分支
    const branches = [...new Set(configs.map(config => config.branch))]
    branchList.value = branches

    // 如果只有一个分支，自动选择
    if (branches.length === 1) {
      formData.value.branch = branches[0]
    }

  } catch (error) {
    console.error('加载部署配置失败:', error)
    message.warning('该项目暂无部署配置')
    branchList.value = []
  } finally {
    configLoading.value = false
  }
}

// 重置表单
function resetForm() {
  formData.value = {
    projectId: undefined,
    projectName: '',
    branch: '',
    startTime: new Date().toLocaleString('zh-CN', {
      hour12: false,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    }).replace(/\//g, '-'),
    status: 'running'
  }
  branchList.value = []
}

// 关闭弹窗
function handleClose() {
  emit('update:show', false)
}

// 提交表单
async function handleSubmit() {
  try {
    // 表单验证
    await formRef.value?.validate()

    if (!formData.value.projectId || !formData.value.projectName || !formData.value.branch) {
      message.error('请完善部署信息')
      return
    }

    const submitData: CreateDeployRecordParams = {
      projectId: formData.value.projectId,
      projectName: formData.value.projectName,
      branch: formData.value.branch,
      startTime: formData.value.startTime,
      status: 'running'
    }

    await createDeployRecord(submitData)
    message.success('部署已开始')
    emit('success')
    handleClose()
  } catch (error) {
    console.error('创建部署失败:', error)
    message.error('创建部署失败，请重试')
  }
}

// 取消
function handleCancel() {
  handleClose()
}
</script>

<template>
  <NModal
    :show="show"
    :mask-closable="false"
    preset="dialog"
    title="新建部署"
    style="width: 600px"
    @update:show="handleClose"
  >
    <div class="p-4">
      <NSpin :show="projectLoading || configLoading">
        <NForm
          ref="formRef"
          :model="formData"
          :rules="formRules"
          label-placement="left"
          label-width="80px"
          require-mark-placement="right-hanging"
        >
          <NFormItem label="项目" path="projectId">
            <NSelect
              v-model:value="formData.projectId"
              placeholder="请选择项目"
              :options="projectOptions"
              :loading="projectLoading"
              filterable
              clearable
              @update:value="handleProjectChange"
            />
          </NFormItem>

          <NFormItem label="分支" path="branch">
            <NSelect
              v-model:value="formData.branch"
              placeholder="请选择分支"
              :options="branchOptions"
              :loading="configLoading"
              :disabled="!formData.projectId"
              filterable
              clearable
            />
          </NFormItem>
        </NForm>
      </NSpin>
    </div>

    <template #action>
      <NButton @click="handleCancel">
        取消
      </NButton>
      <NButton
        type="primary"
        :loading="projectLoading || configLoading"
        @click="handleSubmit"
      >
        开始部署
      </NButton>
    </template>
  </NModal>
</template>