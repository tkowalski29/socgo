# Post Process Package

Ten katalog zawiera procesy związane z operacjami na postach, w tym regularne posty i operacje appendu.

## Struktura

### Pliki główne
- `post.go` - główny proces tworzenia regularnych postów
- `scheduler.go` - proces schedulera do publikowania postów pending
- `data/interfaces.go` - interfejsy i struktury danych

### Katalog `internal/`
Zawiera uniwersalne taski używane przez różne procesy:

#### Uniwersalne taski:
1. **`validate_request_task.go`** - waliduje żądanie HTTP (metoda POST, multipart/form-data)
2. **`parse_data_task.go`** - parsuje dane formularza (obsługuje zarówno regularne posty jak i append)
3. **`validate_content_task.go`** - waliduje treść (różne limity dla regularnych postów i append)
4. **`handle_file_uploads_task.go`** - obsługuje uploady plików (różne pola dla różnych operacji)
5. **`extract_provider_settings_task.go`** - wyciąga ustawienia providerów
6. **`validate_providers_task.go`** - sprawdza konfigurację providerów (z opcjonalną walidacją append)
7. **`save_post_task.go`** - zapisuje posty do bazy danych z odpowiednim statusem i datą wysyłki

#### Taski dla schedulera:
8. **`find_pending_posts_task.go`** - znajduje wszystkie posty pending gotowe do publikacji
9. **`publish_posts_task.go`** - publikuje znalezione posty pending

## Procesy

### 1. Proces Regularnego Posta (`PostProcess`)

Proces tworzenia nowych postów. Składa się z następujących kroków:

1. **Walidacja żądania** - sprawdza metodę POST i content-type
2. **Parsowanie danych** - wyciąga content, providery, schedule_at
3. **Walidacja treści** - sprawdza długość treści (max 2200 znaków dla regularnych postów)
4. **Obsługa plików** - przetwarza pliki z pola `media`
5. **Ustawienia providerów** - wyciąga specyficzne ustawienia
6. **Walidacja providerów** - sprawdza czy providerzy są skonfigurowani
7. **Zapis do bazy** - zapisuje posty do tabeli `posts` ze statusem "pending"

### 2. Proces Schedulera (`SchedulerProcess`)

Proces publikowania postów pending. Składa się z następujących kroków:

1. **Znajdowanie postów** - znajduje wszystkie posty ze statusem "pending" które są gotowe do wysłania
2. **Publikowanie** - publikuje znalezione posty i aktualizuje ich status

## Uniwersalne Taski

Taski zostały zaprojektowane jako uniwersalne i dostosowują swoje zachowanie na podstawie parametru `operation` w formularzu:

### Obsługa wielu providerów
- Task `parse_data_task.go` obsługuje zarówno comma-separated string jak i multiple fields o tej samej nazwie
- Dla każdego wybranego providera tworzony jest osobny post w bazie danych

### Dynamiczne pola formularza
- **Regularne posty**: `content`, `media`, `schedule_at_native`
- **Append**: `append_content`, `append_media`, `existing_post_id`

### Walidacja treści
- **Regularne posty**: max 2200 znaków
- **Append**: max 10000 znaków

### Obsługa plików
- **Regularne posty**: prefix `post_` w nazwach plików
- **Append**: prefix `append_` w nazwach plików

## Baza danych

### Tabela `posts`
- `ScheduledAt` - data planowanej wysyłki
- `Status` - "pending", "published", "failed"
- `PublishedAt` - data faktycznej publikacji
- `ExternalID` - ID posta na platformie zewnętrznej
- `ExternalURL` - URL do posta na platformie zewnętrznej

### Usunięcie tabeli `scheduled_jobs`
- Scheduler pracuje bezpośrednio z tabelą `posts`
- Nie ma już potrzeby osobnej tabeli dla zadań

## Użycie

### Regularny post
```go
// Tworzenie procesu posta
postProcess := post.NewPostProcess(providerService)

// Tworzenie kontekstu
ctx := &process_post_data.PostContext{
    Request:     r,
    UserID:      userID,
    DB:          dbManager,
    // ... inne pola
}

// Wykonanie procesu
if err := postProcess.Execute(ctx); err != nil {
    // obsługa błędu
}
```

### Scheduler
```go
// Tworzenie procesu schedulera
schedulerProcess := post.NewSchedulerProcess(providerService)

// Tworzenie kontekstu dla schedulera
ctx := &process_post_data.PostContext{
    UserID:       userID,
    DB:           dbManager,
    PendingPosts: []data_database.Post{},
    // ... inne pola
}

// Wykonanie procesu schedulera
if err := schedulerProcess.Execute(ctx); err != nil {
    // obsługa błędu
}
```

## Architektura

Aplikacja używa architektury opartej na taskach:

- **Każdy task ma jedną odpowiedzialność**
- **Taski są uniwersalne** - dostosowują się do typu operacji
- **Procesy składają się z sekwencji tasków**
- **Łatwe testowanie** - każdy task można testować osobno
- **Łatwe rozszerzanie** - nowe taski można dodawać do procesów

## Statusy postów

- **`pending`** - post oczekuje na publikację
- **`published`** - post został opublikowany pomyślnie
- **`failed`** - wystąpił błąd podczas publikacji 