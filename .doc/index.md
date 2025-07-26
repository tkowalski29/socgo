# SocGo - Social Media Management Platform

SocGo to nowoczesna platforma do zarządzania treściami w mediach społecznościowych, obsługująca TikTok, Instagram i Facebook. Umożliwia planowanie, publikowanie i zarządzanie postami w różnych serwisach społecznościowych z jednego miejsca.

## 🚀 Funkcjonalności

- **Multi-Platform Support**: Obsługa TikTok, Instagram i Facebook
- **OAuth Authentication**: Bezpieczne uwierzytelnianie z dostawcami
- **Post Scheduling**: Planowanie publikacji postów
- **File Management**: Obsługa plików multimedialnych
- **Notifications**: System powiadomień o statusie postów
- **Web Interface**: Responsywny interfejs użytkownika

## 🏗️ Architektura

SocGo jest zbudowany w Go z wykorzystaniem:

- **Backend**: Go 1.24.3 z Gorilla Mux
- **Database**: SQLite z GORM ORM
- **Templates**: Templ dla server-side rendering
- **Frontend**: Vanilla JavaScript z CSS

## 📋 Struktura projektu

```
socgo/
├── internal/           # Kod źródłowy aplikacji
│   ├── data/          # Warstwa danych
│   ├── handlers/      # HTTP handlers
│   ├── service/       # Logika biznesowa
│   └── social/        # Integracje z platformami
├── web/               # Frontend (templates, static files)
├── uploads/           # Przesłane pliki
└── .doc/              # Dokumentacja
```

## 🛠️ Quick Start

1. **Klonowanie repozytorium**
   ```bash
   git clone https://github.com/tkowalski/socgo.git
   cd socgo
   ```

2. **Konfiguracja**
   ```bash
   cp config.yml.example config.yml
   # Edytuj config.yml z własnymi kluczami API
   ```

3. **Uruchomienie**
   ```bash
   go run main.go
   ```

4. **Dostęp do aplikacji**
   - Otwórz przeglądarkę na `http://localhost:8080`

## 📚 Dokumentacja

- [Architecture Overview](/architecture/) - Przegląd architektury systemu
- [API Reference](/api/) - Dokumentacja API
- [User Guide](/guide/) - Przewodnik użytkownika
- [Provider Integration](/providers/) - Integracje z platformami społecznościowymi

## 🤝 Contributing

Więcej informacji o współtworzeniu projektu znajdziesz w sekcji [Development](/development/contributing).

## 📄 License

Ten projekt jest na licencji MIT.