#!/usr/bin/env python3
"""
Simple WeChat Search Skill - Direct implementation using available tools
"""

import json
import sys
import os
import re
import subprocess
from urllib.parse import urlparse


def is_valid_wechat_url(url):
    """Check if URL is a valid WeChat Official Account article"""
    try:
        parsed = urlparse(url)
        if 'mp.weixin.qq.com' not in parsed.netloc:
            return False
        return '/s' in parsed.path
    except:
        return False


def extract_account_from_url(url):
    """Extract official account name from WeChat URL"""
    try:
        parsed = urlparse(url)
        if 'mp.weixin.qq.com' in parsed.netloc:
            return "WeChat Official Account"
    except:
        return "Unknown Account"


def parse_tavily_results(tavily_output, max_results):
    """Parse Tavily search results into standardized format"""
    articles = []
    try:
        if "## Sources" in tavily_output:
            sources_section = tavily_output.split("## Sources")[1]
            lines = sources_section.strip().split('\n')
            
            for line in lines:
                if line.strip().startswith('- **') and 'mp.weixin.qq.com' in line:
                    parts = line.split('**')
                    if len(parts) >= 3:
                        title = parts[1].strip()
                        url_match = re.search(r'https?://[^\s)]+', line)
                        if url_match:
                            url = url_match.group(0)
                            if is_valid_wechat_url(url):
                                article = {
                                    'title': title,
                                    'url': url,
                                    'snippet': 'Content from Tavily search',
                                    'official_account': extract_account_from_url(url)
                                }
                                articles.append(article)
                                if len(articles) >= max_results:
                                    break
    except Exception as e:
        print(f"Error parsing Tavily results: {e}", file=sys.stderr)
        
    return articles


def search_wechat_articles(query, max_results=5):
    """Search WeChat articles using Tavily with site restriction"""
    try:
        # Check if Tavily API key is available
        tavily_key = os.environ.get('TAVILY_API_KEY')
        if not tavily_key:
            print("Error: TAVILY_API_KEY environment variable not set", file=sys.stderr)
            return []
            
        # Construct search query with site restriction
        search_query = f'{query} site:mp.weixin.qq.com'
        
        # Call Tavily Node.js script
        result = subprocess.run([
            'node', '/root/.openclaw/workspace/skills/tavily-search/scripts/search.mjs',
            search_query, '-n', str(max_results)
        ], capture_output=True, text=True, timeout=30, 
        env={**os.environ, 'TAVILY_API_KEY': tavily_key})
        
        if result.returncode == 0:
            return parse_tavily_results(result.stdout, max_results)
        else:
            print(f"Tavily search failed: {result.stderr}", file=sys.stderr)
            return []
            
    except Exception as e:
        print(f"Search error: {e}", file=sys.stderr)
        return []


def format_results(articles, brief=False):
    """Format search results for display"""
    if not articles:
        return "No WeChat Official Account articles found for your query."
        
    output = []
    for i, article in enumerate(articles, 1):
        if brief:
            output.append(f"{i}. {article['title']}")
            output.append(f"   {article['url']}")
        else:
            output.append(f"{i}. [{article['official_account']}] {article['title']}")
            snippet_preview = article['snippet'][:100] + "..." if len(article['snippet']) > 100 else article['snippet']
            output.append(f"   {snippet_preview}")
            output.append(f"   {article['url']}")
            output.append("")
            
    return "\n".join(output).strip()


def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="Search WeChat Official Account articles")
    parser.add_argument('query', help="Search query")
    parser.add_argument('--max-results', type=int, default=5, 
                       help="Maximum number of results (default: 5, max: 20)")
    parser.add_argument('--brief', action='store_true', help="Brief output (title and URL only)")
    
    args = parser.parse_args()
    
    if args.max_results < 1 or args.max_results > 20:
        print("Error: --max-results must be between 1 and 20", file=sys.stderr)
        sys.exit(1)
    
    try:
        articles = search_wechat_articles(args.query, args.max_results)
        output = format_results(articles, args.brief)
        print(output)
        
    except KeyboardInterrupt:
        print("\nSearch interrupted by user.", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()