# Configuration Guide

## Configuration Files

Gordon Watcher supports multiple configuration sources (in order of priority):

1. Environment variables
2. Config file (YAML)
3. Default values

## Example Configuration

See `configs/config.example.yaml` for a complete example.

## Environment Variables

All configuration options can be set via environment variables with the prefix `GORDON_WATCHER_`.

Example:
```bash
GORDON_WATCHER_MAX_WORKERS=20
GORDON_WATCHER_LOG_LEVEL=debug
```

## Important Settings

### Worker Pool
- `max_workers`: Number of concurrent file processors (default: 10)
- `max_files_per_second`: Rate limit (default: 100)

### File Matching
- `file_patterns`: Files to process (e.g., ["*.xml", "*.zip"])
- `exclude_patterns`: Files to ignore (e.g., [".*", "*.tmp"])
