# WeChat Search Skill

Search for WeChat Official Account (微信公众号) articles using a compliant, three-layer approach that prioritizes legal search APIs and falls back to respectful web scraping when needed.

## Features
- **Compliant Design**: Prioritizes legal search APIs, respects robots.txt and rate limits
- **Three-Layer Strategy**: 
  - Primary: OpenClaw web_search (Brave Search API)
  - Fallback 1: Tavily Search API (if Brave API unavailable)
  - Fallback 2: Direct web fetch from WeChat search results
- **Recent Results**: Returns the 5 most recent articles by default (configurable)
- **Time Filtering**: Support for date range and recency filters
- **Multiple Output Formats**: Text, JSON, and markdown formats available

## Prerequisites
- **OpenClaw Web Tools**: Requires `web_search` and `web_fetch` tools to be available
- **API Keys** (optional but recommended):
  - Brave Search API Key: Set via `openclaw configure --section web`
  - Tavily API Key: Set as `TAVILY_API_KEY` environment variable

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
- Uses `web_search` tool with `site:mp.weixin.qq.com` filter
- Requires Brave Search API key
- Fastest and most reliable option

### Layer 2: Tavily Search API
- Uses Tavily API with `site:mp.weixin.qq.com` filter  
- Requires TAVILY_API_KEY environment variable
- Good alternative when Brave API is unavailable

### Layer 3: Direct Web Fetch
- Fetches results directly from WeChat search (sogou.com)
- Parses HTML to extract article information
- Slowest but always available as last resort

## Compliance & Ethics
- **Respects robots.txt**: Checks and follows robots.txt directives
- **Rate limiting**: Minimum 5-second delay between requests
- **Transparent identification**: Clear User-Agent string identifying the bot
- **Public content only**: Only accesses publicly available articles
- **No data retention**: Does not store full article content, only metadata

## Error Handling
- Automatic retry on network failures (up to 3 attempts)
- Graceful fallback between search strategies
- Clear error messages for debugging

## Future Enhancements
- RSS feed integration support
- Article content summarization
- Author/subscription management
- Enhanced filtering options

This skill is designed to be both useful and responsible, providing access to valuable WeChat Official Account content while respecting platform rules and legal requirements.