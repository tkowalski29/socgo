# Deployment Guide

Kompletny przewodnik wdrażania SocGo od środowiska lokalnego po production.

## 🎯 Opcje wdrażania

| Metoda | Trudność | Skalowalność | Koszt | Zalecenia |
|--------|----------|--------------|-------|-----------|
| [Binary Deployment](#binary-deployment) | ⭐ | ⭐⭐ | 💰 | Small scale, VPS |
| [Docker Deployment](#docker-deployment) | ⭐⭐ | ⭐⭐⭐ | 💰💰 | Medium scale, containers |
| [Kubernetes](#kubernetes-deployment) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 💰💰💰 | Large scale, cloud |
| [Cloud Services](#cloud-deployment) | ⭐⭐⭐ | ⭐⭐⭐⭐ | 💰💰💰 | Managed services |

## 📦 Binary Deployment

### Przygotowanie aplikacji

```bash
# 1. Build for target platform
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o socgo main.go

# 2. Create deployment package
mkdir socgo-deploy
cp socgo socgo-deploy/
cp -r web socgo-deploy/
cp config.yml.example socgo-deploy/config.yml
tar -czf socgo-linux-amd64.tar.gz socgo-deploy/
```

### Server Setup (Ubuntu/Debian)

```bash
# 1. Create user and directories
sudo useradd --system --home /opt/socgo --shell /bin/false socgo
sudo mkdir -p /opt/socgo/{bin,config,data,logs}
sudo chown -R socgo:socgo /opt/socgo

# 2. Upload and extract application
scp socgo-linux-amd64.tar.gz user@server:/tmp/
ssh user@server
sudo tar -xzf /tmp/socgo-linux-amd64.tar.gz -C /opt/socgo --strip-components=1
sudo chown -R socgo:socgo /opt/socgo

# 3. Configure application
sudo -u socgo nano /opt/socgo/config.yml
```

### Systemd Service

```ini
# /etc/systemd/system/socgo.service
[Unit]
Description=SocGo Social Media Manager
Documentation=https://github.com/tkowalski/socgo
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=socgo
Group=socgo
ExecStart=/opt/socgo/socgo
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
Restart=always
RestartSec=5

# Working directory
WorkingDirectory=/opt/socgo

# Environment
Environment=SOCGO_ENV=production
Environment=SOCGO_CONFIG_PATH=/opt/socgo/config.yml

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/socgo/data /opt/socgo/logs

# Resource limits
LimitNOFILE=1048576
LimitNPROC=1048576

[Install]
WantedBy=multi-user.target
```

### Activation i Management

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable socgo
sudo systemctl start socgo

# Check status
sudo systemctl status socgo
sudo journalctl -u socgo -f

# Management commands
sudo systemctl stop socgo
sudo systemctl restart socgo
sudo systemctl reload socgo
```

## 🐳 Docker Deployment

### Production Dockerfile

```dockerfile
# Multi-stage build for production
FROM golang:1.24.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o socgo main.go

# Production stage
FROM scratch

# Copy certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary and assets
COPY --from=builder /app/socgo /
COPY --from=builder /app/web ./web

# Create non-root user
USER 65534:65534

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD ["/socgo", "-health-check"]

# Expose port
EXPOSE 8080

# Run application
ENTRYPOINT ["/socgo"]
```

### Docker Compose Production

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  socgo:
    build:
      context: .
      dockerfile: Dockerfile
      target: production
    container_name: socgo-app
    restart: unless-stopped
    
    ports:
      - "127.0.0.1:8080:8080"
    
    volumes:
      - ./config/config.yml:/app/config.yml:ro
      - socgo-data:/app/data
      - socgo-uploads:/app/uploads
      - socgo-logs:/app/logs
    
    environment:
      - SOCGO_ENV=production
      - SOCGO_LOG_LEVEL=info
    
    healthcheck:
      test: ["CMD", "/socgo", "-health-check"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.socgo.rule=Host(`socgo.yourdomain.com`)"
      - "traefik.http.routers.socgo.tls=true"
      - "traefik.http.routers.socgo.tls.certresolver=letsencrypt"

  # Reverse Proxy
  traefik:
    image: traefik:v3.0
    container_name: socgo-traefik
    restart: unless-stopped
    
    ports:
      - "80:80"
      - "443:443"
    
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik:/etc/traefik
      - traefik-data:/data
    
    command:
      - --api.dashboard=true
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --certificatesresolvers.letsencrypt.acme.tlschallenge=true
      - --certificatesresolvers.letsencrypt.acme.email=admin@yourdomain.com
      - --certificatesresolvers.letsencrypt.acme.storage=/data/acme.json
    
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.dashboard.rule=Host(`traefik.yourdomain.com`)"
      - "traefik.http.routers.dashboard.tls=true"

  # Database (optional external database)
  postgres:
    image: postgres:16-alpine
    container_name: socgo-postgres
    restart: unless-stopped
    
    environment:
      POSTGRES_DB: socgo
      POSTGRES_USER: socgo
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./postgres/init.sql:/docker-entrypoint-initdb.d/init.sql
    
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U socgo"]
      interval: 30s
      timeout: 10s
      retries: 5

  # Redis (optional caching)
  redis:
    image: redis:7-alpine
    container_name: socgo-redis
    restart: unless-stopped
    
    volumes:
      - redis-data:/data
    
    command: redis-server --appendonly yes
    
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  socgo-data:
  socgo-uploads:
  socgo-logs:
  traefik-data:
  postgres-data:
  redis-data:

networks:
  default:
    name: socgo-network
```

### Docker Deployment Commands

```bash
# Production deployment
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f socgo

# Update application
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d

# Backup data
docker run --rm \
  -v socgo_socgo-data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/socgo-data-$(date +%Y%m%d).tar.gz /data

# Scale application (with load balancer)
docker-compose -f docker-compose.prod.yml up -d --scale socgo=3
```

## ☸️ Kubernetes Deployment

### Namespace and ConfigMap

```yaml
# k8s/namespace.yml
apiVersion: v1
kind: Namespace
metadata:
  name: socgo
  labels:
    name: socgo

---
# k8s/configmap.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: socgo-config
  namespace: socgo
data:
  config.yml: |
    server:
      port: 8080
      host: "0.0.0.0"
    
    database:
      type: postgres
      host: postgres-service
      port: 5432
      database: socgo
      username: socgo
      # password from secret
    
    oauth:
      facebook:
        client_id: "${FACEBOOK_CLIENT_ID}"
        client_secret: "${FACEBOOK_CLIENT_SECRET}"
        redirect_url: "https://socgo.yourdomain.com/oauth/facebook/callback"
```

### Secrets

```yaml
# k8s/secrets.yml
apiVersion: v1
kind: Secret
metadata:
  name: socgo-secrets
  namespace: socgo
type: Opaque
stringData:
  database-password: "your-secure-password"
  facebook-client-secret: "your-facebook-secret"
  instagram-client-secret: "your-instagram-secret"
  tiktok-client-secret: "your-tiktok-secret"
```

### Deployment

```yaml
# k8s/deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: socgo
  namespace: socgo
  labels:
    app: socgo
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: socgo
  template:
    metadata:
      labels:
        app: socgo
    spec:
      containers:
      - name: socgo
        image: socgo:latest
        imagePullPolicy: Always
        
        ports:
        - containerPort: 8080
          name: http
        
        env:
        - name: SOCGO_ENV
          value: "production"
        - name: SOCGO_CONFIG_PATH
          value: "/etc/socgo/config.yml"
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: socgo-secrets
              key: database-password
        
        volumeMounts:
        - name: config
          mountPath: /etc/socgo
          readOnly: true
        - name: uploads
          mountPath: /app/uploads
        
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
        
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      
      volumes:
      - name: config
        configMap:
          name: socgo-config
      - name: uploads
        persistentVolumeClaim:
          claimName: socgo-uploads-pvc
      
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        fsGroup: 65534
```

### Service and Ingress

```yaml
# k8s/service.yml
apiVersion: v1
kind: Service
metadata:
  name: socgo-service
  namespace: socgo
spec:
  selector:
    app: socgo
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  type: ClusterIP

---
# k8s/ingress.yml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: socgo-ingress
  namespace: socgo
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - socgo.yourdomain.com
    secretName: socgo-tls
  rules:
  - host: socgo.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: socgo-service
            port:
              number: 80
```

### Persistent Storage

```yaml
# k8s/pvc.yml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: socgo-uploads-pvc
  namespace: socgo
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: nfs-client
```

### Deployment Commands

```bash
# Apply all configurations
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n socgo
kubectl get services -n socgo
kubectl get ingress -n socgo

# View logs
kubectl logs -f deployment/socgo -n socgo

# Scale deployment
kubectl scale deployment socgo --replicas=5 -n socgo

# Rolling update
kubectl set image deployment/socgo socgo=socgo:v2.0.0 -n socgo

# Check rollout status
kubectl rollout status deployment/socgo -n socgo

# Rollback if needed
kubectl rollout undo deployment/socgo -n socgo
```

## ☁️ Cloud Deployment

### AWS ECS Deployment

```json
{
  "family": "socgo-task",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::account:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "socgo",
      "image": "your-account.dkr.ecr.region.amazonaws.com/socgo:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "SOCGO_ENV",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:socgo/database:password::"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/socgo",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

### Google Cloud Run

```yaml
# cloud-run.yml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: socgo
  namespace: default
  annotations:
    run.googleapis.com/ingress: all
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "10"
        run.googleapis.com/cpu-throttling: "false"
        run.googleapis.com/execution-environment: gen2
    spec:
      containerConcurrency: 100
      timeoutSeconds: 300
      containers:
      - name: socgo
        image: gcr.io/your-project/socgo:latest
        ports:
        - containerPort: 8080
        env:
        - name: SOCGO_ENV
          value: production
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: socgo-secrets
              key: database-url
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
          requests:
            cpu: "0.5"
            memory: "256Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 10
          failureThreshold: 3
```

### Deploy Commands

```bash
# AWS ECS
aws ecs register-task-definition --cli-input-json file://task-definition.json
aws ecs update-service --cluster socgo-cluster --service socgo-service --task-definition socgo-task

# Google Cloud Run
gcloud run deploy socgo \
  --image gcr.io/your-project/socgo:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated

# Azure Container Instances
az container create \
  --resource-group socgo-rg \
  --name socgo \
  --image your-registry.azurecr.io/socgo:latest \
  --cpu 1 \
  --memory 1 \
  --ports 8080 \
  --environment-variables SOCGO_ENV=production
```

## 🔧 Configuration Management

### Environment-specific Configs

```yaml
# config/production.yml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  type: postgres
  host: "${DATABASE_HOST}"
  port: 5432
  database: "${DATABASE_NAME}"
  username: "${DATABASE_USERNAME}"
  password: "${DATABASE_PASSWORD}"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s

oauth:
  facebook:
    client_id: "${FACEBOOK_CLIENT_ID}"
    client_secret: "${FACEBOOK_CLIENT_SECRET}"
    redirect_url: "https://${DOMAIN}/oauth/facebook/callback"

logging:
  level: info
  format: json
  output: stdout

monitoring:
  metrics:
    enabled: true
    port: 9090
  tracing:
    enabled: true
    endpoint: "${JAEGER_ENDPOINT}"

security:
  tls:
    enabled: true
    cert_file: "/etc/certs/tls.crt"
    key_file: "/etc/certs/tls.key"
```

### Environment Variables

```bash
# .env.production
SOCGO_ENV=production
SOCGO_LOG_LEVEL=info

# Database
DATABASE_HOST=postgres.example.com
DATABASE_NAME=socgo
DATABASE_USERNAME=socgo
DATABASE_PASSWORD=secure-password

# OAuth secrets
FACEBOOK_CLIENT_ID=your-facebook-app-id
FACEBOOK_CLIENT_SECRET=your-facebook-secret
INSTAGRAM_CLIENT_ID=your-instagram-app-id
INSTAGRAM_CLIENT_SECRET=your-instagram-secret
TIKTOK_CLIENT_ID=your-tiktok-client-key
TIKTOK_CLIENT_SECRET=your-tiktok-client-secret

# Domain
DOMAIN=socgo.yourdomain.com

# Monitoring
JAEGER_ENDPOINT=http://jaeger-collector:14268/api/traces
```

## 📊 Monitoring i Logging

### Prometheus Metrics

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'socgo'
    static_configs:
      - targets: ['socgo:9090']
    metrics_path: /metrics
    scrape_interval: 30s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "SocGo Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

### ELK Stack Logging

```yaml
# logging/logstash.conf
input {
  beats {
    port => 5044
  }
}

filter {
  if [fields][service] == "socgo" {
    json {
      source => "message"
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "socgo-logs-%{+YYYY.MM.dd}"
  }
}
```

## 🔄 CI/CD Pipeline

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test ./...
      
      - name: Build binary
        run: |
          CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o socgo main.go
      
      - name: Build Docker image
        run: |
          docker build -t socgo:${{ github.ref_name }} .
          docker tag socgo:${{ github.ref_name }} socgo:latest
      
      - name: Push to registry
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker push socgo:${{ github.ref_name }}
          docker push socgo:latest

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: Deploy to production
        run: |
          ssh ${{ secrets.DEPLOY_USER }}@${{ secrets.DEPLOY_HOST }} '
            cd /opt/socgo &&
            docker-compose pull &&
            docker-compose up -d --remove-orphans
          '
```

### GitLab CI/CD

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  image: golang:1.21
  script:
    - go test ./...
  coverage: '/coverage: \d+\.\d+% of statements/'

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG
  only:
    - tags

deploy:production:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache openssh-client
    - eval $(ssh-agent -s)
    - echo "$SSH_PRIVATE_KEY" | tr -d '\r' | ssh-add -
  script:
    - ssh -o StrictHostKeyChecking=no $DEPLOY_USER@$DEPLOY_HOST "
        cd /opt/socgo &&
        docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG &&
        docker-compose up -d
      "
  environment:
    name: production
    url: https://socgo.yourdomain.com
  only:
    - tags
```

## 🔒 Security Considerations

### TLS/SSL Configuration

```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    server_name socgo.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/socgo.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/socgo.yourdomain.com/privkey.pem;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    # Rate limiting
    limit_req zone=api burst=20 nodelay;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Firewall Rules

```bash
# UFW configuration
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# Fail2ban for SSH protection
sudo apt install fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

## 🔄 Backup i Recovery

### Database Backup

```bash
#!/bin/bash
# backup-database.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups"
DATABASE_URL="postgres://user:pass@localhost/socgo"

# Create backup
pg_dump $DATABASE_URL | gzip > $BACKUP_DIR/socgo_backup_$DATE.sql.gz

# Keep only last 30 days
find $BACKUP_DIR -name "socgo_backup_*.sql.gz" -mtime +30 -delete

# Upload to S3 (optional)
aws s3 cp $BACKUP_DIR/socgo_backup_$DATE.sql.gz s3://socgo-backups/
```

### Application Backup

```bash
#!/bin/bash
# backup-app.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups"
APP_DIR="/opt/socgo"

# Create application backup
tar -czf $BACKUP_DIR/socgo_app_$DATE.tar.gz \
  -C $APP_DIR \
  --exclude='logs/*' \
  --exclude='tmp/*' \
  .

# Database backup
docker exec postgres pg_dump -U socgo socgo | gzip > $BACKUP_DIR/socgo_db_$DATE.sql.gz

# Upload to cloud storage
rclone copy $BACKUP_DIR/ remote:socgo-backups/
```

To bardzo kompleksowy przewodnik deployment pokrywający wszystkie główne scenariusze wdrażania SocGo! 🚀