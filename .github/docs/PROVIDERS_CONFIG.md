# Konfiguracja Providerów

## Przegląd

System SocGo obsługuje wiele instancji każdego providera (TikTok, Instagram, Facebook), co pozwala na podłączenie wielu kont do jednej aplikacji. Każda instancja ma własną nazwę, App ID i App Secret.

## Struktura konfiguracji

Konfiguracja providerów jest przechowywana w pliku `config.yml` w głównym katalogu projektu.

### Przykład pliku config.yml

```yaml
server:
  port: "8080"
  host: "localhost"
  base_url: "http://localhost:8080"  # Change this to your real host (e.g., https://9a8d76d5d3ee.ngrok-free.app)

database:
  data_dir: "./data"

providers:
  tiktok:
    - name: "Personal TikTok"
      client_id: "your_tiktok_client_id_1"
      client_secret: "your_tiktok_client_secret_1"
      description: "Personal TikTok account for sharing videos"
    
    - name: "Business TikTok"
      client_id: "your_tiktok_client_id_2"
      client_secret: "your_tiktok_client_secret_2"
      description: "Business TikTok account for brand content"

  instagram:
    - name: "Personal Instagram"
      client_id: "your_instagram_client_id_1"
      client_secret: "your_instagram_client_secret_1"
      description: "Personal Instagram account"
    
    - name: "Business Instagram"
      client_id: "your_instagram_client_id_2"
      client_secret: "your_instagram_client_secret_2"
      description: "Business Instagram account for brand posts"

  facebook:
    - name: "Personal Facebook"
      client_id: "your_facebook_client_id_1"
      client_secret: "your_facebook_client_secret_1"
      description: "Personal Facebook profile"
    
    - name: "Business Facebook"
      client_id: "your_facebook_client_id_2"
      client_secret: "your_facebook_client_secret_2"
      description: "Business Facebook page"
```

## Konfiguracja realnego hosta dla OAuth

Aby providerzy społecznościowi mogli przekierowywać użytkowników z powrotem do aplikacji po autoryzacji, musisz skonfigurować realny host. System obsługuje kilka sposobów konfiguracji:

### 1. Ręczna konfiguracja w config.yml

```yaml
server:
  port: "8080"
  host: "localhost"
  base_url: "https://9a8d76d5d3ee.ngrok-free.app"  # Twój realny host
```

### 2. Zmienna środowiskowa SERVER_BASE_URL

```bash
export SERVER_BASE_URL="https://9a8d76d5d3ee.ngrok-free.app"
go run cmd/main.go
```

### 3. Zmienna środowiskowa NGROK_URL

```bash
export NGROK_URL="https://9a8d76d5d3ee.ngrok-free.app"
go run cmd/main.go
```

### 4. Plik .env

```env
SERVER_BASE_URL=https://9a8d76d5d3ee.ngrok-free.app
# lub
NGROK_URL=https://9a8d76d5d3ee.ngrok-free.app
```

**Uwaga**: `SERVER_BASE_URL` ma pierwszeństwo przed `NGROK_URL`.

## Pola konfiguracji

### Dla każdej instancji providera:

- **name**: Unikalna nazwa instancji (wyświetlana na stronie)
- **client_id**: App ID z platformy społecznościowej
- **client_secret**: App Secret z platformy społecznościowej
- **description**: Opis instancji (opcjonalny)

## Jak uzyskać App ID i App Secret

### TikTok
1. Przejdź do [TikTok for Developers](https://developers.tiktok.com/)
2. Utwórz nową aplikację
3. Skopiuj Client Key (App ID) i Client Secret
4. Ustaw Redirect URI na: `{base_url}/oauth/callback/tiktok`

### Instagram
1. Przejdź do [Facebook Developers](https://developers.facebook.com/)
2. Utwórz nową aplikację
3. Dodaj produkt Instagram Basic Display
4. Skopiuj App ID i App Secret
5. Ustaw Redirect URI na: `{base_url}/oauth/callback/instagram`

### Facebook
1. Przejdź do [Facebook Developers](https://developers.facebook.com/)
2. Utwórz nową aplikację
3. Dodaj produkt Facebook Login
4. Skopiuj App ID i App Secret
5. Ustaw Redirect URI na: `{base_url}/oauth/callback/facebook`

**Uwaga**: Zastąp `{base_url}` rzeczywistym URL-em twojej aplikacji (np. `https://9a8d76d5d3ee.ngrok-free.app`)

## Bezpieczeństwo

⚠️ **WAŻNE**: Nigdy nie commituj prawdziwych App ID i App Secret do repozytorium Git!

### Zalecane praktyki:

1. **Użyj zmiennych środowiskowych**:
   ```bash
   export TIKTOK_CLIENT_ID_1="your_real_client_id"
   export TIKTOK_CLIENT_SECRET_1="your_real_client_secret"
   ```

2. **Użyj pliku .env** (nie commituj go):
   ```env
   TIKTOK_CLIENT_ID_1=your_real_client_id
   TIKTOK_CLIENT_SECRET_1=your_real_client_secret
   ```

3. **W produkcji użyj bezpiecznego menedżera sekretów** (AWS Secrets Manager, HashiCorp Vault, etc.)

## Fallback do zmiennych środowiskowych

Jeśli plik `config.yml` nie istnieje, system automatycznie użyje zmiennych środowiskowych:

```bash
# TikTok
export TIKTOK_CLIENT_ID="your_client_id"
export TIKTOK_CLIENT_SECRET="your_client_secret"

# Instagram
export INSTAGRAM_CLIENT_ID="your_client_id"
export INSTAGRAM_CLIENT_SECRET="your_client_secret"

# Facebook
export FACEBOOK_CLIENT_ID="your_client_id"
export FACEBOOK_CLIENT_SECRET="your_client_secret"
```

## Testowanie konfiguracji

1. Uruchom aplikację: `go run cmd/main.go`
2. Przejdź do: `http://localhost:8080/providers`
3. Sprawdź czy dostępne są wszystkie skonfigurowane instancje providerów
4. Spróbuj podłączyć jeden z providerów

## Rozwiązywanie problemów

### Błąd "Provider configuration not found"
- Sprawdź czy nazwa providera w URL odpowiada nazwie w config.yml
- Upewnij się, że plik config.yml jest poprawnie sformatowany

### Błąd "Invalid client_id"
- Sprawdź czy App ID jest poprawne
- Upewnij się, że aplikacja jest skonfigurowana na platformie deweloperskiej

### Błąd "Redirect URI mismatch"
- Sprawdź czy Redirect URI w aplikacji deweloperskiej odpowiada rzeczywistemu URL-owi aplikacji
- Upewnij się, że używasz poprawnego protokołu (http/https)
- Sprawdź czy BaseURL jest poprawnie skonfigurowany w config.yml lub zmiennych środowiskowych
- Dla ngrok: upewnij się, że używasz HTTPS URL (nie HTTP)



### Sprawdzanie konfiguracji BaseURL
Podczas startu aplikacji zobaczysz logi:
```
Server will start on: localhost:8080
Base URL for OAuth: https://9a8d76d5d3ee.ngrok-free.app
```

Jeśli BaseURL nie jest poprawny, sprawdź:
1. Czy ngrok jest uruchomiony
2. Czy zmienna środowiskowa SERVER_BASE_URL jest ustawiona poprawnie
3. Czy w config.yml jest poprawny base_url 