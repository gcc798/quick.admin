-- Development seed data for the upgraded scaffold.
-- Safe to rerun in the nai-tizi development database.

BEGIN;

-- Drop obsolete business tables from the previous scaffold copy. Current scaffold
-- keeps only generic attachment and MQTT retry business tables.
DROP TABLE IF EXISTS biz_alarm_log CASCADE;
DROP TABLE IF EXISTS biz_device_info CASCADE;
DROP TABLE IF EXISTS biz_device_last_will CASCADE;
DROP TABLE IF EXISTS biz_device_online CASCADE;
DROP TABLE IF EXISTS biz_device_share_record CASCADE;
DROP TABLE IF EXISTS biz_device_state_change_record CASCADE;
DROP TABLE IF EXISTS biz_mqtt_server_info CASCADE;
DROP TABLE IF EXISTS biz_sn CASCADE;
DROP TABLE IF EXISTS biz_task CASCADE;

-- Default auth clients. The web frontend uses e10adc3949ba59abbe56e057f20f883e.
INSERT INTO s_auth_client (
  client_id,
  client_key,
  client_secret,
  grant_type,
  device_type,
  status,
  timeout,
  active_timeout,
  remark,
  create_by,
  update_by,
  created_time,
  updated_time
) VALUES
  ('e10adc3949ba59abbe56e057f20f883e', 'web-admin', '123456', 'password,email', 'web', 0, 604800, 1800, '后台管理 Web 客户端', 1, 1, NOW(), NOW()),
  ('5f4dcc3b5aa765d61d8327deb882cf99', 'mobile-ios', 'password', 'password,email,xcx', 'ios', 0, 604800, 1800, 'iOS 客户端', 1, 1, NOW(), NOW()),
  ('098f6bcd4621d373cade4e832627b4f6', 'mobile-android', 'test', 'password,email,xcx', 'android', 0, 604800, 1800, 'Android 客户端', 1, 1, NOW(), NOW()),
  ('5d41402abc4b2a76b9719d911017c592', 'wechat-xcx', 'hello', 'xcx', 'wechat', 0, 604800, 1800, '微信小程序客户端', 1, 1, NOW(), NOW())
ON CONFLICT (client_id) DO UPDATE SET
  client_key = EXCLUDED.client_key,
  client_secret = EXCLUDED.client_secret,
  grant_type = EXCLUDED.grant_type,
  device_type = EXCLUDED.device_type,
  status = EXCLUDED.status,
  timeout = EXCLUDED.timeout,
  active_timeout = EXCLUDED.active_timeout,
  remark = EXCLUDED.remark,
  update_by = EXCLUDED.update_by,
  updated_time = NOW();

-- Default roles.
INSERT INTO s_role (
  id,
  role_key,
  role_name,
  sort,
  status,
  data_scope,
  is_system,
  remark,
  create_by,
  update_by,
  created_time,
  updated_time
) VALUES
  (1880159541355577349, 'super_admin', '超级管理员', 0, 0, 1, TRUE, '系统内置超级管理员', 1, 1, NOW(), NOW()),
  (1880159541355577350, 'admin', '管理员', 1, 0, 1, TRUE, '系统内置管理员', 1, 1, NOW(), NOW()),
  (1880159541355577351, 'user', '普通用户', 2, 0, 5, FALSE, '系统内置普通用户', 1, 1, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  role_key = EXCLUDED.role_key,
  role_name = EXCLUDED.role_name,
  sort = EXCLUDED.sort,
  status = EXCLUDED.status,
  data_scope = EXCLUDED.data_scope,
  is_system = EXCLUDED.is_system,
  remark = EXCLUDED.remark,
  update_by = EXCLUDED.update_by,
  updated_time = NOW();

-- Default admin account: admin / admin123.
INSERT INTO s_user (
  id,
  user_name,
  nick_name,
  user_type,
  org_id,
  email,
  phonenumber,
  sex,
  avatar,
  password,
  status,
  sort,
  remark,
  create_by,
  update_by,
  created_time,
  updated_time
) VALUES (
  1,
  'admin',
  '管理员',
  0,
  0,
  'admin@example.com',
  '',
  2,
  '',
  '$2a$10$Q55.ONb4ACprCH5Wl9NqouI9uWyvV.wGT4BSRRnCWQXdfJiWgOHzK',
  0,
  0,
  '系统内置管理员账号',
  1,
  1,
  NOW(),
  NOW()
)
ON CONFLICT (id) DO UPDATE SET
  user_name = EXCLUDED.user_name,
  nick_name = EXCLUDED.nick_name,
  user_type = EXCLUDED.user_type,
  org_id = EXCLUDED.org_id,
  email = EXCLUDED.email,
  phonenumber = EXCLUDED.phonenumber,
  sex = EXCLUDED.sex,
  avatar = EXCLUDED.avatar,
  password = EXCLUDED.password,
  status = EXCLUDED.status,
  sort = EXCLUDED.sort,
  remark = EXCLUDED.remark,
  update_by = EXCLUDED.update_by,
  updated_time = NOW();

DELETE FROM m_user_role WHERE user_id = 1;
INSERT INTO m_user_role (user_id, role_id, create_by, update_by, created_time, updated_time)
VALUES (1, 1880159541355577349, 1, 1, NOW(), NOW());

-- Casbin policy for the default admin. Keep the explicit super_admin wildcard,
-- and bind the real admin user id used by s_user.
DELETE FROM casbin_rule WHERE ptype = 'g' AND v0 LIKE 'user::%';
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'role::super_admin';
INSERT INTO casbin_rule (ptype, v0, v1, v2)
VALUES
  ('p', 'role::super_admin', '*', '*'),
  ('g', 'user::1', 'role::super_admin', '');

SELECT setval(pg_get_serial_sequence('s_user', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM s_user), 1), TRUE);
SELECT setval(pg_get_serial_sequence('s_role', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM s_role), 1), TRUE);

COMMIT;
