<template>
  <div :class="`${prefix}-footer`">Copyright © 2024-{{ new Date().getFullYear() }} 积木区DevOps (jimuqu.com). All Rights Reserved</div>
</template>
<script setup lang="ts">
import { prefix } from '@/config/global';
</script>
<style lang="less" scoped>
.@{starter-prefix}-footer {
  color: var(--td-text-color-placeholder);
  line-height: 20px;
  text-align: center;
}
</style>
