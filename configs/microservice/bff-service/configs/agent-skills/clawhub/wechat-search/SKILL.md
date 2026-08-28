---
name: wechat-search
description: Search WeChat Official Account articles using OpenClaw's web search, Tavily API, and web fetch capabilities with compliance-focused design.
---

# WeChat Search Skill

Search for WeChat Official Account (微信公众号) articles using a compliant, three-layer approach that prioritizes legal search APIs and falls back to respectful web scraping when needed.

## Features
- **Compliant Design**: Prioritizes legal search APIs, respects robots.txt and rate limits
- **Three-Layer Strategy**: 
  - Primary: OpenClaw web_search (Brave Search API)
  - Secondary: Tavily Search API (if Brave unavailable)
  - Fallback: Direct page fetching from WeChat search
- **Recent Results**: Returns the 5 most recent articles by default (configurable)
- **Time Filtering**: Support for date range and recency filters
- **Multiple Output Formats**: Text, JSON, and markdown formats available

## Prerequisites
- **OpenClaw Web Tools**: Requires `web_search`, `web_fetch` tools to be available
- **API Keys** (optional but recommended):
  - Brave Search API Key (for primary search)
  - Tavily API Key (for secondary search, already configured in your environment)

## Usage

### Basic Search
```bash
wechat-search "人工智能"
```

### Advanced Options
```bash
# Return 10 results instead of default 5
wechat-search "机器学习" --max-results 10

# Search within past week
wechat-search "大模型" --past-week

# Custom date range
wechat-search "AI应用" --from 2026-01-01 --to 2026-02-01

# JSON output format
wechat-search "开源AI" --output json

# Force specific strategy
wechat-search "最新技术" --strategy tavily_only
```

## Configuration
Create `~/.openclaw/wechat-search-config.json` to customize behavior:

```json
{
  "defaultMaxResults": 5,
  "maxResultsLimit": 20,
  "requestDelayMs": 5000,
  "cacheDurationHours": 1,
  "userAgent": "OpenClaw-WeChat-Search-Bot/1.0 (+https://github.com/your-username/wechat-search-skill)"
}
```

## Search Strategy Details

### Layer 1: OpenClaw Web Search (Brave Search)
- Uses Brave Search API with `site:mp.weixin.qq.com` filter
- Fastest and most reliable when API key is configured
- Respects search engine's indexing and ranking

### Layer 2: Tavily Search API
- Activated when Brave Search is unavailable or fails
- Uses Tavily's AI-powered search with WeChat site restriction
- Provides high-quality, relevant results with good coverage

### Layer 3: Direct Web Fetch
- Final fallback when both APIs are unavailable
- Scrapes WeChat search results directly from搜狗微信搜索
- Implements proper delays and respects robots.txt
- Parses HTML to extract article metadata

## Compliance & Ethics
- **Respects robots.txt**: Checks and follows robots.txt directives
- **Rate limiting**: Minimum 5-second delay between requests
- **Transparent identification**: Clear User-Agent string identifying the bot
- **Public content only**: Only accesses publicly available articles
- **No data retention**: Does not store full article content, only metadata

## Error Handling
- Automatic retry on network failures (up to 3 attempts)
- Graceful fallback between all three search strategies
- Clear error messages for debugging
- Handles API key missing scenarios gracefully

## Future Enhancements
- RSS feed integration support
- Article content summarization
- Author/subscription management
- Enhanced filtering options

This skill is designed to be both useful and responsible, providing access to valuable WeChat Official Account content while respecting platform rules and legal requirements.