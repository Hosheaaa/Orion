# 数据库迁移与上线 SOP

适用于 PostgreSQL 15/16 环境（开发、预发、生产）。生产环境建议优先使用 AWS RDS，并在执行前做好快照或备份。

## 1. 基础结构迁移

```sql
-- activities：活动信息
CREATE TABLE IF NOT EXISTS activities (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    speaker TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    input_language TEXT NOT NULL,
    target_languages JSONB NOT NULL,
    cover_url TEXT,
    status TEXT NOT NULL,
    viewer_url TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_activities_status ON activities (status);

-- activity_tokens：演讲者 / 观众令牌
CREATE TABLE IF NOT EXISTS activity_tokens (
    id UUID PRIMARY KEY,
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    max_audience INT,
    created_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL,
    UNIQUE (activity_id, type, value)
);

CREATE INDEX IF NOT EXISTS idx_activity_tokens_activity ON activity_tokens (activity_id);

-- viewer_entries：观众入口与二维码信息
CREATE TABLE IF NOT EXISTS viewer_entries (
    activity_id UUID PRIMARY KEY REFERENCES activities(id) ON DELETE CASCADE,
    share_url TEXT NOT NULL,
    qr_type TEXT NOT NULL,
    qr_content TEXT NOT NULL,
    status TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

> **提示**：应用启动时会自动执行上述 SQL，手动迁移仅用于首次部署或独立运维场景。

## 2. RDS 初始化流程

1. 登录 AWS 控制台，创建 PostgreSQL RDS 实例（推荐 `db.t3.medium`，开启 Multi-AZ 与自动备份）。
2. 创建数据库 `orion_prod`，记下终端节点。
3. 创建应用账号：
   ```sql
   CREATE ROLE orion_app WITH LOGIN PASSWORD '<强密码>';
   GRANT ALL PRIVILEGES ON DATABASE orion_prod TO orion_app;
   ```
4. 切换到 `orion_prod` 执行迁移脚本；执行完成后，限制 `orion_app` 权限（禁止创建数据库/角色）：
   ```sql
   ALTER ROLE orion_app NOSUPERUSER NOCREATEDB NOCREATEROLE;
   ```
5. 更新应用配置：
   ```
   DATABASE_URL=postgres://orion_app:<PASSWORD>@<RDS_ENDPOINT>:5432/orion_prod?sslmode=require
   DATABASE_MAX_OPEN_CONNS=30
   DATABASE_MAX_IDLE_CONNS=10
   DATABASE_CONN_MAX_LIFETIME=30m
   ```

## 3. 迁移执行 SOP

1. 在预发环境验证：部署最新镜像 → 确认迁移自动执行成功 → 跑通回归测试/联调。
2. 生成生产变更单，包含：迁移 SQL、回滚方案（删除新表或恢复快照）、执行人、时间窗口。
3. 生产执行：
   - 维护窗口内暂停写入请求（可设置维护公告或切流到只读节点）。
   - 运行迁移脚本（或启动应用触发自动迁移）。
   - 查看日志确认成功，无报错后恢复对外服务。
4. 迁移后验证：
   - 通过管理后台创建活动/令牌，确认数据写入 PostgreSQL。
   - 检查 RDS 监控指标、慢查询日志是否正常。

## 4. 回滚策略

若迁移失败或上线后出现严重问题：

1. 使用 `pg_dump` 导出变更前备份（建议迁移前执行）。
2. 若数据量小，可直接删除新建表重新部署：
   ```sql
   DROP TABLE IF EXISTS viewer_entries;
   DROP TABLE IF EXISTS activity_tokens;
   DROP TABLE IF EXISTS activities;
   ```
3. 恢复快照或将流量切回旧版本实例。
4. 修复问题后重新执行迁移。

## 5. 后续规划

- 引入专门的迁移工具（如 `golang-migrate`）和版本化脚本，支持灰度/回滚。
- 根据业务增长为 `activities` 增加更多索引（如 `start_time`、`speaker`）。
- 结合 Redis 实现令牌与字幕热点缓存，注意双写一致性。
