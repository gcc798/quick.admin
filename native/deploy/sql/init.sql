-- 系统管理 / 权限管理菜单初始化
-- 说明：
-- 1. s_menu.perms 仅用于前端权限判断，后端接口鉴权以 s_api_permission.code + Casbin 为准。
-- 2. 本脚本可重复执行；已存在同名同父级菜单时会跳过插入。

INSERT INTO s_menu (
  id,
  menu_name,
  parent_id,
  sort,
  path,
  component,
  query,
  is_frame,
  is_cache,
  menu_type,
  visible,
  status,
  perms,
  icon,
  remark,
  create_by,
  update_by
)
SELECT
  COALESCE((SELECT MAX(id) FROM s_menu), 0) + 1,
  '权限管理',
  system_menu.id,
  4,
  'api-permission',
  'system/apiPermission/index',
  '',
  0,
  0,
  1,
  0,
  0,
  'api_permission.read',
  'api',
  'API 权限管理页面',
  1,
  1
FROM s_menu system_menu
WHERE system_menu.menu_name = '系统管理'
  AND system_menu.parent_id = 0
  AND NOT EXISTS (
    SELECT 1
    FROM s_menu existing
    WHERE existing.menu_name = '权限管理'
      AND existing.parent_id = system_menu.id
  );

INSERT INTO s_menu (
  id,
  menu_name,
  parent_id,
  sort,
  path,
  component,
  query,
  is_frame,
  is_cache,
  menu_type,
  visible,
  status,
  perms,
  icon,
  remark,
  create_by,
  update_by
)
SELECT
  COALESCE((SELECT MAX(id) FROM s_menu), 0) + ROW_NUMBER() OVER (ORDER BY seed.parent_id, seed.sort),
  seed.*
FROM (
  SELECT
    '新增权限' AS menu_name,
    permission_menu.id AS parent_id,
    1 AS sort,
    '' AS path,
    '' AS component,
    '' AS query,
    0 AS is_frame,
    0 AS is_cache,
    2 AS menu_type,
    1 AS visible,
    0 AS status,
    'api_permission.create' AS perms,
    'plus' AS icon,
    'API 权限新增按钮' AS remark,
    1 AS create_by,
    1 AS update_by
  FROM s_menu permission_menu
  INNER JOIN s_menu system_menu ON system_menu.id = permission_menu.parent_id
  WHERE system_menu.menu_name = '系统管理'
    AND system_menu.parent_id = 0
    AND permission_menu.menu_name = '权限管理'

  UNION ALL

  SELECT
    '编辑权限',
    permission_menu.id,
    2,
    '',
    '',
    '',
    0,
    0,
    2,
    1,
    0,
    'api_permission.update',
    'form',
    'API 权限编辑按钮',
    1,
    1
  FROM s_menu permission_menu
  INNER JOIN s_menu system_menu ON system_menu.id = permission_menu.parent_id
  WHERE system_menu.menu_name = '系统管理'
    AND system_menu.parent_id = 0
    AND permission_menu.menu_name = '权限管理'

  UNION ALL

  SELECT
    '删除权限',
    permission_menu.id,
    3,
    '',
    '',
    '',
    0,
    0,
    2,
    1,
    0,
    'api_permission.delete',
    'delete',
    'API 权限删除按钮',
    1,
    1
  FROM s_menu permission_menu
  INNER JOIN s_menu system_menu ON system_menu.id = permission_menu.parent_id
  WHERE system_menu.menu_name = '系统管理'
    AND system_menu.parent_id = 0
    AND permission_menu.menu_name = '权限管理'

  UNION ALL

  SELECT
    '分配API权限',
    permission_menu.id,
    4,
    '',
    '',
    '',
    0,
    0,
    2,
    1,
    0,
    'api_permission.assign',
    'safety',
    '角色和用户 API 权限分配按钮',
    1,
    1
  FROM s_menu permission_menu
  INNER JOIN s_menu system_menu ON system_menu.id = permission_menu.parent_id
  WHERE system_menu.menu_name = '系统管理'
    AND system_menu.parent_id = 0
    AND permission_menu.menu_name = '权限管理'
) seed
WHERE NOT EXISTS (
  SELECT 1
  FROM s_menu existing
  WHERE existing.parent_id = seed.parent_id
    AND existing.perms = seed.perms
);

-- API 权限初始化
-- 规则：
-- 1. 每个业务模块都有一个通配符节点，例如 user.*，角色/用户授权勾选模块节点时写入通配符即可覆盖子权限。
-- 2. 具体权限 code 必须和 native/internal/constants/permission.go 以及路由中间件使用的 code 保持一致。
-- 3. method/path 仅作为管理界面展示和维护参考，真正后端鉴权依据是 code + action。

WITH seed (
  seq,
  parent_code,
  module,
  code,
  name,
  node_type,
  action,
  method,
  path,
  sort,
  status,
  remark
) AS (
  VALUES
    (1000, '', 'user', 'user.*', '用户管理', 0, '*', '*', '/api/v1/user/*', 10, 0, '用户管理模块全部 API 权限'),
    (1010, 'user.*', 'user', 'user.read', '用户查询', 2, 'read', 'GET,POST', '/api/v1/user/page,/api/v1/user/:id', 10, 0, '用户列表和详情查询'),
    (1020, 'user.*', 'user', 'user.create', '用户新增', 2, 'write', 'POST', '/api/v1/user,/api/v1/user/import', 20, 0, '新增用户和导入用户'),
    (1030, 'user.*', 'user', 'user.update', '用户更新', 2, 'write', 'PUT', '/api/v1/user/:id,/api/v1/user/:id/password', 30, 0, '更新用户和重置密码'),
    (1040, 'user.*', 'user', 'user.delete', '用户删除', 2, 'write', 'DELETE', '/api/v1/user/:id,/api/v1/user/batch', 40, 0, '删除用户和批量删除用户'),

    (2000, '', 'role', 'role.*', '角色管理', 0, '*', '*', '/api/v1/role/*', 20, 0, '角色管理模块全部 API 权限'),
    (2010, 'role.*', 'role', 'role.read', '角色查询', 2, 'read', 'GET,POST', '/api/v1/role/page,/api/v1/role/:roleId,/api/v1/role/user,/api/v1/role/:roleId/menus', 10, 0, '角色列表、详情、用户角色和角色菜单查询'),
    (2020, 'role.*', 'role', 'role.create', '角色新增', 2, 'write', 'POST', '/api/v1/role', 20, 0, '新增角色'),
    (2030, 'role.*', 'role', 'role.update', '角色更新', 2, 'write', 'PUT,POST', '/api/v1/role/:roleId,/api/v1/role/:roleId/menus', 30, 0, '更新角色和分配角色菜单'),
    (2040, 'role.*', 'role', 'role.delete', '角色删除', 2, 'write', 'DELETE', '/api/v1/role/:roleId', 40, 0, '删除角色'),
    (2050, 'role.*', 'role', 'role.assign', '用户角色分配', 2, 'write', 'POST,DELETE', '/api/v1/role/assign,/api/v1/role/remove', 50, 0, '为用户分配或移除角色'),
    (2060, 'role.*', 'role', 'role.permission', '角色旧权限接口', 2, 'write', 'GET,POST,DELETE', '/api/v1/role/permission,/api/v1/role/permissions', 60, 0, '兼容旧角色权限接口'),

    (3000, '', 'menu', 'menu.*', '菜单管理', 0, '*', '*', '/api/v1/menu/*', 30, 0, '菜单管理模块全部 API 权限'),
    (3010, 'menu.*', 'menu', 'menu.read', '菜单查询', 2, 'read', 'GET', '/api/v1/menu/tree,/api/v1/menu,/api/v1/menu/:id', 10, 0, '菜单树、列表和详情查询'),
    (3020, 'menu.*', 'menu', 'menu.create', '菜单新增', 2, 'write', 'POST', '/api/v1/menu', 20, 0, '新增菜单'),
    (3030, 'menu.*', 'menu', 'menu.update', '菜单更新', 2, 'write', 'PUT', '/api/v1/menu/:id', 30, 0, '更新菜单'),
    (3040, 'menu.*', 'menu', 'menu.delete', '菜单删除', 2, 'write', 'DELETE', '/api/v1/menu/:id', 40, 0, '删除菜单'),

    (4000, '', 'api_permission', 'api_permission.*', 'API 权限管理', 0, '*', '*', '/api/v1/api-permission/*', 40, 0, 'API 权限管理模块全部 API 权限'),
    (4010, 'api_permission.*', 'api_permission', 'api_permission.read', 'API 权限查询', 2, 'read', 'GET', '/api/v1/api-permission/tree,/api/v1/api-permission', 10, 0, 'API 权限树和列表查询'),
    (4020, 'api_permission.*', 'api_permission', 'api_permission.create', 'API 权限新增', 2, 'write', 'POST', '/api/v1/api-permission', 20, 0, '新增 API 权限'),
    (4030, 'api_permission.*', 'api_permission', 'api_permission.update', 'API 权限更新', 2, 'write', 'PUT', '/api/v1/api-permission/:id', 30, 0, '更新 API 权限'),
    (4040, 'api_permission.*', 'api_permission', 'api_permission.delete', 'API 权限删除', 2, 'write', 'DELETE', '/api/v1/api-permission/:id', 40, 0, '删除 API 权限'),
    (4050, 'api_permission.*', 'api_permission', 'api_permission.assign', 'API 权限授权', 2, 'write', 'GET,POST', '/api/v1/role/:roleId/api-permissions,/api/v1/user/:id/api-permissions', 50, 0, '为角色或用户分配 API 权限'),

    (5000, '', 'org', 'org.*', '组织管理', 0, '*', '*', '/api/v1/org/*', 50, 0, '组织管理模块全部 API 权限'),
    (5010, 'org.*', 'org', 'org.read', '组织查询', 2, 'read', 'GET,POST', '/api/v1/org/page,/api/v1/org/tree,/api/v1/org/:id', 10, 0, '组织树、列表和详情查询'),
    (5020, 'org.*', 'org', 'org.create', '组织新增', 2, 'write', 'POST', '/api/v1/org', 20, 0, '新增组织'),
    (5030, 'org.*', 'org', 'org.update', '组织更新', 2, 'write', 'PUT', '/api/v1/org/:id', 30, 0, '更新组织'),
    (5040, 'org.*', 'org', 'org.delete', '组织删除', 2, 'write', 'DELETE', '/api/v1/org/:id,/api/v1/org/batch', 40, 0, '删除组织和批量删除组织'),

    (6000, '', 'dict', 'dict.*', '字典管理', 0, '*', '*', '/api/v1/dict/*', 60, 0, '字典管理模块全部 API 权限'),
    (6010, 'dict.*', 'dict', 'dict.read', '字典查询', 2, 'read', 'GET,POST', '/api/v1/dict/page,/api/v1/dict/type,/api/v1/dict/label,/api/v1/dict/:id', 10, 0, '字典列表、详情、类型和标签查询'),
    (6020, 'dict.*', 'dict', 'dict.create', '字典新增', 2, 'write', 'POST', '/api/v1/dict', 20, 0, '新增字典'),
    (6030, 'dict.*', 'dict', 'dict.update', '字典更新', 2, 'write', 'PUT', '/api/v1/dict/:id', 30, 0, '更新字典'),
    (6040, 'dict.*', 'dict', 'dict.delete', '字典删除', 2, 'write', 'DELETE', '/api/v1/dict/:id,/api/v1/dict/batch', 40, 0, '删除字典和批量删除字典'),

    (7000, '', 'config', 'config.*', '配置管理', 0, '*', '*', '/api/v1/config/*', 70, 0, '配置管理模块全部 API 权限'),
    (7010, 'config.*', 'config', 'config.read', '配置查询', 2, 'read', 'GET,POST', '/api/v1/config/page,/api/v1/config/code,/api/v1/config/data,/api/v1/config/:id', 10, 0, '配置列表、详情、编码和数据查询'),
    (7020, 'config.*', 'config', 'config.create', '配置新增', 2, 'write', 'POST', '/api/v1/config', 20, 0, '新增配置'),
    (7030, 'config.*', 'config', 'config.update', '配置更新', 2, 'write', 'PUT', '/api/v1/config/:id', 30, 0, '更新配置'),
    (7040, 'config.*', 'config', 'config.delete', '配置删除', 2, 'write', 'DELETE', '/api/v1/config/:id,/api/v1/config/batch', 40, 0, '删除配置和批量删除配置'),

    (8000, '', 'login_log', 'login_log.*', '登录日志', 0, '*', '*', '/api/v1/loginLog/*', 80, 0, '登录日志模块全部 API 权限'),
    (8010, 'login_log.*', 'login_log', 'login_log.read', '登录日志查询', 2, 'read', 'GET,POST', '/api/v1/loginLog/page,/api/v1/loginLog/:id', 10, 0, '登录日志列表和详情查询'),
    (8020, 'login_log.*', 'login_log', 'login_log.create', '登录日志新增', 2, 'write', 'POST', '/api/v1/loginLog', 20, 0, '新增登录日志'),
    (8030, 'login_log.*', 'login_log', 'login_log.update', '登录日志更新', 2, 'write', 'PUT', '/api/v1/loginLog/:id', 30, 0, '更新登录日志'),
    (8040, 'login_log.*', 'login_log', 'login_log.delete', '登录日志删除', 2, 'write', 'POST,DELETE', '/api/v1/loginLog/clean,/api/v1/loginLog/:id,/api/v1/loginLog/batch', 40, 0, '删除、批量删除和清理登录日志'),

    (9000, '', 'oper_log', 'oper_log.*', '操作日志', 0, '*', '*', '/api/v1/operLog/*', 90, 0, '操作日志模块全部 API 权限'),
    (9010, 'oper_log.*', 'oper_log', 'oper_log.read', '操作日志查询', 2, 'read', 'GET,POST', '/api/v1/operLog/page,/api/v1/operLog/:id', 10, 0, '操作日志列表和详情查询'),
    (9020, 'oper_log.*', 'oper_log', 'oper_log.create', '操作日志新增', 2, 'write', 'POST', '/api/v1/operLog', 20, 0, '新增操作日志'),
    (9030, 'oper_log.*', 'oper_log', 'oper_log.update', '操作日志更新', 2, 'write', 'PUT', '/api/v1/operLog/:id', 30, 0, '更新操作日志'),
    (9040, 'oper_log.*', 'oper_log', 'oper_log.delete', '操作日志删除', 2, 'write', 'POST,DELETE', '/api/v1/operLog/clean,/api/v1/operLog/:id,/api/v1/operLog/batch', 40, 0, '删除、批量删除和清理操作日志'),

    (10000, '', 'storage_env', 'storage_env.*', '存储环境', 0, '*', '*', '/api/v1/storage-env/*', 100, 0, '存储环境模块全部 API 权限'),
    (10010, 'storage_env.*', 'storage_env', 'storage_env.read', '存储环境查询', 2, 'read', 'GET,POST', '/api/v1/storage-env/page,/api/v1/storage-env/default,/api/v1/storage-env/:id,/api/v1/storage-env/:id/test', 10, 0, '存储环境列表、详情、默认环境和连接测试'),
    (10020, 'storage_env.*', 'storage_env', 'storage_env.create', '存储环境新增', 2, 'write', 'POST', '/api/v1/storage-env', 20, 0, '新增存储环境'),
    (10030, 'storage_env.*', 'storage_env', 'storage_env.update', '存储环境更新', 2, 'write', 'PUT', '/api/v1/storage-env/:id', 30, 0, '更新存储环境'),
    (10040, 'storage_env.*', 'storage_env', 'storage_env.delete', '存储环境删除', 2, 'write', 'DELETE', '/api/v1/storage-env/:id', 40, 0, '删除存储环境'),
    (10050, 'storage_env.*', 'storage_env', 'storage_env.manage', '存储环境管理', 2, 'write', 'POST', '/api/v1/storage-env/default', 50, 0, '设置默认存储环境'),

    (11000, '', 'attachment', 'attachment.*', '附件管理', 0, '*', '*', '/api/v1/attachment/*', 110, 0, '附件管理模块全部 API 权限'),
    (11010, 'attachment.*', 'attachment', 'attachment.read', '附件查询', 2, 'read', 'GET,POST', '/api/v1/attachment/page,/api/v1/attachment/business,/api/v1/attachment/:attachmentId,/api/v1/attachment/:attachmentId/url', 10, 0, '附件列表、详情、业务附件和访问 URL 查询'),
    (11020, 'attachment.*', 'attachment', 'attachment.create', '附件新增', 2, 'write', 'POST', '/api/v1/attachment/upload-file', 20, 0, '兼容附件新增权限，实际上传使用 attachment.upload'),
    (11030, 'attachment.*', 'attachment', 'attachment.update', '附件更新', 2, 'write', 'POST', '/api/v1/attachment/:attachmentId/bind', 30, 0, '兼容附件更新权限，实际绑定使用 attachment.bind'),
    (11040, 'attachment.*', 'attachment', 'attachment.delete', '附件删除', 2, 'write', 'DELETE', '/api/v1/attachment/:attachmentId', 40, 0, '删除附件'),
    (11050, 'attachment.*', 'attachment', 'attachment.upload', '附件上传', 2, 'write', 'POST', '/api/v1/attachment/upload-file', 50, 0, '上传附件文件'),
    (11060, 'attachment.*', 'attachment', 'attachment.download', '附件下载', 2, 'write', 'GET', '/api/v1/attachment/:attachmentId/download', 60, 0, '下载附件文件'),
    (11070, 'attachment.*', 'attachment', 'attachment.bind', '附件绑定', 2, 'write', 'POST', '/api/v1/attachment/:attachmentId/bind', 70, 0, '绑定附件到业务对象')
),
numbered AS (
  SELECT
    seed.*,
    COALESCE((SELECT MAX(id) FROM s_api_permission), 0) + ROW_NUMBER() OVER (ORDER BY seed.seq) AS new_id
  FROM seed
),
resolved AS (
  SELECT
    numbered.*,
    COALESCE(existing_parent.id, numbered_parent.new_id, 0) AS resolved_parent_id
  FROM numbered
  LEFT JOIN s_api_permission existing_parent ON existing_parent.code = numbered.parent_code
  LEFT JOIN numbered numbered_parent ON numbered_parent.code = numbered.parent_code
)
INSERT INTO s_api_permission (
  id,
  parent_id,
  module,
  code,
  name,
  node_type,
  action,
  method,
  path,
  sort,
  status,
  remark,
  create_by,
  update_by,
  created_time,
  updated_time
)
SELECT
  resolved.new_id,
  resolved.resolved_parent_id,
  resolved.module,
  resolved.code,
  resolved.name,
  resolved.node_type,
  resolved.action,
  resolved.method,
  resolved.path,
  resolved.sort,
  resolved.status,
  resolved.remark,
  1,
  1,
  NOW(),
  NOW()
FROM resolved
ON CONFLICT (code) DO UPDATE SET
  parent_id = EXCLUDED.parent_id,
  module = EXCLUDED.module,
  name = EXCLUDED.name,
  node_type = EXCLUDED.node_type,
  action = EXCLUDED.action,
  method = EXCLUDED.method,
  path = EXCLUDED.path,
  sort = EXCLUDED.sort,
  status = EXCLUDED.status,
  remark = EXCLUDED.remark,
  update_by = EXCLUDED.update_by,
  updated_time = NOW();
