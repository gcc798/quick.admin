# 分布式ID生成器

## 概述

项目使用 **Sony Sonyflake** 作为分布式ID生成器，替代数据库自增主键。Sonyflake 是 Twitter Snowflake 的 Go 实现，生成64位唯一ID，适合分布式系统和分表场景。

## 特性

- **全局唯一**: 在分布式环境中保证ID唯一性
- **趋势递增**: ID按时间趋势递增，有利于数据库索引
- **高性能**: 内存生成，无需访问数据库
- **分表友好**: 支持后续分库分表和数据迁移
- **64位整数**: 兼容 PostgreSQL BIGINT 类型

## ID结构

Sonyflake 生成的64位ID结构：
```
39位时间戳 + 8位序列号 + 16位机器ID
```

- **时间戳**: 从配置的起始时间开始计算（本项目: 2024-01-01）
- **序列号**: 同一毫秒内的序列号（0-255）
- **机器ID**: 区分不同节点（0-65535）

## 配置

在 `conf.{env}.yaml` 中配置机器ID：

```yaml
database:
  dsn: "..."
  machineId: 1  # 机器ID，范围 0-65535，默认 1
```

**重要**: 在分布式部署时，每个节点必须配置不同的 `machineId`，以保证ID全局唯一。

## 使用方式

### 自动生成（推荐）

所有模型的主键会自动生成ID，无需手动设置：

```go
user := &model.User{
    UserName: "test",
    NickName: "测试用户",
    // UserId 会自动生成，无需设置
}
db.Create(user)
// user.UserId 现在包含生成的ID
```

### 手动生成

如果需要提前生成ID：

```go
import "github.com/force-c/nai-tizi/internal/utils/idgen"

// 生成ID（返回错误）
id, err := idgen.NextID()
if err != nil {
    // 处理错误
}

// 生成ID（失败时panic）
id := idgen.MustNextID()
```

## 实现原理

### 1. ID生成器初始化

在数据库初始化时，自动初始化ID生成器：

```go
// internal/container/container.go
func (c *container) initDB() error {
    // 初始化ID生成器
    if err := idgen.Init(c.config.Database.MachineID); err != nil {
        return err
    }
    // ...
}
```

### 2. GORM插件自动生成

注册了 `IDGenPlugin` 插件，在创建记录前自动生成ID：

```go
// internal/infrastructure/database/idgen_plugin.go
func (p *IDGenPlugin) beforeCreate(db *gorm.DB) {
    // 检查主键字段
    // 如果主键值为0（未设置），则生成新ID
    // 自动设置到模型字段
}
```

### 3. 模型定义

所有模型的主键字段移除了 `autoIncrement` 标签：

```go
type User struct {
    UserId int64 `gorm:"column:user_id;primaryKey" json:"userId"`
    // 不再使用 autoIncrement
}
```

## 数据库迁移

### 新表

新建表时，主键字段定义为 BIGINT，不使用 SERIAL：

```sql
CREATE TABLE s_user (
    user_id BIGINT PRIMARY KEY,  -- 不使用 SERIAL
    user_name VARCHAR(30) NOT NULL,
    ...
);
```

### 现有表

如果已有使用 SERIAL 的表，需要修改：

```sql
-- 1. 移除默认值（SERIAL 自动创建的序列）
ALTER TABLE s_user ALTER COLUMN user_id DROP DEFAULT;

-- 2. 删除序列（可选）
DROP SEQUENCE IF EXISTS s_user_user_id_seq;
```

## 性能考虑

- **生成速度**: 单机每秒可生成 256,000 个ID（256 * 1000）
- **内存占用**: 极小，仅维护序列号状态
- **无锁设计**: 使用原子操作，高并发友好

## 分布式部署

在多节点部署时：

1. **配置不同的机器ID**:
   ```yaml
   # 节点1: conf.prod.yaml
   database:
     machineId: 1
   
   # 节点2: conf.prod.yaml
   database:
     machineId: 2
   ```

2. **机器ID分配策略**:
   - 手动分配: 在配置文件中指定
   - 自动分配: 可基于IP、主机名等生成（需自行实现）

3. **ID冲突检测**:
   - 不同机器ID保证不会冲突
   - 同一机器ID的不同实例会冲突，需避免

## 注意事项

1. **时钟回拨**: Sonyflake 对时钟回拨有一定容忍度（10秒），超过会返回错误
2. **机器ID唯一性**: 分布式环境必须保证每个节点的机器ID不同
3. **ID不连续**: 生成的ID不是严格连续的，但保证趋势递增
4. **起始时间**: 配置的起始时间（2024-01-01）不应修改，否则可能产生重复ID

## 故障排查

### ID生成失败

```
错误: 生成ID失败: sonyflake初始化失败
```

**原因**: ID生成器未正确初始化

**解决**: 检查配置文件中的 `machineId` 是否正确

### 时钟回拨错误

```
错误: clock moved backwards
```

**原因**: 系统时钟被回拨超过10秒

**解决**: 
1. 等待时钟追上
2. 使用 NTP 同步时间
3. 重启应用

## 参考资料

- [Sonyflake GitHub](https://github.com/sony/sonyflake)
- [Snowflake ID 算法](https://en.wikipedia.org/wiki/Snowflake_ID)
