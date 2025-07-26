# Provider Setup

Przewodnik po konfiguracji i zarządzaniu połączeniami z platformami społecznościowymi w SocGo.

## 🔗 Obsługiwane platformy

### Facebook
- **Typ**: Social Media Platform
- **Funkcje**: Posty tekstowe, obrazy, wideo
- **Wymagania**: Facebook App, OAuth 2.0
- **Limity**: 25 postów/dzień na stronę

### Instagram
- **Typ**: Photo/Video Sharing
- **Funkcje**: Posty z obrazami, Stories (w przygotowaniu)
- **Wymagania**: Instagram Business Account
- **Limity**: 25 postów/dzień

### TikTok
- **Typ**: Video Platform
- **Funkcje**: Wideo, opisy, hashtagi
- **Wymagania**: TikTok for Business
- **Limity**: 10 wideo/dzień

## ⚙️ Konfiguracja OAuth

### Facebook App Setup
1. Przejdź do [Facebook Developers](https://developers.facebook.com)
2. Utwórz nową aplikację
3. Dodaj **Facebook Login** product
4. Skonfiguruj Valid OAuth Redirect URIs:
   ```
   http://localhost:8080/oauth/facebook/callback
   ```
5. Skopiuj App ID i App Secret

### Instagram Basic Display
1. W tej samej aplikacji Facebook dodaj **Instagram Basic Display**
2. Skonfiguruj Valid OAuth Redirect URIs:
   ```
   http://localhost:8080/oauth/instagram/callback
   ```
3. Skopiuj Instagram App ID i App Secret

### TikTok for Developers
1. Przejdź do [TikTok for Developers](https://developers.tiktok.com)
2. Utwórz nową aplikację
3. Skonfiguruj Redirect URI:
   ```
   http://localhost:8080/oauth/tiktok/callback
   ```
4. Skopiuj Client Key i Client Secret

## 🔧 Konfiguracja w SocGo

### Plik config.yml
```yaml
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
```

### Zmienne środowiskowe (opcjonalnie)
```bash
export FACEBOOK_CLIENT_ID="your_facebook_app_id"
export FACEBOOK_CLIENT_SECRET="your_facebook_app_secret"
export INSTAGRAM_CLIENT_ID="your_instagram_app_id"
export INSTAGRAM_CLIENT_SECRET="your_instagram_app_secret"
export TIKTOK_CLIENT_ID="your_tiktok_app_id"
export TIKTOK_CLIENT_SECRET="your_tiktok_app_secret"
```

## 🔐 Połączenie kont

### Krok 1: Przejdź do Settings
1. Kliknij **"Settings"** w menu nawigacji
2. Wybierz zakładkę **"Providers"**

### Krok 2: Wybierz platformę
1. Znajdź platformę (Facebook/Instagram/TikTok)
2. Kliknij przycisk **"Connect"**

### Krok 3: Autoryzacja
1. Zostaniesz przekierowany do platformy
2. Zaloguj się na swoje konto
3. Zatwierdź uprawnienia dla SocGo
4. Wrócisz do aplikacji

### Krok 4: Konfiguracja
1. Nadaj nazwę połączeniu (np. "Moja firma - Facebook")
2. Wybierz strony/konta do zarządzania
3. Kliknij **"Save"**

## 📊 Zarządzanie połączeniami

### Status połączeń
- **Connected** - Aktywne połączenie
- **Disconnected** - Rozłączone
- **Error** - Błąd połączenia
- **Expired** - Token wygasł

### Akcje na połączeniach
- **Edit** - Zmień nazwę lub ustawienia
- **Reconnect** - Odśwież token
- **Delete** - Usuń połączenie

## 🔄 Odświeżanie tokenów

### Automatyczne odświeżanie
- Tokeny są automatycznie odświeżane
- Powiadomienia o wygasających tokenach
- Bezpieczne przechowywanie w bazie danych

### Ręczne odświeżanie
1. Przejdź do Settings > Providers
2. Znajdź połączenie z błędem
3. Kliknij **"Reconnect"**
4. Przejdź przez proces autoryzacji

## 🛡️ Bezpieczeństwo

### Przechowywanie tokenów
- Tokeny są szyfrowane w bazie danych
- Dostęp tylko przez aplikację
- Automatyczne czyszczenie wygasłych tokenów

### Uprawnienia aplikacji
- **Facebook**: publish_to_groups, pages_manage_posts
- **Instagram**: instagram_basic, instagram_content_publish
- **TikTok**: video.upload, video.publish

### Best Practices
- Używaj różnych aplikacji dla różnych środowisk
- Regularnie sprawdzaj uprawnienia
- Monitoruj logi autoryzacji

## ❓ Często zadawane pytania

### Dlaczego połączenie się nie udaje?
1. Sprawdź klucze API w config.yml
2. Upewnij się, że URLs callback są poprawne
3. Sprawdź czy aplikacja jest w trybie Development/Production

### Jak zmienić nazwę połączenia?
1. Przejdź do Settings > Providers
2. Kliknij **"Edit"** przy połączeniu
3. Zmień nazwę i zapisz

### Co jeśli token wygasł?
Kliknij **"Reconnect"** przy połączeniu. Token zostanie automatycznie odświeżony.

### Czy mogę połączyć wiele kont?
Tak, możesz połączyć wiele kont na tej samej platformie.

## 🔧 Rozwiązywanie problemów

### Błędy OAuth
- **"Invalid redirect URI"**: Sprawdź konfigurację w aplikacji platformy
- **"App not approved"**: Przełącz aplikację w tryb Development
- **"Scope not allowed"**: Sprawdź uprawnienia aplikacji

### Problemy z połączeniem
- Sprawdź połączenie internetowe
- Sprawdź firewall/antywirus
- Sprawdź logi aplikacji

### Błędy API
- Sprawdź limity API platformy
- Sprawdź format danych
- Sprawdź status platformy

## 🚀 Zaawansowane funkcje

### Multiple Accounts
- Zarządzanie wieloma kontami
- Przełączanie między kontami
- Masowe operacje

### API Monitoring
- Monitorowanie limitów API
- Alerty o błędach
- Statystyki użycia

### Custom Providers
- Dodawanie własnych integracji
- Custom OAuth flows
- Rozszerzanie funkcjonalności 