# 管理后台模板

## 📋 模板概述

本目录包含完整的管理后台页面模板，涵盖用户管理、角色管理、菜单管理等后台管理功能。

## 📁 模板结构

```
manage/
├── user/
│   ├── index.vue              # 用户列表页面
│   ├── user-detail/
│   │   └── [id].vue          # 用户详情页面
│   └── modules/
│       ├── user-search.vue    # 用户搜索组件
│       └── user-operate-drawer.vue # 用户操作抽屉
├── role/
│   ├── index.vue              # 角色列表页面
│   └── modules/
│       ├── role-search.vue    # 角色搜索组件
│       ├── role-operate-drawer.vue # 角色操作抽屉
│       ├── menu-auth-modal.vue    # 菜单权限配置
│       └── button-auth-modal.vue  # 按钮权限配置
└── menu/
    ├── index.vue              # 菜单列表页面
    └── modules/
        └── menu-operate-modal.vue # 菜单操作模态框
```

## 🎯 核心特性

### 用户管理 (user/index.vue)
- **CRUD操作**: 完整的用户增删改查功能
- **高级搜索**: 支持多条件组合搜索
- **批量操作**: 支持批量删除、批量启用/禁用
- **分页显示**: 数据分页和每页数量控制
- **导入导出**: 用户数据的导入导出功能

### 角色管理 (role/index.vue)
- **权限配置**: 灵活的权限分配机制
- **菜单权限**: 基于角色的菜单访问控制
- **按钮权限**: 细粒度的操作权限控制
- **角色继承**: 支持角色权限继承
- **权限预览**: 权限配置实时预览

### 菜单管理 (menu/index.vue)
- **树形结构**: 菜单的层级结构管理
- **拖拽排序**: 支持菜单拖拽排序
- **图标选择**: 丰富的图标选择器
- **路由配置**: 菜单与路由的关联配置
- **显示控制**: 菜单的显示/隐藏控制

## 🔧 技术实现

### 核心技术栈
- **Vue 3 Composition API**: 现代化的组件开发方式
- **Naive UI**: 企业级UI组件库
- **TypeScript**: 类型安全的开发体验
- **Pinia**: 状态管理
- **Vue Router**: 路由管理

### 关键实现
```typescript
// 用户列表管理
const {
  loading,
  data,
  columns,
  pagination,
  searchParams,
  handleSearch,
  handleReset,
  handleEdit,
  handleDelete,
  handleBatchDelete
} = useTable();

// 表格列定义
const columns = [
  { title: '用户名', key: 'userName', width: 120 },
  { title: '邮箱', key: 'email', width: 200 },
  { title: '角色', key: 'roles', width: 150 },
  { title: '状态', key: 'status', width: 100 },
  { title: '创建时间', key: 'createTime', width: 180 },
  { title: '操作', key: 'actions', width: 200, fixed: 'right' }
];

// 搜索参数管理
const searchParams = reactive({
  userName: '',
  email: '',
  status: null,
  role: '',
  dateRange: null
});
```

### 权限控制实现
```typescript
// 权限检查
const hasPermission = (permission: string) => {
  const { userInfo } = useAuthStore();
  return userInfo.permissions.includes(permission);
};

// 操作按钮权限控制
const actionColumns = computed(() => [
  {
    title: '操作',
    key: 'actions',
    render: (row) => {
      return h('div', { class: 'flex gap-2' }, [
        hasPermission('user:edit') &&
        h(NButton, { size: 'small', onClick: () => handleEdit(row) }, '编辑'),
        hasPermission('user:delete') &&
        h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
          default: () => '确认删除该用户？',
          trigger: () => h(NButton, { size: 'small', type: 'error' }, '删除')
        })
      ]);
    }
  }
]);
```

## 🎨 UI设计特点

- **专业界面**: 企业级管理后台的设计风格
- **数据密集**: 适合大量数据展示和操作
- **操作便捷**: 丰富的快捷操作和批量操作
- **响应式**: 适配不同屏幕尺寸
- **主题适配**: 支持深色/浅色主题

## 📊 组件复用

### 通用组件
- **SearchForm**: 搜索表单组件
- **DataTable**: 数据表格组件
- **OperateDrawer**: 操作抽屉组件
- **ConfirmModal**: 确认模态框组件

### 业务组件
- **UserSearch**: 用户专用搜索组件
- **RoleSelector**: 角色选择器组件
- **PermissionTree**: 权限树组件

## 🚀 使用指南

### AI模型使用示例
```
请基于 templates/manage/user/index.vue 模板，为我创建一个项目管理页面，要求：
1. 保持相同的CRUD操作流程
2. 替换用户相关的字段为项目字段（项目名称、描述、状态、负责人等）
3. 保持搜索、分页、批量操作功能
4. 适配项目管理的权限控制逻辑
```

### 适配其他管理场景
1. **字段替换**: 根据业务需求调整表格字段
2. **搜索条件**: 修改搜索表单的字段和验证规则
3. **操作按钮**: 根据权限调整可操作按钮
4. **API接口**: 替换为对应的业务API接口

## ⚠️ 重要说明

**本模板仅供参考使用，未经允许不得直接修改！**

---
**模板来源**: ui/src/views/manage/
**最后更新**: 2025-01-30