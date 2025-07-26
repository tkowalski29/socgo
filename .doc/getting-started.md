# Getting Started

Szybki przewodnik instalacji i pierwszego uruchomienia SocGo.

## 📋 Wymagania systemowe

### Minimalne wymagania
- **Go**: 1.24.3 lub nowszy
- **System**: Windows, macOS, Linux
- **RAM**: 512 MB
- **Dysk**: 100 MB wolnego miejsca
- **Port**: 8080 (domyślnie)

### Wymagane narzędzia
- Git (do klonowania repozytorium)
- Edytor tekstu (do konfiguracji)

## ⚡ Szybka instalacja

### 1. Pobierz kod źródłowy
```bash
git clone https://github.com/tkowalski/socgo.git
cd socgo
```

### 2. Skopiuj konfigurację
```bash
cp config.yml.example config.yml
```

### 3. Uruchom aplikację
```bash
go run main.go
```

### 4. Otwórz w przeglądarce
```
http://localhost:8080
```

## ⚙️ Konfiguracja

### Plik config.yml

Podstawowa konfiguracja w `config.yml`:

```yaml
server:
  port: 8080
  host: "localhost"
  
database:
  path: "database.db"
  
oauth:
  facebook:
    client_id: "your_facebook_app_id"
    client_secret: "your_facebook_app_secret"
    redirect_url: "http://localhost:8080/oauth/facebook/callback"
  
  instagram:
    client_id: "your_instagram_app_id" 
    client_secret: "your_instagram_app_secret"
    redirect_url: "http://localhost:8080/oauth/instagram/callback"
    
  tiktok:
    client_id: "your_tiktok_app_id"
    client_secret: "your_tiktok_app_secret"
    redirect_url: "http://localhost:8080/oauth/tiktok/callback"

uploads:
  path: "./uploads"
  max_size: 10485760  # 10MB in bytes
```

### Konfiguracja OAuth

#### Facebook App
1. Przejdź do [Facebook Developers](https://developers.facebook.com)
2. Utwórz nową aplikację
3. Dodaj **Facebook Login** product
4. Skonfiguruj Valid OAuth Redirect URIs:
   ```
   http://localhost:8080/oauth/facebook/callback
   ```
5. Skopiuj App ID i App Secret do `config.yml`

#### Instagram Basic Display
1. W tej samej aplikacji Facebook dodaj **Instagram Basic Display**
2. Skonfiguruj Valid OAuth Redirect URIs:
   ```
   http://localhost:8080/oauth/instagram/callback
   ```
3. Skopiuj Instagram App ID i App Secret

#### TikTok for Developers
1. Przejdź do [TikTok for Developers](https://developers.tiktok.com)
2. Utwórz nową aplikację
3. Skonfiguruj Redirect URI:
   ```
   http://localhost:8080/oauth/tiktok/callback
   ```
4. Skopiuj Client Key i Client Secret

## 🚀 Pierwsze uruchomienie

### 1. Start aplikacji
```bash
go run main.go
```

Powinieneś zobaczyć:
```
2024/01/15 10:30:00 Starting SocGo server on :8080
2024/01/15 10:30:00 Database migrations completed
2024/01/15 10:30:00 Server is ready to handle requests
```

### 2. Sprawdź działanie
Otwórz `http://localhost:8080` - powinna pojawić się strona główna SocGo.

### 3. Połącz pierwsze konto
1. Kliknij **"Settings"**
2. Wybierz platformę
3. Kliknij **"Connect"**
4. Autoryzuj aplikację

## 🐳 Docker (opcjonalnie)

### Uruchomienie przez Docker
```bash
# Build image
docker build -t socgo .

# Run container
docker run -p 8080:8080 -v $(pwd)/config.yml:/app/config.yml socgo
```

### Docker Compose
```yaml
version: '3.8'
services:
  socgo:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config.yml:/app/config.yml
      - ./uploads:/app/uploads
      - ./database.db:/app/database.db
```

Uruchom przez:
```bash
docker-compose up
```

## 🔧 Rozwiązywanie problemów

### Port 8080 jest zajęty
Zmień port w `config.yml`:
```yaml
server:
  port: 8081
```

### Błąd połączenia z bazą danych
```bash
# Sprawdź uprawnienia do katalogu
chmod 755 .
chmod 644 database.db  # jeśli plik istnieje
```

### Błędy OAuth
1. Sprawdź czy URLs callback są poprawne
2. Upewnij się, że aplikacje są w trybie Development/Production
3. Sprawdź klucze API w `config.yml`

### Błąd "Module not found"
```bash
# Pobierz zależności
go mod download
go mod tidy
```

## 📊 Weryfikacja instalacji

### Sprawdzenie endpointów
```bash
# Health check
curl http://localhost:8080/

# API status
curl http://localhost:8080/api/status
```

### Logi aplikacji
Aplikacja loguje informacje do stdout. Sprawdź:
- Startup messages
- HTTP requests
- Błędy publikacji
- OAuth callbacks

## 🔄 Aktualizacja

### Aktualizacja kodu
```bash
git pull origin main
go mod download
go run main.go
```

### Migracje bazy danych
Migracje są automatyczne przy starcie aplikacji.

## 📁 Struktura plików po instalacji

```
socgo/
├── config.yml          # Konfiguracja
├── database.db         # Baza SQLite (tworzona automatycznie)
├── uploads/            # Przesłane pliki
├── main.go            # Punkt wejścia aplikacji
├── internal/          # Kod źródłowy
└── web/              # Frontend
```

## ⭐ Następne kroki

Po udanej instalacji:

1. **[Skonfiguruj dostawców](./guide/providers.md)** - Połącz konta social media
2. **[Utwórz pierwszy post](./guide/posts.md)** - Przetestuj publikację
3. **[Zaplanuj treść](./guide/scheduling.md)** - Użyj kalendarza
4. **[Sprawdź API](./api/)** - Jeśli planujesz integracje

## 🆘 Pomoc

Jeśli napotkasz problemy:

1. Sprawdź [FAQ w przewodniku użytkownika](./guide/#często-zadawane-pytania)
2. Przejrzyj logi aplikacji
3. Sprawdź konfigurację OAuth
4. Zgłoś issue na GitHub (jeśli publicznie dostępne)