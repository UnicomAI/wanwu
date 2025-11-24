#!/usr/bin/env python3
# -*- coding: utf-8 -*-

with open('configs/microservice/mcp-service/configs/mcp_config.yaml', 'r', encoding='utf-8') as f:
    content = f.read()

# Translations for remaining 12306 content
replacements = [
    ("车次筛选条件，默认为空，即不筛选。支持多个标志同时筛选。例如用户说"高铁票"，则应使用 \"G\"。可选标志：[G(高铁/城际),D(动车),Z(直达特快),T(特快),K(快速),O(其他),F(复兴号),S(智能动车组)]", "Train filter conditions, default is empty, meaning no filtering. Supports multiple flags for simultaneous filtering. For example, if the user says \"high-speed train ticket\", use \"G\". Optional flags: [G(High-speed/Intercity),D(EMU),Z(Direct Express),T(Express),K(Fast),O(Other),F(Fuxing),S(Smart EMU)]"),
    ("排序方式，默认为空，即不排序。仅支持单一标识。可选标志：[startTime(出发时间从早到晚), arriveTime(抵达时间从早到晚), duration(历时从短到长)]", "Sort method, default is empty, meaning no sorting. Only supports single identifier. Optional flags: [startTime(Departure time from early to late), arriveTime(Arrival time from early to late), duration(Duration from short to long)]"),
    ("车次筛选条件，默认为空。从以下标志中选取多个条件组合[G(高铁/城际),D(动车),Z(直达特快),T(特快),K(快速),O(其他),F(复兴号),S(智能动车组)]", "Train filter conditions, default is empty. Select multiple conditions from the following flags [G(High-speed/Intercity),D(EMU),Z(Direct Express),T(Express),K(Fast),O(Other),F(Fuxing),S(Smart EMU)]"),
]

for chinese, english in replacements:
    content = content.replace(chinese, english)

with open('configs/microservice/mcp-service/configs/mcp_config.yaml', 'w', encoding='utf-8') as f:
    f.write(content)

print("Remaining translations done")

