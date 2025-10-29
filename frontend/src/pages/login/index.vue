<template>
  <div class="login-wrapper">
    <div class="background-gradient"></div>

    <div class="login-container">
      <div class="title-container">
        <h1 class="title">积木区DevOps</h1>
        <p class="sub-title">Jimuqu DevOps Platform - 一站式DevOps自动化部署平台</p>
      </div>

      <login />
    </div>

    <footer class="copyright">Copyright @ 2024-2025 积木区DevOps (jimuqu.com). All Rights Reserved</footer>
  </div>
</template>
<script setup lang="ts">
import Login from './components/Login.vue';

defineOptions({
  name: 'LoginIndex',
});
</script>
<style lang="less" scoped>
@import './index.less';
</style>
