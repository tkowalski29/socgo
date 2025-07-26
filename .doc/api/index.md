# API Reference

SocGo udostępnia REST API dla zarządzania postami, planowania publikacji i integracji z platformami społecznościowymi.

## Base URL

```
http://localhost:8080
```

## Authentyfikacja

SocGo używa systemu opartego na sesji z domyślnym użytkownikiem `default_user`. Wszystkie żądania muszą zawierać odpowiednie nagłówki sesji.

## Formaty odpowiedzi

API zwraca dane w formacie JSON lub HTML w zależności od nagłówka `Accept`:

- `application/json` - Odpowiedź JSON (API)
- `text/html` - Odpowiedź HTML (HTMX)

## Endpointy

### [OAuth Endpoints](./oauth.md)
Zarządzanie połączeniami z dostawcami OAuth (Facebook, Instagram, TikTok).

### [Posts Management](./posts.md) 
Tworzenie, publikowanie i zarządzanie postami.

### [Calendar & Scheduling](./calendar.md)
Funkcje kalendarza i planowania publikacji.

### [Notifications](./notifications.md)
System powiadomień o statusie postów.

### [Settings](./settings.md)
Konfiguracja dostawców i ustawienia aplikacji.

## Kody statusów HTTP

| Kod | Znaczenie | Opis |
|-----|-----------|------|
| 200 | OK | Żądanie przetworzono pomyślnie |
| 201 | Created | Zasób został utworzony |
| 400 | Bad Request | Nieprawidłowe żądanie |
| 401 | Unauthorized | Brak autoryzacji |
| 404 | Not Found | Zasób nie został znaleziony |
| 405 | Method Not Allowed | Niedozwolona metoda HTTP |
| 500 | Internal Server Error | Błąd serwera |

## Struktury błędów

```json
{
  "error": "Error message",
  "details": "Additional error details"
}
```

## Rate Limiting

API nie implementuje obecnie rate limiting, ale zaleca się rozważne korzystanie z endpointów.