# Posts Management API

Endpointy do zarządzania postami - tworzenie, publikowanie, planowanie i historia.

## Create Post

Tworzy i publikuje lub planuje post.

```http
POST /posts
Content-Type: application/json
```

### Request Body

```json
{
  "provider_id": 1,
  "content": "Tekst posta do opublikowania",
  "schedule_at": "2024-01-15T14:30:00Z"
}
```

### Parametry

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `provider_id` | uint | Tak | ID skonfigurowanego dostawcy |
| `content` | string | Tak | Treść posta |
| `schedule_at` | string | Nie | Data publikacji (ISO8601) lub "now" |

### Odpowiedź sukces (publikacja natychmiastowa)

```json
{
  "id": 123,
  "status": "published",
  "provider_id": 1,
  "content": "Tekst posta do opublikowania",
  "created_at": "2024-01-15T10:30:00Z",
  "message": "Post published successfully. Post ID: fb_123456"
}
```

### Odpowiedź sukces (post zaplanowany)

```json
{
  "id": 124,
  "status": "pending",
  "provider_id": 1,
  "content": "Tekst posta do opublikowania",
  "created_at": "2024-01-15T10:30:00Z",
  "message": "Post scheduled successfully for 2024-01-15T14:30:00Z"
}
```

## Get Post History

Pobiera historię postów z paginacją.

```http
GET /posts/history?page=1
```

### Parametry query

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `page` | int | Nie | Numer strony (domyślnie 1) |

### Headers

- `Accept: application/json` - Zwraca JSON
- `Accept: text/html` - Zwraca HTML (HTMX)

### Odpowiedź JSON

```json
{
  "posts": [
    {
      "id": 123,
      "content": "Treść posta",
      "provider_id": 1,
      "provider": {
        "id": 1,
        "name": "My Instagram",
        "type": "instagram"
      },
      "scheduled_at": "2024-01-15T14:30:00Z",
      "created_at": "2024-01-15T10:30:00Z",
      "status": "pending"
    }
  ],
  "page": 1,
  "total": 50
}
```

## Get Post Details

Pobiera szczegóły konkretnego posta.

```http
GET /posts/details/{id}?type={type}
```

### Parametry

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `id` | uint | Tak | ID posta |
| `type` | string | Nie | Typ: "published" lub "scheduled" |

### Odpowiedź

HTML z szczegółami posta (dla HTMX).

## Delete Post

Usuwa post lub anuluje zaplanowany post.

```http
DELETE /posts/{id}?type={type}
```

### Parametry

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `id` | uint | Tak | ID posta |
| `type` | string | Nie | Typ: "published" lub "scheduled" |

### Odpowiedź sukces

```json
{
  "status": "success",
  "message": "Post deleted successfully"
}
```

## File Preview

Generuje podgląd plików przed wysłaniem.

```http
POST /posts/file-preview
Content-Type: application/json
```

### Request Body

```json
{
  "files": [
    {
      "index": 0,
      "fileName": "image.jpg",
      "fileType": "image/jpeg",
      "dataURL": "data:image/jpeg;base64,..."
    }
  ]
}
```

### Odpowiedź

HTML z podglądem plików (dla HTMX).

## Statistics Endpoints

### Providers Count
```http
GET /posts/providers/count
```

### Published Count
```http
GET /posts/published/count
```

### Scheduled Count
```http
GET /posts/scheduled/count
```

### Monthly Count
```http
GET /posts/monthly/count
```

Zwracają liczbę jako tekst.

## Provider Options

Pobiera dostępnych dostawców do wyboru.

```http
GET /posts/providers/options
```

### Odpowiedź

HTML z checkboxami dostawców (dla HTMX).

## Provider Settings

Pobiera ustawienia specyficzne dla dostawcy.

```http
GET /posts/provider/settings?type={type}&name={name}
```

### Parametry query

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `type` | string | Tak | Typ dostawcy |
| `name` | string | Tak | Nazwa dostawcy |

### Odpowiedź

HTML z ustawieniami dostawcy (dla HTMX).

## Provider Tabs

Generuje zakładki dla wybranych dostawców.

```http
GET /posts/provider/tabs?providers={ids}&active={index}
```

### Parametry query

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `providers` | string | Tak | Lista ID dostawców (rozdzielone przecinkami) |
| `active` | int | Nie | Indeks aktywnej zakładki |

### Odpowiedź

HTML z zakładkami dostawców (dla HTMX).

## Kody błędów

| Kod | Błąd | Opis |
|-----|------|------|
| 400 | `provider_id_required` | Brak ID dostawcy |
| 400 | `content_required` | Brak treści posta |
| 400 | `invalid_schedule_format` | Nieprawidłowy format daty |
| 400 | `schedule_in_past` | Data w przeszłości |
| 404 | `provider_not_found` | Dostawca nie został znaleziony |
| 400 | `provider_not_configured` | Dostawca nie jest skonfigurowany |
| 500 | `publish_failed` | Błąd publikacji |