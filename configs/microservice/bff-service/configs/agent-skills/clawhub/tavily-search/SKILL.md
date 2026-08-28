---
name: tavily-search
description: Web search using Tavily's LLM-optimized API. Returns relevant results with content snippets, scores, and metadata.
homepage: https://tavily.com
metadata: {"clawhub":{"emoji":"🔍","requires":{"bins":["node"],"env":["TAVILY_API_KEY"]},"primaryEnv":"TAVILY_API_KEY"}}
---

# Tavily Search

Search the web and get relevant results optimized for LLM consumption.

## Authentication

Get your API key at https://tavily.com and add to your OpenClaw config:

```json
{
  "skills": {
    "entries": {
      "tavily-search": {
        "enabled": true,
        "apiKey": "tvly-YOUR_API_KEY_HERE"
      }
    }
  }
}
```

Or set the environment variable:
```bash
export TAVILY_API_KEY="tvly-YOUR_API_KEY_HERE"
```

## Quick Start

### Using the Script

```bash
node {baseDir}/scripts/search.mjs "query"
node {baseDir}/scripts/search.mjs "query" -n 10
node {baseDir}/scripts/search.mjs "query" --deep
node {baseDir}/scripts/search.mjs "query" --topic news
```

### Examples

```bash
# Basic search
node {baseDir}/scripts/search.mjs "python async patterns"

# With more results
node {baseDir}/scripts/search.mjs "React hooks tutorial" -n 10

# Advanced search
node {baseDir}/scripts/search.mjs "machine learning" --deep

# News search
node {baseDir}/scripts/search.mjs "AI news" --topic news

# Domain-filtered search
node {baseDir}/scripts/search.mjs "Python docs" --include-domains docs.python.org
```

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `-n <count>` | Number of results (1-20) | 10 |
| `--depth <mode>` | Search depth: `ultra-fast`, `fast`, `basic`, `advanced` | `basic` |
| `--topic <topic>` | Topic: `general` or `news` | `general` |
| `--time-range <range>` | Time range: `day`, `week`, `month`, `year` | - |
| `--include-domains <domains>` | Comma-separated domains to include | - |
| `--exclude-domains <domains>` | Comma-separated domains to exclude | - |
| `--raw-content` | Include full page content | false |
| `--json` | Output raw JSON | false |

## Search Depth

| Depth | Latency | Relevance | Use Case |
|-------|---------|-----------|----------|
| `ultra-fast` | Lowest | Lower | Real-time chat, autocomplete |
| `fast` | Low | Good | Need chunks but latency matters |
| `basic` | Medium | High | General-purpose, balanced |
| `advanced` | Higher | Highest | Precision matters, research |

## Tips

- **Keep queries under 400 characters** - Think search query, not prompt
- **Break complex queries into sub-queries** - Better results than one massive query
- **Use `--include-domains`** to focus on trusted sources
- **Use `--time-range`** for recent information
- **Filter by `score`** (0-1) to get highest relevance results