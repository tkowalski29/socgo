# Notifications API

System powiadomień SocGo dla informowania użytkowników o statusie publikacji, błędach i ważnych wydarzeniach.

## 📬 Overview

Notifications API umożliwia:
- **Pobieranie powiadomień** użytkownika
- **Oznaczanie jako przeczytane**
- **Zarządzanie preferencjami**
- **Real-time updates** (jeśli włączone)

## 🔔 Get Notifications

Pobiera listę powiadomień dla użytkownika z paginacją.

```http
GET /notifications?page=1&limit=20&unread_only=false
```

### Query Parameters

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `page` | int | Nie | Numer strony (domyślnie 1) |
| `limit` | int | Nie | Liczba powiadomień na stronę (max 100) |
| `unread_only` | bool | Nie | Tylko nieprzeczytane (domyślnie false) |
| `type` | string | Nie | Filtr typu: info/success/warning/error |

### Response

```json
{
  "notifications": [
    {
      "id": 123,
      "type": "success",
      "title": "Post Published Successfully",
      "message": "Your post has been published to Facebook: My Business Page",
      "metadata": {
        "post_id": 456,
        "provider_id": 1,
        "provider_name": "My Business Page",
        "provider_type": "facebook",
        "external_id": "123456789_987654321",
        "external_url": "https://facebook.com/123456789/posts/987654321"
      },
      "is_read": false,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 124,
      "type": "error",
      "title": "Post Publishing Failed",
      "message": "Failed to publish post to Instagram: Invalid access token",
      "metadata": {
        "post_id": 457,
        "provider_id": 2,
        "provider_name": "My Instagram",
        "provider_type": "instagram",
        "error_code": "OAuthException",
        "error_message": "The access token has expired"
      },
      "is_read": false,
      "created_at": "2024-01-15T10:25:00Z",
      "updated_at": "2024-01-15T10:25:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  },
  "summary": {
    "total_unread": 12,
    "unread_by_type": {
      "info": 2,
      "success": 5,
      "warning": 3,
      "error": 2
    }
  }
}
```

## ✅ Mark as Read

Oznacza powiadomienia jako przeczytane.

```http
PUT /notifications/read
```

### Request Body

```json
{
  "notification_ids": [123, 124, 125],
  "mark_all": false
}
```

### Parameters

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `notification_ids` | array | Nie* | Lista ID powiadomień do oznaczenia |
| `mark_all` | bool | Nie* | Oznacz wszystkie jako przeczytane |

*Wymagane jedno z `notification_ids` lub `mark_all`

### Response

```json
{
  "status": "success",
  "message": "Notifications marked as read",
  "updated_count": 3
}
```

## 🗑️ Delete Notifications

Usuwa powiadomienia.

```http
DELETE /notifications
```

### Request Body

```json
{
  "notification_ids": [123, 124],
  "delete_all": false,
  "delete_read_only": false
}
```

### Parameters

| Parametr | Typ | Wymagane | Opis |
|----------|-----|----------|------|
| `notification_ids` | array | Nie* | Lista ID powiadomień do usunięcia |
| `delete_all` | bool | Nie* | Usuń wszystkie powiadomienia |
| `delete_read_only` | bool | Nie* | Usuń tylko przeczytane |

### Response

```json
{
  "status": "success",
  "message": "Notifications deleted",
  "deleted_count": 2
}
```

## 📊 Get Notification Summary

Pobiera podsumowanie powiadomień użytkownika.

```http
GET /notifications/summary
```

### Response

```json
{
  "total_notifications": 45,
  "unread_count": 12,
  "by_type": {
    "info": {
      "total": 15,
      "unread": 2
    },
    "success": {
      "total": 20,
      "unread": 5
    },
    "warning": {
      "total": 7,
      "unread": 3
    },
    "error": {
      "total": 3,
      "unread": 2
    }
  },
  "recent_notifications": [
    {
      "id": 125,
      "type": "warning",
      "title": "Provider Token Expires Soon",
      "message": "Your Facebook access token will expire in 7 days",
      "created_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

## ⚙️ Notification Preferences

Zarządzanie preferencjami powiadomień użytkownika.

### Get Preferences

```http
GET /notifications/preferences
```

### Response

```json
{
  "email_notifications": {
    "enabled": true,
    "email": "user@example.com",
    "types": {
      "post_published": true,
      "post_failed": true,
      "token_expiring": true,
      "daily_summary": false
    }
  },
  "push_notifications": {
    "enabled": false,
    "types": {
      "post_published": false,
      "post_failed": true,
      "token_expiring": true
    }
  },
  "in_app_notifications": {
    "enabled": true,
    "auto_delete_after": "30d",
    "types": {
      "post_published": true,
      "post_failed": true,
      "token_expiring": true,
      "system_updates": true
    }
  }
}
```

### Update Preferences

```http
PUT /notifications/preferences
```

### Request Body

```json
{
  "email_notifications": {
    "enabled": true,
    "email": "user@example.com",
    "types": {
      "post_published": false,
      "post_failed": true,
      "token_expiring": true
    }
  },
  "in_app_notifications": {
    "auto_delete_after": "7d"
  }
}
```

## 🔄 Real-time Notifications

### WebSocket Connection

```javascript
// Połączenie WebSocket dla real-time powiadomień
const ws = new WebSocket('ws://localhost:8080/notifications/ws');

ws.onopen = function() {
    console.log('Connected to notifications');
    
    // Autoryzacja (jeśli wymagana)
    ws.send(JSON.stringify({
        type: 'auth',
        token: 'your-auth-token'
    }));
};

ws.onmessage = function(event) {
    const notification = JSON.parse(event.data);
    console.log('New notification:', notification);
    
    // Handle notification
    displayNotification(notification);
};

ws.onclose = function() {
    console.log('Disconnected from notifications');
    // Attempt reconnection
    setTimeout(connectToNotifications, 5000);
};
```

### WebSocket Message Format

```json
{
  "type": "notification",
  "data": {
    "id": 126,
    "type": "success",
    "title": "Post Published",
    "message": "Your post has been published to Instagram",
    "metadata": {
      "post_id": 458,
      "provider_name": "My Instagram"
    },
    "created_at": "2024-01-15T11:30:00Z"
  }
}
```

### Heartbeat/Ping

```json
{
  "type": "ping",
  "timestamp": "2024-01-15T11:30:00Z"
}
```

Response:
```json
{
  "type": "pong",
  "timestamp": "2024-01-15T11:30:00Z"
}
```

## 📱 HTML Endpoints (HTMX)

### Notification Bell Component

```http
GET /notifications/bell
```

Zwraca HTML komponent dzwonka z liczbą nieprzeczytanych powiadomień.

### Response

```html
<div class="notification-bell" hx-poll="30s">
    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
              d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9">
        </path>
    </svg>
    <span class="notification-count">12</span>
</div>
```

### Notification List Component

```http
GET /notifications/list?limit=10
```

Zwraca HTML listę powiadomień.

### Response

```html
<div class="notifications-list">
    <div class="notification-item unread success">
        <div class="notification-icon">✅</div>
        <div class="notification-content">
            <h4>Post Published Successfully</h4>
            <p>Your post has been published to Facebook: My Business Page</p>
            <span class="timestamp">2 minutes ago</span>
        </div>
        <button class="mark-read-btn" onclick="markAsRead(123)">Mark as read</button>
    </div>
    
    <div class="notification-item unread error">
        <div class="notification-icon">❌</div>
        <div class="notification-content">
            <h4>Post Publishing Failed</h4>
            <p>Failed to publish post to Instagram: Invalid access token</p>
            <span class="timestamp">5 minutes ago</span>
        </div>
        <button class="mark-read-btn" onclick="markAsRead(124)">Mark as read</button>
    </div>
</div>
```

## 📋 Notification Types

### Success Notifications

```json
{
  "type": "success",
  "triggers": [
    "post_published",
    "provider_connected",
    "token_refreshed",
    "backup_completed"
  ]
}
```

### Error Notifications

```json
{
  "type": "error",
  "triggers": [
    "post_publish_failed",
    "provider_connection_failed",
    "token_refresh_failed",
    "api_rate_limit_exceeded"
  ]
}
```

### Warning Notifications

```json
{
  "type": "warning",
  "triggers": [
    "token_expiring_soon",
    "storage_quota_warning",
    "api_quota_warning",
    "provider_deprecated"
  ]
}
```

### Info Notifications

```json
{
  "type": "info",
  "triggers": [
    "post_scheduled",
    "system_maintenance",
    "feature_update",
    "daily_summary"
  ]
}
```

## 🔧 Notification Metadata

### Post-related Notifications

```json
{
  "metadata": {
    "post_id": 456,
    "provider_id": 1,
    "provider_name": "My Business Page",
    "provider_type": "facebook",
    "external_id": "123456789_987654321",
    "external_url": "https://facebook.com/123456789/posts/987654321",
    "scheduled_at": "2024-01-15T10:30:00Z",
    "published_at": "2024-01-15T10:30:15Z"
  }
}
```

### OAuth-related Notifications

```json
{
  "metadata": {
    "provider_id": 2,
    "provider_name": "My Instagram",
    "provider_type": "instagram",
    "token_expires_at": "2024-01-22T10:30:00Z",
    "days_until_expiry": 7,
    "reconnect_url": "/oauth/instagram/connect?name=My%20Instagram"
  }
}
```

### System Notifications

```json
{
  "metadata": {
    "version": "2.1.0",
    "features": ["bulk_publishing", "analytics"],
    "maintenance_window": {
      "start": "2024-01-20T02:00:00Z",
      "end": "2024-01-20T04:00:00Z"
    },
    "changelog_url": "https://github.com/tkowalski/socgo/releases/tag/v2.1.0"
  }
}
```

## 📈 Notification Analytics

### Get Notification Stats

```http
GET /notifications/stats?period=7d
```

### Response

```json
{
  "period": "7d",
  "total_sent": 156,
  "by_type": {
    "success": 89,
    "error": 12,
    "warning": 23,
    "info": 32
  },
  "by_channel": {
    "in_app": 156,
    "email": 45,
    "push": 23
  },
  "engagement": {
    "read_rate": 0.78,
    "click_rate": 0.23,
    "delete_rate": 0.15
  },
  "daily_breakdown": [
    {
      "date": "2024-01-15",
      "total": 23,
      "by_type": {
        "success": 12,
        "error": 2,
        "warning": 4,
        "info": 5
      }
    }
  ]
}
```

## 🚨 Error Handling

### Notification Errors

| Kod | Błąd | Opis |
|-----|------|------|
| 400 | `invalid_notification_ids` | Nieprawidłowe ID powiadomień |
| 404 | `notification_not_found` | Powiadomienie nie zostało znalezione |
| 429 | `too_many_requests` | Za dużo żądań (rate limiting) |

### Przykład błędu

```json
{
  "error": "notification_not_found",
  "message": "Notification with ID 999 not found",
  "details": {
    "notification_id": 999,
    "user_id": "default_user"
  }
}
```

## 🔄 Webhook Notifications

### Webhook Configuration

```json
{
  "webhook_url": "https://yourapp.com/webhooks/socgo",
  "events": [
    "notification.created",
    "notification.read",
    "notification.deleted"
  ],
  "secret": "your-webhook-secret"
}
```

### Webhook Payload

```json
{
  "event": "notification.created",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "notification": {
      "id": 127,
      "type": "success",
      "title": "Post Published",
      "message": "Your post has been published successfully",
      "user_id": "default_user",
      "created_at": "2024-01-15T10:30:00Z"
    }
  }
}
```