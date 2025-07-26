# Calendar & Scheduling API

Endpointy do zarządzania kalendarzem i planowania publikacji postów.

## Calendar View

Pobiera widok kalendarza z liczbą postów dla każdego dnia.

```http
GET /calendar?year={year}&month={month}
```

### Parametry query

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `year` | int | Nie | Rok (domyślnie bieżący) |
| `month` | int | Nie | Miesiąc 1-12 (domyślnie bieżący) |

### Headers

- `Accept: application/json` - Zwraca JSON
- `Accept: text/html` - Zwraca HTML grid (HTMX)

### Odpowiedź JSON

```json
{
  "year": 2024,
  "month": 1,
  "days": [
    {
      "day": 1,
      "has_posts": true,
      "post_count": 3
    },
    {
      "day": 2,
      "has_posts": false,
      "post_count": 0
    }
  ]
}
```

### Odpowiedź HTML

HTML grid calendar z oznaczonymi dniami z postami.

## Week View

Pobiera widok tygodniowy z szczegółami postów dla każdej godziny.

```http
GET /calendar/week?start={date}
```

### Parametry query

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `start` | string | Nie | Data początkowa (YYYY-MM-DD) |

### Odpowiedź JSON

```json
{
  "start_date": "2024-01-15T00:00:00Z",
  "end_date": "2024-01-22T00:00:00Z",
  "days": [
    {
      "date": "2024-01-15T00:00:00Z",
      "day": 15,
      "hours": {
        "9": [
          {
            "id": 123,
            "content": "Post content",
            "provider_id": 1,
            "provider": {
              "id": 1,
              "name": "My Instagram",
              "type": "instagram"
            },
            "scheduled_at": "2024-01-15T09:00:00Z",
            "created_at": "2024-01-15T08:00:00Z",
            "status": "pending",
            "type": "scheduled",
            "hour": 9
          }
        ]
      }
    }
  ]
}
```

### Struktura WeekPost

| Pole | Typ | Opis |
|------|-----|------|
| `id` | uint | ID posta |
| `content` | string | Treść posta |
| `provider_id` | uint | ID dostawcy |
| `provider` | object | Dane dostawcy |
| `scheduled_at` | string | Data planowana (dla scheduled) |
| `published_at` | string | Data publikacji (dla published) |
| `created_at` | string | Data utworzenia |
| `status` | string | Status: pending/published/failed |
| `external_id` | string | ID u dostawcy |
| `external_url` | string | URL posta u dostawcy |
| `error_message` | string | Komunikat błędu |
| `type` | string | Typ: scheduled/published |
| `hour` | int | Godzina (0-23) |

## Calendar Page

Strona kalendarza z interfejsem użytkownika.

```http
GET /calendar
```

### Parametry query

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `flash` | string | Nie | Wiadomość flash |
| `flash_type` | string | Nie | Typ: success/error/info |

### Odpowiedź

HTML strona z kalendarzem.

## Funkcje kalendarza

### Nawigacja miesięczna
- Przewijanie między miesiącami
- Oznaczanie dni z postami
- Liczba postów na dzień

### Widok tygodniowy
- Szczegółowy widok godzinowy (0-23)
- Posty pogrupowane według godzin
- Rozwijane sekcje dla wielu postów
- Możliwość dodawania postów w przyszłych slotach

### Interaktywność
- Kliknięcie na post - szczegóły
- Hover na przyszłe sloty - opcja dodania
- Rozwijanie/zwijanie postów

## Przykład implementacji HTMX

```html
<!-- Pobieranie kalendarza -->
<div hx-get="/calendar?year=2024&month=1" 
     hx-trigger="load"
     hx-target="#calendar-grid">
</div>

<!-- Pobieranie tygodnia -->
<div hx-get="/calendar/week?start=2024-01-15"
     hx-trigger="load" 
     hx-target="#week-view">
</div>
```

## JavaScript funkcje

### openPostSidebarWithDateTime(date, hour)
Otwiera sidebar tworzenia posta z predefiniowaną datą i godziną.

### showCalendarPostSidebar(postId, type)
Pokazuje szczegóły posta w sidebarze.

### toggleHourPosts(dayIndex, hour)
Przełącza widoczność dodatkowych postów w danej godzinie.

## Ograniczenia

- Kalendarz pokazuje maksymalnie 2 posty na godzinę (pozostałe ukryte)
- Dodawanie postów możliwe tylko w przyszłych slotach czasowych
- Widok tygodniowy zaczyna się od poniedziałku