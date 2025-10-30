# 表单组件模板

## 📋 模板概述

本目录包含各种类型的表单组件模板，从基础表单到复杂的多步骤表单，涵盖所有常见的表单使用场景。

## 📁 模板结构

```
form/
├── basic/
│   └── index.vue          # 基础表单模板
├── query/
│   └── index.vue          # 查询表单模板
└── step/
    └── index.vue          # 步骤表单模板
```

## 🎯 模板特性

### 基础表单 (basic/index.vue)
- **多种输入类型**: 文本、数字、选择、日期、文件等
- **表单验证**: 完整的前端验证规则
- **响应式布局**: 自适应不同屏幕尺寸
- **提交处理**: 表单提交和错误处理
- **重置功能**: 表单数据重置

### 查询表单 (query/index.vue)
- **条件搜索**: 多条件组合搜索
- **高级搜索**: 可展开的高级搜索选项
- **快捷操作**: 常用搜索条件的快捷按钮
- **搜索历史**: 搜索条件历史记录
- **实时搜索**: 输入时实时搜索建议

### 步骤表单 (step/index.vue)
- **向导模式**: 多步骤表单向导
- **进度指示**: 清晰的步骤进度显示
- **步骤验证**: 每个步骤的独立验证
- **上一步/下一步**: 灵活的步骤导航
- **数据保存**: 步骤间的数据暂存

## 🔧 技术实现

### 核心技术栈
- **Vue 3 Composition API**: 现代化的表单处理
- **Naive UI**: 丰富的表单组件库
- **TypeScript**: 类型安全的表单定义
- **Yup**: 表单验证规则定义

### 基础表单实现
```typescript
// 表单数据定义
interface FormState {
  name: string;
  email: string;
  age: number | null;
  gender: string | null;
  birthday: string | null;
  hobby: string[];
  description: string;
  avatar: string | null;
}

// 表单实例
const formRef = ref<FormInst | null>(null);
const formValue = reactive<FormState>({
  name: '',
  email: '',
  age: null,
  gender: null,
  birthday: null,
  hobby: [],
  description: '',
  avatar: null
});

// 验证规则
const rules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { min: 2, max: 20, message: '姓名长度为2-20个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  age: [
    { required: true, message: '请输入年龄', trigger: 'blur' },
    { type: 'number', min: 18, max: 100, message: '年龄必须在18-100之间', trigger: 'blur' }
  ]
};

// 表单提交
const handleSubmit = () => {
  formRef.value?.validate(async (errors) => {
    if (!errors) {
      try {
        loading.value = true;
        // API调用
        await createForm(formValue);
        $message.success('提交成功');
        handleReset();
      } catch (error) {
        $message.error('提交失败');
      } finally {
        loading.value = false;
      }
    }
  });
};
```

### 查询表单实现
```typescript
// 搜索参数
const searchParams = reactive({
  keyword: '',
  category: null,
  status: null,
  dateRange: null,
  priceRange: [null, null] as [number | null, number | null]
});

// 高级搜索展开状态
const showAdvanced = ref(false);

// 搜索处理
const handleSearch = () => {
  emit('search', { ...searchParams });
};

// 重置搜索
const handleReset = () => {
  Object.assign(searchParams, {
    keyword: '',
    category: null,
    status: null,
    dateRange: null,
    priceRange: [null, null]
  });
  emit('reset');
};

// 快捷搜索
const quickSearchOptions = [
  { label: '最近7天', value: 'last7days' },
  { label: '最近30天', value: 'last30days' },
  { label: '本月', value: 'thisMonth' }
];

const handleQuickSearch = (value: string) => {
  // 设置快捷搜索条件
  switch (value) {
    case 'last7days':
      searchParams.dateRange = [
        dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
        dayjs().format('YYYY-MM-DD')
      ];
      break;
    // ... 其他快捷选项
  }
  handleSearch();
};
```

### 步骤表单实现
```typescript
// 当前步骤
const currentStep = ref(0);

// 步骤配置
const steps = [
  { title: '基本信息', description: '填写基本信息' },
  { title: '详细信息', description: '填写详细信息' },
  { title: '确认信息', description: '确认并提交' }
];

// 每个步骤的表单数据
const stepForms = reactive([
  { name: '', email: '', phone: '' },
  { address: '', company: '', position: '' },
  { remarks: '', attachments: [] }
]);

// 步骤验证
const validateStep = async (step: number) => {
  const formRef = refs[`stepForm${step}`] as FormInst;
  return new Promise((resolve) => {
    formRef?.validate((errors) => {
      resolve(!errors);
    });
  });
};

// 下一步
const handleNext = async () => {
  const isValid = await validateStep(currentStep.value);
  if (isValid) {
    if (currentStep.value < steps.length - 1) {
      currentStep.value++;
    }
  }
};

// 上一步
const handlePrev = () => {
  if (currentStep.value > 0) {
    currentStep.value--;
  }
};

// 提交所有步骤
const handleSubmit = async () => {
  const allData = stepForms.reduce((acc, form, index) => {
    return { ...acc, ...form };
  }, {});

  try {
    await submitMultiStepForm(allData);
    $message.success('提交成功');
  } catch (error) {
    $message.error('提交失败');
  }
};
```

## 🎨 UI设计特点

- **清晰布局**: 表单项的合理排列和分组
- **视觉层次**: 重要信息的突出显示
- **交互反馈**: 实时的验证反馈和状态提示
- **无障碍**: 支持键盘导航和屏幕阅读器
- **主题适配**: 支持深色/浅色主题

## 📱 响应式设计

- **桌面端**: 多列布局，充分利用屏幕空间
- **平板端**: 自适应列数调整
- **手机端**: 单列布局，优化触控体验

## 🔍 表单验证

### 验证规则类型
- **必填验证**: required字段验证
- **格式验证**: email、url、phone等格式验证
- **长度验证**: 字符串长度限制
- **数值验证**: 数值范围验证
- **自定义验证**: 复杂业务逻辑验证

### 验证触发时机
- **实时验证**: 输入时实时验证
- **失焦验证**: 字段失焦时验证
- **提交验证**: 表单提交时统一验证

## 🚀 使用指南

### AI模型使用示例
```
请基于 templates/form/basic/index.vue 模板，为我创建一个项目创建表单，要求：
1. 包含项目名称、描述、Git仓库地址、部署环境等字段
2. 添加相应的验证规则
3. 保持现有的表单布局和交互逻辑
4. 适配DevOps平台的业务需求
```

### 适配建议
1. **字段调整**: 根据业务需求添加/删除/修改表单字段
2. **验证规则**: 设置符合业务逻辑的验证规则
3. **API接口**: 替换为对应的业务API接口
4. **样式定制**: 根据设计稿调整表单样式

## ⚠️ 重要说明

**本模板仅供参考使用，未经允许不得直接修改！**

---
**模板来源**: ui/src/views/pro-naive/form/
**最后更新**: 2025-01-30