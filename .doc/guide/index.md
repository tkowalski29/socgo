# User Guide

Przewodnik użytkownika SocGo - platformy do zarządzania treściami w mediach społecznościowych.

## 🚀 Pierwsze kroki

### 1. Dostęp do aplikacji
Po uruchomieniu SocGo, otwórz przeglądarkę i przejdź do:
```
http://localhost:8080
```

### 2. Główne funkcje
- **📝 Tworzenie postów** - Pisanie i publikowanie treści
- **📅 Planowanie** - Zaplanowanie publikacji na określony czas
- **🔗 Połączenia** - Konfiguracja kont mediów społecznościowych
- **📊 Historia** - Przeglądanie opublikowanych i zaplanowanych postów
- **🔔 Powiadomienia** - Monitorowanie statusu publikacji

## 📱 Obsługiwane platformy

### Facebook
- Publikowanie postów tekstowych
- Załączanie obrazów
- Planowanie publikacji
- Zarządzanie stronami Facebook

### Instagram
- Posty z obrazami
- Stories (w przygotowaniu)
- Planowanie treści
- Instagram Business accounts

### TikTok
- Przesyłanie wideo
- Opisy do filmów
- Planowanie publikacji
- TikTok for Business

## 🎯 Szybki start

### Krok 1: Połącz pierwsze konto
1. Kliknij **"Settings"** w menu nawigacji
2. Wybierz platformę (Facebook/Instagram/TikTok)
3. Kliknij **"Connect"**
4. Zaloguj się i autoryzuj aplikację
5. Nadaj nazwę połączeniu

### Krok 2: Utwórz pierwszy post
1. Przejdź do strony głównej
2. Wypełnij pole **"What's on your mind?"**
3. Wybierz dostawców (można wybrać kilku)
4. Opcjonalnie: załącz pliki
5. Wybierz **"Publish now"** lub ustaw czas publikacji
6. Kliknij **"Publish Post"**

### Krok 3: Sprawdź kalendarz
1. Kliknij **"Calendar"** w menu
2. Zobacz zaplanowane posty w widoku miesięcznym
3. Przełącz na **"Week View"** dla szczegółów godzinowych
4. Kliknij na post aby zobaczyć szczegóły

## 📚 Szczegółowe przewodniki

### [Managing Posts](./posts.md)
Tworzenie, edytowanie i zarządzanie postami.

### [Scheduling Content](./scheduling.md) 
Planowanie publikacji i zarządzanie kalendarzem.

### [Provider Setup](./providers.md)
Konfiguracja połączeń z platformami społecznościowymi.

### [Notifications](./notifications.md)
System powiadomień i monitorowanie statusu.

## 💡 Wskazówki i triki

### Planowanie treści
- **Najlepsze godziny**: Publikuj gdy Twoja publiczność jest aktywna
- **Konsystentność**: Utrzymuj regularny harmonogram publikacji
- **Różnorodność**: Mieszaj typy treści (teksty, obrazy, wideo)

### Zarządzanie wieloma kontami
- Nadawaj opisowe nazwy połączeniom (`Moja firma - Facebook`, `Blog - Instagram`)
- Grupuj podobne treści dla różnych platform
- Dostosowuj treść do specyfiki każdej platformy

### Optymalizacja wydajności
- Przygotowuj treści z wyprzedzeniem
- Używaj szablonów dla powtarzających się typów postów
- Monitoruj wyniki i dostosowuj strategię

## ❓ Często zadawane pytania

### Jak anulować zaplanowany post?
1. Przejdź do **"Calendar"** lub sprawdź historię postów
2. Znajdź zaplanowany post (status "pending")
3. Kliknij na post i wybierz **"Delete"**

### Czy mogę edytować zaplanowany post?
Obecnie nie ma opcji edycji. Anuluj post i utwórz nowy z poprawionymi danymi.

### Co jeśli publikacja się nie powiedzie?
1. Sprawdź powiadomienia (🔔)
2. Sprawdź połączenie z dostawcą w Settings
3. Sprawdź czy token OAuth nie wygasł
4. Spróbuj ponownie połączyć konto

### Ile postów mogę zaplanować?
Nie ma limitów na liczbę zaplanowanych postów.

### Czy dane są bezpieczne?
- Tokeny OAuth są bezpiecznie przechowywane
- Pliki są zapisywane lokalnie na serwerze
- Zalecamy regularne backupy bazy danych

## 🆘 Pomoc i wsparcie

### Problemy z połączeniem
1. Sprawdź klucze API w konfiguracji
2. Upewnij się, że aplikacja ma odpowiednie uprawnienia
3. Sprawdź czy URLs callback są poprawnie skonfigurowane

### Problemy z publikacją
1. Sprawdź format i rozmiar plików
2. Upewnij się, że treść spełnia wymagania platformy
3. Sprawdź limity API dostawcy

### Logowanie błędów
Sprawdź logi aplikacji w konsoli serwera:
```bash
go run main.go
```

## 🔧 Zaawansowane funkcje

### API Integration
SocGo udostępnia REST API - sprawdź [API Reference](/api/) dla programistów.

### Bulk Operations
W przygotowaniu: import wielu postów z pliku CSV.

### Analytics
W przygotowaniu: statystyki publikacji i engagement.

### Custom Providers
W przygotowaniu: możliwość dodawania własnych integracji.