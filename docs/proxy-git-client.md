# Repository Proxy Git Client

This document describes the Repository Proxy Git Client feature, which allows Semaphore runners to clone repositories via the Semaphore server instead of directly from git servers.

## Overview

In some environments, Semaphore runners may not have direct access to git servers due to network restrictions, firewalls, or security policies. The Repository Proxy Git Client feature addresses this by allowing runners to request repository data from the Semaphore server, which clones the repository and serves it as a compressed archive.

## How It Works

1. **Server-side**: The Semaphore server clones repositories using standard git clients (`cmd_git` or `go_git`)
2. **Archive Creation**: The server creates a tar.gz archive of the cloned repository
3. **Transfer**: The archive is sent to the runner via HTTP API
4. **Extraction**: The runner extracts the archive to the local filesystem

## Configuration

### Server Configuration

Set the git client type to `proxy_git` in your Semaphore server configuration:

**JSON Configuration:**
```json
{
  "git_client": "proxy_git"
}
```

**Environment Variable:**
```bash
SEMAPHORE_GIT_CLIENT=proxy_git
```

### Runner Configuration

No special runner configuration is required. Runners will automatically use the proxy mode when the server is configured with `git_client: "proxy_git"`.

## Supported Git Client Types

Semaphore supports three git client types:

- **`cmd_git`** (default): Uses system git binary for direct repository access
- **`go_git`**: Uses Go git library for direct repository access  
- **`proxy_git`**: Requests repositories from Semaphore server (new)

## Use Cases

The proxy git client is useful in environments where:

- Runners are deployed in isolated networks without direct git server access
- Corporate firewalls block git protocol access from runner instances
- You want to centralize git access through the Semaphore server
- Runners are ephemeral and you want to minimize their network dependencies

## Limitations

- **SSH Key Support**: Currently limited for repositories requiring SSH authentication
- **Performance**: Slight overhead due to archive creation and transfer
- **Storage**: Server temporarily stores cloned repositories during archive creation
- **Branch Operations**: Limited support for advanced git operations (checkout, etc.)

## Example Configuration

### Docker Compose Example

```yaml
version: '3'
services:
  semaphore:
    image: semaphoreui/semaphore:latest
    environment:
      - SEMAPHORE_GIT_CLIENT=proxy_git
      - SEMAPHORE_DB_DIALECT=sqlite
    volumes:
      - ./data:/etc/semaphore
    ports:
      - "3000:3000"
      
  runner:
    image: semaphoreui/runner:latest
    environment:
      - SEMAPHORE_RUNNER_REGISTRATION_TOKEN=your-token-here
    # Runner will automatically use proxy mode
```

### JSON Configuration File

```json
{
  "git_client": "proxy_git",
  "web_host": "https://your-semaphore-server.com",
  "dialect": "sqlite",
  "sqlite": {
    "host": "/etc/semaphore/database.db"
  }
}
```

## API Endpoint

The proxy git client uses the following internal API endpoint:

```
POST /api/internal/repositories/archive
```

**Request:**
```json
{
  "git_url": "https://github.com/user/repo.git",
  "git_branch": "main",
  "ssh_key_id": 123
}
```

**Response:**
```json
{
  "hash": "commit-hash",
  "message": "commit message", 
  "archive": "base64-encoded-tar.gz-data"
}
```

## Security Considerations

- Repository archives are temporarily stored on the server during processing
- Network traffic between runner and server contains repository data
- Server requires access to all git repositories that runners need
- Authentication credentials are handled server-side

## Troubleshooting

### Common Issues

1. **Repository Clone Failures**: Ensure the server has access to the git repository
2. **Archive Too Large**: Large repositories may cause memory issues during archive creation
3. **Network Timeouts**: Increase timeout settings for large repository transfers

### Logs

Monitor server logs for repository cloning and archive creation:

```bash
# Docker logs
docker logs semaphore-server

# Look for entries like:
# "Repository archive served to runner"
# "Failed to clone repository"
```

## Backward Compatibility

The proxy git client feature is fully backward compatible. Existing configurations using `cmd_git` or `go_git` will continue to work unchanged.