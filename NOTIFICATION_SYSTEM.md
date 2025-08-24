# New Notification System

## Overview

Semaphore now includes a new, more secure notification system that uses a service approach with a common interface for all notification methods. This new system provides better security by separating tokens from URLs and using dedicated configuration fields.

## Key Features

- **Service-oriented architecture**: Common interface for all notification methods
- **Enhanced security**: Separate token and channel/URL parameters (no more tokens in URLs)
- **Backward compatibility**: Old notification methods continue to work
- **Unified configuration**: All notification settings in one place
- **Error handling**: Better error reporting and logging

## Supported Notification Methods

1. **Telegram** - Uses bot token and chat ID
2. **Slack** - Uses webhook URL
3. **Gotify** - Uses server URL and application token
4. **DingTalk** - Uses webhook URL
5. **RocketChat** - Uses webhook URL
6. **Microsoft Teams** - Uses webhook URL

## Configuration

### New Configuration Format

Add the `notifications` section to your `config.json`:

```json
{
  "notifications": {
    "telegram": {
      "enabled": true,
      "token": "your_telegram_bot_token",
      "channel": "your_telegram_chat_id"
    },
    "slack": {
      "enabled": true,
      "url": "https://hooks.slack.com/services/your/webhook/url"
    },
    "gotify": {
      "enabled": true,
      "token": "your_gotify_app_token",
      "url": "https://gotify.example.com"
    },
    "dingtalk": {
      "enabled": true,
      "url": "https://oapi.dingtalk.com/robot/send?access_token=your_token"
    },
    "rocketchat": {
      "enabled": true,
      "url": "https://rocketchat.example.com/hooks/webhook_id/webhook_token"
    },
    "microsoft_teams": {
      "enabled": true,
      "url": "https://outlook.office.com/webhook/your/webhook/url"
    }
  }
}
```

### Configuration Fields

Each notification method supports the following fields:

- `enabled` (boolean): Whether this notification method is active
- `token` (string): Authentication token (for Telegram and Gotify)
- `channel` (string): Channel/Chat ID (for Telegram)
- `url` (string): Webhook or server URL (for all other methods)

## Security Improvements

### Before (Old System)
```json
{
  "gotify_alert": true,
  "gotify_url": "https://gotify.example.com/message?token=insecure_token_in_url"
}
```

### After (New System)
```json
{
  "notifications": {
    "gotify": {
      "enabled": true,
      "token": "secure_token_separate_from_url",
      "url": "https://gotify.example.com"
    }
  }
}
```

## Backward Compatibility

The old notification system continues to work alongside the new one. You can:

1. **Use only the old system**: Keep your existing configuration
2. **Use only the new system**: Add the `notifications` section
3. **Use both systems**: Both will send notifications (not recommended for production)

### Migration Path

1. **Keep existing config** for immediate backward compatibility
2. **Add new `notifications` section** with your preferred settings
3. **Test the new system** to ensure it works as expected
4. **Remove old notification settings** once you're satisfied

## Implementation Details

### Architecture

The new system uses a service-oriented architecture with:

- **Common Interface**: `Notifier` interface that all notification methods implement
- **Service Class**: `NotificationService` that manages all notifiers
- **Configuration**: Centralized configuration in `util.NotificationConfigs`
- **Template System**: Reuses existing notification templates

### Code Structure

```
services/tasks/
├── notification_service.go     # Main service implementation
├── notification_service_test.go # Unit tests
└── alert.go                   # Integration with existing system
```

### Key Components

1. **Notifier Interface**:
   ```go
   type Notifier interface {
       Send(alert Alert, config util.NotificationConfig) error
       GetTemplateName() string
   }
   ```

2. **NotificationService**:
   - Manages all notification methods
   - Handles enabled/disabled states
   - Provides error logging

3. **Individual Notifiers**:
   - TelegramNotifier
   - SlackNotifier
   - GotifyNotifier
   - DingTalkNotifier
   - RocketChatNotifier
   - MicrosoftTeamsNotifier

## Usage Examples

### Telegram Setup

1. Create a Telegram bot via @BotFather
2. Get your bot token
3. Get your chat ID (can be user ID or group ID)
4. Configure:
   ```json
   {
     "notifications": {
       "telegram": {
         "enabled": true,
         "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
         "channel": "-1001234567890"
       }
     }
   }
   ```

### Slack Setup

1. Create a Slack webhook in your workspace
2. Copy the webhook URL
3. Configure:
   ```json
   {
     "notifications": {
       "slack": {
         "enabled": true,
         "url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
       }
     }
   }
   ```

### Gotify Setup

1. Install Gotify server
2. Create an application in Gotify
3. Get the application token
4. Configure:
   ```json
   {
     "notifications": {
       "gotify": {
         "enabled": true,
         "token": "your_gotify_app_token",
         "url": "https://gotify.yourdomain.com"
       }
     }
   }
   ```

## Testing

Run the notification tests:

```bash
go test -v ./services/tasks -run "Test.*Notifier"
```

## Troubleshooting

### Common Issues

1. **No notifications received**:
   - Check if `enabled: true` is set
   - Verify token/URL configuration
   - Check Semaphore logs for error messages

2. **Telegram not working**:
   - Ensure bot token is correct
   - Verify chat ID (use negative ID for groups)
   - Bot must be added to the group/channel

3. **Webhook URLs not working**:
   - Verify URL is accessible from Semaphore server
   - Check if webhook is still active in the service
   - Test webhook manually with curl

### Logging

The new system provides detailed error logging. Check Semaphore logs for messages like:
- `telegram notification failed: ...`
- `slack notification failed: ...`
- etc.

## Future Enhancements

Planned improvements:
- Environment variable support for notification configs
- Template customization per notification method
- Rate limiting and retry mechanisms
- Additional notification providers (Discord, Matrix, etc.)