<script setup lang="ts">
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

// 自动跳转到DevOps仪表盘
onMounted(() => {
  router.replace('/devops/dashboard');
});
</script>

<template>
  <div class="flex items-center justify-center h-200px">
    <NSpin size="large">
      <template #description>
        正在跳转到DevOps仪表盘...
      </template>
    </NSpin>
  </div>
</template>
