#!/bin/bash

# Update system
apt-get update
apt-get upgrade -y

# Install MySQL
apt-get install -y mysql-server

# Install additional tools
apt-get install -y htop iotop nethogs

# Secure MySQL installation (automated)
mysql -e "UPDATE mysql.user SET Password = PASSWORD('${project_name}_secure_password') WHERE User = 'root'"
mysql -e "DELETE FROM mysql.user WHERE User=''"
mysql -e "DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1')"
mysql -e "DROP DATABASE IF EXISTS test"
mysql -e "DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%'"
mysql -e "FLUSH PRIVILEGES"

# Create application database
mysql -u root -p${project_name}_secure_password -e "CREATE DATABASE ${project_name}_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p${project_name}_secure_password -e "CREATE USER '${project_name}_user'@'%' IDENTIFIED BY '${project_name}_db_password';"
mysql -u root -p${project_name}_secure_password -e "GRANT ALL PRIVILEGES ON ${project_name}_db.* TO '${project_name}_user'@'%';"
mysql -u root -p${project_name}_secure_password -e "FLUSH PRIVILEGES;"

# Configure MySQL for remote connections
sed -i 's/bind-address\s*=\s*127.0.0.1/bind-address = 0.0.0.0/' /etc/mysql/mysql.conf.d/mysqld.cnf

# Optimize MySQL configuration for the instance
cat >> /etc/mysql/mysql.conf.d/mysqld.cnf << EOF

# ${project_name} specific optimizations
innodb_buffer_pool_size = 512M
innodb_log_file_size = 64M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT
max_connections = 100
query_cache_size = 64M
query_cache_type = 1
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
EOF

# Create backup script
cat > /usr/local/bin/${project_name}_backup.sh << EOF
#!/bin/bash
BACKUP_DIR="/backup/mysql"
DATE=\$(date +%Y%m%d_%H%M%S)
mkdir -p \$BACKUP_DIR

# Create database backup
mysqldump -u root -p${project_name}_secure_password ${project_name}_db > \$BACKUP_DIR/${project_name}_db_\$DATE.sql

# Compress backup
gzip \$BACKUP_DIR/${project_name}_db_\$DATE.sql

# Keep only last 7 days of backups
find \$BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete

echo "\$(date): Database backup completed" >> /var/log/${project_name}/backup.log
EOF

chmod +x /usr/local/bin/${project_name}_backup.sh

# Create backup directory
mkdir -p /backup/mysql

# Add cron job for daily backups
echo "0 2 * * * /usr/local/bin/${project_name}_backup.sh" | crontab -

# Configure firewall
ufw allow ssh
ufw allow 3306/tcp
ufw --force enable

# Restart MySQL
systemctl restart mysql
systemctl enable mysql

# Create log directory
mkdir -p /var/log/${project_name}

# Create monitoring script
cat > /usr/local/bin/${project_name}_monitor.sh << EOF
#!/bin/bash
LOG_FILE="/var/log/${project_name}/monitor.log"

# Check MySQL status
if systemctl is-active --quiet mysql; then
    echo "\$(date): MySQL is running" >> \$LOG_FILE
else
    echo "\$(date): MySQL is down - attempting restart" >> \$LOG_FILE
    systemctl restart mysql
fi

# Check disk space
DISK_USAGE=\$(df /var/lib/mysql | awk 'NR==2 {print \$5}' | sed 's/%//')
if [ \$DISK_USAGE -gt 80 ]; then
    echo "\$(date): WARNING - Disk usage is at \$DISK_USAGE%" >> \$LOG_FILE
fi

# Check MySQL connections
CONNECTIONS=\$(mysql -u root -p${project_name}_secure_password -e "SHOW STATUS LIKE 'Threads_connected'" | awk 'NR==2 {print \$2}')
echo "\$(date): Active MySQL connections: \$CONNECTIONS" >> \$LOG_FILE
EOF

chmod +x /usr/local/bin/${project_name}_monitor.sh

# Add monitoring cron job (every 5 minutes)
echo "*/5 * * * * /usr/local/bin/${project_name}_monitor.sh" | crontab -

# Log the completion
echo "$(date): Database server setup completed for ${project_name}" >> /var/log/${project_name}/setup.log

# Create database initialization script
cat > /tmp/init_${project_name}_db.sql << EOF
USE ${project_name}_db;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id VARCHAR(128) PRIMARY KEY,
    user_id INT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO users (username, email, password_hash) VALUES
('admin', 'admin@${project_name}.com', SHA2('admin_password', 256)),
('demo', 'demo@${project_name}.com', SHA2('demo_password', 256));
EOF

# Initialize database
mysql -u root -p${project_name}_secure_password < /tmp/init_${project_name}_db.sql

echo "$(date): Database initialized with sample data" >> /var/log/${project_name}/setup.log