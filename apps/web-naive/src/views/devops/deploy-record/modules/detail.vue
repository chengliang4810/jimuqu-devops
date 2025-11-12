<script setup lang="ts">
import { computed, ref } from 'vue'
import { NModal, NCard, NDescriptions, NDescriptionsItem, NTag, NButton, NSpace, NAlert, NDivider } from 'naive-ui'
import { formatDuration, statusConfig } from '../data'
import type { DeployRecord } from '#/api/deploy-record'

interface Props {
  show: boolean
  record: DeployRecord | null
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  record: null,
})

const emit = defineEmits<Emits>()

const showModal = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value),
})

const statusTagType = computed(() => {
  if (!props.record?.status) return 'default'
  return statusConfig[props.record.status]?.type || 'default'
})

const statusTagText = computed(() => {
  if (!props.record?.status) return '未知'
  return statusConfig[props.record.status]?.label || '未知'
})

function handleCopyLogPath() {
  if (props.record?.logPath) {
    navigator.clipboard.writeText(props.record?.logPath).then(() => {
      // 这里可以添加成功提示
    })
  }
}
</script>

<template>
  <NModal v-model:show="showModal" preset="card" :style="{ width: '800px' }" title="部署记录详情">
    <template v-if="record">
      <!-- 基本信息 -->
      <NDescriptions title="基本信息" :column="2" bordered>
        <NDescriptionsItem label="记录ID">
          {{ record.id }}
        </NDescriptionsItem>
        <NDescriptionsItem label="项目名称">
          {{ record.projectName }}
        </NDescriptionsItem>
        <NDescriptionsItem label="分支">
          <code>{{ record.branch }}</code>
        </NDescriptionsItem>
        <NDescriptionsItem label="部署状态">
          <NTag :type="statusTagType">
            {{ statusTagText }}
          </NTag>
        </NDescriptionsItem>
        <NDescriptionsItem label="开始时间">
          {{ record.startTime }}
        </NDescriptionsItem>
        <NDescriptionsItem label="耗时">
          {{ formatDuration(record.duration) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="创建时间" span="2">
          {{ record.createdAt }}
        </NDescriptionsItem>
      </NDescriptions>

      <NDivider />

      <!-- 日志信息 -->
      <div v-if="record.logPath">
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-lg font-medium">
            日志信息
          </h3>
          <NButton size="small" @click="handleCopyLogPath">
            复制路径
          </NButton>
        </div>
        <NAlert type="info" show-icon>
          <template #icon>
            📄
          </template>
          <div class="font-mono text-sm">
            {{ record.logPath }}
          </div>
        </NAlert>
      </div>

      <div v-else>
        <NAlert type="warning" show-icon>
          该部署记录没有日志文件
        </NAlert>
      </div>

      <NDivider />

      <!-- 更新信息 -->
      <NDescriptions title="更新信息" :column="1" bordered>
        <NDescriptionsItem label="最后更新时间">
          {{ record.updatedAt }}
        </NDescriptionsItem>
      </NDescriptions>
    </template>

    <template v-else>
      <div class="text-center py-8">
        <div class="text-gray-500">
          没有选择的记录
        </div>
      </div>
    </template>

    <!-- 底部操作 -->
    <template #footer>
      <NSpace justify="end">
        <NButton @click="showModal = false">
          关闭
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>