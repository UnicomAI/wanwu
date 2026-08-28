#!/usr/bin/env python3
"""
WeChat Search Skill - Fixed version with proper OpenClaw tool integration
Returns the most recent 5 articles by default, with configurable options.
"""

import json
import sys
import os
import re
from datetime import datetime
import subprocess
import urllib.parse


class WeChatSearch:
    def __init__(self, config_path=None):
        self.config = self.load_config(config_path)
        
    def load_config(self, config_path):
        """Load configuration from file or use defaults"""
        default_config = {
            "max_results": 5,
            "search_strategy": "auto",
            "cache_duration_hours": 1,
            "request_delay_ms": 5000,
            "user_agent": "OpenClaw-WeChat-Search-Bot/1.0"
        }
        
        if config_path and os.path.exists(config_path):
            try:
                with open(config_path, 'r') as f:
                    user_config = json.load(f)
                    default_config.update(user_config)
            except Exception as e:
                print(f"Warning: Failed to load config file: {e}", file=sys.stderr)
                
        return default_config

    def search_wechat_articles(self, query, max_results=5):
        """
        Main search function with three-tier fallback strategy:
        1. Tavily Search (if available)
        2. Web Fetch (direct scraping)
        Returns list of articles with title, url, snippet, source
        """
        articles = []
        
        # Tier 1: Try Tavily Search
        print("🔍 Trying Tavily Search...", file=sys.stderr)
        try:
            articles = self.tavily_search_wechat(query, max_results)
            if articles:
                print(f"✅ Tavily Search found {len(articles)} articles", file=sys.stderr)
                return articles
        except Exception as e:
            print(f"🔄 Tavily Search failed: {e}", file=sys.stderr)
        
        # Tier 2: Fallback to Web Fetch
        print("🔄 Tavily failed, trying direct web_fetch...", file=sys.stderr)
        try:
            articles = self.web_fetch_wechat(query, max_results)
            if articles:
                print(f"✅ Web fetch found {len(articles)} articles", file=sys.stderr)
                return articles
        except Exception as e:
            print(f"❌ Web fetch failed: {e}", file=sys.stderr)
        
        return []

    def tavily_search_wechat(self, query, max_results=5):
        """Use Tavily Search API to search WeChat articles"""
        try:
            # Construct search query with site restriction
            search_query = f'{query} site:mp.weixin.qq.com'
            
            # Call Tavily search script directly
            env = os.environ.copy()
            if 'TAVILY_API_KEY' not in env:
                # Try to get from common locations
                tavily_config_path = os.path.expanduser('~/.openclaw/tavily-config.json')
                if os.path.exists(tavily_config_path):
                    with open(tavily_config_path, 'r') as f:
                        config = json.load(f)
                        env['TAVILY_API_KEY'] = config.get('api_key', '')
            
            result = subprocess.run([
                'node',
                '/root/.openclaw/workspace/skills/tavily-search/scripts/search.mjs',
                search_query,
                '-n', str(max_results)
            ], capture_output=True, text=True, timeout=30, env=env)
            
            if result.returncode == 0:
                # Parse Tavily output
                return self.parse_tavily_results(result.stdout, max_results)
            else:
                raise Exception(f"Tavily search failed: {result.stderr}")
                
        except Exception as e:
            raise Exception(f"Tavily search error: {e}")

    def web_fetch_wechat(self, query, max_results=5):
        """Use OpenClaw's web_fetch tool to fetch WeChat search results"""
        try:
            # Construct WeChat search URL (using Sogou as proxy)
            encoded_query = urllib.parse.quote(query)
            search_url = f"https://weixin.sogou.com/weixin?type=2&query={encoded_query}"
            
            result = subprocess.run([
                'openclaw', 'tool', 'web_fetch',
                '--url', search_url,
                '--extract-mode', 'markdown'
            ], capture_output=True, text=True, timeout=30)
            
            if result.returncode == 0:
                content = result.stdout
                return self.parse_sogou_results(content, max_results)
            else:
                raise Exception(f"Web fetch failed: {result.stderr}")
                
        except Exception as e:
            raise Exception(f"Web fetch error: {e}")

    def parse_tavily_results(self, tavily_output, max_results):
        """Parse Tavily search results into standardized format"""
        articles = []
        try:
            # Tavily output contains markdown with sources section
            if '## Sources' in tavily_output:
                sources_section = tavily_output.split('## Sources')[1]
                lines = sources_section.strip().split('\n')
                
                current_title = ""
                current_url = ""
                current_snippet = ""
                
                for line in lines:
                    line = line.strip()
                    if line.startswith('- **') and '**' in line:
                        # Extract title
                        title_part = line.split('**')[1]
                        current_title = title_part
                    elif line.startswith('  ') and 'http' in line:
                        # Extract URL
                        url_start = line.find('http')
                        url_end = line.find(')', url_start)
                        if url_end == -1:
                            url_end = len(line)
                        current_url = line[url_start:url_end].strip()
                        
                        # Create article
                        if current_title and current_url:
                            article = {
                                'title': current_title.strip(),
                                'url': current_url.strip(),
                                'snippet': current_snippet.strip() if current_snippet else 'No snippet available',
                                'source': 'tavily',
                                'official_account': self.extract_account_from_title(current_title)
                            }
                            if self.is_valid_wechat_url(article['url']):
                                articles.append(article)
                                if len(articles) >= max_results:
                                    break
                        # Reset for next article
                        current_title = ""
                        current_url = ""
                        current_snippet = ""
                    elif line.startswith('  # ') and current_title:
                        # Extract snippet/content
                        current_snippet = line[4:].strip()
                        
        except Exception as e:
            print(f"Error parsing Tavily results: {e}", file=sys.stderr)
            
        return articles[:max_results]

    def parse_sogou_results(self, content, max_results):
        """Parse Sogou WeChat search results"""
        articles = []
        try:
            # Look for article patterns in the content
            lines = content.split('\n')
            for i, line in enumerate(lines):
                if '### [' in line and '](' in line:
                    # Extract title and URL
                    title_start = line.find('[') + 1
                    title_end = line.find(']')
                    url_start = line.find('(') + 1
                    url_end = line.find(')')
                    
                    if title_start > 0 and title_end > title_start and url_start > 0 and url_end > url_start:
                        title = line[title_start:title_end].strip()
                        url = line[url_start:url_end].strip()
                        
                        # Get snippet from next lines
                        snippet = ""
                        if i + 1 < len(lines):
                            snippet = lines[i + 1].strip()
                        
                        article = {
                            'title': title,
                            'url': url,
                            'snippet': snippet,
                            'source': 'sogou',
                            'official_account': self.extract_account_from_title(title)
                        }
                        if self.is_valid_wechat_url(article['url']):
                            articles.append(article)
                            if len(articles) >= max_results:
                                break
        except Exception as e:
            print(f"Error parsing Sogou results: {e}", file=sys.stderr)
            
        return articles[:max_results]

    def extract_account_from_title(self, title):
        """Extract official account name from title"""
        if ' - ' in title:
            return title.split(' - ')[-1]
        elif '｜' in title:
            return title.split('｜')[-1]
        elif '|' in title:
            return title.split('|')[-1]
        else:
            return "Unknown Account"

    def is_valid_wechat_url(self, url):
        """Check if URL is a valid WeChat Official Account article"""
        try:
            if not url.startswith('http'):
                return False
            if 'mp.weixin.qq.com' in url:
                return True
            if 'weixin.sogou.com/link' in url:
                return True
            return False
        except:
            return False

    def format_results(self, articles, brief=False):
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
    parser.add_argument('--config', help="Path to configuration file")
    
    args = parser.parse_args()
    
    # Validate max_results
    if args.max_results < 1 or args.max_results > 20:
        print("Error: --max-results must be between 1 and 20", file=sys.stderr)
        sys.exit(1)
    
    try:
        searcher = WeChatSearch(args.config)
        articles = searcher.search_wechat_articles(args.query, args.max_results)
        output = searcher.format_results(articles, args.brief)
        print(output)
        
    except KeyboardInterrupt:
        print("\nSearch interrupted by user.", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()