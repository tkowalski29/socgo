# Settings API

API do zarządzania ustawieniami aplikacji, konfiguracją dostawców i preferencjami użytkownika.

## ⚙️ Overview

Settings API umożliwia:
- **Zarządzanie dostawcami** (providers)
- **Konfigurację OAuth**
- **Ustawienia aplikacji**
- **Preferencje użytkownika**

## 🏠 Settings Page

Główna strona ustawień z interfejsem użytkownika.

```http
GET /settings
```

### Query Parameters

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `flash` | string | Nie | Wiadomość flash do wyświetlenia |
| `flash_type` | string | Nie | Typ wiadomości: success/error/info |

### Response

HTML strona z interfejsem ustawień zawierająca:
- Listę połączonych dostawców
- Formularz dodawania nowych dostawców
- Ustawienia aplikacji
- Preferencje powiadomień

## 👥 Provider Management

### Get Connected Providers

```http
GET /settings/providers
```

### Response JSON

```json
{
  "providers": [
    {
      "id": 1,
      "name": "My Business Page",
      "type": "facebook",
      "is_active": true,
      "connected_at": "2024-01-10T15:30:00Z",
      "last_used": "2024-01-15T10:30:00Z",
      "token_expires_at": "2024-03-15T15:30:00Z",
      "posts_count": 25,
      "status": "connected",
      "capabilities": [
        "text_posts",
        "image_posts",
        "video_posts",
        "scheduling"
      ]
    },
    {
      "id": 2,
      "name": "My Instagram",
      "type": "instagram",
      "is_active": true,
      "connected_at": "2024-01-12T09:15:00Z",
      "last_used": "2024-01-15T08:45:00Z",
      "token_expires_at": "2024-03-12T09:15:00Z",
      "posts_count": 18,
      "status": "connected",
      "capabilities": [
        "image_posts",
        "video_posts",
        "scheduling"
      ]
    }
  ],
  "summary": {
    "total_providers": 2,
    "active_providers": 2,
    "total_posts": 43
  }
}
```

### Response HTML (HTMX)

```html
<div class="providers-list">
    <div class="provider-card connected facebook">
        <div class="provider-header">
            <div class="provider-logo facebook-logo"></div>
            <div class="provider-info">
                <h3>My Business Page</h3>
                <p class="provider-type">Facebook Page</p>
                <span class="status-badge connected">Connected</span>
            </div>
        </div>
        <div class="provider-stats">
            <div class="stat">
                <span class="value">25</span>
                <span class="label">Posts</span>
            </div>
            <div class="stat">
                <span class="value">7d</span>
                <span class="label">Last used</span>
            </div>
        </div>
        <div class="provider-actions">
            <button onclick="reconnectProvider(1)" class="btn-secondary">Reconnect</button>
            <button onclick="disconnectProvider(1)" class="btn-danger">Disconnect</button>
        </div>
    </div>
</div>
```

### Add New Provider

```http
POST /settings/providers
```

### Request Body

```json
{
  "type": "facebook",
  "name": "My Second Page"
}
```

### Response

```json
{
  "status": "success",
  "message": "Provider configuration started",
  "oauth_url": "https://www.facebook.com/v18.0/dialog/oauth?client_id=...",
  "provider": {
    "id": 3,
    "name": "My Second Page",
    "type": "facebook",
    "status": "pending_auth"
  }
}
```

### Update Provider

```http
PUT /settings/providers/{id}
```

### Request Body

```json
{
  "name": "Updated Provider Name",
  "is_active": true,
  "settings": {
    "auto_publish": true,
    "default_privacy": "public",
    "notification_preferences": {
      "success": true,
      "errors": true
    }
  }
}
```

### Delete Provider

```http
DELETE /settings/providers/{id}
```

### Response

```json
{
  "status": "success",
  "message": "Provider disconnected successfully",
  "deleted_provider": {
    "id": 3,
    "name": "My Second Page",
    "type": "facebook"
  }
}
```

## 🔑 Provider Configuration

### Get Provider Settings

```http
GET /settings/providers/{id}/config
```

### Response

```json
{
  "provider": {
    "id": 1,
    "name": "My Business Page",
    "type": "facebook"
  },
  "oauth_config": {
    "client_id": "your_app_id",
    "scopes": ["pages_manage_posts", "pages_read_engagement"],
    "redirect_url": "http://localhost:8080/oauth/facebook/callback"
  },
  "provider_settings": {
    "auto_publish": true,
    "default_privacy": "public",
    "image_quality": 95,
    "video_quality": "high",
    "max_retries": 3,
    "notification_preferences": {
      "success": true,
      "errors": true,
      "warnings": false
    }
  },
  "rate_limits": {
    "posts_per_hour": 25,
    "posts_per_day": 200,
    "current_usage": {
      "hour": 5,
      "day": 23
    }
  }
}
```

### Update Provider Settings

```http
PUT /settings/providers/{id}/config
```

### Request Body

```json
{
  "provider_settings": {
    "auto_publish": false,
    "default_privacy": "friends",
    "image_quality": 90,
    "notification_preferences": {
      "success": false,
      "errors": true,
      "warnings": true
    }
  }
}
```

## 🔧 Application Settings

### Get App Settings

```http
GET /settings/application
```

### Response

```json
{
  "general": {
    "app_name": "SocGo",
    "app_version": "1.0.0",
    "timezone": "UTC",
    "date_format": "YYYY-MM-DD",
    "time_format": "24h"
  },
  "posting": {
    "default_schedule_time": "09:00",
    "auto_save_drafts": true,
    "max_file_size": 10485760,
    "allowed_file_types": ["jpg", "png", "gif", "mp4", "mov"],
    "image_compression": {
      "enabled": true,
      "quality": 95,
      "max_width": 1920,
      "max_height": 1920
    }
  },
  "scheduling": {
    "enabled": true,
    "advance_notice": "5m",
    "retry_failed": true,
    "max_retries": 3,
    "retry_delay": "1h"
  },
  "notifications": {
    "enabled": true,
    "email_notifications": false,
    "push_notifications": false,
    "cleanup_after": "30d"
  },
  "security": {
    "session_timeout": "24h",
    "auto_logout": true,
    "require_2fa": false
  }
}
```

### Update App Settings

```http
PUT /settings/application
```

### Request Body

```json
{
  "general": {
    "timezone": "Europe/Warsaw",
    "date_format": "DD/MM/YYYY"
  },
  "posting": {
    "default_schedule_time": "10:00",
    "image_compression": {
      "quality": 90
    }
  },
  "notifications": {
    "cleanup_after": "7d"
  }
}
```

## 👤 User Preferences

### Get User Preferences

```http
GET /settings/user
```

### Response

```json
{
  "user_id": "default_user",
  "profile": {
    "display_name": "User",
    "email": "user@example.com",
    "avatar_url": "/uploads/avatars/default.png"
  },
  "preferences": {
    "language": "en",
    "theme": "light",
    "dashboard": {
      "show_stats": true,
      "show_recent_posts": true,
      "posts_per_page": 20
    },
    "calendar": {
      "default_view": "week",
      "start_day": "monday",
      "show_weekends": true
    },
    "editor": {
      "auto_save": true,
      "spell_check": true,
      "word_count": true
    }
  },
  "privacy": {
    "analytics": true,
    "crash_reports": true,
    "usage_stats": false
  }
}
```

### Update User Preferences

```http
PUT /settings/user
```

### Request Body

```json
{
  "preferences": {
    "theme": "dark",
    "language": "pl",
    "dashboard": {
      "posts_per_page": 50
    },
    "calendar": {
      "default_view": "month"
    }
  },
  "privacy": {
    "analytics": false
  }
}
```

## 🔄 OAuth Management

### Refresh Provider Token

```http
POST /settings/providers/{id}/refresh-token
```

### Response

```json
{
  "status": "success",
  "message": "Token refreshed successfully",
  "token_info": {
    "expires_at": "2024-04-15T15:30:00Z",
    "scopes": ["pages_manage_posts", "pages_read_engagement"],
    "valid": true
  }
}
```

### Test Provider Connection

```http
POST /settings/providers/{id}/test
```

### Response

```json
{
  "status": "success",
  "message": "Connection test successful",
  "test_results": {
    "token_valid": true,
    "api_accessible": true,
    "permissions_valid": true,
    "rate_limit_ok": true,
    "response_time": 245
  },
  "provider_info": {
    "name": "My Business Page",
    "id": "123456789",
    "followers": 1542,
    "posts_count": 89
  }
}
```

### Reconnect Provider

```http
POST /settings/providers/{id}/reconnect
```

### Response

```json
{
  "status": "success",
  "message": "Reconnection initiated",
  "oauth_url": "https://www.facebook.com/v18.0/dialog/oauth?client_id=...",
  "expires_in": 300
}
```

## 📊 Settings Analytics

### Get Usage Statistics

```http
GET /settings/analytics?period=30d
```

### Response

```json
{
  "period": "30d",
  "summary": {
    "total_posts": 156,
    "successful_posts": 148,
    "failed_posts": 8,
    "success_rate": 0.948
  },
  "by_provider": {
    "facebook": {
      "posts": 89,
      "success_rate": 0.966,
      "avg_engagement": 45.2
    },
    "instagram": {
      "posts": 67,
      "success_rate": 0.925,
      "avg_engagement": 78.5
    }
  },
  "posting_patterns": {
    "best_hours": [9, 10, 14, 18],
    "best_days": ["monday", "wednesday", "friday"],
    "avg_posts_per_day": 5.2
  },
  "errors": {
    "token_expired": 3,
    "rate_limit": 2,
    "invalid_content": 2,
    "network_error": 1
  }
}
```

## 🔧 System Information

### Get System Info

```http
GET /settings/system
```

### Response

```json
{
  "application": {
    "name": "SocGo",
    "version": "1.0.0",
    "build": "2024-01-15-abc123",
    "go_version": "1.21.5",
    "uptime": "168h30m45s"
  },
  "database": {
    "type": "sqlite",
    "version": "3.44.0",
    "size": "15.2 MB",
    "tables": 6,
    "records": {
      "posts": 156,
      "providers": 3,
      "notifications": 89,
      "scheduled_jobs": 12
    }
  },
  "storage": {
    "uploads_path": "/app/uploads",
    "total_files": 234,
    "total_size": "156.7 MB",
    "available_space": "8.9 GB"
  },
  "providers": {
    "facebook": {
      "status": "available",
      "api_version": "v18.0",
      "last_check": "2024-01-15T10:30:00Z"
    },
    "instagram": {
      "status": "available",
      "api_version": "v18.0",
      "last_check": "2024-01-15T10:30:00Z"
    },
    "tiktok": {
      "status": "available",
      "api_version": "v2",
      "last_check": "2024-01-15T10:30:00Z"
    }
  }
}
```

## 🔒 Security Settings

### Get Security Settings

```http
GET /settings/security
```

### Response

```json
{
  "authentication": {
    "session_timeout": "24h",
    "auto_logout_enabled": true,
    "max_failed_attempts": 5,
    "lockout_duration": "15m"
  },
  "oauth_security": {
    "token_encryption": true,
    "auto_refresh_tokens": true,
    "revoke_on_disconnect": true,
    "secure_storage": true
  },
  "data_protection": {
    "encrypt_sensitive_data": true,
    "backup_encryption": true,
    "gdpr_compliance": true,
    "data_retention": "1y"
  },
  "api_security": {
    "rate_limiting": true,
    "requests_per_minute": 60,
    "csrf_protection": true,
    "secure_headers": true
  }
}
```

### Update Security Settings

```http
PUT /settings/security
```

### Request Body

```json
{
  "authentication": {
    "session_timeout": "12h",
    "max_failed_attempts": 3
  },
  "api_security": {
    "requests_per_minute": 30
  }
}
```

## 📤 Export/Import Settings

### Export Settings

```http
GET /settings/export
```

### Response

```json
{
  "export_date": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "settings": {
    "application": { /* app settings */ },
    "user_preferences": { /* user prefs */ },
    "providers": [
      {
        "name": "My Business Page",
        "type": "facebook",
        "settings": { /* provider settings */ }
        // Note: OAuth tokens are NOT exported for security
      }
    ]
  }
}
```

### Import Settings

```http
POST /settings/import
```

### Request Body

```json
{
  "settings": {
    "application": { /* app settings */ },
    "user_preferences": { /* user prefs */ }
  },
  "merge_strategy": "overwrite"  // overwrite, merge, skip_existing
}
```

### Response

```json
{
  "status": "success",
  "message": "Settings imported successfully",
  "imported": {
    "application": true,
    "user_preferences": true,
    "providers": false  // Requires manual OAuth setup
  },
  "warnings": [
    "Provider OAuth tokens need to be reconfigured manually"
  ]
}
```

## 🔄 Backup & Restore

### Create Backup

```http
POST /settings/backup
```

### Response

```json
{
  "status": "success",
  "message": "Backup created successfully",
  "backup": {
    "id": "backup_20240115_103000",
    "created_at": "2024-01-15T10:30:00Z",
    "size": "2.3 MB",
    "includes": [
      "database",
      "uploads",
      "configuration"
    ]
  }
}
```

### List Backups

```http
GET /settings/backups
```

### Response

```json
{
  "backups": [
    {
      "id": "backup_20240115_103000",
      "created_at": "2024-01-15T10:30:00Z",
      "size": "2.3 MB",
      "type": "manual"
    },
    {
      "id": "backup_20240114_020000",
      "created_at": "2024-01-14T02:00:00Z",
      "size": "2.1 MB",
      "type": "automatic"
    }
  ]
}
```

### Restore Backup

```http
POST /settings/restore/{backup_id}
```

### Response

```json
{
  "status": "success",
  "message": "Backup restored successfully",
  "restored_at": "2024-01-15T10:35:00Z",
  "requires_restart": true
}
```

## 🚨 Error Handling

### Settings Errors

| Kod | Błąd | Opis |
|-----|------|------|
| 400 | `invalid_provider_type` | Nieobsługiwany typ dostawcy |
| 400 | `invalid_settings_format` | Nieprawidłowy format ustawień |
| 401 | `unauthorized` | Brak autoryzacji |
| 404 | `provider_not_found` | Dostawca nie został znaleziony |
| 409 | `provider_already_exists` | Dostawca już istnieje |
| 422 | `validation_failed` | Walidacja danych nie powiodła się |

### Przykład błędu

```json
{
  "error": "validation_failed",
  "message": "Invalid configuration provided",
  "details": {
    "field": "image_quality",
    "value": 150,
    "constraint": "must be between 1 and 100"
  }
}
```