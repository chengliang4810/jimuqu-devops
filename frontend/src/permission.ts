import 'nprogress/nprogress.css'; // progress bar style

import NProgress from 'nprogress'; // progress bar

import router from '@/router';

NProgress.configure({ showSpinner: false });

router.beforeEach(async (to, from, next) => {
  NProgress.start();

  // 简化权限控制，直接允许访问所有路由
  next();
});

router.afterEach(() => {
  NProgress.done();
});
