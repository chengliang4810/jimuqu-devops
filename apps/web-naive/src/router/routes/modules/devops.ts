import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'HostManagement',
    path: '/host',
    component: () => import('#/views/devops/host/index.vue'),
    meta: {
      title: '主机管理',
      icon: 'mdi:server',
      order: 2,
    },
  },
  {
    name: 'ProjectManagement',
    path: '/project',
    component: () => import('#/views/devops/project/index.vue'),
    meta: {
      title: '项目管理',
      icon: 'mdi:source-repository',
      order: 3,
    },
  },
  {
    name: 'DeployConfig',
    path: '/deploy-config',
    component: () => import('#/views/devops/deploy-config/index.vue'),
    meta: {
      title: '部署配置',
      icon: 'mdi:cog-outline',
      order: 4,
    },
  },
  {
    name: 'DeployRecord',
    path: '/deploy-record',
    component: () => import('#/views/devops/deploy-record/index.vue'),
    meta: {
      title: '部署记录',
      icon: 'mdi:history',
      order: 5,
    },
  },
];

export default routes;
