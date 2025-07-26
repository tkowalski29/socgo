# Installation Guide

Szczegółowy przewodnik instalacji SocGo w różnych środowiskach.

## 🎯 Wybierz metodę instalacji

| Metoda | Trudność | Czas | Zalecenia |
|--------|----------|------|-----------|
| [Go Run](#go-run-development) | ⭐ | 5 min | Development, testowanie |
| [Binary Release](#binary-release) | ⭐⭐ | 10 min | Production, prosty deployment |
| [Docker](#docker-deployment) | ⭐⭐⭐ | 15 min | Izolacja, skalowalność |
| [Source Build](#build-from-source) | ⭐⭐⭐⭐ | 20 min | Customizacja, development |

## 🔧 Go Run (Development)

Najszybsza metoda do testowania i developmentu.

### Wymagania
- Go 1.24.3+
- Git

### Kroki
```bash
# 1. Klonuj repozytorium
git clone https://github.com/tkowalski/socgo.git
cd socgo

# 2. Pobierz zależności
go mod download

# 3. Skopiuj konfigurację
cp config.yml.example config.yml

# 4. Edytuj konfigurację
nano config.yml  # lub vim, code, etc.

# 5. Uruchom
go run main.go
```

### Zalety
- ✅ Szybkie zmiany kodu
- ✅ Łatwe debugging
- ✅ Automatyczne recompilation

### Wady
- ❌ Tylko do developmentu
- ❌ Wymaga Go runtime
- ❌ Brak optymalizacji

## 📦 Binary Release

Idealny do production deploymentu.

### Kroki
```bash
# 1. Pobierz release (przykład dla Linux)
wget https://github.com/tkowalski/socgo/releases/latest/download/socgo-linux-amd64.tar.gz

# 2. Rozpakuj
tar -xzf socgo-linux-amd64.tar.gz
cd socgo

# 3. Skopiuj konfigurację
cp config.yml.example config.yml

# 4. Edytuj konfigurację
nano config.yml

# 5. Uruchom
./socgo
```

### Dostępne binarie
- `socgo-linux-amd64.tar.gz` - Linux 64-bit
- `socgo-linux-arm64.tar.gz` - Linux ARM64
- `socgo-darwin-amd64.tar.gz` - macOS Intel
- `socgo-darwin-arm64.tar.gz` - macOS Apple Silicon
- `socgo-windows-amd64.zip` - Windows 64-bit

### Zalety
- ✅ Gotowy do production
- ✅ Nie wymaga Go runtime
- ✅ Optymalizowany
- ✅ Łatwy deployment

### Wady
- ❌ Brak hot reload
- ❌ Wymagana rekompilacja dla zmian

## 🐳 Docker Deployment

Pełna izolacja i łatwa skalowalność.

### Docker Run
```bash
# 1. Utwórz katalog dla konfiguracji
mkdir socgo-docker
cd socgo-docker

# 2. Pobierz przykładową konfigurację
curl -O https://raw.githubusercontent.com/tkowalski/socgo/main/config.yml.example
mv config.yml.example config.yml

# 3. Edytuj konfigurację
nano config.yml

# 4. Uruchom kontener
docker run -d \
  --name socgo \
  -p 8080:8080 \
  -v $(pwd)/config.yml:/app/config.yml \
  -v $(pwd)/uploads:/app/uploads \
  -v $(pwd)/database.db:/app/database.db \
  tkowalski/socgo:latest
```

### Docker Compose
Utwórz `docker-compose.yml`:

```yaml
version: '3.8'

services:
  socgo:
    image: tkowalski/socgo:latest
    container_name: socgo
    ports:
      - "8080:8080"
    volumes:
      - ./config.yml:/app/config.yml:ro
      - ./uploads:/app/uploads
      - ./database.db:/app/database.db
    environment:
      - SOCGO_ENV=production
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Opcjonalnie: Nginx reverse proxy
  nginx:
    image: nginx:alpine
    container_name: socgo-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - socgo
    restart: unless-stopped
```

Uruchom przez:
```bash
docker-compose up -d
```

### Build własnego image
```bash
# 1. Klonuj repozytorium
git clone https://github.com/tkowalski/socgo.git
cd socgo

# 2. Build image
docker build -t my-socgo .

# 3. Uruchom
docker run -p 8080:8080 my-socgo
```

### Zalety
- ✅ Pełna izolacja
- ✅ Łatwa skalowalność
- ✅ Consistent environment
- ✅ Easy backup/restore

### Wady
- ❌ Wymaga Docker knowledge
- ❌ Dodatkowy overhead
- ❌ Debugging może być trudniejszy

## 🔨 Build from Source

Maksymalna kontrola i customizacja.

### Wymagania
- Go 1.24.3+
- Git
- Make (opcjonalnie)

### Kroki
```bash
# 1. Klonuj z full history
git clone --depth=1 https://github.com/tkowalski/socgo.git
cd socgo

# 2. Install dependencies
go mod download
go mod verify

# 3. Run tests (opcjonalnie)
go test ./...

# 4. Build for current platform
go build -o socgo main.go

# Lub build dla różnych platform
GOOS=linux GOARCH=amd64 go build -o socgo-linux-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -o socgo-darwin-amd64 main.go
GOOS=windows GOARCH=amd64 go build -o socgo-windows-amd64.exe main.go

# 5. Lub użyj Makefile
make build        # Local build
make build-all    # All platforms
make install      # Install to $GOPATH/bin
```

### Makefile targets
```bash
make help         # Show all available targets
make build        # Build for current platform
make build-all    # Build for all platforms  
make test         # Run tests
make clean        # Clean build artifacts
make install      # Install to system
make docker       # Build Docker image
make release      # Create release packages
```

### Build flags
```bash
# Optimized production build
go build -ldflags="-w -s" -o socgo main.go

# With version info
go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD)" -o socgo main.go

# Debug build with race detection
go build -race -o socgo-debug main.go
```

### Zalety
- ✅ Pełna kontrola
- ✅ Custom build flags
- ✅ Latest features
- ✅ Development workflow

### Wady
- ❌ Wymaga Go knowledge
- ❌ Więcej kroków
- ❌ Potential instability

## ⚙️ Post-Installation Setup

### 1. Konfiguracja systemd (Linux)

Utwórz `/etc/systemd/system/socgo.service`:

```ini
[Unit]
Description=SocGo Social Media Manager
After=network.target

[Service]
Type=simple
User=socgo
Group=socgo
WorkingDirectory=/opt/socgo
ExecStart=/opt/socgo/socgo
Restart=always
RestartSec=5
Environment=SOCGO_ENV=production

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/socgo

[Install]
WantedBy=multi-user.target
```

Aktywuj:
```bash
sudo systemctl daemon-reload
sudo systemctl enable socgo
sudo systemctl start socgo
sudo systemctl status socgo
```

### 2. Nginx Reverse Proxy

Utwórz `/etc/nginx/sites-available/socgo`:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /uploads/ {
        alias /opt/socgo/uploads/;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }
}
```

### 3. SSL z Let's Encrypt
```bash
sudo certbot --nginx -d your-domain.com
```

### 4. Firewall
```bash
# UFW (Ubuntu)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# iptables
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

## 📊 Verification

### Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "5m30s"
}
```

### Performance Test
```bash
# Install hey if not available
go install github.com/rakyll/hey@latest

# Basic load test
hey -n 1000 -c 10 http://localhost:8080/
```

### Database Check
```bash
# SQLite
sqlite3 database.db "SELECT COUNT(*) FROM posts;"

# GORM logs (in development)
# Check application logs for SQL queries
```

## 🔄 Updates

### Binary Update
```bash
# 1. Backup
cp database.db database.db.backup

# 2. Download new version
wget https://github.com/tkowalski/socgo/releases/latest/download/socgo-linux-amd64.tar.gz

# 3. Replace binary
tar -xzf socgo-linux-amd64.tar.gz
sudo systemctl stop socgo
cp socgo /opt/socgo/
sudo systemctl start socgo
```

### Docker Update
```bash
docker-compose pull
docker-compose up -d
```

### Source Update
```bash
git pull origin main
go build -o socgo main.go
sudo systemctl restart socgo
```

## 🐛 Troubleshooting

### Common Issues

#### Port already in use
```bash
# Find process using port 8080
sudo lsof -i :8080

# Kill process
sudo kill -9 <PID>

# Or change port in config.yml
```

#### Permission denied
```bash
# Fix file permissions
chmod +x socgo
chown -R socgo:socgo /opt/socgo
```

#### Database locked
```bash
# Check for hanging processes
ps aux | grep socgo

# Kill all instances
sudo killall socgo

# Start fresh
rm database.db-wal database.db-shm  # SQLite WAL files
```

#### OAuth redirect mismatch
1. Check `config.yml` URLs
2. Verify OAuth app settings
3. Ensure domain/port match

## 📚 Next Steps

Po pomyślnej instalacji:

1. **[Getting Started](./getting-started.md)** - Pierwsze uruchomienie
2. **[Configuration](./configuration/)** - Zaawansowana konfiguracja  
3. **[User Guide](./guide/)** - Przewodnik użytkownika
4. **[API Documentation](./api/)** - REST API reference