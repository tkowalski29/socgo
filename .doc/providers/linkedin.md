# LinkedIn Integration

## Przegląd

Integracja LinkedIn pozwala użytkownikom na:
- Autoryzację OAuth 2.0 z LinkedIn
- Publikowanie postów tekstowych na LinkedIn
- Pobieranie statusu opublikowanych postów
- Zarządzanie tokeny dostępu

## Konfiguracja

### LinkedIn Developer App

1. Utwórz aplikację w [LinkedIn Developer Portal](https://developer.linkedin.com/)
2. Skonfiguruj następujące scopes:
   - `openid` - Use your name and photo
   - `profile` - Use your name and photo  
   - `w_member_social` - Create, modify, and delete posts, comments, and reactions on your behalf
   - `email` - Use the primary email address associated with your LinkedIn account

3. Ustaw Redirect URLs:
   - `{BASE_URL}/oauth/callback/linkedin`

### Konfiguracja aplikacji

Dodaj następującą konfigurację do `config.yml`:

```yaml
providers:
  linkedin:
    - name: "LinkedIn Business"
      client_id: "your_linkedin_client_id"
      client_secret: "your_linkedin_client_secret"
      description: "LinkedIn Business Profile"
```

## Architektura

### Pliki implementacji

- `internal/social/linkedin/oauth.go` - OAuth 2.0 authentication
- `internal/social/linkedin/post.go` - Publishing functionality
- `internal/social/linkedin/oauth_test.go` - OAuth tests
- `internal/social/linkedin/post_test.go` - Publishing tests

### Interfejsy

LinkedIn implementuje następujące interfejsy:
- `oauth.ProviderAuth` - OAuth authentication
- `provider.Provider` - Post publishing

### Przepływ OAuth

1. Użytkownik klika "Connect LinkedIn"
2. Przekierowanie do LinkedIn OAuth
3. Użytkownik autoryzuje aplikację
4. LinkedIn przekierowuje z authorization code
5. Aplikacja wymienia code na access token
6. Aplikacja pobiera informacje o użytkowniku
7. Konto zostaje zapisane w bazie danych

## API Endpoints

### LinkedIn API

- **Authorization**: `https://www.linkedin.com/oauth/v2/authorization`
- **Token Exchange**: `https://www.linkedin.com/oauth/v2/accessToken`  
- **User Info**: `https://api.linkedin.com/v2/userinfo`
- **UGC Posts**: `https://api.linkedin.com/v2/ugcPosts`

### Aplikacja

- **Connect**: `GET /oauth/connect/linkedin?name={provider_name}`
- **Callback**: `GET /oauth/callback/linkedin`

## Publikowanie Postów

### Wspierane formaty

✅ **Tekstowe posty**
- Pełne wsparcie poprzez LinkedIn UGC Posts API
- Maksymalna długość zgodna z LinkedIn

⚠️ **Posty z mediami**
- Obecnie częściowo zaimplementowane
- Media są dodawane jako tekst w opisie posta
- Planowane: pełne wsparcie dla obrazów przez LinkedIn Media Upload API

### Przykład użycia

```go
// Publikowanie prostego posta
postID, err := provider.Publish(ctx, dbProvider, "Hello LinkedIn!", nil)

// Publikowanie z "mediami" (obecnie jako tekst)
media := []provider.Media{{FileName: "image.jpg", MimeType: "image/jpeg"}}
postID, err := provider.Publish(ctx, dbProvider, "Post with image", media)
```

## Statusy Postów

LinkedIn zwraca następujące statusy:
- `PUBLISHED` - Post opublikowany
- `DRAFT` - Post w wersji roboczej
- Inne - mapowane na `pending`

## Ograniczenia

1. **Rate Limits**: LinkedIn ma limity API calls per day/month
2. **Media Upload**: Obecnie nie w pełni zaimplementowane
3. **Company Pages**: Obecnie tylko personal profiles
4. **Post Scheduling**: Nie wspierane (LinkedIn API limitation)

## Rozwiązywanie Problemów

### Częste błędy

**Invalid Access Token**
- Token wygasł - użyj refresh token
- Niepoprawne scopes - sprawdź konfigurację app

**User consent required**
- Użytkownik musi ponownie autoryzować aplikację
- Sprawdź czy wymagane scopes są zatwierdzone

**API Rate Limit**
- Zbyt wiele requestów - implementuj exponential backoff
- Sprawdź LinkedIn API quotas

### Debug

Włącz logi debug dla szczegółowych informacji:
```go
log.Printf("LinkedIn OAuth URL: %s", connectURL)
log.Printf("LinkedIn API response: %s", responseBody)
```

## Testowanie

Uruchom testy jednostkowe:
```bash
go test ./internal/social/linkedin/ -v
```

Uwaga: Niektóre testy wymagają mockowania HTTP client dla pełnej funkcjonalności.

## Przyszłe Ulepszenia

1. **Pełne wsparcie dla mediów**
   - Implementacja LinkedIn Media Upload API
   - Wsparcie dla obrazów i wideo

2. **Company Pages**
   - Rozszerzenie o publikowanie na behalf of company pages
   - Zarządzanie wieloma profilami firmowymi

3. **Advanced Features**
   - Post scheduling (jeśli LinkedIn API będzie wspierać)
   - Post analytics i metrics
   - Comment management

## Bezpieczeństwo

- Tokeny są szyfrowane w bazie danych
- Client secret nie jest eksponowany w frontend
- HTTPS wymagane dla production
- State parameter zabezpiecza przed CSRF