-- 修复MySQL用户权限，允许远程IP连接
-- 需要在MySQL服务器上执行这些命令

-- 方案1：允许特定IP连接（推荐）
-- 将 '113.90.157.137' 替换为实际的客户端IP地址
GRANT ALL PRIVILEGES ON `wails-contract-warn`.* TO 'wails-contract-warn'@'113.90.157.137' IDENTIFIED BY 'Jp5eMyTTfc6ffcY8';
FLUSH PRIVILEGES;

-- 方案2：允许所有IP连接（安全性较低，但方便）
-- 如果客户端IP会变化，可以使用 '%' 允许所有IP
GRANT ALL PRIVILEGES ON `wails-contract-warn`.* TO 'wails-contract-warn'@'%' IDENTIFIED BY 'Jp5eMyTTfc6ffcY8';
FLUSH PRIVILEGES;

-- 方案3：允许IP段连接（如果IP在某个范围内）
-- 例如允许 113.90.157.% 网段
GRANT ALL PRIVILEGES ON `wails-contract-warn`.* TO 'wails-contract-warn'@'113.90.157.%' IDENTIFIED BY 'Jp5eMyTTfc6ffcY8';
FLUSH PRIVILEGES;

-- 查看当前用户权限
SELECT User, Host FROM mysql.user WHERE User = 'wails-contract-warn';

-- 查看用户权限详情
SHOW GRANTS FOR 'wails-contract-warn'@'%';




