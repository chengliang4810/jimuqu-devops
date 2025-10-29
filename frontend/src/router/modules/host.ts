import { ServerIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/host',
    name: 'host',
    component: Layout,
    redirect: '/host/management',
    meta: {
      title: {
        zh_CN: '主机管理',
        en_US: 'Host Management',
      },
      icon: shallowRef(ServerIcon),
      orderNo: 1,
    },
    children: [
      {
        path: 'management',
        name: 'host-management',
        component: () => import('@/pages/host/index.vue'),
        meta: {
          title: {
            zh_CN: '主机管理',
            en_US: 'Host Management',
          },
          icon: shallowRef(ServerIcon),
          orderNo: 1,
        },
      },
    ],
  },
];