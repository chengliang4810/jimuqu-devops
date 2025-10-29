<template>
  <div class="login-wrapper">
    <div class="background-gradient"></div>

    <div class="login-container">
      <div class="title-container">
        <h1 class="title">DevOps 自动化部署平台</h1>
        <p class="sub-title">请输入您的账号密码登录系统</p>
      </div>

      <login />
    </div>

    <footer class="copyright">Copyright @ 2021-2025 DevOps Platform. All Rights Reserved</footer>
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
