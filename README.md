# Remnawave Migration Tools

This repository contains a collection of tools for migrating users from various VPN panels to Remnawave panel.

## Available Migration Tools

### [Marzban Migration Tool](./marzban)

Migrate users from Marzban panel to Remnawave panel. Supports batch processing, selective migration of recent users, and custom traffic reset strategies.

Features:

- Migrate user credentials and settings
- Batch processing
- Selective migration of recent users
- Customizable traffic reset strategy
- Environment variables support

[Learn more about Marzban Migration →](./marzban)

## General Information

All migration tools in this repository follow these principles:

- Safe and non-destructive migration
- Configurable through CLI flags and environment variables
- Detailed logging and error handling
- Respect for existing users and data
- Clear documentation and usage examples

## Contributing

We welcome contributions for new migration tools or improvements to existing ones. If you'd like to add support for migrating from another panel:

## Contribute

1. **Fork & Branch**: Fork this repository and create a branch for your work.
2. **Create a new directory** for your panel tool (e.g., 3xui for 3X-UI migration)
3. **Follow the existing code structure** and documentation patterns
4. **Submit a pull request** with your changes
5. **Feedback**: Be prepared to engage with feedback and further refine your contribution.
