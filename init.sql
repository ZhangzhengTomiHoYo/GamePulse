-- 1. 创建数据库
CREATE DATABASE IF NOT EXISTS bluebell;

-- 2. 切换到 mysql 系统库
USE mysql;

-- 3. 修正认证方式和密码
-- 注意：不需要 CREATE USER，因为 Docker 容器启动时已经自动创建了 root 用户
ALTER USER 'root'@'%' IDENTIFIED WITH mysql_native_password BY '123456';

-- 4. 刷新权限
FLUSH PRIVILEGES;