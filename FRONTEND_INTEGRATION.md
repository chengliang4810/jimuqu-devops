# DevOps平台 - lithe-admin前端集成指南

## 🎨 已选择的开源框架

**lithe-admin** - 现代化的Vue3管理后台模板

### ✨ 技术栈
- ✅ **Vue 3.5.22** + Composition API
- ✅ **TypeScript 5.9.3** - 完整类型支持
- ✅ **Vite 7.1.19** - 极速构建工具
- ✅ **TailwindCSS 4.1.16** - 现代化样式框架
- ✅ **Naive UI 2.43.1** - Vue3最佳组件库
- ✅ **Pinia** - 状态管理
- ✅ **Vue Router** - 路由管理
- ✅ **ECharts** - 数据图表库

### 🎯 框架特性
- 磨砂质感和纹理效果
- 韵滑的过渡动画
- 响应式设计
- 灵活主题定制
- 完整的TypeScript支持

---

## 📦 集成步骤

### 步骤1: 前端项目准备

```bash
# 进入前端目录
cd frontend

# 安装依赖 (推荐使用pnpm)
pnpm install

# 或使用npm
npm install

# 启动开发服务器
pnpm dev
```

### 步骤2: 项目结构

```
frontend/src/
├── components/          # 通用组件
├── views/              # 页面视图
│   ├── devops/         # DevOps平台页面
│   │   ├── dashboard/  # 仪表盘
│   │   ├── projects/   # 项目管理
│   │   └── deployments/# 部署记录
│   ├── sign-in/        # 登录页
│   └── error-page/     # 错误页
├── router/             # 路由配置
├── stores/             # Pinia状态管理
├── utils/              # 工具函数
└── layout/             # 布局组件
```

### 步骤3: 核心文件创建

需要创建的11个核心文件：

#### 1. 路由配置
**frontend/src/router/index.ts**
```typescript
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { devopsRoutes } from './devops'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/dashboard' },
  {
    path: '/sign-in',
    name: 'signIn',
    component: () => import('@/views/sign-in/index.vue')
  },
  ...devopsRoutes,
  {
    name: 'errorPage',
    path: '/:pathMatch(.*)*',
    component: () => import('@/views/error-page/index.vue')
  }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
```

**frontend/src/router/devops.ts**
```typescript
import type { RouteRecordRaw } from 'vue-router'

export const devopsRoutes: RouteRecordRaw[] = [
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/devops/dashboard/index.vue'),
    meta: { title: '仪表盘', icon: 'dashboard' }
  },
  {
    path: '/projects',
    name: 'Projects',
    component: () => import('@/views/devops/projects/index.vue'),
    meta: { title: '项目管理', icon: 'project' }
  },
  {
    path: '/deployments',
    name: 'Deployments',
    component: () => import('@/views/devops/deployments/index.vue'),
    meta: { title: '部署记录', icon: 'deployment' }
  }
]
```

#### 2. Pinia状态管理

**frontend/src/stores/projects.ts**
```typescript
import { defineStore } from 'pinia'
import { api } from '@/utils/api'

export const useProjectsStore = defineStore('projects', {
  state: () => ({
    projects: [],
    loading: false
  }),

  actions: {
    async fetchProjects() {
      this.loading = true
      try {
        const response = await api.get('/api/projects')
        this.projects = response
      } finally {
        this.loading = false
      }
    },

    async createProject(data: any) {
      const response = await api.post('/api/projects', data)
      this.projects.push(response)
      return response
    },

    async updateProject(id: number, data: any) {
      const response = await api.put(`/api/projects/${id}`, data)
      const index = this.projects.findIndex(p => p.id === id)
      if (index !== -1) this.projects[index] = response
      return response
    },

    async deleteProject(id: number) {
      await api.delete(`/api/projects/${id}`)
      this.projects = this.projects.filter(p => p.id !== id)
    },

    async testConnection(id: number) {
      return api.post(`/api/projects/${id}/test-connection`, {})
    }
  }
})
```

**frontend/src/stores/deployments.ts**
```typescript
import { defineStore } from 'pinia'
import { api } from '@/utils/api'

export const useDeploymentsStore = defineStore('deployments', {
  state: () => ({
    deployments: [],
    loading: false
  }),

  actions: {
    async fetchDeployments(projectId?: number) {
      this.loading = true
      try {
        const params = projectId ? `?project_id=${projectId}` : ''
        const response = await api.get(`/api/deployments${params}`)
        this.deployments = response.deployments || []
      } finally {
        this.loading = false
      }
    },

    async executeDeployment(id: number) {
      return api.post(`/api/deployments/${id}/execute`, {})
    }
  }
})
```

#### 3. API客户端

**frontend/src/utils/api.ts**
```typescript
class ApiClient {
  private baseURL = window.location.origin
  private auth: string | null = null

  constructor() {
    this.auth = localStorage.getItem('auth')
  }

  private async request(endpoint: string, options: RequestInit = {}) {
    const url = `${this.baseURL}${endpoint}`
    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    }
    if (this.auth) headers['Authorization'] = `Basic ${this.auth}`

    const response = await fetch(url, { ...options, headers })
    const data = await response.json()

    if (!response.ok) throw new Error(data.detail || '请求失败')
    return data
  }

  async get(endpoint: string) { return this.request(endpoint) }
  async post(endpoint: string, data?: any) { return this.request(endpoint, { method: 'POST', body: JSON.stringify(data) }) }
  async put(endpoint: string, data?: any) { return this.request(endpoint, { method: 'PUT', body: JSON.stringify(data) }) }
  async delete(endpoint: string) { return this.request(endpoint, { method: 'DELETE' }) }

  async login(username: string, password: string) {
    this.auth = btoa(`${username}:${password}`)
    localStorage.setItem('auth', this.auth)
    try {
      return await this.get('/api/me')
    } catch {
      this.auth = null
      localStorage.removeItem('auth')
      throw new Error('登录失败')
    }
  }

  getWebSocketUrl(deploymentId: number): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}/ws/deployments/${deploymentId}`
  }
}

export const api = new ApiClient()
```

### 步骤4: 页面组件

#### 仪表盘页面 (frontend/src/views/devops/dashboard/index.vue)
```vue
<template>
  <div class="p-6">
    <n-grid :cols="5" :x-gap="24">
      <n-gi>
        <n-card>
          <n-statistic label="总项目数" :value="stats.total_projects" />
        </n-card>
      </n-gi>
      <!-- 更多统计卡片 -->
    </n-grid>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NGrid, NGi, NStatistic } from 'naive-ui'
import { useDashboardStore } from '@/stores/dashboard'

const dashboardStore = useDashboardStore()
const stats = ref({ total_projects: 0, active_projects: 0, today_deployments: 0, success_rate: 0, average_duration: 0 })

onMounted(async () => {
  const data = await dashboardStore.getDashboardStats()
  stats.value = data
})
</script>
```

### 步骤5: 修改App.vue

在App.vue中集成认证和路由：
```vue
<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/utils/api'

const router = useRouter()

onMounted(() => {
  // 检查是否已登录
  if (!api.auth && router.currentRoute.value.path !== '/sign-in') {
    router.push('/sign-in')
  }
})
</script>
```

---

## 🚀 启动流程

### 开发模式

```bash
# 启动前端开发服务器 (端口5173)
cd frontend
pnpm dev

# 启动后端服务 (端口8000)
cd ..
docker-compose up -d

# 访问 http://localhost:5173
```

### 生产构建

```bash
# 构建前端
cd frontend
pnpm build

# 复制到后端静态目录
cp -r dist/* ../app/static/

# 重启服务
cd ..
docker-compose restart web
```

---

## 🔗 后端API对接

所有API接口已在 **frontend/src/utils/api.ts** 中配置，对接FastAPI后端：

### 认证接口
- `GET /api/me` - 获取当前用户

### 项目接口
- `GET /api/projects` - 获取项目列表
- `POST /api/projects` - 创建项目
- `PUT /api/projects/{id}` - 更新项目
- `DELETE /api/projects/{id}` - 删除项目
- `POST /api/projects/{id}/test-connection` - 测试连接

### 部署接口
- `GET /api/deployments` - 获取部署列表
- `POST /api/deployments` - 创建部署
- `POST /api/deployments/{id}/execute` - 执行部署
- `GET /api/deployments/stats/dashboard` - 仪表盘统计

### WebSocket
- `ws://host/ws/deployments/{id}` - 实时日志

---

## 🎨 UI组件使用

基于Naive UI组件库，已包含：

### 常用组件
- `n-card` - 卡片容器
- `n-button` - 按钮
- `n-input` - 输入框
- `n-select` - 选择器
- `n-data-table` - 数据表格
- `n-modal` - 模态框
- `n-message` - 消息提示
- `n-statistic` - 统计数字

### 图标库
- `@vicons/antd` - Ant Design图标
- `@vicons/ionicons5` - Ionicons图标
- `@vicons/fluent` - Fluent图标

---

## 🔄 实时日志实现

WebSocket实时日志：

```typescript
// 组件中
const connectWebSocket = () => {
  const wsUrl = api.getWebSocketUrl(deploymentId)
  const ws = new WebSocket(wsUrl)

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.type === 'log') {
      logs.value.push({
        type: data.log_type,
        message: data.message,
        timestamp: data.timestamp
      })
    }
  }
}
```

---

## 📝 开发建议

### 1. 状态管理
使用Pinia进行状态管理，每个功能模块一个store

### 2. 类型安全
所有接口使用TypeScript类型定义，确保类型安全

### 3. 组件复用
抽取通用组件到 `src/components/` 目录

### 4. 样式规范
遵循TailwindCSS原子化CSS原则

### 5. 错误处理
统一使用Naive UI的Message组件进行错误提示

---

## 🚀 下一步

1. 创建11个核心文件
2. 配置路由和状态管理
3. 开发页面组件
4. 测试API对接
5. 构建生产版本

---

## 💡 优势

使用lithe-admin的优势：
- ✅ 现代化技术栈（Vue3 + TypeScript）
- ✅ 完整的类型支持
- ✅ 丰富的组件库（Naive UI）
- ✅ 响应式设计
- ✅ 主题定制能力
- ✅ 快速开发体验

比原生HTML/JS的优势：
- 更好的开发体验（类型提示、智能感知）
- 更强的可维护性（组件化、模块化）
- 更丰富的UI组件
- 更好的状态管理
- 更灵活的主题定制

---

**现在启动开发服务器即可开始使用！**
