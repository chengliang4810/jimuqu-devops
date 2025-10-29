import { DashboardIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/dashboard',
    name: 'dashboard',
    component: Layout,
    redirect: '/dashboard/home',
    meta: {
      title: {
        zh_CN: '仪表盘',
        en_US: 'Dashboard',
      },
      icon: shallowRef(DashboardIcon),
      orderNo: 0,
    },
    children: [
      {
        path: 'home',
        name: 'dashboard-home',
        component: () => import('@/pages/dashboard/base/index.vue'),
        meta: {
          title: {
            zh_CN: '仪表盘',
            en_US: 'Dashboard',
          },
          icon: shallowRef(DashboardIcon),
          orderNo: 0,
          hideInMenu: true, // 隐藏在菜单中，但作为默认显示页面
        },
      },
    ],
  },
];
