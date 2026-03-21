import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { App, Button, Card, Form, Input, Space, Typography } from 'antd';
import { LockOutlined, ReloadOutlined, UserOutlined } from '@ant-design/icons';
import { authApi } from '@/api/auth';
import { useAuthStore } from '@/store/auth';

interface LoginFormValues {
  username: string;
  password: string;
  code?: string;
}

export default function LoginPage() {
  const [form] = Form.useForm<LoginFormValues>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { message } = App.useApp();
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn);
  const login = useAuthStore((state) => state.login);
  const [loading, setLoading] = useState(false);
  const [showCaptcha, setShowCaptcha] = useState(false);
  const [captchaId, setCaptchaId] = useState('');
  const [captchaImage, setCaptchaImage] = useState('');

  const loadCaptcha = async () => {
    try {
      const data = await authApi.getImageCaptcha();
      setCaptchaId(data.id);
      setCaptchaImage(String(data.data.image ?? ''));
    } catch {
      message.error('加载验证码失败');
    }
  };

  useEffect(() => {
    if (isLoggedIn) {
      navigate('/', { replace: true });
    }
  }, [isLoggedIn, navigate]);

  useEffect(() => {
    void (async () => {
      try {
        const types = await authApi.getCaptchaTypes();
        const enabled = types.includes('image');
        setShowCaptcha(enabled);

        if (enabled) {
          await loadCaptcha();
        }
      } catch {
        setShowCaptcha(false);
      }
    })();
  }, []);

  const handleFinish = async (values: LoginFormValues) => {
    setLoading(true);
    try {
      // 登录参数严格对齐 sys-api/auth.api，不沿用旧前端里可能出现的残留兼容字段。
      await login({
        clientKey: import.meta.env.VITE_CLIENT_KEY,
        clientSecret: import.meta.env.VITE_CLIENT_SECRET,
        grantType: 'password',
        username: values.username,
        password: values.password,
        code: values.code,
        uuid: captchaId || undefined,
      });

      message.success('登录成功');
      // 登录后优先回到用户原本想访问的页面，和旧后台行为保持一致。
      const redirect = searchParams.get('redirect') || '/';
      navigate(redirect, { replace: true });
    } catch (error) {
      message.error(error instanceof Error ? error.message : '登录失败');
      if (showCaptcha) {
        form.setFieldValue('code', '');
        await loadCaptcha();
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <Card className="login-card" variant="borderless">
        <Space direction="vertical" size={8} style={{ width: '100%', marginBottom: 24 }}>
          <Typography.Title level={2} style={{ marginBottom: 0 }}>
            {import.meta.env.VITE_APP_TITLE}
          </Typography.Title>
          <Typography.Text type="secondary">
            React 重写版后台。保持功能等价，同时保证代码结构清晰、便于学习。
          </Typography.Text>
        </Space>

        <Form<LoginFormValues>
          form={form}
          layout="vertical"
          onFinish={(values) => void handleFinish(values)}
        >
          <Form.Item
            label="用户名"
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              allowClear
              prefix={<UserOutlined />}
              placeholder="请输入用户名"
              size="large"
            />
          </Form.Item>

          <Form.Item
            label="密码"
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              allowClear
              prefix={<LockOutlined />}
              placeholder="请输入密码"
              size="large"
            />
          </Form.Item>

          {showCaptcha ? (
            <Form.Item
              label="验证码"
              name="code"
              rules={[{ required: true, message: '请输入验证码' }]}
            >
              <div className="login-captcha">
                <Input placeholder="请输入验证码" size="large" />
                <div
                  className="login-captcha-image"
                  onClick={() => void loadCaptcha()}
                >
                  {captchaImage ? (
                    <img alt="验证码" src={captchaImage} style={{ width: '100%' }} />
                  ) : (
                    <ReloadOutlined />
                  )}
                </div>
              </div>
            </Form.Item>
          ) : null}

          <Form.Item style={{ marginBottom: 0 }}>
            <Button block htmlType="submit" loading={loading} size="large" type="primary">
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
