# Notification System Refactoring

This document describes the new secure, service-oriented notification system that has been implemented to replace the previous URL parameter-based approach.

## Overview

The notification system has been refactored to provide:

- **Secure token-based configurations** instead of insecure URL parameters
- **Common interface** for all notification services  
- **Project-specific settings** support
- **Backward compatibility** with existing configurations
- **Extensible architecture** for adding new notification services

## Architecture

### Core Components

1. **NotificationService Interface** - Common interface for all notification services
2. **NotificationManager** - Manages and coordinates all notification services
3. **Service Implementations** - Individual services (Telegram, Slack, Gotify, DingTalk)
4. **Backward Compatibility Layer** - Ensures existing configurations continue to work

### New Configuration Structure

```json
{
  "notifications": {
    "telegram": {
      "enabled": true,
      "token": "your_bot_token_here",
      "channel": "default_chat_id",
      "config": {
        "parse_mode": "Markdown"
      }
    },
    "slack": {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
      "config": {
        "username": "Semaphore",
        "icon_emoji": ":robot_face:"
      }
    },
    "gotify": {
      "enabled": true,
      "url": "https://your-gotify-server.com",
      "token": "your_gotify_app_token",
      "config": {
        "priority": 5
      }
    },
    "dingtalk": {
      "enabled": true,
      "url": "https://oapi.dingtalk.com/robot/send?access_token=your_token",
      "config": {
        "msgtype": "markdown"
      }
    }
  }
}
```

## Supported Notification Services

### Telegram
- **Token-based authentication** - No more URL parameters
- **Project-specific chat IDs** - Override default chat per project
- **Secure token storage** - Tokens stored in configuration, not URLs

### Slack
- **Webhook URLs** - Secure webhook configuration
- **Project-specific webhooks** - Different webhooks per project
- **Rich message formatting** - Supports Slack's attachment format

### Gotify
- **Server URL + Token** - Separate URL and token configuration
- **Priority levels** - Configurable message priority
- **Project overrides** - Per-project server and token settings

### DingTalk
- **Webhook authentication** - Token-based webhook URLs
- **Markdown support** - Rich message formatting
- **Project-specific webhooks** - Per-project configuration

## Backward Compatibility

The system maintains full backward compatibility with existing configurations:

```json
{
  "telegram_alert": true,
  "telegram_token": "legacy_bot_token",
  "telegram_chat": "legacy_chat_id",
  "slack_alert": true,
  "slack_url": "https://hooks.slack.com/services/LEGACY/WEBHOOK"
}
```

### Migration Strategy

1. **Parallel Operation** - Both old and new systems run simultaneously
2. **Graceful Fallback** - New system falls back to legacy configuration if needed
3. **No Breaking Changes** - Existing configurations continue to work unchanged
4. **Gradual Migration** - Users can migrate at their own pace

## Project-Specific Configuration

Projects can override notification settings without modifying the global configuration:

- **Telegram**: Override chat ID per project using `project.AlertChat`
- **Slack**: Project-specific webhook URLs via project config
- **Gotify**: Per-project server URLs and tokens
- **DingTalk**: Project-specific webhook URLs

## Environment Variables

The new system supports environment-based configuration:

```bash
SEMAPHORE_NOTIFICATIONS='{"telegram":{"enabled":true,"token":"bot_token","channel":"chat_id"}}'
```

## Security Improvements

### Before (Insecure)
```
https://api.telegram.org/bot{TOKEN}/sendMessage?chat_id={CHAT_ID}&text=message
```

### After (Secure)
```json
{
  "notifications": {
    "telegram": {
      "token": "secure_token_in_config",
      "channel": "chat_id_in_config"
    }
  }
}
```

## Implementation Details

### Service Interface
```go
type NotificationService interface {
    Send(ctx context.Context, notification *Notification) error
    GetName() string
    IsConfigured() bool
    SupportsProjectOverride() bool
}
```

### Notification Structure
```go
type Notification struct {
    TemplateName          string
    TaskStatus            task_logger.TaskStatus
    TaskMessage           string
    AuthorName            string
    Recipients            []string
    ProjectConfig         map[string]interface{}
    SuppressSuccessAlerts bool
}
```

## Testing

The test notification functionality has been updated to use the new system:

```go
// Send test notifications
err := notifications.SendTestNotifications(ctx, &project)
```

## Benefits

1. **Security** - Tokens no longer exposed in URLs
2. **Flexibility** - Easy to add new notification services
3. **Maintainability** - Common interface reduces code duplication
4. **Configuration** - Centralized, structured configuration
5. **Extensibility** - Project-specific overrides without code changes
6. **Compatibility** - No breaking changes for existing users

## Future Enhancements

The new architecture enables future enhancements:

- **Microsoft Teams** support (already partially implemented)
- **Discord** notifications
- **Email** service refactoring
- **Webhook** generic service
- **SMS** notifications
- **Push notifications**

## Troubleshooting

### Common Issues

1. **Service not sending** - Check if service is enabled and configured
2. **Legacy fallback** - Verify both new and legacy configurations
3. **Project overrides** - Ensure project-specific settings are correct

### Debugging

Enable debug logging to see notification attempts:
```json
{
  "log": {
    "level": "debug"
  }
}
```

## Migration Examples

### From Legacy Telegram
```json
// Old
{
  "telegram_alert": true,
  "telegram_token": "bot_token",
  "telegram_chat": "chat_id"
}

// New
{
  "notifications": {
    "telegram": {
      "enabled": true,
      "token": "bot_token",
      "channel": "chat_id"
    }
  }
}
```

### From Legacy Slack
```json
// Old
{
  "slack_alert": true,
  "slack_url": "webhook_url"
}

// New
{
  "notifications": {
    "slack": {
      "enabled": true,
      "url": "webhook_url"
    }
  }
}
```

This refactoring provides a solid foundation for secure, extensible notifications while maintaining full backward compatibility with existing configurations.