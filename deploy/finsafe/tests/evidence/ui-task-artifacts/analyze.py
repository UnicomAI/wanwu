import csv
from collections import defaultdict
from decimal import Decimal

rows = [
    ("上海", "软件", 12, 800),
    ("北京", "软件", 8, 900),
    ("深圳", "服务", 5, 1200),
    ("上海", "服务", 3, 1500),
    ("北京", "服务", 4, 1100),
    ("深圳", "软件", 10, 750),
]

with open("sales.csv", "w", newline="", encoding="utf-8") as f:
    w = csv.writer(f)
    w.writerow(["city", "category", "quantity", "unit_price"])
    w.writerows(rows)

city_rev = defaultdict(int)
cat_rev = defaultdict(int)
total = 0
for city, cat, qty, price in rows:
    rev = qty * price
    city_rev[city] += rev
    cat_rev[cat] += rev
    total += rev

city_sorted = sorted(city_rev.items(), key=lambda x: x[1], reverse=True)
with open("summary.csv", "w", newline="", encoding="utf-8") as f:
    w = csv.writer(f)
    w.writerow(["city", "revenue"])
    w.writerows(city_sorted)

lines = []
lines.append("# 销售数据分析简报")
lines.append("")
lines.append("> 说明：本报告基于**虚构测试数据**，仅用于演示数据分析流程，不代表任何真实业务情况。")
lines.append("")
lines.append("## 一、数据概览")
lines.append("")
lines.append(f"- 数据行数：{len(rows)} 行")
lines.append(f"- 总销售额：{total:,} 元")
lines.append("")
lines.append("## 二、城市销售额排名")
lines.append("")
lines.append("| 排名 | 城市 | 销售额（元） | 占比 |")
lines.append("| --- | --- | ---: | ---: |")
for i, (city, rev) in enumerate(city_sorted, 1):
    lines.append(f"| {i} | {city} | {rev:,} | {rev / total:.1%} |")
lines.append("")
lines.append("## 三、按类别汇总")
lines.append("")
lines.append("| 类别 | 销售额（元） | 占比 |")
lines.append("| --- | ---: | ---: |")
for cat, rev in sorted(cat_rev.items(), key=lambda x: x[1], reverse=True):
    lines.append(f"| {cat} | {rev:,} | {rev / total:.1%} |")
lines.append("")

with open("report.md", "w", encoding="utf-8") as f:
    f.write("\n".join(lines))

print("city:", dict(city_rev))
print("category:", dict(cat_rev))
print("total:", total)
print("lines:", len(rows))
