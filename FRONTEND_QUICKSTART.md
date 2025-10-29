# 前端开发 - 5分钟快速上手

## 🚀 方式一：一键启动（推荐）

```bash
./dev-start.sh
```

自动完成：
- ✅ 检查Node.js环境
- ✅ 安装pnpm
- ✅ 启动后端服务
- ✅ 启动前端开发服务器

访问：**http://localhost:5173**

---

## 🛠️ 方式二：手动启动

### 1. 启动后端服务
```bash
docker-compose up -d redis
```

### 2. 启动前端开发服务器
```bash
cd frontend
pnpm install
pnpm dev
```

访问：**http://localhost:5173**

---

## 📁 目录结构

```
frontend/
├── src/
│   ├── views/          # 页面组件
│   │   ├── devops/
│   │   │   ├── dashboard/    # 仪表盘
│   │   │   ├── projects/     # 项目管理
│   │   │   └── deployments/  # 部署记录
│   │   ├── sign-in/    # 登录页
│   │   └── error-page/ # 错误页
│   ├── router/         # 路由配置
│   ├── stores/         # 状态管理
│   ├── utils/          # 工具函数
│   └── components/     # 通用组件
├── package.json        # 依赖配置
└── vite.config.ts      # Vite配置
```

---

## 🎨 开发流程

### 1. 创建新页面

例如：创建"用户管理"页面

```bash
# 创建目录
mkdir -p frontend/src/views/users

# 创建页面文件
touch frontend/src/views/users/index.vue
```

### 2. 添加路由

**frontend/src/router/users.ts**
```typescript
import type { RouteRecordRaw } from 'vue-router'

export const userRoutes: RouteRecordRaw[] = [
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/users/index.vue'),
    meta: { title: '用户管理', icon: 'user' }
  }
]
```

### 3. 注册路由

**frontend/src/router/index.ts**
```typescript
import { userRoutes } from './users'

const routes: RouteRecordRaw[] = [
  // ...
  ...userRoutes
]
```

### 4. 创建状态管理

**frontend/src/stores/users.ts**
```typescript
import { defineStore } from 'pinia'
import { api } from '@/utils/api'

export const useUsersStore = defineStore('users', {
  state: () => ({
    users: [],
    loading: false
  }),

  actions: {
    async fetchUsers() {
      this.loading = true
      try {
        this.users = await api.get('/api/users')
      } finally {
        this.loading = false
      }
    }
  }
})
```

---

## 🔧 常用命令

```bash
# 安装依赖
cd frontend
pnpm install

# 启动开发服务器
pnpm dev

# 构建生产版本
pnpm build

# 预览生产版本
pnpm preview

# 代码检查
pnpm lint:check

# 代码修复
pnpm lint:fix

# 格式检查
pnpm format:check

# 格式修复
pnpm format:fix
```

---

## 📚 技术文档

### Vue 3 文档
- https://cn.vuejs.org/
- Composition API: https://cn.vuejs.org/guide/extras/composition-api-faq.html

### TypeScript 文档
- https://www.typescriptlang.org/docs/
- Vue 3 + TS: https://cn.vuejs.org/guide/typescript/overview.html

### Naive UI 文档
- https://www.naiveui.com/
- 中文文档: https://www.naiveui.com/zh-CN/os-theme

### TailwindCSS 文档
- https://tailwindcss.com/docs
- 中文文档: https://tailwindcss.cn/

### Vite 文档
- https://vitejs.dev/
- Vue + Vite: https://cn.vitejs.dev/

### Pinia 文档
- https://pinia.vuejs.org/
- 中文文档: https://pinia.vuejs.org/zh/

---

## 🎯 开发规范

### 1. 命名规范

- **文件命名**: kebab-case (短横线)
  ```
  user-profile.vue
  deployment-log.vue
  ```

- **组件命名**: PascalCase
  ```
  UserProfile
  DeploymentLog
  ```

- **变量命名**: camelCase
  ```typescript
  const userName = 'admin'
  const deploymentId = 123
  ```

### 2. 目录规范

```
src/
├── components/     # 通用组件
│   ├── ui/         # 基础UI组件
│   └── forms/      # 表单组件
├── views/          # 页面组件
│   ├── module/     # 功能模块
│   │   ├── index.vue    # 页面入口
│   │   └── components/  # 页面专用组件
├── router/         # 路由配置
├── stores/         # 状态管理
├── utils/          # 工具函数
└── types/          # 类型定义
```

### 3. 代码风格

使用ESLint + Prettier规范代码

```bash
# 检查代码
pnpm lint:check

# 自动修复
pnpm lint:fix
```

### 4. 提交规范

```
feat: 新功能
fix: 修复
docs: 文档
style: 格式
refactor: 重构
test: 测试
chore: 构建
```

示例：
```
feat: 添加用户管理页面
fix: 修复项目列表加载失败
docs: 更新API文档
```

---

## 🐛 常见问题

### Q: 前端启动失败？

A: 检查Node.js版本（需要20+）：
```bash
node -v
```

### Q: 依赖安装失败？

A: 清理缓存重试：
```bash
cd frontend
rm -rf node_modules pnpm-lock.yaml
pnpm install
```

### Q: API请求失败？

A: 检查后端服务是否启动：
```bash
docker-compose ps
# 应该看到 redis 服务正在运行
```

### Q: WebSocket连接失败？

A: 检查后端WebSocket服务是否正常，访问：
```
http://localhost:8000/docs
```

### Q: 如何修改API地址？

A: 修改 `frontend/src/utils/api.ts`：
```typescript
// 默认使用当前域名
private baseURL = window.location.origin

// 或指定其他地址
private baseURL = 'http://localhost:8000'
```

---

## 💡 最佳实践

### 1. 状态管理

使用Pinia进行状态管理，避免滥用全局状态

```typescript
// ✅ 正确：每个模块独立store
export const useProjectStore = defineStore('projects', { ... })
export const useUserStore = defineStore('users', { ... })

// ❌ 错误：所有状态堆在一个store
export const useAppStore = defineStore('app', {
  state: () => ({
    projects: [],
    users: [],
    deployments: []
  })
})
```

### 2. API调用

封装API调用，避免重复代码

```typescript
// ✅ 正确：在stores中封装
export const useProjectStore = defineStore('projects', {
  actions: {
    async fetchProjects() {
      return api.get('/api/projects')
    }
  }
})

// 在组件中使用
const projectStore = useProjectStore()
await projectStore.fetchProjects()

// ❌ 错误：在组件中直接调用API
const projects = await api.get('/api/projects')
```

### 3. 组件拆分

合理拆分组件，提高复用性

```vue
<!-- ✅ 正确：拆分基础组件 -->
<template>
  <ProjectForm
    v-model:show="show"
    :project="currentProject"
    @success="handleSuccess"
  />
</template>

<!-- ❌ 错误：一个组件做所有事情 -->
<template>
  <div class="project-manager">
    <!-- 表单、列表、弹窗都在一个组件 -->
  </div>
</template>
```

### 4. 类型安全

使用TypeScript确保类型安全

```typescript
// ✅ 正确：定义类型
interface Project {
  id: number
  name: string
  git_url: string
  language: 'java' | 'python' | 'node' | 'go'
}

export const useProjectStore = defineStore('projects', {
  state: () => ({
    projects: [] as Project[]
  })
})

// ❌ 错误：不使用类型
export const useProjectStore = defineStore('projects', {
  state: () => ({
    projects: [] // 不知道是什么类型
  })
})
```

---

## 🎉 开始开发

现在你可以：

1. **启动开发服务器**
   ```bash
   ./dev-start.sh
   ```

2. **访问前端**
   ```
   http://localhost:5173
   ```

3. **查看API文档**
   ```
   http://localhost:8000/docs
   ```

4. **开始开发**
   - 修改页面: `frontend/src/views/devops/`
   - 添加路由: `frontend/src/router/`
   - 管理状态: `frontend/src/stores/`

---

**祝你开发愉快！** 🚀
