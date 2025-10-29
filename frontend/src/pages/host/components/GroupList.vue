<template>
  <div class="host-group-list">
    <div class="group-header">
      <h3>主机分组</h3>
      <t-button
        size="small"
        variant="text"
        @click="showCreateDialog = true"
      >
        <template #icon>
          <add-icon />
        </template>
        新建分组
      </t-button>
    </div>

    <div class="group-content">
      <draggable
        v-model="groupList"
        item-key="id"
        ghost-class="ghost"
        chosen-class="chosen"
        drag-class="drag"
        @end="handleSortEnd"
      >
        <template #item="{ element }">
          <div
            :class="['group-item', { active: selectedGroupId === element.id }]"
            @click="selectGroup(element.id)"
            @contextmenu.prevent="showContextMenu($event, element)"
          >
            <div class="group-info">
              <div class="group-name">{{ element.name }}</div>
              <div class="group-count">{{ element.hostCount }} 台主机</div>
            </div>
            <div class="group-actions">
              <t-dropdown>
                <t-button variant="text" size="small">
                  <template #icon>
                    <more-icon />
                  </template>
                </t-button>
                <t-dropdown-menu>
                  <t-dropdown-item @click="editGroup(element)">
                    <template #icon>
                      <edit-icon />
                    </template>
                    编辑
                  </t-dropdown-item>
                  <t-dropdown-item @click="deleteGroup(element)" theme="danger">
                    <template #icon>
                      <delete-icon />
                    </template>
                    删除
                  </t-dropdown-item>
                </t-dropdown-menu>
              </t-dropdown>
            </div>
          </div>
        </template>
      </draggable>
    </div>

    <!-- 右键菜单 -->
    <t-dropdown
      v-model:visible="contextMenuVisible"
      :popup-props="{ x: contextMenuX, y: contextMenuY }"
      trigger="context-menu"
      attach="body"
    >
      <t-dropdown-menu>
        <t-dropdown-item @click="editGroup(contextMenuGroup)">
          <template #icon>
            <edit-icon />
          </template>
          编辑分组
        </t-dropdown-item>
        <t-dropdown-item @click="deleteGroup(contextMenuGroup)" theme="danger">
          <template #icon>
            <delete-icon />
          </template>
          删除分组
        </t-dropdown-item>
      </t-dropdown-menu>
    </t-dropdown>

    <!-- 新建/编辑分组对话框 -->
    <t-dialog
      v-model:visible="showCreateDialog"
      :header="editingGroup ? '编辑分组' : '新建分组'"
      width="500px"
      @confirm="handleGroupSubmit"
    >
      <t-form ref="groupFormRef" :data="groupForm" :rules="groupRules">
        <t-form-item label="分组名称" name="name">
          <t-input v-model="groupForm.name" placeholder="请输入分组名称" />
        </t-form-item>
        <t-form-item label="分组描述" name="description">
          <t-textarea
            v-model="groupForm.description"
            placeholder="请输入分组描述（可选）"
            :maxlength="200"
          />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 删除确认对话框 -->
    <t-dialog
      v-model:visible="showDeleteDialog"
      header="删除分组"
      width="400px"
      @confirm="confirmDelete"
    >
      <div class="delete-confirm">
        <p>确定要删除分组 "{{ deletingGroup?.name }}" 吗？</p>
        <p v-if="deletingGroup?.hostCount > 0" class="warning-text">
          注意：该分组下还有 {{ deletingGroup.hostCount }} 台主机，删除分组后这些主机将被移至默认分组。
        </p>
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { AddIcon, EditIcon, DeleteIcon, MoreIcon } from 'tdesign-icons-vue-next';
import draggable from 'vuedraggable';

import { hostGroupApi } from '@/api/host';
import type { HostGroup, HostGroupForm } from '@/types/host';

const emit = defineEmits<{
  groupSelect: [groupId: string];
}>();

// 响应式数据
const groupList = ref<HostGroup[]>([]);
const selectedGroupId = ref<string>('default');
const showCreateDialog = ref(false);
const showDeleteDialog = ref(false);
const editingGroup = ref<HostGroup | null>(null);
const deletingGroup = ref<HostGroup | null>(null);
const groupFormRef = ref();

// 右键菜单
const contextMenuVisible = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const contextMenuGroup = ref<HostGroup | null>(null);

// 表单数据
const groupForm = reactive<HostGroupForm>({
  name: '',
  description: '',
});

// 表单验证规则
const groupRules = {
  name: [
    { required: true, message: '请输入分组名称' },
    { min: 1, max: 50, message: '分组名称长度为1-50个字符' },
  ],
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

// 选择分组
const selectGroup = (groupId: string) => {
  selectedGroupId.value = groupId;
  emit('groupSelect', groupId);
};

// 显示右键菜单
const showContextMenu = (event: MouseEvent, group: HostGroup) => {
  contextMenuX.value = event.clientX;
  contextMenuY.value = event.clientY;
  contextMenuGroup.value = group;
  contextMenuVisible.value = true;
};

// 编辑分组
const editGroup = (group: HostGroup) => {
  if (group.id === 'default') {
    MessagePlugin.warning('默认分组不能编辑');
    return;
  }

  editingGroup.value = group;
  groupForm.name = group.name;
  groupForm.description = group.description || '';
  showCreateDialog.value = true;
  contextMenuVisible.value = false;
};

// 删除分组
const deleteGroup = (group: HostGroup) => {
  if (group.id === 'default') {
    MessagePlugin.warning('默认分组不能删除');
    return;
  }

  deletingGroup.value = group;
  showDeleteDialog.value = true;
  contextMenuVisible.value = false;
};

// 确认删除
const confirmDelete = async () => {
  if (!deletingGroup.value) return;

  try {
    const res = await hostGroupApi.delete(deletingGroup.value.id);
    if (res.success) {
      MessagePlugin.success('删除成功');
      await getGroupList();

      // 如果删除的是当前选中的分组，切换到默认分组
      if (selectedGroupId.value === deletingGroup.value.id) {
        selectGroup('default');
      }
    }
  } catch (error) {
    console.error('删除分组失败:', error);
    MessagePlugin.error('删除失败');
  }

  showDeleteDialog.value = false;
  deletingGroup.value = null;
};

// 处理分组提交
const handleGroupSubmit = async () => {
  const valid = await groupFormRef.value?.validate();
  if (!valid) return;

  try {
    if (editingGroup.value) {
      // 更新分组
      const res = await hostGroupApi.update(editingGroup.value.id, groupForm);
      if (res.success) {
        MessagePlugin.success('更新成功');
      }
    } else {
      // 创建分组
      const res = await hostGroupApi.create(groupForm);
      if (res.success) {
        MessagePlugin.success('创建成功');
      }
    }

    await getGroupList();
    showCreateDialog.value = false;
    resetForm();
  } catch (error) {
    console.error('操作失败:', error);
    MessagePlugin.error('操作失败');
  }
};

// 处理拖拽排序
const handleSortEnd = async () => {
  try {
    const sortData = groupList.value
      .filter(group => group.id !== 'default')
      .map((group, index) => ({
        id: group.id,
        sort: index + 1,
      }));

    const res = await hostGroupApi.updateSort(sortData);
    if (res.success) {
      MessagePlugin.success('排序更新成功');
    }
  } catch (error) {
    console.error('更新排序失败:', error);
    MessagePlugin.error('更新排序失败');
    // 恢复排序
    await getGroupList();
  }
};

// 重置表单
const resetForm = () => {
  groupForm.name = '';
  groupForm.description = '';
  editingGroup.value = null;
};

// 组件挂载时获取数据
onMounted(() => {
  getGroupList();
  // 默认选择默认分组
  selectGroup('default');
});
</script>

<style lang="less" scoped>
.host-group-list {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--td-bg-color-container);
  border-right: 1px solid var(--td-component-border);
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--td-component-border);

  h3 {
    margin: 0;
    font: var(--td-font-title-medium);
    color: var(--td-text-color-primary);
  }
}

.group-content {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
}

.group-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  margin-bottom: 4px;
  border-radius: var(--td-radius-default);
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;

  &:hover {
    background: var(--td-bg-color-container-hover);
  }

  &.active {
    background: var(--td-brand-color-1);
    border: 1px solid var(--td-brand-color);
  }

  &.ghost {
    opacity: 0.5;
    background: var(--td-bg-color-container-select);
  }

  &.chosen {
    background: var(--td-brand-color-1);
    border: 1px solid var(--td-brand-color);
  }

  &.drag {
    background: var(--td-bg-color-container-hover);
    transform: rotate(5deg);
  }
}

.group-info {
  flex: 1;
}

.group-name {
  font: var(--td-font-body-medium);
  color: var(--td-text-color-primary);
  margin-bottom: 4px;
}

.group-count {
  font: var(--td-font-body-small);
  color: var(--td-text-color-secondary);
}

.group-actions {
  opacity: 0;
  transition: opacity 0.2s;
}

.group-item:hover .group-actions {
  opacity: 1;
}

.delete-confirm {
  .warning-text {
    color: var(--td-error-color);
    margin-top: 8px;
  }
}
</style>