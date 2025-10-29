# Login 登录模板

## 概述

Login 模板用于用户认证、注册和密码重置页面，提供了完整的用户登录流程和安全验证机制。

## 功能特点

- 用户登录/注册切换
- 表单验证和错误处理
- 第三方登录集成
- 密码强度检测
- 记住登录状态
- 安全验证码

## 组件结构

```
login/
├── index.vue              # 主登录页面
├── components/            # 子组件
│   ├── Header.vue        # 页面头部组件
│   ├── Login.vue         # 登录表单组件
│   ├── Register.vue      # 注册表单组件
│   └── ThirdParty.vue    # 第三方登录组件
├── index.less            # 样式文件
└── README.md             # 说明文档
```

## 使用场景

- 用户登录页面
- 用户注册页面
- 密码重置页面
- 用户认证流程

## 最佳实践

### 主页面结构
```vue
<template>
  <div class="login-wrapper">
    <!-- 页面头部 -->
    <login-header />

    <!-- 登录容器 -->
    <div class="login-container">
      <div class="title-container">
        <h1 class="title">{{ t('pages.login.loginTitle') }}</h1>
        <div class="sub-title">
          <p class="tip">{{ type === 'register' ? '已有账号？' : '没有账号？' }}</p>
          <p class="tip" @click="switchType(type === 'register' ? 'login' : 'register')">
            {{ type === 'register' ? '立即登录' : '注册账号' }}
          </p>
        </div>
      </div>

      <!-- 登录/注册表单 -->
      <login v-if="type === 'login'" />
      <register v-else @register-success="switchType('login')" />
    </div>

    <!-- 页面底部 -->
    <footer class="copyright">Copyright @ 2024 Company Name</footer>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { t } from '@/locales';

import LoginHeader from './components/Header.vue';
import Login from './components/Login.vue';
import Register from './components/Register.vue';

defineOptions({
  name: 'LoginIndex',
});

const type = ref<'login' | 'register'>('login');

const switchType = (val: 'login' | 'register') => {
  type.value = val;
};
</script>
```

### 登录表单组件
```vue
<template>
  <t-form
    ref="formRef"
    class="login-form"
    :data="formData"
    :rules="FORM_RULES"
    label-width="0"
    @submit="handleSubmit"
  >
    <t-form-item name="username">
      <t-input
        v-model="formData.username"
        size="large"
        placeholder="请输入用户名"
        :prefix-icon="userIcon"
      />
    </t-form-item>

    <t-form-item name="password">
      <t-input
        v-model="formData.password"
        size="large"
        type="password"
        placeholder="请输入密码"
        :prefix-icon="lockIcon"
      />
    </t-form-item>

    <t-form-item name="verifyCode" v-if="showVerifyCode">
      <t-input
        v-model="formData.verifyCode"
        size="large"
        placeholder="请输入验证码"
        :prefix-icon="protectIcon"
      >
        <template #suffix>
          <img
            :src="verifyCodeUrl"
            class="verify-code-img"
            @click="refreshVerifyCode"
            alt="验证码"
          />
        </template>
      </t-input>
    </t-form-item>

    <div class="options-container">
      <t-checkbox v-model="formData.rememberMe">记住密码</t-checkbox>
      <t-link hover="color" @click="handleForgotPassword">忘记密码？</t-link>
    </div>

    <t-button
      type="submit"
      size="large"
      theme="primary"
      :loading="loading"
      block
    >
      登录
    </t-button>

    <!-- 第三方登录 -->
    <div class="third-party-login">
      <t-divider>或使用以下方式登录</t-divider>
      <div class="third-party-icons">
        <t-button
          v-for="item in thirdPartyList"
          :key="item.type"
          variant="text"
          shape="square"
          size="large"
          @click="handleThirdPartyLogin(item.type)"
        >
          <component :is="item.icon" />
        </t-button>
      </div>
    </div>
  </t-form>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';

import { login, getUserInfo } from '@/api/auth';
import { useUserStore } from '@/store';

defineOptions({
  name: 'LoginForm',
});

const router = useRouter();
const userStore = useUserStore();

const formRef = ref();
const loading = ref(false);
const showVerifyCode = ref(false);

const formData = reactive({
  username: '',
  password: '',
  verifyCode: '',
  rememberMe: false,
});

const FORM_RULES = {
  username: [
    { required: true, message: '请输入用户名', type: 'error' },
    { min: 3, max: 20, message: '用户名长度为3-20个字符', type: 'warning' },
  ],
  password: [
    { required: true, message: '请输入密码', type: 'error' },
    { min: 6, message: '密码长度至少6个字符', type: 'warning' },
  ],
  verifyCode: [
    { required: true, message: '请输入验证码', type: 'error' },
    { len: 4, message: '验证码为4位', type: 'error' },
  ],
};

const handleSubmit = async ({ validateResult, firstError }: any) => {
  if (validateResult === true) {
    loading.value = true;
    try {
      await login(formData);
      const userInfo = await getUserInfo();
      userStore.setUserInfo(userInfo);

      MessagePlugin.success('登录成功');
      router.push('/dashboard');
    } catch (error: any) {
      MessagePlugin.error(error.message || '登录失败');

      // 显示验证码
      if (error.code === 'NEED_VERIFY_CODE') {
        showVerifyCode.value = true;
        refreshVerifyCode();
      }
    } finally {
      loading.value = false;
    }
  } else {
    MessagePlugin.warning(firstError);
  }
};

const handleForgotPassword = () => {
  router.push('/forgot-password');
};

const handleThirdPartyLogin = (type: string) => {
  // 处理第三方登录
  console.log('第三方登录:', type);
};
</script>
```

### 注册表单组件
```vue
<template>
  <t-form
    ref="formRef"
    class="register-form"
    :data="formData"
    :rules="FORM_RULES"
    label-width="0"
    @submit="handleSubmit"
  >
    <t-form-item name="username">
      <t-input
        v-model="formData.username"
        size="large"
        placeholder="请输入用户名"
        :prefix-icon="userIcon"
      />
    </t-form-item>

    <t-form-item name="email">
      <t-input
        v-model="formData.email"
        size="large"
        placeholder="请输入邮箱"
        :prefix-icon="mailIcon"
      />
    </t-form-item>

    <t-form-item name="password">
      <t-input
        v-model="formData.password"
        size="large"
        type="password"
        placeholder="请输入密码"
        :prefix-icon="lockIcon"
        @input="checkPasswordStrength"
      />
      <!-- 密码强度指示器 -->
      <div class="password-strength" v-if="formData.password">
        <div class="strength-bar">
          <div
            class="strength-fill"
            :class="passwordStrength.level"
            :style="{ width: passwordStrength.width }"
          ></div>
        </div>
        <span class="strength-text">{{ passwordStrength.text }}</span>
      </div>
    </t-form-item>

    <t-form-item name="confirmPassword">
      <t-input
        v-model="formData.confirmPassword"
        size="large"
        type="password"
        placeholder="请确认密码"
        :prefix-icon="lockIcon"
      />
    </t-form-item>

    <t-form-item name="agreement">
      <t-checkbox v-model="formData.agreement">
        我已阅读并同意
        <t-link hover="color" @click="showUserAgreement">用户协议</t-link>
        和
        <t-link hover="color" @click="showPrivacyPolicy">隐私政策</t-link>
      </t-checkbox>
    </t-form-item>

    <t-button
      type="submit"
      size="large"
      theme="primary"
      :loading="loading"
      block
    >
      注册
    </t-button>
  </t-form>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';

import { register } from '@/api/auth';

const emit = defineEmits<{
  registerSuccess: [];
}>();

defineOptions({
  name: 'RegisterForm',
});

const formRef = ref();
const loading = ref(false);

const formData = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  agreement: false,
});

const passwordStrength = ref({
  level: '',
  width: '0%',
  text: '',
});

const FORM_RULES = {
  username: [
    { required: true, message: '请输入用户名', type: 'error' },
    { pattern: /^[a-zA-Z0-9_]{3,20}$/, message: '用户名只能包含字母、数字和下划线，长度3-20位', type: 'error' },
  ],
  email: [
    { required: true, message: '请输入邮箱', type: 'error' },
    { email: true, message: '请输入有效的邮箱地址', type: 'error' },
  ],
  password: [
    { required: true, message: '请输入密码', type: 'error' },
    { min: 8, message: '密码长度至少8位', type: 'warning' },
    { pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,}$/, message: '密码必须包含大小写字母和数字', type: 'warning' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', type: 'error' },
    {
      validator: (value: string) => value === formData.password,
      message: '两次输入的密码不一致',
      type: 'error'
    },
  ],
  agreement: [
    {
      validator: (value: boolean) => value === true,
      message: '请同意用户协议和隐私政策',
      type: 'error'
    },
  ],
};

// 密码强度检测
const checkPasswordStrength = (password: string) => {
  let strength = 0;

  // 长度检测
  if (password.length >= 8) strength++;
  if (password.length >= 12) strength++;

  // 复杂度检测
  if (/[a-z]/.test(password)) strength++;
  if (/[A-Z]/.test(password)) strength++;
  if (/\d/.test(password)) strength++;
  if (/[@$!%*?&]/.test(password)) strength++;

  if (strength <= 2) {
    passwordStrength.value = { level: 'weak', width: '33%', text: '弱' };
  } else if (strength <= 4) {
    passwordStrength.value = { level: 'medium', width: '66%', text: '中' };
  } else {
    passwordStrength.value = { level: 'strong', width: '100%', text: '强' };
  }
};

const handleSubmit = async ({ validateResult, firstError }: any) => {
  if (validateResult === true) {
    loading.value = true;
    try {
      await register(formData);
      MessagePlugin.success('注册成功');
      emit('registerSuccess');
    } catch (error: any) {
      MessagePlugin.error(error.message || '注册失败');
    } finally {
      loading.value = false;
    }
  } else {
    MessagePlugin.warning(firstError);
  }
};
</script>
```

## 样式规范

### 登录页面样式
```less
.login-wrapper {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);

  .login-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: var(--td-comp-paddingTB-xxl);
  }

  .title-container {
    text-align: center;
    margin-bottom: var(--td-comp-margin-xxxl);

    .title {
      font-size: 32px;
      font-weight: 600;
      color: #fff;
      margin: 0;
    }

    .sub-title {
      margin-top: var(--td-comp-margin-m);
      color: rgba(255, 255, 255, 0.8);

      .tip {
        cursor: pointer;
        transition: color 0.3s;

        &:hover {
          color: #fff;
        }
      }
    }
  }

  .login-form,
  .register-form {
    width: 100%;
    max-width: 400px;
    background: rgba(255, 255, 255, 0.95);
    border-radius: 8px;
    padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);
    backdrop-filter: blur(10px);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  }

  .options-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin: var(--td-comp-margin-m) 0;
  }

  .third-party-login {
    margin-top: var(--td-comp-margin-xl);
    text-align: center;

    .third-party-icons {
      display: flex;
      justify-content: center;
      gap: var(--td-comp-margin-m);
      margin-top: var(--td-comp-margin-m);
    }
  }

  .password-strength {
    margin-top: var(--td-comp-margin-s);

    .strength-bar {
      height: 4px;
      background: #f0f0f0;
      border-radius: 2px;
      overflow: hidden;

      .strength-fill {
        height: 100%;
        transition: all 0.3s;

        &.weak { background: #ff4d4f; }
        &.medium { background: #faad14; }
        &.strong { background: #52c41a; }
      }
    }

    .strength-text {
      font-size: 12px;
      margin-left: var(--td-comp-margin-s);
    }
  }

  .verify-code-img {
    height: 32px;
    cursor: pointer;
    border-radius: 4px;
  }
}

.copyright {
  text-align: center;
  padding: var(--td-comp-padding-m);
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;
}

// 响应式设计
@media (max-width: 768px) {
  .login-wrapper {
    .login-container {
      padding: var(--td-comp-padding-xl);
    }

    .title-container .title {
      font-size: 24px;
    }

    .login-form,
    .register-form {
      max-width: 100%;
      padding: var(--td-comp-padding-xl);
    }
  }
}
```

## 安全考虑

### 1. 密码安全
```typescript
// 密码加密传输
import { encrypt } from '@/utils/crypto';

const handleSubmit = async () => {
  const encryptedPassword = encrypt(formData.password);

  await login({
    ...formData,
    password: encryptedPassword,
  });
};

// 密码强度验证
const validatePasswordStrength = (password: string) => {
  const checks = {
    length: password.length >= 8,
    lowercase: /[a-z]/.test(password),
    uppercase: /[A-Z]/.test(password),
    number: /\d/.test(password),
    special: /[@$!%*?&]/.test(password),
  };

  return Object.values(checks).filter(Boolean).length;
};
```

### 2. 验证码安全
```typescript
// 验证码刷新
const refreshVerifyCode = () => {
  const timestamp = Date.now();
  verifyCodeUrl.value = `/api/verify-code?t=${timestamp}`;
};

// 验证码校验
const validateVerifyCode = async (code: string) => {
  try {
    await checkVerifyCode(code);
    return true;
  } catch {
    return false;
  }
};
```

### 3. 防暴力破解
```typescript
// 登录失败次数限制
let loginAttempts = 0;
const MAX_ATTEMPTS = 5;
const LOCK_TIME = 15 * 60 * 1000; // 15分钟

const handleLoginFailure = () => {
  loginAttempts++;

  if (loginAttempts >= MAX_ATTEMPTS) {
    // 锁定账户
    const lockUntil = Date.now() + LOCK_TIME;
    localStorage.setItem('accountLocked', lockUntil.toString());

    MessagePlugin.error('登录失败次数过多，账户已锁定15分钟');
  }
};

const isAccountLocked = () => {
  const lockUntil = localStorage.getItem('accountLocked');
  return lockUntil && Date.now() < parseInt(lockUntil);
};
```

## 最佳实践

### ✅ 推荐做法
- 实现完整的表单验证
- 提供密码强度指示
- 支持记住登录状态
- 添加安全验证码
- 实现第三方登录集成
- 响应式设计适配移动端

### ❌ 避免做法
- 明文传输密码
- 缺少登录失败处理
- 忽略安全性验证
- 缺少用户协议确认

## 测试建议

### 单元测试
```typescript
import { mount } from '@vue/test-utils';
import LoginForm from '../components/Login.vue';

describe('LoginForm', () => {
  it('should validate form correctly', async () => {
    const wrapper = mount(LoginForm);

    // 测试空表单提交
    await wrapper.find('[data-testid="submit-btn"]').trigger('click');
    expect(wrapper.find('.t-form-item--error').exists()).toBe(true);
  });

  it('should handle login success', async () => {
    const wrapper = mount(LoginForm);

    await wrapper.find('[data-testid="username-input"]').setValue('testuser');
    await wrapper.find('[data-testid="password-input"]').setValue('password123');
    await wrapper.find('[data-testid="submit-btn"]').trigger('click');

    // 验证登录逻辑
  });
});
```

### 安全测试
- 测试密码强度验证
- 验证验证码功能
- 检查防暴力破解机制
- 测试第三方登录安全性