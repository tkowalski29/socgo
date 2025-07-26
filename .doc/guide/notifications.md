# Notifications

Przewodnik po systemie powiadomień i monitorowaniu statusu publikacji w SocGo.

## 🔔 System powiadomień

### Typy powiadomień
- **Success** - Pomyślna publikacja
- **Error** - Błąd podczas publikacji
- **Warning** - Ostrzeżenie (np. wygasający token)
- **Info** - Informacje systemowe

### Kanały powiadomień
- **In-app** - Powiadomienia w aplikacji
- **Email** - Powiadomienia email (w przygotowaniu)
- **Webhook** - Integracja zewnętrzna (w przygotowaniu)

## 📊 Monitorowanie statusu

### Statusy publikacji
- **Pending** - W trakcie publikacji
- **Published** - Opublikowany pomyślnie
- **Failed** - Błąd podczas publikacji
- **Scheduled** - Zaplanowany do publikacji

### Śledzenie postów
- **Real-time updates** - Aktualizacje w czasie rzeczywistym
- **History** - Historia wszystkich publikacji
- **Analytics** - Statystyki sukcesu/niepowodzenia

## 🔍 Przeglądanie powiadomień

### Panel powiadomień
1. Kliknij ikonę dzwonka 🔔 w menu
2. Zobacz listę najnowszych powiadomień
3. Kliknij na powiadomienie aby zobaczyć szczegóły

### Filtrowanie
- **All** - Wszystkie powiadomienia
- **Errors** - Tylko błędy
- **Success** - Tylko sukcesy
- **Unread** - Nieprzeczytane

### Akcje na powiadomieniach
- **Mark as read** - Oznacz jako przeczytane
- **Dismiss** - Odrzuć powiadomienie
- **View details** - Zobacz szczegóły

## ⚠️ Obsługa błędów

### Typowe błędy
- **"Invalid content"** - Nieprawidłowa treść
- **"File too large"** - Plik za duży
- **"Rate limit exceeded"** - Przekroczony limit API
- **"Token expired"** - Token wygasł
- **"Network error"** - Błąd sieci

### Rozwiązywanie błędów
1. **Sprawdź powiadomienie** - Kliknij aby zobaczyć szczegóły
2. **Sprawdź logi** - Przejrzyj logi aplikacji
3. **Sprawdź połączenie** - Upewnij się, że dostawca jest dostępny
4. **Spróbuj ponownie** - Kliknij "Retry" jeśli dostępne

## 🔄 Automatyczne powiadomienia

### Powiadomienia systemowe
- **Startup** - Aplikacja uruchomiona
- **Shutdown** - Aplikacja zatrzymana
- **Database migration** - Migracje bazy danych
- **Token refresh** - Odświeżanie tokenów

### Powiadomienia o harmonogramie
- **Scheduled post reminder** - Przypomnienie o zaplanowanym poście
- **Failed scheduled post** - Błąd zaplanowanego posta
- **Bulk operation complete** - Zakończenie operacji masowej

## 📈 Statystyki i raporty

### Dashboard powiadomień
- **Total notifications** - Łączna liczba powiadomień
- **Success rate** - Procent udanych publikacji
- **Error breakdown** - Podział błędów według typu
- **Recent activity** - Ostatnia aktywność

### Eksport danych
- **CSV export** - Eksport do CSV
- **JSON export** - Eksport do JSON
- **PDF report** - Raport PDF (w przygotowaniu)

## ⚙️ Konfiguracja powiadomień

### Ustawienia aplikacji
```yaml
notifications:
  enabled: true
  retention_days: 30
  max_notifications: 1000
  email_enabled: false
  webhook_enabled: false
```

### Personalizacja
- **Email frequency** - Częstotliwość powiadomień email
- **Notification sounds** - Dźwięki powiadomień
- **Auto-dismiss** - Automatyczne odrzucanie po czasie

## 🎯 Najlepsze praktyki

### Monitorowanie
- Regularnie sprawdzaj powiadomienia
- Reaguj na błędy natychmiast
- Analizuj wzorce błędów

### Proaktywne zarządzanie
- Sprawdzaj status dostawców
- Monitoruj wygasające tokeny
- Sprawdzaj limity API

### Dokumentacja
- Dokumentuj rozwiązania błędów
- Twórz procedury dla typowych problemów
- Szkol zespół w obsłudze powiadomień

## ❓ Często zadawane pytania

### Jak wyłączyć powiadomienia?
Przejdź do Settings > Notifications i wyłącz odpowiednie opcje.

### Czy mogę otrzymywać powiadomienia email?
Funkcja powiadomień email jest w przygotowaniu.

### Jak długo są przechowywane powiadomienia?
Domyślnie 30 dni. Można to zmienić w konfiguracji.

### Czy mogę eksportować powiadomienia?
Tak, możesz eksportować do CSV lub JSON.

## 🔧 Rozwiązywanie problemów

### Powiadomienia nie działają
1. Sprawdź czy są włączone w ustawieniach
2. Sprawdź logi aplikacji
3. Sprawdź połączenie z bazą danych

### Błędy w powiadomieniach
- Sprawdź format danych
- Sprawdź uprawnienia do bazy danych
- Sprawdź konfigurację aplikacji

### Problemy z wydajnością
- Ogranicz liczbę powiadomień
- Włącz automatyczne czyszczenie
- Zoptymalizuj zapytania do bazy danych

## 🚀 Zaawansowane funkcje

### Webhook Integration
- Integracja z Slack
- Integracja z Discord
- Custom webhook endpoints

### Advanced Filtering
- Filtrowanie według dostawcy
- Filtrowanie według typu błędu
- Filtrowanie według daty

### Notification Templates
- Custom templates dla różnych typów
- Personalizowane wiadomości
- Multi-language support

### Analytics Integration
- Integracja z Google Analytics
- Custom event tracking
- Performance monitoring 