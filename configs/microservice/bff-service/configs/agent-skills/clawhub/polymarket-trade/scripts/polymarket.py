#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "requests>=2.28.0",
# ]
# ///
"""
Polymarket prediction market data.

Enhanced with:
- Watchlist + Alerts
- Resolution Calendar
- Momentum Scanner
- Category Digests
- Paper Trading Portfolio
"""

import argparse
import json
import os
import re
import sys
from datetime import datetime, timezone, timedelta
from pathlib import Path
from urllib.parse import urlparse

import requests

BASE_URL = "https://gamma-api.polymarket.com"
DATA_DIR = Path.home() / ".polymarket"


def ensure_data_dir():
    """Ensure data directory exists."""
    DATA_DIR.mkdir(parents=True, exist_ok=True)


def load_json(filename: str, default=None):
    """Load JSON file from data dir."""
    path = DATA_DIR / filename
    if path.exists():
        try:
            return json.loads(path.read_text())
        except:
            pass
    return default if default is not None else {}


def save_json(filename: str, data):
    """Save JSON file to data dir."""
    ensure_data_dir()
    path = DATA_DIR / filename
    path.write_text(json.dumps(data, indent=2, default=str))


def fetch(endpoint: str, params: dict = None) -> dict:
    """Fetch from Gamma API."""
    url = f"{BASE_URL}{endpoint}"
    resp = requests.get(url, params=params, timeout=30)
    resp.raise_for_status()
    return resp.json()


def format_price(price) -> str:
    """Format price as percentage."""
    if price is None:
        return "N/A"
    try:
        pct = float(price) * 100
        return f"{pct:.1f}%"
    except:
        return str(price)


def format_volume(volume) -> str:
    """Format volume in human readable form."""
    if volume is None:
        return "N/A"
    try:
        v = float(volume)
        if v >= 1_000_000:
            return f"${v/1_000_000:.1f}M"
        elif v >= 1_000:
            return f"${v/1_000:.1f}K"
        else:
            return f"${v:.0f}"
    except:
        return str(volume)


def format_change(change) -> str:
    """Format price change with arrow."""
    if change is None:
        return ""
    try:
        c = float(change) * 100
        if c > 0:
            return f"↑{c:.1f}%"
        elif c < 0:
            return f"↓{abs(c):.1f}%"
        else:
            return "→0%"
    except:
        return ""


def format_time_remaining(end_date: str) -> str:
    """Format time remaining until end date."""
    if not end_date:
        return ""
    try:
        dt = datetime.fromisoformat(end_date.replace('Z', '+00:00'))
        now = datetime.now(timezone.utc)
        delta = dt - now
        
        if delta.days < 0:
            return "Ended"
        elif delta.days == 0:
            hours = delta.seconds // 3600
            if hours == 0:
                mins = delta.seconds // 60
                return f"Ends in {mins}m"
            return f"Ends in {hours}h"
        elif delta.days == 1:
            return "Ends tomorrow"
        elif delta.days < 7:
            return f"Ends in {delta.days}d"
        elif delta.days < 30:
            weeks = delta.days // 7
            return f"Ends in {weeks}w"
        else:
            return dt.strftime('%b %d, %Y')
    except:
        return ""


def extract_slug_from_url(url_or_slug: str) -> str:
    """Extract slug from Polymarket URL or return as-is if already a slug."""
    if 'polymarket.com' in url_or_slug:
        parsed = urlparse(url_or_slug)
        path = parsed.path.strip('/')
        if path.startswith('event/'):
            return path.replace('event/', '')
        return path
    return url_or_slug


def get_market_price(market: dict) -> float:
    """Get current Yes price from market."""
    prices = market.get('outcomePrices')
    if prices:
        if isinstance(prices, str):
            try:
                prices = json.loads(prices)
            except:
                return 0
        if prices and len(prices) >= 1:
            try:
                return float(prices[0])
            except:
                pass
    return 0


def format_market(market: dict, verbose: bool = False) -> str:
    """Format a single market for display."""
    lines = []
    
    question = market.get('question') or market.get('title', 'Unknown')
    lines.append(f"📊 **{question}**")
    
    prices = market.get('outcomePrices')
    if prices:
        if isinstance(prices, str):
            try:
                prices = json.loads(prices)
            except:
                prices = None
        
        if prices and len(prices) >= 2:
            yes_price = format_price(prices[0])
            no_price = format_price(prices[1])
            
            day_change = format_change(market.get('oneDayPriceChange'))
            change_str = f" ({day_change})" if day_change else ""
            
            lines.append(f"   Yes: {yes_price}{change_str} | No: {no_price}")
    
    bid = market.get('bestBid')
    ask = market.get('bestAsk')
    if bid is not None and ask is not None:
        spread = float(ask) - float(bid)
        if spread > 0:
            lines.append(f"   Spread: {spread*100:.1f}% (Bid: {format_price(bid)} / Ask: {format_price(ask)})")
    
    volume = market.get('volume') or market.get('volumeNum')
    if volume:
        vol_str = f"   Volume: {format_volume(volume)}"
        vol_24h = market.get('volume24hr')
        if vol_24h and float(vol_24h) > 0:
            vol_str += f" (24h: {format_volume(vol_24h)})"
        lines.append(vol_str)
    
    end_date = market.get('endDate') or market.get('endDateIso')
    time_left = format_time_remaining(end_date)
    if time_left:
        lines.append(f"   ⏰ {time_left}")
    
    if verbose:
        week_change = format_change(market.get('oneWeekPriceChange'))
        month_change = format_change(market.get('oneMonthPriceChange'))
        if week_change or month_change:
            lines.append(f"   📈 1w: {week_change or 'N/A'} | 1m: {month_change or 'N/A'}")
        
        liquidity = market.get('liquidityNum') or market.get('liquidity')
        if liquidity:
            lines.append(f"   💧 Liquidity: {format_volume(liquidity)}")
    
    slug = market.get('slug') or market.get('market_slug')
    if slug:
        lines.append(f"   🔗 polymarket.com/event/{slug}")
    
    return '\n'.join(lines)


def format_event(event: dict, show_all_markets: bool = False) -> str:
    """Format an event with its markets."""
    lines = []
    
    title = event.get('title', 'Unknown Event')
    lines.append(f"🎯 **{title}**")
    
    volume = event.get('volume')
    if volume:
        vol_str = f"   Volume: {format_volume(volume)}"
        vol_24h = event.get('volume24hr')
        if vol_24h and float(vol_24h) > 0:
            vol_str += f" (24h: {format_volume(vol_24h)})"
        lines.append(vol_str)
    
    end_date = event.get('endDate')
    time_left = format_time_remaining(end_date)
    if time_left:
        lines.append(f"   ⏰ {time_left}")
    
    markets = event.get('markets', [])
    if markets:
        market_prices = []
        for m in markets:
            yes_price = get_market_price(m)
            if not m.get('active', True) and m.get('volumeNum', 0) == 0:
                continue
            market_prices.append((m, yes_price))
        
        market_prices.sort(key=lambda x: x[1], reverse=True)
        
        lines.append(f"   Markets: {len(market_prices)}")
        
        display_count = len(market_prices) if show_all_markets else min(10, len(market_prices))
        for m, price in market_prices[:display_count]:
            name = m.get('groupItemTitle') or m.get('question', '')[:40]
            vol = m.get('volumeNum', 0)
            day_change = format_change(m.get('oneDayPriceChange'))
            change_str = f" {day_change}" if day_change else ""
            
            if price > 0:
                lines.append(f"   • {name}: {format_price(price)}{change_str} ({format_volume(vol)})")
            else:
                lines.append(f"   • {name}")
        
        if len(market_prices) > display_count:
            lines.append(f"   ... and {len(market_prices) - display_count} more")
    
    slug = event.get('slug')
    if slug:
        lines.append(f"   🔗 polymarket.com/event/{slug}")
    
    return '\n'.join(lines)


# ==================== ORIGINAL COMMANDS ====================

def cmd_trending(args):
    """Get trending/active markets."""
    params = {
        'order': 'volume24hr',
        'ascending': 'false',
        'closed': 'false',
        'limit': args.limit
    }
    
    data = fetch('/events', params)
    
    print(f"🔥 **Trending on Polymarket**\n")
    
    for event in data:
        print(format_event(event))
        print()


def cmd_featured(args):
    """Get featured markets."""
    params = {
        'closed': 'false',
        'featured': 'true',
        'limit': args.limit
    }
    
    data = fetch('/events', params)
    
    print(f"⭐ **Featured Markets**\n")
    
    if not data:
        params = {
            'order': 'volume',
            'ascending': 'false',
            'closed': 'false',
            'limit': args.limit
        }
        data = fetch('/events', params)
        print("(Showing highest volume markets)\n")
    
    for event in data:
        print(format_event(event))
        print()


def expand_query(query: str) -> list:
    """Expand query with synonyms and variations."""
    query = query.lower().strip()
    expansions = set([query])
    words = query.split()
    
    # Synonym mappings
    synonyms = {
        'championship': ['champion', 'winner', 'tournament', 'title', 'finals'],
        'trade': ['traded', 'next team', 'destination', 'move'],
        'win': ['winner', 'won', 'wins', 'winning'],
        'election': ['president', 'presidential', 'vote'],
        'fed': ['federal reserve', 'interest rate', 'fomc'],
        'bitcoin': ['btc', 'crypto'],
        'btc': ['bitcoin', 'crypto'],
        'ethereum': ['eth', 'crypto'],
        'eth': ['ethereum', 'crypto'],
    }
    
    sport_leagues = {
        'nba': ['basketball'], 'nfl': ['football'], 'mlb': ['baseball'],
        'nhl': ['hockey'], 'ncaa': ['college', 'tournament'],
    }
    
    for key, values in synonyms.items():
        if key in query:
            for v in values:
                expansions.add(query.replace(key, v))
                expansions.add(v)
    
    for league, sports in sport_leagues.items():
        if league in query:
            for s in sports:
                expansions.add(query.replace(league, s))
    
    if len(words) >= 2:
        for word in words:
            if len(word) >= 3:
                expansions.add(word)
    
    expansions.add(query.replace(' ', '-'))
    
    return list(expansions)


def cmd_search(args):
    """Search markets with fuzzy matching."""
    query = args.query.lower()
    queries = expand_query(query)
    
    slug_guess = query.replace(' ', '-')
    try:
        data = fetch('/events', {'slug': slug_guess, 'closed': 'false'})
        if data:
            print(f"🔍 **Found: '{args.query}'**\n")
            for event in data[:args.limit]:
                print(format_event(event, show_all_markets=args.all))
                print()
            return
    except:
        pass
    
    try:
        data = fetch('/events', {'closed': 'false', 'limit': 500})
        matches = []
        
        for event in data:
            slug = event.get('slug', '').lower()
            title = event.get('title', '').lower()
            desc = event.get('description', '').lower()
            
            found = False
            for q in queries:
                if q in slug or q in title or q in desc:
                    matches.append(event)
                    found = True
                    break
            
            if found:
                continue
            
            for m in event.get('markets', []):
                mq = m.get('question', '').lower()
                item = m.get('groupItemTitle', '').lower()
                for q in queries:
                    if q in mq or q in item:
                        matches.append(event)
                        found = True
                        break
                if found:
                    break
        
        print(f"🔍 **Search: '{args.query}'**\n")
        
        if not matches:
            print("No markets found.")
            return
        
        for event in matches[:args.limit]:
            print(format_event(event, show_all_markets=args.all))
            print()
            
    except Exception as e:
        print(f"Search error: {e}")


def cmd_event(args):
    """Get specific event by slug or URL."""
    slug = extract_slug_from_url(args.slug)
    
    try:
        data = fetch('/events', {'slug': slug})
        
        if not data:
            all_events = fetch('/events', {'closed': 'false', 'limit': 200})
            slug_lower = slug.lower()
            matches = [e for e in all_events if slug_lower in e.get('slug', '').lower()]
            
            if matches:
                data = matches
            else:
                print(f"❌ Event not found: {slug}")
                return
        
        event = data[0] if isinstance(data, list) and data else data
        print(format_event(event, show_all_markets=True))
        
    except requests.HTTPError as e:
        if e.response.status_code == 404:
            print(f"❌ Event not found: {slug}")
        else:
            raise


def cmd_market(args):
    """Get specific market outcome within an event."""
    slug = extract_slug_from_url(args.slug)
    outcome = args.outcome.lower() if args.outcome else None
    
    try:
        data = fetch('/events', {'slug': slug})
        
        if not data:
            print(f"❌ Event not found: {slug}")
            return
        
        event = data[0] if isinstance(data, list) else data
        markets = event.get('markets', [])
        
        if not outcome:
            print(f"🎯 **{event.get('title')}**\n")
            for m in markets:
                print(format_market(m, verbose=True))
                print()
            return
        
        for m in markets:
            name = m.get('groupItemTitle', '').lower()
            question = m.get('question', '').lower()
            if outcome in name or outcome in question:
                print(format_market(m, verbose=True))
                return
        
        print(f"❌ Outcome '{args.outcome}' not found")
        print(f"\nAvailable outcomes:")
        for m in markets[:15]:
            name = m.get('groupItemTitle') or m.get('question', '')[:40]
            print(f"  • {name}")
                
    except requests.HTTPError as e:
        if e.response.status_code == 404:
            print(f"❌ Event not found: {slug}")
        else:
            raise


def cmd_category(args):
    """Get markets by category."""
    categories = {
        'politics': ['politics', 'election', 'trump', 'biden', 'congress'],
        'crypto': ['crypto', 'bitcoin', 'ethereum', 'btc', 'eth'],
        'sports': ['sports', 'nba', 'nfl', 'mlb', 'soccer'],
        'tech': ['tech', 'ai', 'apple', 'google', 'microsoft'],
        'entertainment': ['entertainment', 'movie', 'oscar', 'grammy'],
        'science': ['science', 'space', 'nasa', 'climate'],
        'business': ['business', 'fed', 'interest', 'stock', 'market']
    }
    
    tags = categories.get(args.category.lower(), [args.category.lower()])
    
    data = fetch('/events', {
        'closed': 'false',
        'limit': 100,
        'order': 'volume24hr',
        'ascending': 'false'
    })
    
    matches = []
    for event in data:
        title = event.get('title', '').lower()
        event_tags = [t.get('label', '').lower() for t in event.get('tags', [])]
        
        for tag in tags:
            if tag in title or tag in ' '.join(event_tags):
                matches.append(event)
                break
    
    print(f"📁 **Category: {args.category.title()}**\n")
    
    if not matches:
        print(f"No markets found for '{args.category}'")
        return
    
    for event in matches[:args.limit]:
        print(format_event(event))
        print()


# ==================== NEW: WATCHLIST ====================

def cmd_watch(args):
    """Add/remove markets from watchlist."""
    watchlist = load_json('watchlist.json', {'markets': []})
    
    if args.action == 'add':
        slug = extract_slug_from_url(args.slug)
        
        # Fetch current price
        try:
            data = fetch('/events', {'slug': slug})
            if not data:
                print(f"❌ Event not found: {slug}")
                return
            event = data[0] if isinstance(data, list) else data
        except:
            print(f"❌ Could not fetch event: {slug}")
            return
        
        # Get price from first market or specified outcome
        price = 0
        market_name = event.get('title', slug)
        markets = event.get('markets', [])
        
        if args.outcome and markets:
            for m in markets:
                name = m.get('groupItemTitle', '').lower()
                if args.outcome.lower() in name:
                    price = get_market_price(m)
                    market_name = m.get('groupItemTitle', market_name)
                    break
        elif markets:
            price = get_market_price(markets[0])
            if len(markets) == 1:
                market_name = markets[0].get('question', market_name)
        
        entry = {
            'slug': slug,
            'outcome': args.outcome,
            'name': market_name,
            'added_at': datetime.now(timezone.utc).isoformat(),
            'added_price': price,
            'alert_at': args.alert_at / 100 if args.alert_at else None,
            'alert_change': args.alert_change / 100 if args.alert_change else None,
        }
        
        # Check if already watching
        existing = [w for w in watchlist['markets'] if w['slug'] == slug and w.get('outcome') == args.outcome]
        if existing:
            watchlist['markets'] = [w for w in watchlist['markets'] if not (w['slug'] == slug and w.get('outcome') == args.outcome)]
        
        watchlist['markets'].append(entry)
        save_json('watchlist.json', watchlist)
        
        alert_str = ""
        if args.alert_at:
            alert_str += f" (alert at {args.alert_at}%)"
        if args.alert_change:
            alert_str += f" (alert on {args.alert_change}% change)"
        
        print(f"👁️ Now watching: **{market_name}**")
        print(f"   Current: {format_price(price)}{alert_str}")
        print(f"   Slug: {slug}")
        
    elif args.action == 'remove':
        slug = extract_slug_from_url(args.slug)
        before = len(watchlist['markets'])
        watchlist['markets'] = [w for w in watchlist['markets'] if w['slug'] != slug]
        save_json('watchlist.json', watchlist)
        
        if len(watchlist['markets']) < before:
            print(f"✅ Removed {slug} from watchlist")
        else:
            print(f"❌ {slug} not in watchlist")
            
    elif args.action == 'list':
        if not watchlist['markets']:
            print("📋 Watchlist is empty")
            print("\nAdd markets with: polymarket watch add <slug>")
            return
        
        print(f"👁️ **Watchlist** ({len(watchlist['markets'])} markets)\n")
        
        for w in watchlist['markets']:
            try:
                data = fetch('/events', {'slug': w['slug']})
                if data:
                    event = data[0] if isinstance(data, list) else data
                    markets = event.get('markets', [])
                    
                    current_price = 0
                    if w.get('outcome') and markets:
                        for m in markets:
                            if w['outcome'].lower() in m.get('groupItemTitle', '').lower():
                                current_price = get_market_price(m)
                                break
                    elif markets:
                        current_price = get_market_price(markets[0])
                    
                    added_price = w.get('added_price', 0)
                    change = current_price - added_price
                    change_str = f" ({format_change(change)})" if change != 0 else ""
                    
                    print(f"• **{w['name']}**")
                    print(f"  Current: {format_price(current_price)}{change_str}")
                    if w.get('alert_at'):
                        print(f"  Alert at: {w['alert_at']*100:.0f}%")
                    if w.get('alert_change'):
                        print(f"  Alert on: ±{w['alert_change']*100:.0f}% change")
                    print()
            except Exception as e:
                print(f"• {w['name']} (error fetching: {e})")
                print()


def cmd_alerts(args):
    """Check watchlist for alerts (for cron jobs)."""
    watchlist = load_json('watchlist.json', {'markets': []})
    
    if not watchlist['markets']:
        if not args.quiet:
            print("No markets in watchlist")
        return
    
    alerts = []
    
    for w in watchlist['markets']:
        try:
            data = fetch('/events', {'slug': w['slug']})
            if not data:
                continue
            
            event = data[0] if isinstance(data, list) else data
            markets = event.get('markets', [])
            
            current_price = 0
            if w.get('outcome') and markets:
                for m in markets:
                    if w['outcome'].lower() in m.get('groupItemTitle', '').lower():
                        current_price = get_market_price(m)
                        break
            elif markets:
                current_price = get_market_price(markets[0])
            
            added_price = w.get('added_price', 0)
            change = current_price - added_price
            
            triggered = False
            reason = ""
            
            # Check alert_at threshold
            if w.get('alert_at'):
                if current_price >= w['alert_at']:
                    triggered = True
                    reason = f"reached {format_price(current_price)} (threshold: {w['alert_at']*100:.0f}%)"
            
            # Check alert_change threshold
            if w.get('alert_change') and added_price > 0:
                pct_change = abs(change) / added_price
                if pct_change >= w['alert_change']:
                    triggered = True
                    direction = "up" if change > 0 else "down"
                    reason = f"moved {direction} {format_change(change)} (threshold: ±{w['alert_change']*100:.0f}%)"
            
            if triggered:
                alerts.append({
                    'name': w['name'],
                    'slug': w['slug'],
                    'price': current_price,
                    'reason': reason,
                })
                
        except Exception as e:
            continue
    
    if alerts:
        print(f"🚨 **Polymarket Alerts** ({len(alerts)})\n")
        for a in alerts:
            print(f"• **{a['name']}**")
            print(f"  {a['reason']}")
            print(f"  🔗 polymarket.com/event/{a['slug']}")
            print()
    elif not args.quiet:
        print("✅ No alerts triggered")


# ==================== NEW: CALENDAR ====================

def cmd_calendar(args):
    """Show markets resolving soon."""
    days = args.days
    
    data = fetch('/events', {
        'closed': 'false',
        'limit': 200,
        'order': 'endDate',
        'ascending': 'true'
    })
    
    now = datetime.now(timezone.utc)
    cutoff = now + timedelta(days=days)
    
    upcoming = []
    for event in data:
        end_date = event.get('endDate')
        if not end_date:
            continue
        
        try:
            dt = datetime.fromisoformat(end_date.replace('Z', '+00:00'))
            if now <= dt <= cutoff:
                upcoming.append((dt, event))
        except:
            continue
    
    upcoming.sort(key=lambda x: x[0])
    
    print(f"📅 **Resolving in {days} days** ({len(upcoming)} markets)\n")
    
    if not upcoming:
        print("No markets resolving in this timeframe.")
        return
    
    current_date = None
    for dt, event in upcoming[:args.limit]:
        date_str = dt.strftime('%a %b %d')
        if date_str != current_date:
            current_date = date_str
            print(f"\n**{date_str}**")
        
        title = event.get('title', 'Unknown')[:60]
        vol = format_volume(event.get('volume', 0))
        time_str = dt.strftime('%I:%M %p')
        
        # Get lead outcome
        markets = event.get('markets', [])
        lead = ""
        if markets:
            sorted_markets = sorted(markets, key=lambda m: get_market_price(m), reverse=True)
            if sorted_markets:
                top = sorted_markets[0]
                top_name = top.get('groupItemTitle', 'Yes')[:20]
                top_price = get_market_price(top)
                lead = f" → {top_name} {format_price(top_price)}"
        
        print(f"  {time_str} | {title}{lead} ({vol})")


# ==================== NEW: MOVERS ====================

def cmd_movers(args):
    """Find biggest price movers."""
    timeframe = args.timeframe
    min_volume = args.min_volume * 1000 if args.min_volume else 10000
    
    data = fetch('/events', {
        'closed': 'false',
        'limit': 300,
    })
    
    movers = []
    
    for event in data:
        vol = float(event.get('volume24hr', 0) or 0)
        if vol < min_volume:
            continue
        
        markets = event.get('markets', [])
        for m in markets:
            if timeframe == '24h':
                change = m.get('oneDayPriceChange')
            elif timeframe == '1w':
                change = m.get('oneWeekPriceChange')
            elif timeframe == '1m':
                change = m.get('oneMonthPriceChange')
            else:
                change = m.get('oneDayPriceChange')
            
            if change is None:
                continue
            
            try:
                change_val = abs(float(change))
            except:
                continue
            
            if change_val > 0.01:  # At least 1% move
                movers.append({
                    'event': event.get('title', ''),
                    'market': m.get('groupItemTitle') or m.get('question', ''),
                    'change': float(change),
                    'price': get_market_price(m),
                    'volume': vol,
                    'slug': event.get('slug', ''),
                })
    
    # Sort by absolute change
    movers.sort(key=lambda x: abs(x['change']), reverse=True)
    
    print(f"📈 **Biggest Movers ({timeframe})**\n")
    
    if not movers:
        print("No significant movers found.")
        return
    
    for m in movers[:args.limit]:
        direction = "🟢" if m['change'] > 0 else "🔴"
        change_pct = m['change'] * 100
        
        name = m['market'] or m['event']
        if len(name) > 50:
            name = name[:47] + "..."
        
        print(f"{direction} **{name}**")
        print(f"   {change_pct:+.1f}% → Now {format_price(m['price'])} (Vol: {format_volume(m['volume'])})")
        print()


# ==================== NEW: DIGEST ====================

def cmd_digest(args):
    """Category digest with summary."""
    category = args.category.lower()
    
    categories = {
        'politics': ['politics', 'election', 'trump', 'biden', 'congress', 'senate'],
        'crypto': ['crypto', 'bitcoin', 'ethereum', 'btc', 'eth', 'solana'],
        'sports': ['sports', 'nba', 'nfl', 'mlb', 'soccer', 'ufc', 'ncaa'],
        'tech': ['tech', 'ai', 'apple', 'google', 'microsoft', 'openai'],
        'business': ['business', 'fed', 'interest', 'stock', 'economy', 'recession'],
    }
    
    tags = categories.get(category, [category])
    
    data = fetch('/events', {
        'closed': 'false',
        'limit': 200,
        'order': 'volume24hr',
        'ascending': 'false'
    })
    
    matches = []
    for event in data:
        title = event.get('title', '').lower()
        desc = event.get('description', '').lower()
        
        for tag in tags:
            if tag in title or tag in desc:
                matches.append(event)
                break
    
    if not matches:
        print(f"No markets found for '{category}'")
        return
    
    # Calculate stats
    total_volume = sum(float(e.get('volume', 0) or 0) for e in matches)
    total_24h = sum(float(e.get('volume24hr', 0) or 0) for e in matches)
    
    # Find biggest movers in category
    movers = []
    for event in matches:
        for m in event.get('markets', []):
            change = m.get('oneDayPriceChange')
            if change:
                try:
                    movers.append({
                        'name': m.get('groupItemTitle') or event.get('title', ''),
                        'change': float(change),
                        'price': get_market_price(m),
                    })
                except:
                    pass
    
    movers.sort(key=lambda x: abs(x['change']), reverse=True)
    
    # Find upcoming resolutions
    now = datetime.now(timezone.utc)
    week_out = now + timedelta(days=7)
    upcoming = []
    for event in matches:
        end = event.get('endDate')
        if end:
            try:
                dt = datetime.fromisoformat(end.replace('Z', '+00:00'))
                if now <= dt <= week_out:
                    upcoming.append((dt, event))
            except:
                pass
    upcoming.sort(key=lambda x: x[0])
    
    # Print digest
    print(f"📊 **{category.title()} Digest**\n")
    print(f"Markets: {len(matches)} | Volume: {format_volume(total_volume)} | 24h: {format_volume(total_24h)}")
    print()
    
    if movers:
        print("**🔥 Biggest Movers (24h)**")
        for m in movers[:5]:
            direction = "↑" if m['change'] > 0 else "↓"
            print(f"  {direction} {m['name'][:40]}: {m['change']*100:+.1f}%")
        print()
    
    if upcoming:
        print("**⏰ Resolving This Week**")
        for dt, event in upcoming[:5]:
            print(f"  {dt.strftime('%a %b %d')}: {event.get('title', '')[:40]}")
        print()
    
    print("**📈 Top by Volume**")
    for event in matches[:5]:
        print(format_event(event))
        print()


# ==================== NEW: PORTFOLIO ====================

def cmd_portfolio(args):
    """Show paper trading portfolio."""
    portfolio = load_json('portfolio.json', {'positions': [], 'history': [], 'cash': 10000})
    
    if not portfolio['positions']:
        print("📈 **Paper Portfolio**\n")
        print(f"Cash: ${portfolio['cash']:,.2f}")
        print("\nNo positions. Start with:")
        print("  polymarket buy <slug> <amount>")
        return
    
    print("📈 **Paper Portfolio**\n")
    
    total_value = portfolio['cash']
    total_cost = 0
    
    for pos in portfolio['positions']:
        try:
            data = fetch('/events', {'slug': pos['slug']})
            if data:
                event = data[0] if isinstance(data, list) else data
                markets = event.get('markets', [])
                
                current_price = 0
                if pos.get('outcome') and markets:
                    for m in markets:
                        if pos['outcome'].lower() in m.get('groupItemTitle', '').lower():
                            current_price = get_market_price(m)
                            break
                elif markets:
                    current_price = get_market_price(markets[0])
                
                shares = pos['shares']
                cost_basis = pos['cost_basis']
                current_value = shares * current_price
                pnl = current_value - cost_basis
                pnl_pct = (pnl / cost_basis * 100) if cost_basis > 0 else 0
                
                total_value += current_value
                total_cost += cost_basis
                
                direction = "🟢" if pnl >= 0 else "🔴"
                print(f"{direction} **{pos['name'][:40]}**")
                print(f"   {shares:.0f} shares @ {format_price(pos['entry_price'])} → {format_price(current_price)}")
                print(f"   Value: ${current_value:,.2f} | P&L: ${pnl:+,.2f} ({pnl_pct:+.1f}%)")
                print()
        except Exception as e:
            print(f"• {pos['name']} (error: {e})")
            print()
    
    total_pnl = total_value - 10000  # Starting cash
    print(f"**Summary**")
    print(f"Cash: ${portfolio['cash']:,.2f}")
    print(f"Positions: ${total_value - portfolio['cash']:,.2f}")
    print(f"Total: ${total_value:,.2f} (P&L: ${total_pnl:+,.2f})")


def cmd_buy(args):
    """Paper buy a position."""
    portfolio = load_json('portfolio.json', {'positions': [], 'history': [], 'cash': 10000})
    
    slug = extract_slug_from_url(args.slug)
    amount = args.amount
    
    if amount > portfolio['cash']:
        print(f"❌ Insufficient cash. Have: ${portfolio['cash']:,.2f}, Need: ${amount:,.2f}")
        return
    
    try:
        data = fetch('/events', {'slug': slug})
        if not data:
            print(f"❌ Event not found: {slug}")
            return
        
        event = data[0] if isinstance(data, list) else data
        markets = event.get('markets', [])
        
        price = 0
        market_name = event.get('title', slug)
        outcome = args.outcome
        
        if outcome and markets:
            for m in markets:
                name = m.get('groupItemTitle', '').lower()
                if outcome.lower() in name:
                    price = get_market_price(m)
                    market_name = m.get('groupItemTitle', market_name)
                    break
            if price == 0:
                print(f"❌ Outcome '{outcome}' not found")
                return
        elif markets:
            price = get_market_price(markets[0])
            if len(markets) == 1:
                market_name = markets[0].get('question', market_name)
        
        if price <= 0:
            print("❌ Could not get price")
            return
        
        shares = amount / price
        
        # Check if already have position
        existing = None
        for p in portfolio['positions']:
            if p['slug'] == slug and p.get('outcome') == outcome:
                existing = p
                break
        
        if existing:
            # Average in
            total_shares = existing['shares'] + shares
            total_cost = existing['cost_basis'] + amount
            existing['shares'] = total_shares
            existing['cost_basis'] = total_cost
            existing['entry_price'] = total_cost / total_shares
        else:
            portfolio['positions'].append({
                'slug': slug,
                'outcome': outcome,
                'name': market_name,
                'shares': shares,
                'entry_price': price,
                'cost_basis': amount,
                'bought_at': datetime.now(timezone.utc).isoformat(),
            })
        
        portfolio['cash'] -= amount
        portfolio['history'].append({
            'action': 'buy',
            'slug': slug,
            'outcome': outcome,
            'shares': shares,
            'price': price,
            'amount': amount,
            'at': datetime.now(timezone.utc).isoformat(),
        })
        
        save_json('portfolio.json', portfolio)
        
        print(f"✅ Bought {shares:.1f} shares of **{market_name}**")
        print(f"   Price: {format_price(price)} | Cost: ${amount:,.2f}")
        print(f"   Cash remaining: ${portfolio['cash']:,.2f}")
        
    except Exception as e:
        print(f"❌ Error: {e}")


def cmd_sell(args):
    """Paper sell a position."""
    portfolio = load_json('portfolio.json', {'positions': [], 'history': [], 'cash': 10000})
    
    slug = extract_slug_from_url(args.slug)
    
    # Find position
    pos = None
    for p in portfolio['positions']:
        if p['slug'] == slug:
            pos = p
            break
    
    if not pos:
        print(f"❌ No position in {slug}")
        return
    
    try:
        data = fetch('/events', {'slug': slug})
        if not data:
            print(f"❌ Event not found: {slug}")
            return
        
        event = data[0] if isinstance(data, list) else data
        markets = event.get('markets', [])
        
        price = 0
        if pos.get('outcome') and markets:
            for m in markets:
                if pos['outcome'].lower() in m.get('groupItemTitle', '').lower():
                    price = get_market_price(m)
                    break
        elif markets:
            price = get_market_price(markets[0])
        
        if price <= 0:
            print("❌ Could not get price")
            return
        
        shares = pos['shares']
        proceeds = shares * price
        pnl = proceeds - pos['cost_basis']
        
        portfolio['cash'] += proceeds
        portfolio['positions'] = [p for p in portfolio['positions'] if p['slug'] != slug]
        portfolio['history'].append({
            'action': 'sell',
            'slug': slug,
            'shares': shares,
            'price': price,
            'proceeds': proceeds,
            'pnl': pnl,
            'at': datetime.now(timezone.utc).isoformat(),
        })
        
        save_json('portfolio.json', portfolio)
        
        direction = "🟢" if pnl >= 0 else "🔴"
        print(f"{direction} Sold {shares:.1f} shares of **{pos['name']}**")
        print(f"   Price: {format_price(price)} | Proceeds: ${proceeds:,.2f}")
        print(f"   P&L: ${pnl:+,.2f}")
        print(f"   Cash: ${portfolio['cash']:,.2f}")
        
    except Exception as e:
        print(f"❌ Error: {e}")


# ==================== MAIN ====================

def main():
    parser = argparse.ArgumentParser(description="Polymarket prediction markets")
    parser.add_argument("--limit", "-l", type=int, default=5, help="Number of results")
    parser.add_argument("--json", "-j", action="store_true", help="Output raw JSON")
    parser.add_argument("--all", "-a", action="store_true", help="Show all markets in event")
    
    subparsers = parser.add_subparsers(dest="command", required=True)
    
    # Original commands
    subparsers.add_parser("trending", help="Get trending markets")
    subparsers.add_parser("featured", help="Get featured markets")
    
    search_parser = subparsers.add_parser("search", help="Search markets")
    search_parser.add_argument("query", help="Search query")
    search_parser.add_argument("--all", "-a", action="store_true", help="Show all outcomes")
    
    event_parser = subparsers.add_parser("event", help="Get event by slug or URL")
    event_parser.add_argument("slug", help="Event slug or polymarket.com URL")
    
    market_parser = subparsers.add_parser("market", help="Get specific market outcome")
    market_parser.add_argument("slug", help="Event slug or URL")
    market_parser.add_argument("outcome", nargs="?", help="Outcome name")
    
    cat_parser = subparsers.add_parser("category", help="Markets by category")
    cat_parser.add_argument("category", help="Category: politics, crypto, sports, tech, etc.")
    
    # NEW: Watch commands
    watch_parser = subparsers.add_parser("watch", help="Manage watchlist")
    watch_parser.add_argument("action", choices=['add', 'remove', 'list'], help="Action")
    watch_parser.add_argument("slug", nargs="?", help="Event slug")
    watch_parser.add_argument("--outcome", "-o", help="Specific outcome to watch")
    watch_parser.add_argument("--alert-at", type=float, help="Alert when price reaches X%")
    watch_parser.add_argument("--alert-change", type=float, help="Alert on X% change from entry")
    
    # NEW: Alerts (for cron)
    alerts_parser = subparsers.add_parser("alerts", help="Check watchlist for alerts")
    alerts_parser.add_argument("--quiet", "-q", action="store_true", help="Only output if alerts triggered")
    
    # NEW: Calendar
    calendar_parser = subparsers.add_parser("calendar", help="Markets resolving soon")
    calendar_parser.add_argument("--days", "-d", type=int, default=7, help="Days to look ahead")
    
    # NEW: Movers
    movers_parser = subparsers.add_parser("movers", help="Biggest price movers")
    movers_parser.add_argument("--timeframe", "-t", default="24h", choices=["24h", "1w", "1m"], help="Timeframe")
    movers_parser.add_argument("--min-volume", type=float, default=10, help="Min 24h volume in $K")
    
    # NEW: Digest
    digest_parser = subparsers.add_parser("digest", help="Category digest summary")
    digest_parser.add_argument("category", help="Category: politics, crypto, sports, tech, business")
    
    # NEW: Portfolio
    subparsers.add_parser("portfolio", help="Show paper portfolio")
    
    # NEW: Buy
    buy_parser = subparsers.add_parser("buy", help="Paper buy position")
    buy_parser.add_argument("slug", help="Event slug")
    buy_parser.add_argument("amount", type=float, help="Amount in dollars")
    buy_parser.add_argument("--outcome", "-o", help="Specific outcome")
    
    # NEW: Sell
    sell_parser = subparsers.add_parser("sell", help="Paper sell position")
    sell_parser.add_argument("slug", help="Event slug")
    
    args = parser.parse_args()
    
    commands = {
        "trending": cmd_trending,
        "featured": cmd_featured,
        "search": cmd_search,
        "event": cmd_event,
        "market": cmd_market,
        "category": cmd_category,
        "watch": cmd_watch,
        "alerts": cmd_alerts,
        "calendar": cmd_calendar,
        "movers": cmd_movers,
        "digest": cmd_digest,
        "portfolio": cmd_portfolio,
        "buy": cmd_buy,
        "sell": cmd_sell,
    }
    
    try:
        commands[args.command](args)
    except requests.RequestException as e:
        print(f"❌ API Error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"❌ Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
