# PostgreSQL 慢 SQL 监控 - 快速开始

## 5 分钟快速配置

### 步骤 1：配置 PostgreSQL

编辑 `postgresql.conf`（通常在 `/usr/local/var/postgresql@14/` 或 `/etc/postgresql/14/main/`）：

```ini
# 添加以下配置
shared_preload_libraries = 'pg_stat_statements'
pg_stat_statements.max = 10000
pg_stat_statements.track = all
log_min_duration_statement = 100
```

### 步骤 2：重启 PostgreSQL

```bash
# macOS
brew services restart postgresql@14

# Linux
sudo systemctl restart postgresql
```

### 步骤 3：启用监控扩展

```bash
psql -U postgres -d nai_tizi -f scripts/sql/enable_slow_query_monitoring.sql
```

### 步骤 4：配置应用

编辑 `conf.dev.yaml`：

```yaml
database:
  dsn: "host=127.0.0.1 user=postgres password=post123 dbname=nai-tizi port=5433 sslmode=disable TimeZone=Asia/Shanghai"
  slowThreshold: 100  # 慢 SQL 阈值：100 毫秒
```

### 步骤 5：启动应用

```bash
go run cmd/api/main.go
```

查看日志，应该看到：

```
INFO    slow query plugin initialized   {"threshold": "100ms"}
```

### 步骤 6：测试监控

运行测试脚本：

```bash
./scripts/test_slow_query.sh
```

或手动测试：

```bash
# 查看最慢的 SQL
psql -U postgres -d nai_tizi -c "SELECT * FROM v_slow_queries LIMIT 5;"

# 查看应用日志
tail -f logs/app.log | grep "slow query detected"
```

## 验证监控是否生效

### 1. 检查扩展

```sql
SELECT * FROM pg_available_extensions WHERE name = 'pg_stat_statements';
```

应该看到 `installed_version` 列有值。

### 2. 检查视图

```sql
\dv v_slow_queries
```

应该看到视图已创建。

### 3. 模拟慢查询

```sql
SELECT pg_sleep(0.2), COUNT(*) FROM s_users;
```

然后查看应用日志：

```bash
tail -f logs/app.log | grep "slow query"
```

应该看到类似输出：

```json
{
  "level": "warn",
  "ts": "2024-12-28T22:00:00.000+0800",
  "msg": "slow query detected",
  "sql": "SELECT COUNT(*) FROM s_users",
  "duration": "205.123ms",
  "rows_affected": 1,
  "table": "s_users"
}
```

### 4. 查看统计信息

```sql
-- 最慢的 5 条 SQL
SELECT 
    LEFT(short_query, 80) AS query,
    calls,
    ROUND(mean_exec_time::numeric, 2) AS avg_ms,
    ROUND(max_exec_time::numeric, 2) AS max_ms
FROM v_slow_queries
LIMIT 5;
```

## 常见问题

### Q1: 扩展启用失败

**错误信息：**
```
ERROR:  could not load library "pg_stat_statements": ...
```

**解决方案：**
1. 确认 `shared_preload_libraries` 配置正确
2. 重启 PostgreSQL 服务
3. 检查 PostgreSQL 版本是否支持该扩展

### Q2: 应用日志没有慢 SQL 记录

**可能原因：**
1. 慢 SQL 阈值设置过高
2. 没有执行超过阈值的 SQL
3. 日志级别配置不正确

**解决方案：**
```yaml
# 降低阈值测试
database:
  slowThreshold: 10  # 10ms
```

### Q3: 视图查询为空

**可能原因：**
1. 还没有执行过 SQL
2. 所有 SQL 都很快（< 100ms）
3. 统计信息被重置

**解决方案：**
```sql
-- 执行一些查询
SELECT * FROM s_users LIMIT 100;

-- 再次查看
SELECT * FROM v_slow_queries;
```

## 下一步

1. 阅读完整文档：[PostgreSQL慢SQL监控方案.md](./PostgreSQL慢SQL监控方案.md)
2. 配置告警规则
3. 定期分析和优化慢 SQL
4. 建立性能基线

## 相关文档

- [PostgreSQL慢SQL监控方案.md](./PostgreSQL慢SQL监控方案.md) - 完整方案
- [生产环境部署清单.md](./生产环境部署清单.md) - 部署检查
- [健康检查和优雅关闭说明.md](./健康检查和优雅关闭说明.md) - 监控配置

---

**文档版本：** 1.0  
**创建时间：** 2024-12-28  
**维护人员：** Kiro AI Assistant
