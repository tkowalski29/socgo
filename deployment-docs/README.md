# SocGo Deployment Guide

Ten przewodnik opisuje jak wdrożyć aplikację SocGo na serwerze produkcyjnym używając Docker.

## Wymagania

- Docker Engine 20.10+ 
- Docker Compose 2.0+
- Minimum 1GB RAM
- Minimum 5GB miejsca na dysku
- Otwarte porty: 40378

## Opcje wdrożenia

### 1. Proste wdrożenie (bez reverse proxy)

Idealne dla testowania lub małych instalacji.

```bash
# Sklonuj repozytorium lub pobierz pliki
git clone https://github.com/tkowalski/socgo.git
cd socgo

# Skopiuj przykładową konfigurację
mkdir -p config
cp config.yml.example config/config.yml

# Edytuj konfigurację
nano config/config.yml

# Uruchom aplikację
docker-compose -f docker-compose.simple.yml up -d
```

Aplikacja będzie dostępna na `http://your-server-ip:40378`

### 2. Produkcyjne wdrożenie

Dla produkcyjnych instalacji.

```bash
# Sklonuj repozytorium
git clone https://github.com/tkowalski/socgo.git
cd socgo

# Przygotuj konfigurację
mkdir -p config
cp config.yml.example config/config.yml

# Edytuj konfigurację
nano config/config.yml

# Uruchom aplikację
docker-compose -f docker-compose.prod.yml up -d
```

Aplikacja będzie dostępna na `http://your-server-ip:40378`

## Szczegółowa instrukcja krok po kroku

### Krok 1: Przygotowanie serwera

```bash
# Aktualizuj system (Ubuntu/Debian)
sudo apt update && sudo apt upgrade -y

# Zainstaluj Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Dodaj użytkownika do grupy docker
sudo usermod -aG docker $USER

# Wyloguj się i zaloguj ponownie lub wykonaj:
newgrp docker

# Sprawdź instalację
docker --version
docker-compose --version
```

### Krok 2: Przygotowanie aplikacji

```bash
# Stwórz katalog dla aplikacji
mkdir -p /opt/socgo
cd /opt/socgo

# Pobierz pliki docker-compose
wget https://raw.githubusercontent.com/tkowalski/socgo/main/docker-compose.simple.yml
# LUB dla pełnej wersji:
wget https://raw.githubusercontent.com/tkowalski/socgo/main/docker-compose.prod.yml

# Stwórz katalog konfiguracji
mkdir -p config

# Pobierz przykładową konfigurację
wget -O config/config.yml https://raw.githubusercontent.com/tkowalski/socgo/main/config.yml.example
```

### Krok 3: Konfiguracja

Edytuj plik `config/config.yml`:

```yaml
server:
  port: "8081"
  host: "0.0.0.0"
  base_url: "https://yourdomain.com"  # Zmień na swoją domenę

database:
  data_dir: "/app/data"

providers:
  # Skonfiguruj OAuth credentials dla platform społecznościowych
  instagram:
    - name: "Your App"
      client_id: "your_instagram_client_id"
      client_secret: "your_instagram_client_secret"
  
  tiktok:
    - name: "Your App"  
      client_id: "your_tiktok_client_id"
      client_secret: "your_tiktok_client_secret"
  
  # Dodaj inne platformy według potrzeb
```

### Krok 4: Uruchomienie

#### Proste wdrożenie:
```bash
docker-compose -f docker-compose.simple.yml up -d
```

#### Produkcyjne wdrożenie:
```bash
# Uruchom
docker-compose -f docker-compose.prod.yml up -d
```

### Krok 5: Weryfikacja

```bash
# Sprawdź status kontenerów
docker-compose ps

# Sprawdź logi
docker-compose logs -f socgo

# Test healthcheck
curl http://localhost:40378/
```

## Zarządzanie

### Aktualizacja aplikacji

```bash
# Pobierz najnowszy obraz
docker-compose pull

# Restart z nowym obrazem
docker-compose up -d
```

### Backup danych

```bash
# Backup volume'ów
docker run --rm -v socgo_data:/data -v socgo_uploads:/uploads -v $(pwd):/backup busybox tar czf /backup/socgo-backup-$(date +%Y%m%d).tar.gz /data /uploads
```

### Restore danych

```bash
# Restore z backup
docker run --rm -v socgo_data:/data -v socgo_uploads:/uploads -v $(pwd):/backup busybox tar xzf /backup/socgo-backup-YYYYMMDD.tar.gz
```

### Monitoring logów

```bash
# Logi aplikacji
docker-compose logs -f socgo

# Logi wszystkich usług
docker-compose logs -f

# Logi z timestamp
docker-compose logs -f -t
```

## Troubleshooting

### Typowe problemy

1. **Port 40378 już zajęty**
   ```bash
   # Sprawdź co używa portu
   sudo netstat -tlnp | grep :40378
   # Zatrzymaj usługę lub zmień port w docker-compose
   ```

2. **Brak uprawnień do Docker**
   ```bash
   sudo usermod -aG docker $USER
   newgrp docker
   ```

3. **Aplikacja nie odpowiada**
   - Sprawdź czy port 40378 jest otwarty w firewall
   - Sprawdź czy kontener jest uruchomiony: `docker-compose ps`

4. **Aplikacja nie startuje**
   ```bash
   # Sprawdź logi
   docker-compose logs socgo
   
   # Sprawdź konfigurację
   docker-compose config
   ```

### Przydatne komendy

```bash
# Status wszystkich kontenerów
docker ps

# Restart pojedynczego serwisu
docker-compose restart socgo

# Rebuild i restart
docker-compose up -d --build

# Zatrzymanie wszystkich serwisów
docker-compose down

# Zatrzymanie z usunięciem volume'ów (OSTROŻNIE!)
docker-compose down -v

# Shell do kontenera
docker-compose exec socgo sh

# Sprawdzenie zasobów
docker stats
```

## Bezpieczeństwo

### Zalecenia

1. **Firewall**
   ```bash
   # Tylko potrzebne porty
   sudo ufw allow 22      # SSH
   sudo ufw allow 40378   # SocGo aplikacja
   sudo ufw enable
   ```

2. **HTTPS (opcjonalne)**
   - Użyj reverse proxy (nginx, Traefik) z certyfikatami SSL
   - Lub skonfiguruj aplikację za load balancerem z SSL

3. **Backup**
   - Regularnie rób backup volume'ów
   - Testuj restore procedure

4. **Monitoring**
   - Monitoruj logi aplikacji
   - Ustaw alerty dla błędów

## Skalowanie

Dla większego ruchu:

1. **Load Balancer**
   ```yaml
   socgo:
     deploy:
       replicas: 3  # Więcej instancji
   ```

2. **Zewnętrzna baza danych**
   - Użyj zewnętrznej bazy PostgreSQL/MySQL
   - Skonfiguruj w `config.yml`

3. **CDN**
   - Użyj CDN dla static assets
   - Konfiguruj caching headers

---

Potrzebujesz pomocy? Sprawdź [GitHub Issues](https://github.com/tkowalski/socgo/issues) lub stwórz nowe zgłoszenie.