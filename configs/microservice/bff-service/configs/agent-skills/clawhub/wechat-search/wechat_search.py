#!/usr/bin/env python3
"""
WeChat Search Skill - Search WeChat Official Account articles using Tavily Search API.
Returns the most recent 5 articles by default, with configurable options.

This skill uses Tavily Search API with site restriction to find WeChat articles.
It requires TAVILY_API_KEY environment variable to be set.
"""

import json
import sys
import os
import re
from urllib.parse import urlparse, parse_qs
import subprocess


class WeChatSearch:
    def __init__(self, config_path=None):
        self.config = self.load_config(config_path)
        
    def load_config(self, config_path):
        """Load configuration from file or use defaults"""
        default_config = {
            "max_results": 5,
            "cache_duration_hours": 1,
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

    def is_valid_wechat_url(self, url):
        """Check if URL is a valid WeChat Official Account article"""
        try:
            parsed = urlparse(url)
            if 'mp.weixin.qq.com' not in parsed.netloc:
                return False
            # Valid WeChat URLs have /s in the path
            return '/s' in parsed.path
        except:
            return False

    def extract_account_from_url(self, url):
        """Extract official account name from WeChat URL"""
        try:
            parsed = urlparse(url)
            if 'mp.weixin.qq.com' in parsed.netloc:
                return "Unknown Account"
        except:
            return "Unknown Account"

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
                date_info = "Unknown date"
                output.append(f"{i}. [{article['official_account']}] {article['title']} - {date_info}")
                snippet_preview = article['snippet'][:100] + "..." if len(article['snippet']) > 100 else article['snippet']
                output.append(f"   {snippet_preview}")
                output.append(f"   {article['url']}")
                output.append("")
                
        return "\n".join(output).strip()

    def search(self, query, max_results=5):
        """Search using Tavily API with site restriction"""
        try:
            # Check if Tavily API key is available
            tavily_key = os.environ.get('TAVILY_API_KEY')
            if not tavily_key:
                print("Error: TAVILY_API_KEY environment variable not set", file=sys.stderr)
                return []
                
            # Construct search query with site restriction
            search_query = f'{query} site:mp.weixin.qq.com'
            
            # Use subprocess to call Tavily Node.js script
            result = subprocess.run([
                'node', '/root/.openclaw/workspace/skills/tavily-search/scripts/search.mjs',
                search_query, '-n', str(max_results)
            ], capture_output=True, text=True, timeout=30, 
            env={**os.environ, 'TAVILY_API_KEY': tavily_key})
            
            if result.returncode == 0:
                return self._parse_tavily_results(result.stdout, max_results)
            else:
                print(f"Tavily search failed: {result.stderr}", file=sys.stderr)
                return []
                
        except Exception as e:
            print(f"Tavily search error: {e}", file=sys.stderr)
            return []

    def _parse_tavily_results(self, tavily_output, max_results):
        """Parse Tavily search results into standardized format"""
        articles = []
        try:
            # Tavily returns markdown with "## Sources" section
            if "## Sources" in tavily_output:
                sources_section = tavily_output.split("## Sources")[1]
                lines = sources_section.strip().split('\n')
                
                for line in lines:
                    if line.strip().startswith('- **') and 'mp.weixin.qq.com' in line:
                        # Extract title and URL from Tavily format
                        parts = line.split('**')
                        if len(parts) >= 3:
                            title = parts[1].strip()
                            # Find URL in the line
                            url_match = re.search(r'https?://[^\s)]+', line)
                            if url_match:
                                url = url_match.group(0)
                                article = {
                                    'title': title,
                                    'url': url,
                                    'snippet': 'Content from Tavily search',
                                    'source': 'tavily',
                                    'published_at': None,
                                    'official_account': self.extract_account_from_url(url)
                                }
                                if self.is_valid_wechat_url(url):
                                    articles.append(article)
                                    if len(articles) >= max_results:
                                        break
        except Exception as e:
            print(f"Error parsing Tavily results: {e}", file=sys.stderr)
            
        return articles


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
        articles = searcher.search(args.query, args.max_results)
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