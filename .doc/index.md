# Code Generation Manager

Zaawansowany system zarządzania generowaniem kodu z AI napisany w Go.

## 🎯 Dlaczego Code Generation Manager?

Code Generation Manager to zaawansowany system, który umożliwia:

- **🚀 Automatyzację** procesów generowania kodu
- **📊 Monitorowanie** kosztów i użycia AI
- **⚙️ Zarządzanie** zadaniami i kolejkami
- **🔗 Integrację** z różnymi executorami
- **🛡️ Bezpieczeństwo** i kontrolę dostępu

## 🏗️ Architektura systemu

```mermaid
graph TB
    subgraph "Frontend"
        UI[🌐 Web Interface]
        API[🔌 API Client]
    end
    
    subgraph "Backend Core"
        Main[🎯 Main Application]
        Router[🛣️ HTTP Router]
        DI[🔧 Dependency Injection]
    end
    
    subgraph "Services"
        TaskMgr[📋 Task Manager]
        QueueSvc[🔄 Queue Service]
        ChatSvc[💬 Chat Service]
        MonitorSvc[📊 Monitoring Service]
    end
    
    subgraph "Executors"
        DockerExec[🐳 Docker Executor]
        ClaudeExec[🤖 Claude Executor]
        CustomExec[⚙️ Custom Executor]
    end
    
    subgraph "Storage"
        FileStorage[💾 File Storage]
        ConfigStorage[⚙️ Config Storage]
        LogStorage[📝 Log Storage]
    end
    
    subgraph "External"
        ClaudeAPI[🤖 Claude API]
        DockerAPI[🐳 Docker API]
        Webhooks[🔗 Webhooks]
    end
    
    UI --> API
    API --> Router
    Router --> Main
    Main --> DI
    DI --> TaskMgr
    DI --> QueueSvc
    DI --> ChatSvc
    DI --> MonitorSvc
    
    TaskMgr --> DockerExec
    TaskMgr --> ClaudeExec
    TaskMgr --> CustomExec
    
    ChatSvc --> ClaudeExec
    MonitorSvc --> ClaudeAPI
    
    DockerExec --> DockerAPI
    ClaudeExec --> ClaudeAPI
    
    TaskMgr --> FileStorage
    Main --> ConfigStorage
    Main --> LogStorage
    
    MonitorSvc --> Webhooks
```

## 🔄 Przepływ danych

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant API as 🔌 API
    participant TM as 📋 Task Manager
    participant QS as 🔄 Queue Service
    participant EX as ⚙️ Executor
    participant S as 💾 Storage
    
    U->>API: 📝 Create Task
    API->>TM: 🔄 Process Task
    TM->>QS: 📥 Add to Queue
    QS->>EX: 🚀 Execute Task
    EX->>S: 💾 Save Results
    EX->>API: ✅ Return Status
    API->>U: 📊 Task Complete
```

## 🚀 Szybki start

```bash
# 📥 Klonuj repozytorium
git clone https://github.com/your-username/code-gen-manager.git
cd code-gen-manager

# ⚙️ Skopiuj konfigurację
cp config.example.yml config.yml

# 🚀 Uruchom aplikację
go run main.go
```

## 📖 Dokumentacja

### 🏗️ Architektura i design
- **[🏗️ Przegląd architektury](./architecture/)** - Struktura systemu i wzorce projektowe
- **[🔧 Dependency Injection](./architecture/overview.md)** - Zarządzanie zależnościami

### 🔌 API i integracje
- **[🔌 API Reference](./api/)** - Pełna dokumentacja REST API
- **[📋 Endpointy](./api/endpoints.md)** - Lista wszystkich endpointów

### ⚙️ Konfiguracja i deployment
- **[⚙️ Konfiguracja](./configuration/)** - Ustawienia systemu
- **[📄 Plik konfiguracyjny](./configuration/config-file.md)** - Opis pliku `config.yml`
- **[🚀 Deployment](./deployment/)** - Wdrażanie systemu
- **[📦 Instalacja](./deployment/installation.md)** - Instrukcje instalacji

### 🛠️ Development i contributing
- **[🛠️ Development](./development/)** - Rozwój systemu
- **[🏗️ Struktura kodu](./development/code-structure.md)** - Organizacja kodu
- **[🧪 Testy](./development/testing.md)** - Pisanie i uruchamianie testów
- **[🤝 Contributing](./development/contributing.md)** - Jak współtworzyć projekt

### 🎯 Funkcjonalności
- **[🎯 Funkcjonalności](./features/)** - Przegląd wszystkich funkcji
- **[📋 Zarządzanie zadaniami](./features/task-management/)** - Tworzenie i zarządzanie zadaniami
- **[🔄 System kolejek](./features/queue-system/)** - Kolejkowanie i przetwarzanie
- **[💬 Chat z AI](./features/chat-system/)** - Interaktywny chat
- **[📊 Monitorowanie Claude](./features/claude-monitoring/)** - Statystyki użycia
- **[🔍 Monitorowanie wzorców](./features/pattern-monitoring/)** - Analiza wzorców

## 🛠️ Technologie

| Kategoria | Technologia | Opis |
|-----------|-------------|------|
| **🔧 Backend** | Go (Golang) | Główny język aplikacji |
| **🌐 Frontend** | Vue.js + Templ | Interfejs użytkownika |
| **🤖 AI** | Claude API, GPT | Modele AI |
| **🐳 Containerization** | Docker | Izolacja środowisk |
| **💾 Database** | File-based storage | Przechowywanie danych |
| **📚 Documentation** | VitePress | Dokumentacja techniczna |

## 📊 Statystyki projektu

```mermaid
pie title Struktura kodu
    "🔧 Internal Services" : 35
    "🌐 Web Interface" : 25
    "📋 Task Management" : 20
    "🔄 Queue System" : 15
    "📊 Monitoring" : 5
```

## 🤝 Contributing

Zapraszamy do współpracy! Sprawdź **[🤝 Contributing Guide](./development/contributing.md)** aby dowiedzieć się więcej.

