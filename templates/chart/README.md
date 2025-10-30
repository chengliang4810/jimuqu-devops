# 图表组件模板

## 📋 模板概述

本目录包含各种类型的图表组件模板，基于主流图表库实现，涵盖数据可视化、统计报表、实时监控等应用场景。

## 📁 模板结构

```
chart/
├── echarts/
│   └── index.vue          # ECharts图表模板
├── antv/
│   ├── index.vue          # AntV图表模板
│   └── modules/
│       └── antv-flow.vue  # AntV流程图
└── vchart/
    └── index.vue          # VChart商业图表模板
```

## 🎯 图表类型

### ECharts图表 (echarts/index.vue)
- **柱状图**: 单列、多列、堆叠柱状图
- **折线图**: 单线、多线、面积图
- **饼图**: 基础饼图、环形图、玫瑰图
- **散点图**: 基础散点图、气泡图
- **仪表盘**: 进度指示器
- **组合图**: 多种图表类型组合

### AntV图表 (antv/index.vue)
- **G2图表**: 统计图表、分布图、关联图
- **G6图表**: 关系图、流程图、拓扑图
- **可视化效果**: 动画效果、交互效果

### VChart图表 (vchart/index.vue)
- **商业报表**: 专业财务报表
- **数据透视**: 多维度数据分析
- **高级图表**: 瀑布图、漏斗图、桑基图

## 🔧 技术实现

### ECharts实现
```typescript
import * as echarts from 'echarts/core';
import { BarChart, LineChart, PieChart } from 'echarts/charts';
import { TitleComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';

// 注册必需的组件
echarts.use([
  BarChart, LineChart, PieChart,
  TitleComponent, TooltipComponent, LegendComponent,
  CanvasRenderer
]);

// 图表配置
const chartOptions = computed(() => ({
  title: {
    text: '部署统计',
    left: 'center'
  },
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'shadow'
    }
  },
  legend: {
    data: ['成功部署', '失败部署'],
    bottom: 0
  },
  xAxis: {
    type: 'category',
    data: chartData.value.labels
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      name: '成功部署',
      type: 'bar',
      data: chartData.value.success,
      itemStyle: {
        color: '#52c41a'
      }
    },
    {
      name: '失败部署',
      type: 'bar',
      data: chartData.value.failed,
      itemStyle: {
        color: '#ff4d4f'
      }
    }
  ]
}));

// 图表实例
const chartRef = ref<HTMLDivElement>();
let chartInstance: echarts.ECharts | null = null;

// 初始化图表
const initChart = () => {
  if (chartRef.value) {
    chartInstance = echarts.init(chartRef.value);
    chartInstance.setOption(chartOptions.value);
  }
};

// 响应式更新
const updateChart = () => {
  if (chartInstance) {
    chartInstance.setOption(chartOptions.value);
  }
};

// 窗口大小变化时重绘
const handleResize = () => {
  chartInstance?.resize();
};

onMounted(() => {
  initChart();
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  chartInstance?.dispose();
});
```

### AntV G2实现
```typescript
import { Chart } from '@antv/g2';

// 图表实例
const chartRef = ref<HTMLDivElement>();
let chartInstance: Chart | null = null;

// 初始化G2图表
const initG2Chart = () => {
  if (chartRef.value) {
    chartInstance = new Chart({
      container: chartRef.value,
      autoFit: true,
      height: 400
    });

    // 数据处理
    chartInstance.data(chartData.value);

    // 坐标轴
    chartInstance.scale('date', {
      type: 'time',
      tickCount: 8
    });

    chartInstance.scale('value', {
      nice: true
    });

    // 绘制折线图
    chartInstance
      .line()
      .position('date*value')
      .color('type')
      .shape('smooth');

    // 添加点
    chartInstance
      .point()
      .position('date*value')
      .color('type')
      .shape('circle')
      .style({
        stroke: '#fff',
        lineWidth: 2
      });

    // 添加图例
    chartInstance.legend({
      position: 'bottom'
    });

    // 添加tooltip
    chartInstance.tooltip({
      showCrosshairs: true,
      shared: true
    });

    chartInstance.render();
  }
};

// 更新数据
const updateG2Chart = () => {
  if (chartInstance) {
    chartInstance.changeData(chartData.value);
  }
};
```

### VChart实现
```typescript
import VChart from 'vue-echarts';

// 图表配置
const vchartOptions = computed(() => ({
  title: {
    text: '项目监控仪表盘',
    left: 'center'
  },
  tooltip: {
    trigger: 'item'
  },
  radar: {
    indicator: [
      { name: '代码质量', max: 100 },
      { name: '部署频率', max: 100 },
      { name: '成功率', max: 100 },
      { name: '响应时间', max: 100 },
      { name: '稳定性', max: 100 }
    ]
  },
  series: [{
    type: 'radar',
    data: [
      {
        value: [85, 90, 95, 80, 88],
        name: '当前状态',
        areaStyle: {
          color: 'rgba(24, 144, 255, 0.3)'
        }
      }
    ]
  }]
}));
```

## 📊 数据处理

### 数据格式化
```typescript
// 图表数据接口
interface ChartData {
  labels: string[];
  datasets: {
    label: string;
    data: number[];
    backgroundColor?: string;
    borderColor?: string;
  }[];
}

// 数据转换
const transformDataForChart = (rawData: any[]): ChartData => {
  const labels = rawData.map(item => item.date);
  const successData = rawData.map(item => item.successCount);
  const failedData = rawData.map(item => item.failedCount);

  return {
    labels,
    datasets: [
      {
        label: '成功部署',
        data: successData,
        backgroundColor: 'rgba(82, 196, 26, 0.8)',
        borderColor: '#52c41a'
      },
      {
        label: '失败部署',
        data: failedData,
        backgroundColor: 'rgba(255, 77, 79, 0.8)',
        borderColor: '#ff4d4f'
      }
    ]
  };
};
```

### 实时数据更新
```typescript
// WebSocket实时数据
const ws = ref<WebSocket | null>(null);

const connectWebSocket = () => {
  ws.value = new WebSocket('ws://localhost:8000/ws/chart-data');

  ws.value.onmessage = (event) => {
    const newData = JSON.parse(event.data);
    chartData.value.push(newData);

    // 保持最近30个数据点
    if (chartData.value.length > 30) {
      chartData.value.shift();
    }

    updateChart();
  };
};

// 定时刷新
const refreshInterval = ref<NodeJS.Timeout | null>(null);

const startAutoRefresh = () => {
  refreshInterval.value = setInterval(async () => {
    try {
      const response = await api.getChartData();
      chartData.value = response.data;
      updateChart();
    } catch (error) {
      console.error('数据刷新失败:', error);
    }
  }, 5000); // 每5秒刷新一次
};
```

## 🎨 交互功能

### 图表事件
```typescript
// ECharts事件处理
const handleChartClick = (params: any) => {
  console.log('图表点击:', params);
  // 可以实现点击图表元素后的业务逻辑
};

const handleChartMouseOver = (params: any) => {
  // 鼠标悬停效果
};

// 图表缩放
const handleDataZoom = (params: any) => {
  // 数据范围缩放处理
};
```

### 图表配置
```typescript
// 主题配置
const chartTheme = {
  backgroundColor: '#fff',
  color: ['#1890ff', '#52c41a', '#faad14', '#f5222d'],
  textStyle: {
    fontFamily: 'Arial, sans-serif'
  }
};

// 响应式配置
const responsiveConfig = {
  baseOption: chartOptions.value,
  media: [
    {
      query: { maxWidth: 768 },
      option: {
        legend: { bottom: 10 },
        grid: { bottom: 80 }
      }
    }
  ]
};
```

## 📱 响应式设计

- **自适应尺寸**: 图表自动适应容器大小
- **移动端优化**: 移动端图表显示优化
- **交互适配**: 触摸设备的交互优化

## 🔍 性能优化

- **按需加载**: 只加载需要的图表组件
- **数据缓存**: 避免重复数据处理
- **防抖处理**: 频繁数据更新的防抖
- **懒加载**: 图表组件的懒加载

## 🚀 使用指南

### AI模型使用示例
```
请基于 templates/chart/echarts/index.vue 模板，为我创建一个部署监控仪表板，要求：
1. 显示近30天的部署成功率和失败率趋势
2. 包含项目部署量的柱状图对比
3. 添加实时部署状态饼图
4. 支持图表的时间范围筛选
5. 保持现有的图表交互和动画效果
```

### 适配建议
1. **数据结构**: 根据业务数据调整图表数据格式
2. **图表类型**: 选择合适的图表类型展示数据
3. **颜色主题**: 根据品牌调整图表颜色
4. **交互功能**: 添加符合业务需求的交互功能

## ⚠️ 重要说明

**本模板仅供参考使用，未经允许不得直接修改！**

---
**模板来源**: ui/src/views/plugin/charts/
**最后更新**: 2025-01-30