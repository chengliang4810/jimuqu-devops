import type { RouteRecordRaw } from 'vue-router';

// 首页路由已移至 core.ts 作为核心路由
const routes: RouteRecordRaw[] = [
  {
    name: 'Home',
    path: '/home',
    component: () => import('#/views/dashboard/analytics/index.vue'),
    meta: {
      title: '首页',
      icon: 'mdi:home-outline',
      order: 1,
    },
  },
];

export default routes;
