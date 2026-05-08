#!/bin/bash

# Update system
apt-get update
apt-get upgrade -y

# Install nginx
apt-get install -y nginx

# Install Node.js for modern web applications
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
apt-get install -y nodejs

# Install PM2 for process management
npm install -g pm2

# Configure nginx
cat > /etc/nginx/sites-available/default << EOF
server {
    listen 80 default_server;
    listen [::]:80 default_server;

    root /var/www/${project_name};
    index index.html index.htm index.nginx-debian.html;

    server_name _;

    location / {
        try_files \$uri \$uri/ @backend;
    }

    location @backend {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF

# Create web directory
mkdir -p /var/www/${project_name}

# Create a simple index page
cat > /var/www/${project_name}/index.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>${project_name} - Web Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; }
        .status { background: #e8f5e8; padding: 10px; border-radius: 4px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Welcome to ${project_name}</h1>
        <div class="status">
            <strong>Status:</strong> Web server is running successfully!
        </div>
        <p>This is a demo web application deployed using Terraform and Airtel Cloud.</p>
        <h2>Server Information</h2>
        <ul>
            <li><strong>Project:</strong> ${project_name}</li>
            <li><strong>Server:</strong> Web Server</li>
            <li><strong>Technology:</strong> Nginx + Node.js</li>
            <li><strong>Cloud Provider:</strong> Airtel Cloud</li>
        </ul>
    </div>
</body>
</html>
EOF

# Set permissions
chown -R www-data:www-data /var/www/${project_name}

# Enable and start nginx
systemctl enable nginx
systemctl start nginx

# Configure firewall
ufw allow 'Nginx Full'
ufw allow ssh
ufw --force enable

# Create log directory
mkdir -p /var/log/${project_name}

# Log the completion
echo "$(date): Web server setup completed for ${project_name}" >> /var/log/${project_name}/setup.log