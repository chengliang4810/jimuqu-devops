# 数据表格模板

## 📋 模板概述

本目录包含功能强大的数据表格组件模板，支持远程数据加载、行内编辑、高级搜索、批量操作等企业级应用需求。

## 📁 模板结构

```
table/
├── remote/
│   └── index.vue          # 远程数据表格模板
└── row-edit/
    └── index.vue          # 行内编辑表格模板
```

## 🎯 核心特性

### 远程数据表格 (remote/index.vue)
- **远程数据**: 支持从服务器异步加载数据
- **分页支持**: 完整的分页功能
- **排序功能**: 多列排序支持
- **筛选功能**: 列筛选和全局搜索
- **刷新机制**: 手动和自动数据刷新
- **加载状态**: 数据加载时的loading状态

### 行内编辑表格 (row-edit/index.vue)
- **行内编辑**: 直接在表格中编辑数据
- **单元格编辑**: 单个单元格的独立编辑
- **批量编辑**: 批量选择和编辑
- **验证支持**: 编辑时的数据验证
- **撤销功能**: 编辑撤销和恢复
- **保存策略**: 手动保存和自动保存

## 🔧 技术实现

### 核心技术栈
- **Vue 3 Composition API**: 现代化的表格处理
- **Naive UI**: 功能丰富的数据表格组件
- **TypeScript**: 类型安全的表格数据定义
- **异步处理**: Promise/async-await数据处理

### 远程数据表格实现
```typescript
// 表格数据接口
interface TableData {
  id: string;
  name: string;
  status: 'active' | 'inactive';
  createTime: string;
  updateTime: string;
  // ... 其他字段
}

// 表格状态管理
const {
  data,
  loading,
  columns,
  pagination,
  searchParams,
  handleSearch,
  handleReset,
  handleSort,
  handleFilter,
  handlePaginationChange
} = useRemoteTable<TableData>();

// 表格列定义
const columns = [
  {
    title: 'ID',
    key: 'id',
    width: 80,
    fixed: 'left',
    sorter: true
  },
  {
    title: '名称',
    key: 'name',
    width: 200,
    sorter: true,
    filterable: true,
    filterOptions: [
      { label: '包含A', value: 'contains_a' },
      { label: '包含B', value: 'contains_b' }
    ]
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const statusMap = {
        active: { color: 'success', text: '活跃' },
        inactive: { color: 'error', text: '非活跃' }
      };
      const status = statusMap[row.status];
      return h(NTag, { type: status.color }, status.text);
    }
  },
  {
    title: '创建时间',
    key: 'createTime',
    width: 180,
    sorter: true,
    render: (row) => dayjs(row.createTime).format('YYYY-MM-DD HH:mm:ss')
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    fixed: 'right',
    render: (row) => {
      return h('div', { class: 'flex gap-2' }, [
        h(NButton, { size: 'small', onClick: () => handleEdit(row) }, '编辑'),
        h(NPopconfirm, {
          onPositiveClick: () => handleDelete(row.id)
        }, {
          default: () => '确认删除？',
          trigger: () => h(NButton, { size: 'small', type: 'error' }, '删除')
        })
      ]);
    }
  }
];

// 分页配置
const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  showQuickJumper: true
});

// 远程数据加载
const loadTableData = async () => {
  loading.value = true;
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...searchParams,
      ...sortParams
    };
    const response = await api.getTableData(params);
    data.value = response.data.list;
    pagination.itemCount = response.data.total;
  } catch (error) {
    $message.error('数据加载失败');
  } finally {
    loading.value = false;
  }
};
```

### 行内编辑表格实现
```typescript
// 编辑状态管理
const editingKey = ref<string | null>(null);
const editingData = reactive<Record<string, any>>({});

// 开始编辑
const handleEdit = (row: TableData) => {
  editingKey.value = row.id;
  Object.assign(editingData, row);
};

// 保存编辑
const handleSave = async (rowId: string) => {
  try {
    await api.updateData(rowId, editingData);
    $message.success('保存成功');
    editingKey.value = null;
    loadTableData(); // 重新加载数据
  } catch (error) {
    $message.error('保存失败');
  }
};

// 取消编辑
const handleCancel = () => {
  editingKey.value = null;
  Object.keys(editingData).forEach(key => {
    delete editingData[key];
  });
};

// 编辑单元格渲染
const renderCell = (row: TableData, column: TableColumn) => {
  const isEditing = editingKey.value === row.id;

  if (isEditing) {
    switch (column.key) {
      case 'name':
        return h(NInput, {
          value: editingData.name,
          onUpdateValue: (value) => { editingData.name = value; }
        });
      case 'status':
        return h(NSelect, {
          value: editingData.status,
          options: [
            { label: '活跃', value: 'active' },
            { label: '非活跃', value: 'inactive' }
          ],
          onUpdateValue: (value) => { editingData.status = value; }
        });
      default:
        return row[column.key];
    }
  }

  return row[column.key];
};

// 操作列渲染
const renderActions = (row: TableData) => {
  const isEditing = editingKey.value === row.id;

  if (isEditing) {
    return h('div', { class: 'flex gap-2' }, [
      h(NButton, {
        size: 'small',
        type: 'primary',
        onClick: () => handleSave(row.id)
      }, '保存'),
      h(NButton, {
        size: 'small',
        onClick: handleCancel
      }, '取消')
    ]);
  }

  return h(NButton, {
    size: 'small',
    onClick: () => handleEdit(row)
  }, '编辑');
};
```

## 🎨 高级特性

### 搜索和筛选
```typescript
// 搜索参数
const searchParams = reactive({
  keyword: '',
  status: null,
  dateRange: null,
  category: []
});

// 搜索表单
const SearchForm = () => h('div', { class: 'mb-4 p-4 bg-gray-50 rounded' }, [
  h(NRow, { gutter: 16 }, [
    h(NCol, { span: 6 }, [
      h(NInput, {
        value: searchParams.keyword,
        placeholder: '搜索关键词',
        onUpdateValue: (value) => { searchParams.keyword = value; }
      })
    ]),
    h(NCol, { span: 6 }, [
      h(NSelect, {
        value: searchParams.status,
        placeholder: '选择状态',
        options: statusOptions,
        onUpdateValue: (value) => { searchParams.status = value; }
      })
    ]),
    h(NCol, { span: 12 }, [
      h('div', { class: 'flex gap-2' }, [
        h(NButton, { type: 'primary', onClick: handleSearch }, '搜索'),
        h(NButton, { onClick: handleReset }, '重置')
      ])
    ])
  ])
]);
```

### 批量操作
```typescript
// 选中行管理
const checkedRowKeys = ref<string[]>([]);

// 批量操作处理
const handleBatchDelete = async () => {
  if (checkedRowKeys.value.length === 0) {
    $message.warning('请选择要删除的数据');
    return;
  }

  try {
    await NDialog.warning({
      title: '批量删除确认',
      content: `确认删除选中的 ${checkedRowKeys.value.length} 条数据？`,
      positiveText: '确认删除',
      onPositiveClick: async () => {
        await api.batchDelete(checkedRowKeys.value);
        $message.success('删除成功');
        checkedRowKeys.value = [];
        loadTableData();
      }
    });
  } catch (error) {
    $message.error('删除失败');
  }
};
```

## 📱 响应式适配

- **列宽适配**: 根据屏幕宽度调整列宽
- **固定列**: 重要列的固定显示
- **横向滚动**: 数据过多时的横向滚动
- **移动端优化**: 移动端的操作优化

## 🔍 性能优化

- **虚拟滚动**: 大数据量的性能优化
- **分页加载**: 分批加载减少内存占用
- **数据缓存**: 避免重复请求
- **防抖搜索**: 搜索输入的防抖处理

## 🚀 使用指南

### AI模型使用示例
```
请基于 templates/table/remote/index.vue 模板，为我创建一个部署记录表格，要求：
1. 显示部署ID、项目名称、分支、状态、部署时间等信息
2. 支持按项目、状态、时间范围搜索
3. 支持状态列的筛选和排序
4. 包含查看详情、重新部署、回滚等操作按钮
5. 保持现有的分页和加载状态功能
```

### 适配建议
1. **数据结构**: 根据业务数据调整接口和数据结构
2. **列定义**: 修改表格列以适应业务需求
3. **搜索条件**: 根据业务特点设置搜索字段
4. **操作按钮**: 根据权限和业务流程设置操作按钮

## ⚠️ 重要说明

**本模板仅供参考使用，未经允许不得直接修改！**

---
**模板来源**: ui/src/views/pro-naive/table/
**最后更新**: 2025-01-30