# 双Token认证机制说明

## 概述

系统已升级为**双Token认证机制**（AccessToken + RefreshToken），提供更高的安全性和更好的用户体验。

## Token类型

### 1. AccessToken（短期访问令牌）

- **用途**：用于日常API请求的身份验证
- **格式**：JWT（自包含，无需查询数据库）
- **过期时间**：由 `SysClient.ActiveTimeout` 控制
  - Web管理后台：30分钟
  - iOS/Android移动端：1小时
  - 微信小程序：2小时
- **存储位置**：客户端本地存储
- **传输方式**：HTTP Header `Authorization: Bearer {accessToken}`

### 2. RefreshToken（长期刷新令牌）

- **用途**：用于刷新AccessToken，延长登录会话
- **格式**：随机字符串（Base64编码）
- **过期时间**：由 `SysClient.Timeout` 控制
  - Web管理后台：7天
  - iOS/Android移动端：30天
  - 微信小程序：90天
- **存储位置**：
  - 服务端：Redis（`refresh_token:{userId}:{clientId}`）
  - 客户端：安全存储（如Keychain、SharedPreferences加密存储）
- **安全特性**：
  - 每次刷新后轮换（生成新的RefreshToken，旧的失效）
  - 绑定到特定设备（clientId）
  - 支持并发登录控制

## 字段语义重新定义

| 字段 | 原含义 | 新含义 |
|------|--------|--------|
| `timeout` | Token固定过期时间 | **RefreshToken过期时间**（长期） |
| `active_timeout` | 活动超时时间 | **AccessToken过期时间**（短期） |

## API接口

### 1. 登录接口

**请求**：
```http
POST /login
Content-Type: application/json

{
  "clientKey": "web-admin",
  "clientSecret": "web-secret-2024",
  "grantType": "password",
  "username": "admin",
  "password": "admin123"
}
```

**响应**：
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4=",
    "expires_in": 1800,
    "refresh_expires_in": 604800,
    "user_info": {
      "userId": 1,
      "username": "admin",
      "nickname": "系统管理员",
      "email": "admin@example.com",
      "avatar": "",
      "userType": "sys_user"
    }
  }
}
```

### 2. 刷新Token接口

**请求**：
```http
POST /auth/refresh
Content-Type: application/json

{
  "refreshToken": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4=",
  "clientKey": "web-admin",
  "clientSecret": "web-secret-2024"
}
```

**响应**：
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "bmV3IHJlZnJlc2ggdG9rZW4=",
    "expires_in": 1800,
    "refresh_expires_in": 604800
  }
}
```

### 3. 登出接口

**请求**：
```http
POST /logout
Authorization: Bearer {accessToken}
```

**响应**：
```json
{
  "code": 200,
  "msg": "success",
  "data": "ok"
}
```

## 客户端集成指南

### 1. 登录流程

```javascript
// 1. 用户登录
const loginResponse = await fetch('/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    clientKey: 'web-admin',
    clientSecret: 'web-secret-2024',
    grantType: 'password',
    username: 'admin',
    password: 'admin123'
  })
});

const { access_token, refresh_token, expires_in } = loginResponse.data;

// 2. 存储Token
localStorage.setItem('accessToken', access_token);
localStorage.setItem('refreshToken', refresh_token);
localStorage.setItem('tokenExpireTime', Date.now() + expires_in * 1000);
```

### 2. API请求拦截器

```javascript
// 请求拦截器
axios.interceptors.request.use(async (config) => {
  const accessToken = localStorage.getItem('accessToken');
  const expireTime = localStorage.getItem('tokenExpireTime');
  
  // 检查Token是否即将过期（提前5分钟刷新）
  if (Date.now() > expireTime - 5 * 60 * 1000) {
    await refreshAccessToken();
  }
  
  config.headers.Authorization = `Bearer ${accessToken}`;
  return config;
});

// 响应拦截器
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Token无效，尝试刷新
      const success = await refreshAccessToken();
      if (success) {
        // 重试原请求
        return axios.request(error.config);
      } else {
        // 刷新失败，跳转登录页
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);
```

### 3. 刷新Token函数

```javascript
async function refreshAccessToken() {
  try {
    const refreshToken = localStorage.getItem('refreshToken');
    const response = await fetch('/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        refreshToken,
        clientKey: 'web-admin',
        clientSecret: 'web-secret-2024'
      })
    });
    
    const { access_token, refresh_token, expires_in } = response.data;
    
    // 更新存储
    localStorage.setItem('accessToken', access_token);
    localStorage.setItem('refreshToken', refresh_token);
    localStorage.setItem('tokenExpireTime', Date.now() + expires_in * 1000);
    
    return true;
  } catch (error) {
    console.error('刷新Token失败', error);
    return false;
  }
}
```

### 4. 登出流程

```javascript
async function logout() {
  const accessToken = localStorage.getItem('accessToken');
  
  await fetch('/logout', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  
  // 清除本地存储
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
  localStorage.removeItem('tokenExpireTime');
  
  // 跳转登录页
  window.location.href = '/login';
}
```

## 安全特性

### 1. RefreshToken轮换

每次刷新AccessToken时，系统会生成新的RefreshToken并使旧的失效，防止RefreshToken被盗用。

### 2. 设备绑定

RefreshToken绑定到特定的客户端（clientId），不同设备的RefreshToken相互独立。

### 3. 并发登录控制

通过配置 `auth.allowConcurrent` 和 `auth.shareToken` 控制：

- `allowConcurrent=false`：不允许并发登录，新登录会使旧Token失效
- `allowConcurrent=true, shareToken=true`：允许并发，共享Token
- `allowConcurrent=true, shareToken=false`：允许并发，每个设备独立Token

### 4. 短期AccessToken

AccessToken有效期短（30分钟-2小时），即使泄露影响也较小。

## Redis存储结构

### RefreshToken存储

```
Key: refresh_token:{userId}:{clientId}
Type: Hash
Fields:
  - token: RefreshToken字符串
  - userId: 用户ID
  - userName: 用户名
  - clientId: 客户端ID
  - deviceType: 设备类型
  - createdAt: 创建时间戳
TTL: timeout（7-90天）
```

### RefreshToken索引

```
Key: refresh_token_index:{tokenHash}
Type: String
Value: refresh_token:{userId}:{clientId}
TTL: timeout（7-90天）
```

用于快速查找RefreshToken对应的用户信息。

## 配置说明

### 客户端配置

| 客户端 | AccessToken过期 | RefreshToken过期 | 授权类型 |
|--------|----------------|-----------------|---------|
| Web管理后台 | 30分钟 | 7天 | password,email,refresh |
| iOS移动端 | 1小时 | 30天 | password,sms,email,refresh |
| Android移动端 | 1小时 | 30天 | password,sms,email,refresh |
| 微信小程序 | 2小时 | 90天 | xcx,refresh |

### 认证配置（conf.dev.yaml）

```yaml
auth:
  tokenHeader: "Authorization"  # Token请求头名称
  allowConcurrent: false         # 是否允许并发登录
  shareToken: false              # 并发登录时是否共享Token
```

## 迁移指南

### 从旧版本升级

1. **数据库迁移**：
   ```sql
   -- 更新客户端配置，添加 refresh 授权类型
   UPDATE sys_client 
   SET grant_type = grant_type || ',refresh' 
   WHERE grant_type NOT LIKE '%refresh%';
   ```

2. **客户端代码更新**：
   - 更新登录响应处理，保存 `refresh_token`
   - 添加Token刷新逻辑
   - 更新API请求拦截器

3. **测试验证**：
   - 测试登录流程
   - 测试Token刷新
   - 测试Token过期处理
   - 测试并发登录控制

## 常见问题

### Q1: AccessToken过期后会发生什么？

A: 客户端应该在AccessToken过期前主动刷新（建议提前5分钟）。如果AccessToken已过期，API请求会返回401错误，客户端应使用RefreshToken刷新后重试。

### Q2: RefreshToken过期后会发生什么？

A: RefreshToken过期后，用户需要重新登录。建议在RefreshToken即将过期时提醒用户。

### Q3: 如何处理多设备登录？

A: 通过配置 `allowConcurrent` 和 `shareToken` 控制：
- 不允许多设备：`allowConcurrent=false`
- 允许多设备，共享Token：`allowConcurrent=true, shareToken=true`
- 允许多设备，独立Token：`allowConcurrent=true, shareToken=false`

### Q4: RefreshToken是否需要加密存储？

A: 是的，RefreshToken应该安全存储：
- Web：使用HttpOnly Cookie或加密的localStorage
- 移动端：使用Keychain（iOS）或Keystore（Android）
- 小程序：使用加密的storage

### Q5: 如何强制用户重新登录？

A: 调用 `/logout` 接口会使RefreshToken失效，用户需要重新登录。

## 性能优化

1. **AccessToken验证**：JWT自包含，无需查询数据库
2. **RefreshToken索引**：使用Token哈希作为索引，快速查找
3. **Redis Pipeline**：批量操作减少网络往返
4. **Token复用**：共享Token模式下减少Token生成次数

## 监控建议

1. **Token刷新频率**：监控刷新接口调用频率
2. **Token过期率**：统计AccessToken和RefreshToken过期情况
3. **登录失败率**：监控认证失败原因
4. **并发登录**：统计多设备登录情况

## 总结

双Token机制提供了安全性和用户体验的最佳平衡：

✅ **安全性**：短期AccessToken降低泄露风险  
✅ **用户体验**：长期RefreshToken减少登录次数  
✅ **灵活性**：支持多种并发登录策略  
✅ **可扩展性**：易于集成第三方认证
