#!/bin/bash
# 直接在 docker-compose.ontology.yaml 中为 data-connection / vega-gateway-pro /
# web / wga-sandbox-ontology 注入 RSA 密钥及相关挂载：
#   - 目标 service 若没有 volumes: 键则新增；已有则按行内容去重追加。
#   - state.json 由 generate-keys.sh 生成；此脚本只在 state.json 存在时把对应挂载写入。
#   - 触发文件不存在则跳过对应挂载（不会反向删除已存在的挂载行，仅打印 WARN）。
# 幂等：mount 行已存在时不重复追加。
# 无论从项目根 (wanwu/) 还是脚本所在目录 (configs/.../vega-server/) 执行，结果一致。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
cd "$PROJECT_ROOT"

TARGET="./docker-compose.ontology.yaml"

BASE="./configs/microservice/ontology/vega-server"
DC_APP_YML="${BASE}/data-connection/application.yml"
PRIV="${BASE}/private_key.pem"
PUB="${BASE}/public_key.pem"
WEB_KEY="${BASE}/public-key.js"
STATE_JSON="${BASE}/state.json"

MISSING=()
for f in "$DC_APP_YML" "$PRIV" "$PUB" "$WEB_KEY" "$STATE_JSON"; do
  [ ! -s "$f" ] && MISSING+=("$f")
done
if [ ${#MISSING[@]} -gt 0 ]; then
  echo "ERROR: required files missing or empty — refusing to inject partial mounts:" >&2
  for f in "${MISSING[@]}"; do echo "  - $f" >&2; done
  echo "" >&2
  echo "Run the prerequisite steps first (see configs/microservice/ontology/vega-server/README.md):" >&2
  echo "  1. ./configs/microservice/ontology/vega-server/generate-keys.sh ./configs/microservice/ontology/vega-server" >&2
  echo "  2. node ./configs/microservice/ontology/vega-server/generate-public-key-js.js" >&2
  exit 1
fi

if [ ! -f "$TARGET" ]; then
  echo "ERROR: $TARGET not found (run from repo root)" >&2
  exit 1
fi

if [ -f "./docker-compose.override.yaml" ]; then
  echo "WARN: ./docker-compose.override.yaml exists. This script no longer generates it; remove it manually if it was created by an earlier version of this script." >&2
fi

DATA_CONNECTION_MOUNTS=()
VEGA_GATEWAY_PRO_MOUNTS=()
WEB_MOUNTS=()
WGA_SANDBOX_MOUNTS=()

[ -s "$DC_APP_YML" ] && DATA_CONNECTION_MOUNTS+=("${DC_APP_YML}:/opt/data-connection/config/application.yml:ro")
[ -s "$PRIV" ]      && DATA_CONNECTION_MOUNTS+=("${PRIV}:/opt/vega/config/private_key.pem:ro")
[ -s "$PUB" ]       && DATA_CONNECTION_MOUNTS+=("${PUB}:/opt/vega/config/public_key.pem:ro")
[ -s "$PRIV" ]      && VEGA_GATEWAY_PRO_MOUNTS+=("${PRIV}:/opt/vega-gateway-pro/config/private_key.pem:ro")
[ -s "$WEB_KEY" ]   && WEB_MOUNTS+=("${WEB_KEY}:/usr/share/nginx/html/vega/config/public-key.js:ro")
[ -s "$STATE_JSON" ] && WGA_SANDBOX_MOUNTS+=("${STATE_JSON}:/root/.ontology/state.json:ro")

inject_service() {
  local svc="$1"; shift
  local mounts=("$@")
  if [ ${#mounts[@]} -eq 0 ]; then
    return 0
  fi

  local mounts_str=""
  local m
  for m in "${mounts[@]}"; do
    mounts_str+="${m}"$'\n'
  done

  local tmp
  tmp=$(mktemp)
  SVC="$svc" MOUNTS="$mounts_str" awk '
    function indent_count(s,    i, c) {
      c = 0
      for (i = 1; i <= length(s); i++) {
        if (substr(s, i, 1) == " ") c++
        else break
      }
      return c
    }
    function extract_mount(s,    t) {
      t = s
      sub(/^[[:space:]]+- /, "", t)
      return t
    }
    function flush(    i, vol_key_idx, vol_end_idx, insert_after, mount) {
      if (!is_target) {
        for (i = 0; i < buf_count; i++) print buf[i]
        return
      }
      vol_key_idx = -1
      vol_end_idx = -1
      delete existing
      for (i = 0; i < buf_count; i++) {
        if (vol_key_idx < 0) {
          if (buf[i] ~ /^    volumes:[[:space:]]*$/) {
            vol_key_idx = i
            vol_end_idx = i
          }
        } else {
          if (buf[i] ~ /^      /) {
            vol_end_idx = i
            if (buf[i] ~ /^      - /) existing[extract_mount(buf[i])] = 1
          } else {
            break
          }
        }
      }
      if (vol_key_idx >= 0) {
        for (i = 0; i <= vol_end_idx; i++) print buf[i]
        for (i = 1; i <= n_want; i++) {
          mount = want[i]
          if (!(mount in existing)) print "      - " mount
        }
        for (i = vol_end_idx + 1; i < buf_count; i++) print buf[i]
      } else {
        insert_after = 0
        for (i = 1; i < buf_count; i++) {
          if (buf[i] ~ /^[[:space:]]*$/) continue
          if (indent_count(buf[i]) >= 4) insert_after = i
        }
        for (i = 0; i <= insert_after; i++) print buf[i]
        print "    volumes:"
        for (i = 1; i <= n_want; i++) print "      - " want[i]
        for (i = insert_after + 1; i < buf_count; i++) print buf[i]
      }
    }
    BEGIN {
      svc = ENVIRON["SVC"]
      raw_n = split(ENVIRON["MOUNTS"], raw, "\n")
      n_want = 0
      for (i = 1; i <= raw_n; i++) {
        if (raw[i] != "") {
          n_want++
          want[n_want] = raw[i]
        }
      }
      buf_count = 0
      is_target = 0
      in_services = 0
    }
    {
      if ($0 ~ /^[a-zA-Z][a-zA-Z0-9_-]*:[[:space:]]*$/) {
        if (buf_count > 0) flush()
        buf_count = 0
        is_target = 0
        top_key = $0
        sub(/:.*$/, "", top_key)
        in_services = (top_key == "services") ? 1 : 0
        buf[buf_count++] = $0
        next
      }
      if (in_services && $0 ~ /^  [a-z][a-z0-9_-]*:[[:space:]]*$/) {
        if (buf_count > 0) flush()
        buf_count = 0
        is_target = 0
        hdr = $0
        sub(/^  /, "", hdr); sub(/:.*$/, "", hdr)
        if (hdr == svc) is_target = 1
        buf[buf_count++] = $0
        next
      }
      buf[buf_count++] = $0
    }
    END {
      if (buf_count > 0) flush()
    }
  ' "$TARGET" > "$tmp"

  if ! cmp -s "$tmp" "$TARGET"; then
    mv "$tmp" "$TARGET"
    echo "Updated ${svc} in ${TARGET}"
  else
    rm "$tmp"
    echo "No change for ${svc} (mounts already present)"
  fi
}

if [ ${#DATA_CONNECTION_MOUNTS[@]} -gt 0 ]; then
  inject_service "data-connection" "${DATA_CONNECTION_MOUNTS[@]}"
fi
if [ ${#VEGA_GATEWAY_PRO_MOUNTS[@]} -gt 0 ]; then
  inject_service "vega-gateway-pro" "${VEGA_GATEWAY_PRO_MOUNTS[@]}"
fi
if [ ${#WEB_MOUNTS[@]} -gt 0 ]; then
  inject_service "web" "${WEB_MOUNTS[@]}"
fi
if [ ${#WGA_SANDBOX_MOUNTS[@]} -gt 0 ]; then
  inject_service "wga-sandbox-ontology" "${WGA_SANDBOX_MOUNTS[@]}"
fi

warn_orphan() {
  local file="$1" mount_substr="$2"
  if [ ! -s "$file" ] && grep -qF -- "$mount_substr" "$TARGET"; then
    echo "WARN: ${file} missing but ${TARGET} still contains '${mount_substr}' — remove manually if no longer needed" >&2
  fi
}

warn_orphan "$DC_APP_YML" "${DC_APP_YML}:"
warn_orphan "$PRIV"      "${PRIV}:/opt/vega/config/private_key.pem:"
warn_orphan "$PUB"       "${PUB}:/opt/vega/config/public_key.pem:"
warn_orphan "$WEB_KEY"   "${WEB_KEY}:/usr/share/nginx/html/vega/config/public-key.js:"
warn_orphan "$STATE_JSON" "${STATE_JSON}:/root/.ontology/state.json:"

echo "Done."
