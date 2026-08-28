#!/usr/bin/env python3

from wechat_search import WeChatSearch

searcher = WeChatSearch()
test_url = "https://mp.weixin.qq.com/s/abc123"
print(f"Testing URL: {test_url}")
result = searcher.is_valid_wechat_url(test_url)
print(f"Result: {result}")

# Test with the exact string from test
test_url2 = "https://mp.weixin.qq.com/s/abc123"
print(f"Testing URL2: {test_url2}")
result2 = searcher.is_valid_wechat_url(test_url2)
print(f"Result2: {result2}")