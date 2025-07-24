# SocGo - Social Media Management Platform

SocGo to platforma do zarządzania treściami w mediach społecznościowych, obsługująca TikTok, Instagram i Facebook.

## ToDo

- [x] facebook post + image
- [ ] instagram post + image
- [ ] tiktok post + image

## Funkcje

- 🔗 **OAuth Integration** - Bezpieczne połączenie z kontami społecznościowymi
- 📝 **Content Management** - Tworzenie i planowanie postów
- 📅 **Scheduling** - Automatyczne publikowanie o określonych godzinach
- 📊 **Analytics** - Śledzenie wydajności postów
- 🔐 **API Tokens** - Bezpieczny dostęp przez API

## Szybki start

### 1. Instalacja

```bash
git clone <repository-url>
cd socgo
go mod download
```

### 2. Konfiguracja

Skopiuj plik konfiguracyjny:
```bash
cp cmd/config.yml.example cmd/config.yml
```

Edytuj `cmd/config.yml` i dodaj swoje App ID i App Secret dla providerów społecznościowych.

### 3. Uruchomienie

#### Opcja A: Lokalny rozwój
```bash
go run cmd/main.go
```

#### Opcja B: Z ngrok (dla OAuth)
```bash
# Terminal 1: Uruchom ngrok
ngrok http 8080

# Terminal 2: Skonfiguruj base_url w config.yml lub zmiennych środowiskowych
# i uruchom aplikację
go run cmd/main.go
```

#### Opcja C: Z Docker
```bash
docker-compose -f docker-compose.dev.yml up
```

### 4. Dostęp

- **Lokalny**: http://localhost:8080
- **Przez ngrok**: https://your-ngrok-url.ngrok-free.app

## Konfiguracja realnego hosta

Aby providerzy społecznościowi mogli przekierowywać użytkowników z powrotem do aplikacji, musisz skonfigurować realny host:

### Ręczna konfiguracja
```yaml
# config.yml
server:
  base_url: "https://your-domain.com"
```

### Zmienne środowiskowe
```bash
export SERVER_BASE_URL="https://your-domain.com"
# lub
export NGROK_URL="https://your-ngrok-url.ngrok-free.app"
```

**Uwaga**: `SERVER_BASE_URL` ma pierwszeństwo przed `NGROK_URL`.

## Konfiguracja providerów

### TikTok
1. Przejdź do [TikTok for Developers](https://developers.tiktok.com/)
2. Utwórz aplikację
3. Ustaw Redirect URI: `{base_url}/oauth/callback/tiktok`

### Instagram
1. Przejdź do [Facebook Developers](https://developers.facebook.com/)
2. Utwórz aplikację z produktem Instagram Basic Display
3. Ustaw Redirect URI: `{base_url}/oauth/callback/instagram`

### Facebook
1. Przejdź do [Facebook Developers](https://developers.facebook.com/)
2. Utwórz aplikację z produktem Facebook Login
3. Ustaw Redirect URI: `{base_url}/oauth/callback/facebook`

## API

### Generowanie tokenu API
```bash
curl -X POST http://localhost:8080/api-tokens
```

### Używanie API
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"content":"Hello World","provider_id":1}' \
     http://localhost:8080/api/posts
```

## Rozwój

### Uruchomienie testów
```bash
go test ./...
```

### Linting
```bash
golangci-lint run
```

### Generowanie szablonów
```bash
templ generate
```

## Struktura projektu

```
socgo/
├── cmd/                    # Główny punkt wejścia
├── internal/              # Logika aplikacji
│   ├── config/           # Konfiguracja
│   ├── database/         # Zarządzanie bazą danych
│   ├── handlers/         # Obsługa żądań HTTP
│   ├── middleware/       # Middleware
│   ├── oauth/           # Integracja OAuth
│   ├── providers/       # Providerzy społecznościowi
│   ├── scheduler/       # Planowanie zadań
│   └── server/          # Serwer HTTP
├── web/                  # Szablony HTML
└── docker-compose.dev.yml # Konfiguracja Docker
```

## Licencja

MIT License