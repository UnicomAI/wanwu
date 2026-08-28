---
name: polymarket
description: Query Polymarket prediction markets. Check odds, find trending markets, search events, track price movements.
homepage: https://polymarket.com
user-invocable: true
disable-model-invocation: true
metadata:
  clawhub:
    emoji: "📊"
    requires:
      bins: [python3]
---
# Step-by-Step Guide
https://telegra.ph/How-Building-a-Weather-Polymarket-Bot-with-OpenClaw-Skill-and-turn-100--8000-Step-by-Step-Guide-02-28-2

# Polymarket
Query [Polymarket](https://polymarket.com) prediction markets. Check odds, find trending markets, search events, track price movements.

## Quick Start

```bash
# Trending markets
python3 {baseDir}/scripts/polymarket.py trending

# Search
python3 {baseDir}/scripts/polymarket.py search "trump"

# Biggest movers
python3 {baseDir}/scripts/polymarket.py movers

# What's resolving soon
python3 {baseDir}/scripts/polymarket.py calendar
```

---

## After Install â€” Suggested Setup

### 1. Add to Morning Briefing
Add Polymarket to your daily cron:
```
polymarket featured + polymarket movers --timeframe 24h
```

### 2. Watch Markets You Care About
```bash
# Watch with price target alert
python3 {baseDir}/scripts/polymarket.py watch add trump-2028 --alert-at 60

# Watch with change alert (Â±10% from current)
python3 {baseDir}/scripts/polymarket.py watch add bitcoin-100k --alert-change 10
```

### 3. Set Up Hourly Alerts (Cron)
```bash
# Check watchlist every hour, only notify on alerts
python3 {baseDir}/scripts/polymarket.py alerts --quiet
```

### 4. Weekly Category Digests
```bash
# Every Sunday: politics digest
python3 {baseDir}/scripts/polymarket.py digest politics
```

### 5. Paper Trade to Track Predictions
```bash
python3 {baseDir}/scripts/polymarket.py buy trump-2028 100  # $100 on Trump
python3 {baseDir}/scripts/polymarket.py portfolio           # Check P&L
```

---

## Commands

### Core

```bash
# Trending markets (by 24h volume)
python3 {baseDir}/scripts/polymarket.py trending

# Featured/high-profile markets
python3 {baseDir}/scripts/polymarket.py featured

# Search markets
python3 {baseDir}/scripts/polymarket.py search "giannis"

# Get event by slug
python3 {baseDir}/scripts/polymarket.py event trump-2028

# Browse by category
python3 {baseDir}/scripts/polymarket.py category politics
```

### Watchlist + Alerts (NEW)

```bash
# Add to watchlist
python3 {baseDir}/scripts/polymarket.py watch add trump-2028
python3 {baseDir}/scripts/polymarket.py watch add bitcoin-100k --alert-at 70
python3 {baseDir}/scripts/polymarket.py watch add fed-rate-cut --alert-change 15

# Watch specific outcome in multi-market
python3 {baseDir}/scripts/polymarket.py watch add giannis-trade --outcome warriors

# List watchlist with current prices
python3 {baseDir}/scripts/polymarket.py watch list

# Remove from watchlist
python3 {baseDir}/scripts/polymarket.py watch remove trump-2028

# Check for alerts (for cron)
python3 {baseDir}/scripts/polymarket.py alerts
python3 {baseDir}/scripts/polymarket.py alerts --quiet  # Only output if triggered
```

### Resolution Calendar (NEW)

```bash
# Markets resolving in next 7 days
python3 {baseDir}/scripts/polymarket.py calendar

# Markets resolving in next 3 days
python3 {baseDir}/scripts/polymarket.py calendar --days 3

# More results
python3 {baseDir}/scripts/polymarket.py calendar --days 14 --limit 20
```

### Momentum Scanner (NEW)

```bash
# Biggest movers (24h)
python3 {baseDir}/scripts/polymarket.py movers

# Weekly movers
python3 {baseDir}/scripts/polymarket.py movers --timeframe 1w

# Monthly movers with volume filter
python3 {baseDir}/scripts/polymarket.py movers --timeframe 1m --min-volume 50
```

### Category Digests (NEW)

```bash
# Politics digest
python3 {baseDir}/scripts/polymarket.py digest politics

# Crypto digest
python3 {baseDir}/scripts/polymarket.py digest crypto

# Sports digest
python3 {baseDir}/scripts/polymarket.py digest sports
```

Categories: `politics`, `crypto`, `sports`, `tech`, `business`

### Paper Trading (NEW)

```bash
# Buy $100 of a market
python3 {baseDir}/scripts/polymarket.py buy trump-2028 100

# Buy specific outcome
python3 {baseDir}/scripts/polymarket.py buy giannis-trade 50 --outcome warriors

# View portfolio
python3 {baseDir}/scripts/polymarket.py portfolio

# Sell position
python3 {baseDir}/scripts/polymarket.py sell trump-2028
```

Starts with $10,000 paper cash. Track your predictions without real money.

---

## Data Storage

Watchlist and portfolio stored in `~/.polymarket/`:
- `watchlist.json` â€” Watched markets and alert thresholds
- `portfolio.json` â€” Paper positions and trade history

---

## Cron Examples

### Hourly Alert Check
```
0 * * * * python3 ~/.../polymarket.py alerts --quiet
```

### Daily Morning Brief
```
0 7 * * * python3 ~/.../polymarket.py movers && python3 ~/.../polymarket.py calendar --days 1
```

### Weekly Digests
```
0 10 * * 0 python3 ~/.../polymarket.py digest politics
0 10 * * 0 python3 ~/.../polymarket.py digest crypto
```

---

## Output Features

Markets show:
- **Current odds** (Yes/No prices)
- **Price momentum** (24h/1wk/1mo changes with arrows)
- **Volume** (total + 24h activity)
- **Time remaining**
- **Bid/ask spread**

---

## API

Uses the public Gamma API (no auth required for reading):
- Base URL: `https://gamma-api.polymarket.com`
- Docs: https://docs.polymarket.com

---

## Security & Permissions

**No API key or authentication required.** This skill uses Polymarket's public Gamma API.

**What this skill does:**
- Makes HTTPS GET requests to `gamma-api.polymarket.com` (public, unauthenticated)
- Reads market data: odds, volumes, event details, price history
- Paper trading is **local simulation only** â€” stored in `~/.polymarket/` as JSON files
- No real money, no wallet, no blockchain transactions

**What this skill does NOT do:**
- Does not connect to any wallet or financial account
- Does not execute real trades or transactions
- Does not require or handle any credentials or API keys
- Does not send any personal data externally
- Cannot be invoked autonomously by the agent (`disable-model-invocation: true`)

**Data stored locally:** `~/.polymarket/watchlist.json`, `~/.polymarket/portfolio.json`

Review `scripts/polymarket.py` before first use to verify behavior.

## Note

This is read-only + paper trading. Real trading requires wallet authentication (not implemented).
