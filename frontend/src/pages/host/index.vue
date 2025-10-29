<template>
  <div class="host-page">
    <div class="host-layout">
      <!-- 左侧分组列表 -->
      <div class="group-panel">
        <GroupList @group-select="handleGroupSelect" />
      </div>

      <!-- 右侧主机管理 -->
      <div class="host-panel">
        <HostManagement
          :selected-group-id="selectedGroupId"
          :group-list="groupList"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';

import GroupList from './components/GroupList.vue';
import HostManagement from './components/HostManagement.vue';
import { hostGroupApi } from '@/api/host';
import type { HostGroup } from '@/types/host';

defineOptions({
  name: 'HostManagement',
});

// 响应式数据
const selectedGroupId = ref<string>('default');
const groupList = ref<HostGroup[]>([]);

// 处理分组选择
const handleGroupSelect = (groupId: string) => {
  selectedGroupId.value = groupId;
};

// 获取分组列表
const getGroupList = async () => {
  try {
    const res = await hostGroupApi.getAll();
    if (res.success) {
      // 添加默认分组
      const defaultGroup: HostGroup = {
        id: 'default',
        name: '默认分组',
        description: '系统默认分组',
        sort: 0,
        hostCount: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      groupList.value = [defaultGroup, ...res.data];
    }
  } catch (error) {
    console.error('获取分组列表失败:', error);
    MessagePlugin.error('获取分组列表失败');
  }
};

// 组件挂载时获取数据
onMounted(() => {
  getGroupList();
});
</script>

<style lang="less" scoped>
.host-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.host-layout {
  flex: 1;
  display: flex;
  height: 100%;
  overflow: hidden;
}

.group-panel {
  width: 280px;
  flex-shrink: 0;
  height: 100%;
}

.host-panel {
  flex: 1;
  height: 100%;
  min-width: 0;
}

@media (max-width: 1200px) {
  .group-panel {
    width: 240px;
  }
}

@media (max-width: 768px) {
  .host-layout {
    flex-direction: column;
  }

  .group-panel {
    width: 100%;
    height: 200px;
  }

  .host-panel {
    height: calc(100% - 200px);
  }
}
</style>