<script lang="ts" setup>
import type { Project } from '#/api/project';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { createProject, updateProject } from '#/api/project';

import { useSchema } from '../data';

const emit = defineEmits(['success']);
const formData = ref<Project>();
const getTitle = computed(() => {
  return formData.value?.id ? '编辑项目' : '添加项目';
});

const [Form, formApi] = useVbenForm({
  layout: 'horizontal',
  schema: useSchema(),
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (valid) {
      modalApi.lock();
      const data = await formApi.getValues();
      try {
        if (formData.value?.id) {
          // 编辑模式：确保使用原始数据中的ID
          const updateData = {
            id: formData.value.id,
            ...data
          };
          await updateProject(updateData);
        } else {
          // 创建模式
          await createProject(data);
        }
        modalApi.close();
        emit('success');
      } finally {
        modalApi.lock(false);
      }
    }
  },
  onOpenChange(isOpen) {
    if (isOpen) {
      const data = modalApi.getData<Project>();
      if (data) {
        formData.value = data;
        formApi.setValues(formData.value);
      }
    }
  },
});
</script>

<template>
  <Modal :title="getTitle">
    <Form class="mx-4" />
  </Modal>
</template>