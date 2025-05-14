# Remnawave Migration Tool

A command-line tool for migrating users from various VPN management panels to Remnawave.

## Supported Source Panels

- Marzban
- Marzneshin

## Overview

This tool helps you migrate user accounts from various VPN management panels to Remnawave. It supports batch processing, selective migration of recent users, custom traffic reset strategies, and full CLI/environment variable configuration.

## Key Features

- Batch processing with configurable batch size
- Migration of selected number of most recent users
- Automatic handling of existing users
- Support for environment variables
- Customizable traffic reset strategy
- Flexible status handling
- Support for custom headers in both source and destination panels

## Migrated User Fields

| Field                | Description                                       |
| -------------------- | ------------------------------------------------- |
| Username             | User's unique identifier                          |
| Status               | User's status (can be preserved or set to ACTIVE) |
| ShortUUID            | Generated from subscription URL hash              |
| TrojanPassword       | Password for Trojan protocol                      |
| VlessUUID            | UUID for VLESS protocol                           |
| SsPassword           | Password for Shadowsocks protocol                 |
| TrafficLimitBytes    | Traffic limit in bytes                            |
| TrafficLimitStrategy | Traffic reset strategy                            |
| ExpireAt             | Account expiration date (UTC)                     |
| Description          | User notes/description                            |

## Configuration

The tool can be configured using command-line flags or environment variables.

| Flag                 | Env Variable       | Description                                                   | Default |
| -------------------- | ------------------ | ------------------------------------------------------------- | ------- |
| --panel-type         | PANEL_TYPE         | Source panel type (marzban or marzneshin)                     | marzban |
| --panel-url          | PANEL_URL          | Source panel URL                                              |         |
| --panel-username     | PANEL_USERNAME     | Source panel admin username                                   |         |
| --panel-password     | PANEL_PASSWORD     | Source panel admin password                                   |         |
| --remnawave-url      | REMNAWAVE_URL      | Destination panel URL                                         |         |
| --remnawave-token    | REMNAWAVE_TOKEN    | Destination panel API token (used as Authorization Bearer)    |         |
| --dest-headers       | DEST_HEADERS       | Additional headers for Remnawave (e.g., X-Api-Key)            |         |
| --source-headers     | SOURCE_HEADERS     | Additional headers for source panel                           |         |
| --batch-size         | BATCH_SIZE         | Number of users to process in one batch                       | 100     |
| --last-users         | LAST_USERS         | Only migrate last N users (0 means all users)                 | 0       |
| --preferred-strategy | PREFERRED_STRATEGY | Preferred traffic reset strategy (NO_RESET, DAY, WEEK, MONTH) |         |
| --preserve-status    | PRESERVE_STATUS    | Preserve user status from source panel                        | false   |

## Usage

### Migrate All Users (sets all to ACTIVE)

```bash
./remnawave-migrate \
  --panel-type=marzban \
  --panel-url="http://marzban.example.com" \
  --panel-username="admin" \
  --panel-password="password" \
  --remnawave-url="http://remnawave.example.com" \
  --remnawave-token="your-token"
```

### Preserve User Status

```bash
./remnawave-migrate \
  [other flags...] \
  --preserve-status
```

### Migrate Only Last N Users

```bash
./remnawave-migrate \
  [other flags...] \
  --last-users=50
```

### Use a Preferred Traffic Reset Strategy

```bash
./remnawave-migrate \
  [other flags...] \
  --preferred-strategy=MONTH
```

**Available strategy values:**

- NO_RESET
- DAY
- WEEK
- MONTH

> Note: If not specified, the original strategy from Marzban will be used. YEAR is converted to NO_RESET.

## Custom Headers

You can provide additional HTTP headers for both the source and destination panels:

### Format:

```
key1:value1,key2:value2,...
```

### Example with destination headers:

```bash
--remnawave-token=eyJ... \
--dest-headers="X-Api-Key:abc123"
```

This will result in:

```
Authorization: Bearer eyJ...
X-Api-Key: abc123
```

If you define `Authorization` inside `--dest-headers`, it overrides the token-based default.

### Example with source headers:

```bash
--source-headers="X-Forwarded-For: 1.2.3.4,X-Custom: true"
```

These will be added to all source panel HTTP requests.

## Using Environment Variables

```bash
export PANEL_TYPE=marzban
export PANEL_URL=http://marzban.example.com
export PANEL_USERNAME=admin
export PANEL_PASSWORD=password
export REMNAWAVE_URL=http://remnawave.example.com
export REMNAWAVE_TOKEN=your-token
export DEST_HEADERS="X-Api-Key:abc123"
export SOURCE_HEADERS="X-Debug:true"

./remnawave-migrate
```

## Contribute

1. Fork & Branch: Fork this repository and create a branch for your work.
2. Implement: Work on your feature or fix, keeping code clean and documented.
3. Test: Ensure your changes maintain or improve current functionality.
4. Commit & PR: Commit changes with clear messages, open a pull request.
5. Feedback: Respond to review feedback and refine as needed.

---

Made with ❤️ by the Remnawave team and contributors.
