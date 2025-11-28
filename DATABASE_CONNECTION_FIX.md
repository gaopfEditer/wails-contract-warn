# MySQL 连接权限问题解决方案

## 错误信息
```
Error 1130: Host '113.90.157.137' is not allowed to connect to this MySQL server
```

## 问题原因
MySQL服务器拒绝了来自IP地址 `113.90.157.137` 的连接请求，因为该IP地址没有被授权访问。

## 解决方案

### 方案1：在MySQL服务器上添加IP权限（推荐）

需要在MySQL服务器（8.155.10.218）上执行以下SQL命令：

#### 1.1 允许特定IP连接
```sql
-- 连接到MySQL服务器（在服务器上执行）
mysql -u root -p

-- 允许特定IP连接
GRANT ALL PRIVILEGES ON `wails-contract-warn`.* TO 'wails-contract-warn'@'113.90.157.137' IDENTIFIED BY 'Jp5eMyTTfc6ffcY8';
FLUSH PRIVILEGES;
```

#### 1.2 允许所有IP连接（如果客户端IP会变化）
```sql
-- 允许所有IP连接（使用通配符 %）
GRANT ALL PRIVILEGES ON `wails-contract-warn`.* TO 'wails-contract-warn'@'%' IDENTIFIED BY 'Jp5eMyTTfc6ffcY8';
FLUSH PRIVILEGES;
```

#### 1.3 允许IP段连接（如果IP在某个范围内）
```sql
-- 允许 113.90.157.% 网段连接
GRANT ALL PRIVILEGES ON `wails-contract-warn`.* TO 'wails-contract-warn'@'113.90.157.%' IDENTIFIED BY 'Jp5eMyTTfc6ffcY8';
FLUSH PRIVILEGES;
```

### 方案2：检查MySQL配置

#### 2.1 检查 bind-address 配置
确保MySQL服务器的 `my.cnf` 或 `my.ini` 配置文件中，`bind-address` 不是只绑定到 `127.0.0.1`：

```ini
# 允许所有IP连接
bind-address = 0.0.0.0

# 或者注释掉 bind-address
# bind-address = 127.0.0.1
```

修改后需要重启MySQL服务。

#### 2.2 检查防火墙
确保MySQL服务器的防火墙允许3306端口的连接：
- Linux: `sudo ufw allow 3306/tcp` 或 `sudo firewall-cmd --add-port=3306/tcp --permanent`
- Windows: 在防火墙设置中允许3306端口

### 方案3：查看当前权限

在MySQL服务器上执行以下命令查看当前用户权限：

```sql
-- 查看用户列表
SELECT User, Host FROM mysql.user WHERE User = 'wails-contract-warn';

-- 查看用户权限详情
SHOW GRANTS FOR 'wails-contract-warn'@'%';
SHOW GRANTS FOR 'wails-contract-warn'@'113.90.157.137';
```

### 方案4：删除旧权限并重新创建

如果用户已存在但权限不正确，可以先删除再重新创建：

```sql
-- 删除旧权限
DROP USER 'wails-contract-warn'@'%';
DROP USER 'wails-contract-warn'@'113.90.157.137';

-- 重新创建用户并授权
CREATE USER 'wails-contract-warn'@'%' IDENTIFIED BY 'Jp5eMyTTfc6ffcY8';
GRANT ALL PRIVILEGES ON `wails-contract-warn`.* TO 'wails-contract-warn'@'%';
FLUSH PRIVILEGES;
```

## 快速修复命令（推荐）

在MySQL服务器上执行以下命令（允许所有IP连接）：

```sql
-- 如果用户不存在，创建用户
CREATE USER IF NOT EXISTS 'wails-contract-warn'@'%' IDENTIFIED BY 'Jp5eMyTTfc6ffcY8';

-- 授权
GRANT ALL PRIVILEGES ON `wails-contract-warn`.* TO 'wails-contract-warn'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- 验证
SHOW GRANTS FOR 'wails-contract-warn'@'%';
```

## 注意事项

1. **安全性**：允许所有IP（`%`）连接安全性较低，建议在测试环境使用。生产环境应该只允许特定IP。
2. **密码**：确保密码正确，如果密码已更改，需要更新 `config/config.go` 中的配置。
3. **网络**：确保客户端能够访问服务器的3306端口（检查防火墙和网络路由）。

## 验证连接

修复权限后，可以在客户端测试连接：

```bash
# 使用mysql客户端测试
mysql -h 8.155.10.218 -P 3306 -u wails-contract-warn -p wails-contract-warn
# 输入密码: Jp5eMyTTfc6ffcY8
```

或者在应用启动后查看日志，应该看到：
```
数据库连接成功，表结构已创建
```




