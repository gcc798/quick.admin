# Go 语言错误处理 vs Java Spring Boot 对比

## 核心理念对比

### Java Spring Boot
```java
@RestControllerAdvice
public class GlobalExceptionHandler {
    @ExceptionHandler(BusinessException.class)
    public ResponseEntity<Result> handleBusiness(BusinessException e) {
        return ResponseEntity.ok(Result.fail(e.getCode(), e.getMessage()));
    }
    
    @ExceptionHandler(Exception.class)
    public ResponseEntity<Result> handleSystem(Exception e) {
        log.error("系统异常", e);
        return ResponseEntity.status(500).body(Result.fail(500, "系统异常"));
    }
}
```

### Go 语言（本项目）
```go
// 在 main.go 或 router 中注册中间件
r.Use(middleware.Recovery(logger))

// 自动捕获所有 panic，无需在每个函数中手动处理
```

---

## 错误处理方式对比

### ❌ 错误方式（不要这样做）

每个函数都写 defer recover，这是**反模式**：

```go
// ❌ 错误示例：不要这样做！
func (s *userService) Create(ctx context.Context, req *request.CreateUserRequest) error {
    defer func() {
        if r := recover(); r != nil {
            // 每个函数都写这段代码太繁琐了！
            logger.Error("panic", zap.Any("error", r))
        }
    }()
    
    // 业务代码
    return nil
}

func (s *userService) Update(ctx context.Context, req *request.UpdateUserRequest) error {
    defer func() {
        if r := recover(); r != nil {
            // 又要复制粘贴一遍！
            logger.Error("panic", zap.Any("error", r))
        }
    }()
    
    // 业务代码
    return nil
}
```

### ✅ 正确方式（推荐）

**全局中间件自动捕获**，开发者专注业务逻辑：

```go
// ✅ 正确示例：Service 层直接写业务逻辑
func (s *userService) Create(ctx context.Context, req *request.CreateUserRequest) error {
    // 检查用户名是否存在
    if exists {
        return apperrors.NewBusiness(apperrors.CodeUserNameExists, "用户名已存在")
    }
    
    // 数据库操作
    if err := s.db.Create(user).Error; err != nil {
        return apperrors.NewInfrastructure(apperrors.CodeDatabaseError, "数据库写入失败", err)
    }
    
    // 可能发生 panic 的代码（无需手动处理）
    arr := []int{1, 2, 3}
    _ = arr[someIndex]  // 如果越界，中间件会自动捕获
    
    return nil
}

// ✅ Controller 层统一错误处理
func (c *userController) Create(ctx *gin.Context) {
    err := c.userService.Create(ctx.Request.Context(), &req)
    if err != nil {
        response.Error(ctx, err)  // 自动识别错误类型并处理
        return
    }
    response.Success(ctx, data)
}
```

---

## 中间件如何工作

### 全局 Panic 恢复中间件

```go
func Recovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 自动记录详细堆栈
                logger.Error("系统 Panic 捕获",
                    zap.Any("error", err),
                    zap.String("path", c.Request.URL.Path),
                    zap.String("stack", string(debug.Stack())))
                
                // 自动返回统一文案
                c.JSON(500, Response{
                    Code: 30001,
                    Msg:  "系统异常，请联系管理员",
                })
                c.Abort()
            }
        }()
        
        c.Next()  // 执行后续的所有 Handler
    }
}
```

### 工作原理

1. **中间件在最外层包裹了 defer recover**
2. **整个请求链路上的任何 panic 都会被捕获**
3. **包括 Controller、Service、Repository 等所有层级**

```
HTTP Request
    ↓
[Recovery 中间件] ← defer func() { recover() }
    ↓
[Auth 中间件]
    ↓
[Controller]
    ↓
[Service] ← 这里发生 panic
    ↓
[Repository]
```

---

## 什么时候需要手动 defer recover？

### 只有 2 种特殊情况

#### 1. Goroutine 内部（中间件无法捕获）

```go
func (s *someService) AsyncTask(ctx context.Context) {
    go func() {
        defer func() {
            if r := recover() {
                // goroutine 内部的 panic 必须手动处理
                logger.Error("goroutine panic", zap.Any("error", r))
            }
        }()
        
        // 异步任务代码
        arr := []int{1, 2, 3}
        _ = arr[10]  // 这里的 panic 中间件捕获不到
    }()
}
```

#### 2. 需要优雅降级的场景

```go
func (s *cacheService) GetFromCache(key string) (value string, err error) {
    defer func() {
        if r := recover() {
            // 缓存读取失败不影响主流程，降级处理
            logger.Warn("cache panic", zap.Any("error", r))
            err = nil  // 不返回错误，继续后续流程
        }
    }()
    
    // 可能 panic 的缓存操作
    value = cache[key]
    return value, nil
}
```

---

## 最佳实践总结

### ✅ 应该做的

1. **在 main.go 注册全局 Recovery 中间件**
2. **Service 层直接抛出业务错误或基础设施错误**
3. **Controller 层统一使用 response.Error(c, err)**
4. **让中间件自动捕获所有 panic**

### ❌ 不应该做的

1. ❌ 在每个 Service 函数中写 defer recover
2. ❌ 在 Controller 层判断错误类型
3. ❌ 手动处理 panic（除非是 goroutine 或降级场景）

---

## 与 Java 的对比

| 特性 | Java Spring Boot | Go (本项目) |
|------|-----------------|-------------|
| 全局异常处理 | @ControllerAdvice | middleware.Recovery |
| 业务异常 | throw new BusinessException() | return apperrors.NewBusiness() |
| 系统异常 | 自动捕获 | 自动捕获（通过中间件） |
| 手动捕获 | try-catch | defer recover（仅 goroutine） |
| 开发体验 | 不需要 try-catch | 不需要 defer recover |

**结论：Go 的错误处理同样优雅，不需要在每个函数中写 defer recover！**
