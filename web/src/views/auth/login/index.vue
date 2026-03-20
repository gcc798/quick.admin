<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>{{ title }}</h1>
        <p>欢迎登录管理后台</p>
      </div>

      <a-form
        :model="formData"
        :rules="rules"
        @finish="handleLogin"
        layout="vertical"
        class="login-form"
      >
        <a-form-item name="username" label="用户名">
          <a-input
            v-model:value="formData.username"
            size="large"
            placeholder="请输入用户名"
            allow-clear
          >
            <template #prefix>
              <user-outlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item name="password" label="密码">
          <a-input-password
            v-model:value="formData.password"
            size="large"
            placeholder="请输入密码"
            allow-clear
          >
            <template #prefix>
              <lock-outlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item v-if="showCaptcha" name="code" label="验证码">
          <div class="captcha-wrapper">
            <a-input
              v-model:value="formData.code"
              size="large"
              placeholder="请输入验证码"
              allow-clear
              style="flex: 1"
            />
            <div class="captcha-image" @click="loadCaptcha">
              <img v-if="captchaImage" :src="captchaImage" alt="验证码" />
              <reload-outlined class="reload-icon" />
            </div>
          </div>
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            :loading="loading"
            block
          >
            登录
          </a-button>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { message } from 'ant-design-vue';
import { UserOutlined, LockOutlined, ReloadOutlined } from '@ant-design/icons-vue';
import { useAuthStore } from '@/stores/auth';
import type { LoginParams } from '@/types/api';
import { captchaApi } from '@/api/captcha';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const title = import.meta.env.VITE_APP_TITLE;
const loading = ref(false);
const showCaptcha = ref(false);
const captchaImage = ref('');
const uuid = ref('');

const formData = reactive({
  username: '',
  password: '',
  code: '',
});

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  code: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
};

const loadCaptcha = async () => {
  try {
    const data = await captchaApi.generateImage();
    uuid.value = data.id;
    captchaImage.value = data.data.image;
  } catch (error: any) {
    message.error('加载验证码失败');
  }
};

const handleLogin = async () => {
  try {
    loading.value = true;

    const params: LoginParams = {
      grantType: 'password',
      username: formData.username,
      password: formData.password,
      clientKey: import.meta.env.VITE_CLIENT_KEY,
      clientSecret: import.meta.env.VITE_CLIENT_SECRET,
    };

    if (showCaptcha.value) {
      params.uuid = uuid.value;
      params.code = formData.code;
    }

    await authStore.login(params);

    message.success('登录成功');

    const redirect = (route.query.redirect as string) || '/';
    router.push(redirect);
  } catch (error: any) {
    message.error(error.message || '登录失败');
    if (showCaptcha.value) {
      await loadCaptcha();
      formData.code = '';
    }
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  try {
    const types = await captchaApi.getEnabledTypes();
    if (types.includes('image')) {
      showCaptcha.value = true;
      await loadCaptcha();
    }
  } catch (error) {
    console.error('Failed to load captcha types:', error);
  }
});
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.login-header p {
  font-size: 14px;
  color: #666;
}

.login-form {
  margin-top: 24px;
}

.captcha-wrapper {
  display: flex;
  gap: 12px;
  align-items: center;
}

.captcha-image {
  position: relative;
  width: 120px;
  height: 40px;
  cursor: pointer;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  overflow: hidden;
  flex-shrink: 0;
}

.captcha-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.captcha-image .reload-icon {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
  color: rgba(0, 0, 0, 0.3);
  opacity: 0;
  transition: opacity 0.2s;
}

.captcha-image:hover .reload-icon {
  opacity: 1;
}
</style>
