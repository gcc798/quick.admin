import { request } from '@/utils/request';
import type { CaptchaData, CaptchaType } from '@/types/api';

export const captchaApi = {
  // 获取启用的验证码类型
  getEnabledTypes: () => request.get<CaptchaType[]>('/captcha/enabled-types'),

  // 生成图形验证码
  generateImage: () => request.get<CaptchaData>('/captcha/image'),

  // 发送短信验证码
  sendSms: (phone: string) => request.post<CaptchaData>('/captcha/sms', { phone }),

  // 发送邮箱验证码
  sendEmail: (email: string) => request.post<CaptchaData>('/captcha/email', { email }),
};
