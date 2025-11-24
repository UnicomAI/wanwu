#!/bin/bash

# Use sed to replace Chinese text with English
sed -i '' \
  's/查询12306余票信息。/Query 12306 available ticket information./g' \
  's/查询日期，格式为 "yyyy-MM-dd"。如果用户提供的是相对日期（如"明天"），请务必先调用 `get-current-date` 接口获取当前日期，并计算出目标日期。/Query date, format is "yyyy-MM-dd". If the user provides a relative date (such as "tomorrow"), be sure to call the `get-current-date` interface first to get the current date and calculate the target date./g' \
  's/出发地的 `station_code` 。必须是通过 `get-station-code-by-names` 或 `get-station-code-of-citys` 接口查询得到的编码，严禁直接使用中文地名。/Departure location `station_code`. Must be obtained through `get-station-code-by-names` or `get-station-code-of-citys` interface, strictly prohibited from using Chinese place names directly./g' \
  's/到达地的 `station_code` 。必须是通过 `get-station-code-by-names` 或 `get-station-code-of-citys` 接口查询得到的编码，严禁直接使用中文地名。/Arrival location `station_code`. Must be obtained through `get-station-code-by-names` or `get-station-code-of-citys` interface, strictly prohibited from using Chinese place names directly./g' \
  's/车次筛选条件，默认为空，即不筛选。支持多个标志同时筛选。例如用户说"高铁票"，则应使用 "G"。可选标志：\[G(高铁\/城际),D(动车),Z(直达特快),T(特快),K(快速),O(其他),F(复兴号),S(智能动车组)\]/Train filter conditions, default is empty, meaning no filtering. Supports multiple flags for simultaneous filtering. For example, if the user says "high-speed train ticket", use "G". Optional flags: [G(High-speed\/Intercity),D(EMU),Z(Direct Express),T(Express),K(Fast),O(Other),F(Fuxing),S(Smart EMU)]/g' \
  's/排序方式，默认为空，即不排序。仅支持单一标识。可选标志：\[startTime(出发时间从早到晚), arriveTime(抵达时间从早到晚), duration(历时从短到长)\]/Sort method, default is empty, meaning no sorting. Only supports single identifier. Optional flags: [startTime(Departure time from early to late), arriveTime(Arrival time from early to late), duration(Duration from short to long)]/g' \
  's/是否逆向排序结果，默认为false。仅在设置了sortFlag时生效。/Whether to reverse sort results, default is false. Only effective when sortFlag is set./g' \
  's/返回的余票数量限制，默认为0，即不限制。/Limit on the number of available tickets returned, default is 0, meaning no limit./g' \
  's/查询12306中转余票信息。尚且只支持查询前十条。/Query 12306 transfer available ticket information. Currently only supports querying the first ten results./g' \
  'configs/microservice/mcp-service/configs/mcp_config.yaml'

echo "Translation completed"
