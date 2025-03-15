# Marzban Migration Tool

A command-line tool for migrating users from Marzban panel to Remnawave panel.

## Overview

This tool helps you migrate user accounts from a Marzban VPN panel to a Remnawave panel. It supports batch processing, selective migration of recent users, and customization of traffic reset strategies.

Key features:

- Batch processing with configurable batch size
- Migration of selected number of most recent users
- Automatic handling of existing users
- Support for environment variables
- Customizable traffic reset strategy
- Flexible status handling

### Migrated User Fields

The following user fields are migrated from Marzban to Remnawave:

| Field                | Description                                       |
| -------------------- | ------------------------------------------------- |
| Username             | User's unique identifier                          |
| Status               | User's status (can be preserved or set to ACTIVE) |
| ShortUUID            | Generated from subscription URL hash              |
| TrojanPassword       | Password for Trojan protocol                      |
| VlessUUID            | UUID for VLESS protocol                           |
| SsPassword           | Password for Shadowsocks protocol                 |
| TrafficLimitBytes    | Traffic limit in bytes                            |
| TrafficLimitStrategy | Traffic reset strategy (can be customized)        |
| ExpireAt             | Account expiration date (UTC)                     |
| Description          | User notes/description                            |

## Configuration

The tool can be configured using command-line flags or environment variables:

| Flag                   | Environment Variable | Description                             | Default  |
| ---------------------- | -------------------- | --------------------------------------- | -------- |
| `--marzban-url`        | `MARZBAN_URL`        | Source Marzban panel URL                | Required |
| `--marzban-username`   | `MARZBAN_USERNAME`   | Marzban admin username                  | Required |
| `--marzban-password`   | `MARZBAN_PASSWORD`   | Marzban admin password                  | Required |
| `--remnawave-url`      | `REMNAWAVE_URL`      | Destination Remnawave panel URL         | Required |
| `--remnawave-token`    | `REMNAWAVE_TOKEN`    | Remnawave API token                     | Required |
| `--batch-size`         | `BATCH_SIZE`         | Number of users to process in one batch | 100      |
| `--last-users`         | `LAST_USERS`         | Only migrate last N users               | 0 (all)  |
| `--preferred-strategy` | `PREFERRED_STRATEGY` | Preferred traffic reset strategy        | (empty)  |
| `--preserve-status`    | `PRESERVE_STATUS`    | Preserve user status from Marzban       | false    |

## Usage

### Basic Usage

```bash
# Migrate all users (sets all users to ACTIVE status)
./marzban-migration-tool \
    --marzban-url="http://marzban.example.com" \
    --marzban-username="admin" \
    --marzban-password="password" \
    --remnawave-url="http://remnawave.example.com" \
    --remnawave-token="your-token"
```

### Preserve User Status

```bash
# Migrate users preserving their original status
./marzban-migration-tool \
    [other flags...] \
    --preserve-status
```

### Migrate Last N Users

```bash
# Migrate only the last 50 users
./marzban-migration-tool \
    [other flags...] \
    --last-users=50
```

### Set Preferred Traffic Reset Strategy

```bash
# Migrate users with a specific reset strategy
./marzban-migration-tool \
    [other flags...] \
    --preferred-strategy=MONTH
```

Available strategy values:

- `NO_RESET` - No traffic limit reset
- `DAY` - Reset daily
- `WEEK` - Reset weekly
- `MONTH` - Reset monthly

**Note:** If not specified, the original strategy from Marzban will be used (with YEAR strategy converted to NO_RESET as Remnawave doesn't support yearly resets).

### Using Environment Variables

```bash
export MARZBAN_URL="http://marzban.example.com"
export MARZBAN_USERNAME="admin"
export MARZBAN_PASSWORD="password"
export REMNAWAVE_URL="http://remnawave.example.com"
export REMNAWAVE_TOKEN="your-token"
export BATCH_SIZE="200"
export LAST_USERS="50"
export PREFERRED_STRATEGY="MONTH"
export PRESERVE_STATUS="true"

./marzban-migration-tool
```

## Contribute

1. **Fork & Branch**: Fork this repository and create a branch for your work.
2. **Implement Changes**: Work on your feature or fix, keeping code clean and well-documented.
3. **Test**: Ensure your changes maintain or improve current functionality, adding tests for new features.
4. **Commit & PR**: Commit your changes with clear messages, then open a pull request detailing your work.
5. **Feedback**: Be prepared to engage with feedback and further refine your contribution.
