#!/bin/sh
# Reproduce on a native Linux x86_64 Docker daemon with supplied verified binaries.
set -eu
phase=${1:-compatibility}
case "$phase" in compatibility|wanwu) ;; *) echo 'Usage: run.sh [compatibility|wanwu]' >&2; exit 2 ;; esac
: "${FINSAFE_RUNTIME_DIR:?Set FINSAFE_RUNTIME_DIR to the complete extracted Linux x64 runtime}"
: "${OPENCODE_BINARY:?Set OPENCODE_BINARY to the Linux x64 OpenCode 1.19.5 executable}"
: "${FINSAFE_CGROUP_ROOT:?Set FINSAFE_CGROUP_ROOT to the dedicated delegated cgroup subtree}"
repo_dir=$(CDPATH='' cd -- "$(dirname -- "$0")/../../.." && pwd)
cd "$repo_dir"
architecture=$(docker info --format '{{.Architecture}}')
case "$architecture" in
  x86_64|amd64) ;;
  *) echo 'Use a native x64 Docker daemon, or the documented full x64 kernel fixture on ARM.' >&2; exit 2 ;;
esac
name=wanwu-finsafe-test-$$
baseline=$name-baseline
cleanup() {
  docker rm -f "$name" "$baseline" >/dev/null 2>&1 || true
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' HUP TERM
image=wanwu-finsafe-compat:local
docker build --platform linux/amd64 -f deploy/finsafe/tests/Dockerfile -t "$image" deploy/finsafe

if [ "$phase" = compatibility ]; then
  docker run -d --name "$baseline" --platform linux/amd64 \
    -v "$OPENCODE_BINARY:/root/.opencode/bin/opencode:ro" \
    -v "$repo_dir/deploy/finsafe/tests/opencode-baseline.conf:/opt/gem/supervisord/opencode.conf:ro" \
    -v "$repo_dir/deploy/finsafe/tests:/tests:ro" "$image"
  docker exec "$baseline" python3 /tests/compatibility.py
  docker rm -f "$baseline" >/dev/null
fi

if [ "$phase" = wanwu ]; then
  : "${WANWU_INTEGRATION_BINARY:?Set WANWU_INTEGRATION_BINARY to the compiled Linux x64 test driver}"
  docker run -d --name "$name" --platform linux/amd64 --cgroupns host --security-opt seccomp=unconfined \
    -v "$FINSAFE_CGROUP_ROOT:/sys/fs/cgroup:rw" \
    -v "$FINSAFE_RUNTIME_DIR:/opt/finsafe:ro" \
    -v "$OPENCODE_BINARY:/root/.opencode/bin/opencode:ro" \
    -v "$repo_dir/deploy/finsafe/tests:/tests:ro" \
    -v "$WANWU_INTEGRATION_BINARY:/usr/local/bin/wanwu-integration:ro" \
    --entrypoint /bin/sh "$image" -c \
      'cp /tests/model-config.json /home/root/.config/opencode/opencode.json; exec supervisord -n -c /etc/supervisor/supervisord.conf'
  docker exec "$name" python3 /tests/compatibility.py --confined
  docker exec -d "$name" python3 /tests/services.py
  docker exec -i "$name" python3 - <<'PY'
import socket, time
for attempt in range(30):
    try:
        with socket.create_connection(("127.0.0.1", 8000), timeout=1):
            break
    except OSError:
        time.sleep(1)
else:
    raise SystemExit("Fixture model did not start")
PY
  docker exec "$name" /usr/local/bin/wanwu-integration
else
  docker run -d --name "$name" --platform linux/amd64 --cgroupns host --security-opt seccomp=unconfined \
    -v "$FINSAFE_CGROUP_ROOT:/sys/fs/cgroup:rw" \
    -v "$FINSAFE_RUNTIME_DIR:/opt/finsafe:ro" \
    -v "$OPENCODE_BINARY:/root/.opencode/bin/opencode:ro" \
    -v "$repo_dir/deploy/finsafe/tests:/tests:ro" "$image"
  docker exec "$name" python3 /tests/compatibility.py --confined
fi
