<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 页面头部 -->
    <div class="flex-y-center justify-between">
      <h1 class="text-18px font-bold">DevOps 仪表盘</h1>
      <NButton type="info" @click="refreshData">
        <Icon icon="material-symbols:refresh" class="mr-4px text-16px" />
        刷新数据
      </NButton>
    </div>

    <!-- 统计卡片区域 -->
    <div class="grid grid-cols-1 gap-16px sm:grid-cols-2 lg:grid-cols-4">
      <div class="card-wrapper p-16px">
        <div class="flex justify-between items-center">
          <div>
            <div class="text-12px text-gray-500 mb-4px">应用总数</div>
            <div class="text-24px font-bold text-blue-600">{{ statistics.totalApplications }}</div>
          </div>
          <div class="w-48px h-48px bg-blue-100 rounded-full flex items-center justify-center">
            <Icon icon="material-symbols:apps" class="text-24px text-blue-600" />
          </div>
        </div>
      </div>

      <div class="card-wrapper p-16px">
        <div class="flex justify-between items-center">
          <div>
            <div class="text-12px text-gray-500 mb-4px">构建总次数</div>
            <div class="text-24px font-bold text-green-600">{{ statistics.totalBuilds }}</div>
          </div>
          <div class="w-48px h-48px bg-green-100 rounded-full flex items-center justify-center">
            <Icon icon="material-symbols:build" class="text-24px text-green-600" />
          </div>
        </div>
      </div>

      <div class="card-wrapper p-16px">
        <div class="flex justify-between items-center">
          <div>
            <div class="text-12px text-gray-500 mb-4px">成功率</div>
            <div class="text-24px font-bold text-green-600">{{ successRate }}%</div>
          </div>
          <div class="w-48px h-48px bg-green-100 rounded-full flex items-center justify-center">
            <Icon icon="material-symbols:check-circle" class="text-24px text-green-600" />
          </div>
        </div>
      </div>

      <div class="card-wrapper p-16px">
        <div class="flex justify-between items-center">
          <div>
            <div class="text-12px text-gray-500 mb-4px">在线主机</div>
            <div class="text-24px font-bold text-purple-600">{{ statistics.onlineHosts }}</div>
          </div>
          <div class="w-48px h-48px bg-purple-100 rounded-full flex items-center justify-center">
            <Icon icon="material-symbols:computer" class="text-24px text-purple-600" />
          </div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="grid grid-cols-1 gap-16px lg:grid-cols-2">
      <!-- 构建趋势图 -->
      <div class="card-wrapper p-16px">
        <h3 class="text-16px font-bold mb-16px">构建趋势 (最近7天)</h3>
        <div ref="buildTrendChart" class="h-300px w-full"></div>
      </div>

      <!-- 成功率分布图 -->
      <div class="card-wrapper p-16px">
        <h3 class="text-16px font-bold mb-16px">构建状态分布</h3>
        <div ref="statusChart" class="h-300px w-full"></div>
      </div>
    </div>

    <!-- 最近构建列表 -->
    <div class="card-wrapper p-16px">
      <div class="flex justify-between items-center mb-16px">
        <h3 class="text-16px font-bold">最近构建</h3>
        <NButton size="small" text type="primary" @click="$router.push('/devops/build')">
          查看全部
          <Icon icon="material-symbols:arrow-forward" class="ml-4px" />
        </NButton>
      </div>
      <NDataTable
        :columns="recentBuildsColumns"
        :data="recentBuilds"
        :pagination="false"
        :scroll-x="800"
        size="small"
      />
    </div>

    <!-- 系统状态 -->
    <div class="grid grid-cols-1 gap-16px lg:grid-cols-2">
      <!-- 主机状态 -->
      <div class="card-wrapper p-16px">
        <h3 class="text-16px font-bold mb-16px">主机状态</h3>
        <div class="space-y-12px">
          <div v-for="host in hostStatus" :key="host.id" class="flex justify-between items-center">
            <div class="flex items-center gap-8px">
              <div 
                :class="[
                  'w-8px h-8px rounded-full',
                  host.status === 'ONLINE' ? 'bg-green-500' : 
                  host.status === 'OFFLINE' ? 'bg-gray-400' : 'bg-red-500'
                ]"
              ></div>
              <span class="text-14px">{{ host.name }}</span>
            </div>
            <NTag :type="getHostStatusType(host.status)" size="small">
              {{ getHostStatusText(host.status) }}
            </NTag>
          </div>
        </div>
      </div>

      <!-- 快速操作 -->
      <div class="card-wrapper p-16px">
        <h3 class="text-16px font-bold mb-16px">快速操作</h3>
        <div class="grid grid-cols-2 gap-12px">
          <NButton type="primary" @click="$router.push('/devops/host')">
            <Icon icon="material-symbols:add" class="mr-4px" />
            添加主机
          </NButton>
          <NButton type="success" @click="$router.push('/devops/application')">
            <Icon icon="material-symbols:add" class="mr-4px" />
            添加应用
          </NButton>
          <NButton type="info" @click="refreshHostStatus">
            <Icon icon="material-symbols:refresh" class="mr-4px" />
            刷新主机状态
          </NButton>
          <NButton type="warning" @click="$router.push('/devops/build')">
            <Icon icon="material-symbols:visibility" class="mr-4px" />
            查看构建
          </NButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, h } from 'vue';
import type { Ref } from 'vue';
import { NButton, NDataTable, NTag, useMessage } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { Icon } from '@iconify/vue';
import * as echarts from 'echarts';

interface Statistics {
  totalApplications: number;
  totalBuilds: number;
  successfulBuilds: number;
  failedBuilds: number;
  onlineHosts: number;
}

interface Build {
  id: number;
  applicationName: string;
  buildNumber: number;
  status: string;
  triggeredBy: string;
  duration: number;
  createTime: string;
}

interface Host {
  id: number;
  name: string;
  status: 'ONLINE' | 'OFFLINE' | 'ERROR';
}

const message = useMessage();

// 图表引用
const buildTrendChart = ref<HTMLElement>();
const statusChart = ref<HTMLElement>();

// 统计数据
const statistics = reactive<Statistics>({
  totalApplications: 0,
  totalBuilds: 0,
  successfulBuilds: 0,
  failedBuilds: 0,
  onlineHosts: 0
});

// 计算成功率
const successRate = computed(() => {
  if (statistics.totalBuilds === 0) return 0;
  return Math.round((statistics.successfulBuilds / statistics.totalBuilds) * 100);
});

// 最近构建
const recentBuilds: Ref<Build[]> = ref([]);

// 主机状态
const hostStatus: Ref<Host[]> = ref([]);

// 最近构建表格列定义
const recentBuildsColumns: DataTableColumns<Build> = [
  {
    title: '应用名称',
    key: 'applicationName',
    width: 150
  },
  {
    title: '构建编号',
    key: 'buildNumber',
    width: 100,
    render: (row) => `#${row.buildNumber}`
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const typeMap = {
        SUCCESS: 'success',
        FAILED: 'error',
        RUNNING: 'info',
        PENDING: 'default'
      };
      const textMap = {
        SUCCESS: '成功',
        FAILED: '失败', 
        RUNNING: '运行中',
        PENDING: '等待中'
      };
      return h(NTag, { type: typeMap[row.status], size: 'small' }, () => textMap[row.status]);
    }
  },
  {
    title: '触发方式',
    key: 'triggeredBy',
    width: 100
  },
  {
    title: '耗时',
    key: 'duration',
    width: 80,
    render: (row) => `${row.duration}s`
  },
  {
    title: '时间',
    key: 'createTime',
    render: (row) => new Date(row.createTime).toLocaleString()
  }
];

// 主机状态相关方法
const getHostStatusType = (status: string) => {
  const typeMap = {
    ONLINE: 'success',
    OFFLINE: 'default',
    ERROR: 'error'
  };
  return typeMap[status] || 'default';
};

const getHostStatusText = (status: string) => {
  const textMap = {
    ONLINE: '在线',
    OFFLINE: '离线',
    ERROR: '错误'
  };
  return textMap[status] || '未知';
};

// 初始化构建趋势图
const initBuildTrendChart = () => {
  if (!buildTrendChart.value) return;

  const chart = echarts.init(buildTrendChart.value);
  
  const option = {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['成功', '失败']
    },
    xAxis: {
      type: 'category',
      data: ['1/15', '1/16', '1/17', '1/18', '1/19', '1/20', '1/21']
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '成功',
        type: 'line',
        data: [12, 19, 15, 22, 18, 25, 20],
        itemStyle: { color: '#18a058' }
      },
      {
        name: '失败',
        type: 'line',
        data: [2, 3, 1, 4, 2, 1, 3],
        itemStyle: { color: '#d03050' }
      }
    ]
  };

  chart.setOption(option);
  
  // 响应式处理
  window.addEventListener('resize', () => {
    chart.resize();
  });
};

// 初始化状态分布图
const initStatusChart = () => {
  if (!statusChart.value) return;

  const chart = echarts.init(statusChart.value);
  
  const option = {
    tooltip: {
      trigger: 'item'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '构建状态',
        type: 'pie',
        radius: '50%',
        data: [
          { value: statistics.successfulBuilds, name: '成功', itemStyle: { color: '#18a058' } },
          { value: statistics.failedBuilds, name: '失败', itemStyle: { color: '#d03050' } },
          { value: 5, name: '运行中', itemStyle: { color: '#2080f0' } },
          { value: 2, name: '等待中', itemStyle: { color: '#909399' } }
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  };

  chart.setOption(option);
  
  // 响应式处理
  window.addEventListener('resize', () => {
    chart.resize();
  });
};

// 获取统计数据
const getStatistics = async () => {
  try {
    // 模拟数据 - 实际应调用后端API
    Object.assign(statistics, {
      totalApplications: 5,
      totalBuilds: 120,
      successfulBuilds: 98,
      failedBuilds: 22,
      onlineHosts: 3
    });
  } catch (error) {
    message.error('获取统计数据失败');
    console.error('获取统计数据失败:', error);
  }
};

// 获取最近构建
const getRecentBuilds = async () => {
  try {
    // 模拟数据
    recentBuilds.value = [
      {
        id: 1,
        applicationName: 'demo-spring-boot',
        buildNumber: 15,
        status: 'SUCCESS',
        triggeredBy: 'WEBHOOK',
        duration: 180,
        createTime: new Date().toISOString()
      },
      {
        id: 2,
        applicationName: 'demo-vue-app',
        buildNumber: 8,
        status: 'RUNNING',
        triggeredBy: 'MANUAL',
        duration: 0,
        createTime: new Date(Date.now() - 60000).toISOString()
      },
      {
        id: 3,
        applicationName: 'demo-spring-boot',
        buildNumber: 14,
        status: 'FAILED',
        triggeredBy: 'WEBHOOK',
        duration: 120,
        createTime: new Date(Date.now() - 300000).toISOString()
      }
    ];
  } catch (error) {
    message.error('获取最近构建失败');
    console.error('获取最近构建失败:', error);
  }
};

// 获取主机状态
const getHostStatus = async () => {
  try {
    // 模拟数据
    hostStatus.value = [
      { id: 1, name: '生产服务器-1', status: 'ONLINE' },
      { id: 2, name: '生产服务器-2', status: 'ONLINE' },
      { id: 3, name: '测试服务器', status: 'OFFLINE' },
      { id: 4, name: '开发服务器', status: 'ONLINE' }
    ];
  } catch (error) {
    message.error('获取主机状态失败');
    console.error('获取主机状态失败:', error);
  }
};

// 刷新数据
const refreshData = async () => {
  const loadingMessage = message.loading('正在刷新数据...', { duration: 0 });
  try {
    await Promise.all([
      getStatistics(),
      getRecentBuilds(),
      getHostStatus()
    ]);
    
    // 重新初始化图表
    await nextTick();
    initBuildTrendChart();
    initStatusChart();
    
    loadingMessage.destroy();
    message.success('数据刷新成功');
  } catch (error) {
    loadingMessage.destroy();
    message.error('数据刷新失败');
    console.error('数据刷新失败:', error);
  }
};

// 刷新主机状态
const refreshHostStatus = async () => {
  const loadingMessage = message.loading('正在刷新主机状态...', { duration: 0 });
  try {
    const response = await fetch('/api/hosts/status/update', {
      method: 'POST'
    });
    const result = await response.json();
    
    loadingMessage.destroy();
    
    if (result.code === 200) {
      message.success('主机状态刷新已开始');
      setTimeout(() => {
        getHostStatus();
        getStatistics();
      }, 3000);
    } else {
      message.error(result.message || '刷新主机状态失败');
    }
  } catch (error) {
    loadingMessage.destroy();
    message.error('刷新主机状态失败');
    console.error('刷新主机状态失败:', error);
  }
};

// 生命周期
onMounted(async () => {
  await refreshData();
});
</script>