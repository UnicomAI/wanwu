---
name: find-skills
description: "Search and discover OpenClaw skills from various sources. Use when: user wants to find available skills, search for specific functionality, or discover new skills to install."
homepage: https://clawhub.com
metadata: { "openclaw": { "emoji": "🔍", "requires": { "bins": [] } } }
---

# Find Skills Skill

Search and discover OpenClaw skills from various sources.

## When to Use

✅ **USE this skill when:**

- "Find skills for [task]"
- "Search for OpenClaw skills"
- "What skills are available?"
- "Discover new skills"
- "Find skills by category"

## When NOT to Use

❌ **DON'T use this skill when:**

- Installing skills → use `clawhub install`
- Managing installed skills → use `openclaw skills list`
- Creating new skills → use skill-creator skill

## Sources for Finding Skills

### 1. ClawHub (Primary)
```bash
# Search skills
npx clawhub search "keyword"

# Browse categories
npx clawhub browse
```

### 2. OpenClaw Directory
- Website: https://www.openclawdirectory.dev/skills
- Browse by category, popularity, or search

### 3. LobeHub Skills Marketplace
- Website: https://lobehub.com/skills
- Community-contributed skills

### 4. GitHub
- Search: `openclaw skill` or `agent-skill`
- Look for repositories with `SKILL.md` files

### 5. Community Forums
- SitePoint: https://www.sitepoint.com/community/
- Discord: https://discord.com/invite/clawd

## Search Strategies

### By Functionality
```bash
# Web search skills
npx clawhub search "web search"

# Weather skills
npx clawhub search "weather"

# Document skills
npx clawhub search "document"
```

### By Provider
```bash
# Tavily skills
npx clawhub search "tavily"

# GitHub skills
npx clawhub search "github"

# Calendar skills
npx clawhub search "calendar"
```

### By Popularity
```bash
# Most installed skills
npx clawhub search --sort installs

# Most starred skills
npx clawhub search --sort stars
```

## Installation Tips

1. **Check requirements** before installing
2. **Read SKILL.md** for usage instructions
3. **Test in isolation** before production use
4. **Check for updates** regularly

## Common Skill Categories

### Core Skills
- `weather` - Weather forecasts
- `skill-creator` - Create new skills
- `healthcheck` - Security audits

### Integration Skills
- `github` - GitHub operations
- `feishu` - Feishu integration
- `notion` - Notion API

### Search Skills
- `tavily-search` - Web search via Tavily
- `web-search-plus` - Enhanced web search

### Agent Skills
- `proactive-agent` - Proactive automation
- `coding-agent` - Code generation

## Troubleshooting

### Rate Limits
If you hit rate limits with clawhub:
1. Wait 1 hour before retrying
2. Use alternative sources (websites)
3. Search manually on GitHub

### Installation Issues
1. Check skill requirements
2. Verify network connectivity
3. Check OpenClaw version compatibility

## Best Practices

1. **Search before creating** - Don't reinvent the wheel
2. **Read documentation** - Understand skill capabilities
3. **Start simple** - Install one skill at a time
4. **Test thoroughly** - Verify skill works as expected
5. **Provide feedback** - Help improve skills

## Related Skills

- `clawhub` - ClawHub CLI tool
- `skill-creator` - Create new skills
- `healthcheck` - System health checks