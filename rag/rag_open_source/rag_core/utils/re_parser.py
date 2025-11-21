import re


def extract_pattern(text, pattern):
    result = []
    matched_strs = re.findall(pattern, text)
    for match in matched_strs:
        matched_strs.append(match.strip())

    return result

if __name__ == "__main__":
    text = "测试使用发布日期：2025年08月08日， 测试2025-02-02"
    date_patterns = [
        # Match "Release date: July 8, 2025" or "Release date: July 8, 2025"
        r'发布日期[：:]\s*\d{4}\s*年\s*\d{1,2}\s*月\s*\d{1,2}\s*日',

        # Match "Release Date: 2025-07-08" or "Release Date: 2025-7-8"
        r'发布日期[：:]\s*\d{4}-\d{1,2}-\d{1,2}',

        # Match "Release Date: 2025/07/08" or "Release Date: 2025/7/8"
        r'发布日期[：:]\s*\d{4}/\d{1,2}/\d{1,2}',

        # Match "Release date: 2025.07.08"
        r'发布日期[：:]\s*\d{4}\.\d{1,2}\.\d{1,2}',

        # Matches independent "July 8, 2025" or "July 8, 2025"
        r'\d{4}\s*年\s*\d{1,2}\s*月\s*\d{1,2}\s*日',

        # Matches "2025-07-08" or "2025-7-8" independently
        r'\b\d{4}-\d{1,2}-\d{1,2}\b',

        # Matches "2025/07/08" or "2025/7/8" independently
        r'\b\d{4}/\d{1,2}/\d{1,2}\b',

        # Matches the independent "2025.07.08"
        r'\b\d{4}\.\d{1,2}\.\d{1,2}\b'
    ]

    # Iterate through all patterns in order
    for pattern in date_patterns:
        matches = extract_pattern(text, pattern)
        for match in matches:
            print(match.strip())

