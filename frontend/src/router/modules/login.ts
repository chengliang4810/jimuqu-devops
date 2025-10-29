export default [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/login/index.vue'),
    meta: {
      title: {
        zh_CN: '登录',
        en_US: 'Login',
      },
      hidden: true, // 不在菜单中显示
    },
  },
];
