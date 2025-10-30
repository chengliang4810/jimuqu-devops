# 认证页面模板

## 📋 模板概述

本目录包含完整的用户认证系统页面模板，涵盖登录、注册、密码重置等所有认证相关功能。

## 📁 模板结构

```
auth/
├── login/
│   ├── index.vue          # 登录主页面
│   └── modules/
│       ├── pwd-login.vue      # 密码登录组件
│       ├── code-login.vue     # 验证码登录组件
│       ├── register.vue       # 注册组件
│       ├── reset-pwd.vue      # 密码重置组件
│       └── bind-wechat.vue    # 微信绑定组件
├── 403/
│   └── index.vue          # 权限不足页面
├── 404/
│   └── index.vue          # 页面不存在
└── 500/
    └── index.vue          # 服务器错误
```

## 🎯 核心特性

### 登录页面 (login/index.vue)
- **多种登录方式**: 支持密码登录、验证码登录、第三方登录
- **表单验证**: 完整的前端表单验证逻辑
- **响应式设计**: 适配桌面端和移动端
- **记住登录**: 支持自动登录和记住密码
- **国际化**: 支持多语言切换

### 认证组件
- **pwd-login.vue**: 用户名密码登录组件
- **code-login.vue**: 手机验证码登录组件
- **register.vue**: 用户注册组件
- **reset-pwd.vue**: 密码重置组件

## 🔧 技术实现

### 核心技术栈
- **Vue 3 Composition API**: 使用最新的组合式API
- **TypeScript**: 完整的类型定义
- **Naive UI**: 现代化的UI组件库
- **Vue Router**: 路由管理
- **Pinia**: 状态管理

### 关键实现
```typescript
// 登录状态管理
const { login, logout } = useAuthStore();

// 表单验证
const rules = {
  username: [
    { required: true, message: '请输入用户名' },
    { min: 3, max: 20, message: '用户名长度为3-20个字符' }
  ],
  password: [
    { required: true, message: '请输入密码' },
    { min: 6, max: 20, message: '密码长度为6-20个字符' }
  ]
};

// 登录处理
const handleLogin = async () => {
  loading.value = true;
  try {
    await login(loginForm.value);
    router.push('/');
  } catch (error) {
    // 错误处理
  } finally {
    loading.value = false;
  }
};
```

## 🎨 UI设计特点

- **现代化界面**: 简洁美观的登录界面设计
- **动画效果**: 平滑的页面切换动画
- **主题适配**: 支持深色/浅色主题
- **交互反馈**: 完善的用户操作反馈

## 📱 响应式支持

- **桌面端**: 宽屏布局，两侧装饰区域
- **平板端**: 自适应布局调整
- **手机端**: 垂直布局，优化触控体验

## 🔐 安全特性

- **前端验证**: 多重表单验证
- **防暴力破解**: 登录失败限制
- **CSRF防护**: 跨站请求伪造防护
- **XSS防护**: 跨站脚本攻击防护

## 🚀 使用指南

### AI模型使用示例
```
请基于 templates/auth/login/index.vue 模板，为我创建一个DevOps平台的专业登录页面，要求：
1. 保持现有的登录流程和验证逻辑
2. 替换背景图片为DevOps相关的技术图标
3. 添加"记住登录"和"自动登录"选项
4. 保持响应式设计和国际化支持
```

### 开发者适配建议
1. 根据业务需求调整表单字段
2. 修改API接口调用方式
3. 自定义主题色彩和样式
4. 添加特定的第三方登录集成

## ⚠️ 重要说明

**本模板仅供参考使用，未经允许不得直接修改！**

---
**模板来源**: ui/src/views/_builtin/
**最后更新**: 2025-01-30