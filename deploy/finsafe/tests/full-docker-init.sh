#!/bin/sh
# Test-only PID 1 in the Docker-hosted full x64 kernel environment.
set -eu
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
mount -t proc proc /proc
mount -t sysfs sysfs /sys
mount -t devtmpfs devtmpfs /dev
exec </dev/console >/dev/console 2>&1
mkdir -p /dev/pts /run /sys/fs/cgroup /sys/kernel/security /var/lib/docker /var/log
mount -t devpts devpts /dev/pts
mount -t tmpfs tmpfs /run
mount -t securityfs securityfs /sys/kernel/security
mount -t cgroup2 cgroup2 /sys/fs/cgroup
ip link set lo up
python3 - <<'PY'
import ctypes, errno, json, os
from pathlib import Path
libc = ctypes.CDLL(None, use_errno=True)
for name in json.loads(Path('/modules/order.json').read_text()):
    fd = os.open('/modules/' + name, os.O_RDONLY)
    try:
        result = libc.syscall(313, fd, b'', 0)
        if result != 0 and ctypes.get_errno() != errno.EEXIST:
            raise OSError(ctypes.get_errno(), name)
    finally:
        os.close(fd)
print('PASS full x64 guest kernel modules loaded', flush=True)
PY
ip addr add 10.0.2.15/24 dev eth0
ip link set eth0 up
ip route add default via 10.0.2.2
printf 'nameserver 10.0.2.3\n' > /etc/resolv.conf
mount -t ext4 /dev/vda /var/lib/docker
printf '+cpu +memory +pids\n' > /sys/fs/cgroup/cgroup.subtree_control
mkdir -p /sys/fs/cgroup/finsafe-wanwu/finsafe.slice
printf '+cpu +memory +pids\n' > /sys/fs/cgroup/finsafe-wanwu/cgroup.subtree_control
printf '+cpu +memory +pids\n' > /sys/fs/cgroup/finsafe-wanwu/finsafe.slice/cgroup.subtree_control
# This test host boots from initramfs, where pivot_root is unavailable.
# Docker's RAMDISK option changes only the test host's container root setup.
DOCKER_RAMDISK=1 dockerd --host=unix:///run/docker.sock --data-root=/var/lib/docker --exec-root=/run/docker \
  --storage-driver=overlay2 --iptables=false --ip6tables=false --bridge=none \
  --ip-forward=false --ip-masq=false --exec-opt native.cgroupdriver=cgroupfs \
  > /var/log/dockerd.log 2>&1 &
attempt=0
until docker info >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "$attempt" -gt 60 ]; then cat /var/log/dockerd.log; exit 1; fi
  sleep 1
done
docker info --format 'GUEST_DOCKER_READY arch={{.Architecture}} driver={{.Driver}} cgroup={{.CgroupVersion}}'
on_exit() {
  code=$?
  echo "FULL_IMAGE_TEST_EXIT=$code"
  docker logs --tail 80 wanwu-wga-finsafe-full 2>&1 || true
  cat /sys/fs/cgroup/finsafe-wanwu/memory.current /sys/fs/cgroup/finsafe-wanwu/memory.events || true
  timeout 20 docker exec wanwu-wga-finsafe-full sh -c 'tail -n 12 /var/log/gem/opencode.log /var/log/gem/nginx-error.log /var/log/gem/nginx-access.log /var/log/gem/gem-server.log /var/log/gem/python-server.log /var/log/gem/mcp-hub.log' || true
  tail -n 30 /var/log/dockerd.log || true
  sync
  while :; do sleep 60; done
}
trap on_exit EXIT
if [ -b /dev/vdb ]; then
  echo 'Importing the supplied sandbox image (existing layers are reused).'
  docker load -i /dev/vdb
elif ! docker image inspect wanwu-wga-finsafe:0.11.19 >/dev/null 2>&1; then
  echo BOOTSTRAP_ONLY_PASS
  sync
  python3 -c 'import ctypes; ctypes.CDLL(None).reboot(0x4321fedc)'
fi
docker rm -f wanwu-wga-finsafe-full >/dev/null 2>&1 || true
model_config=/tests/live-model-config.json
sandbox_memory=0
case "$(cat /proc/cmdline)" in
  *wanwu_low_memory=1*)
    sandbox_memory=1536m
    # Count both Docker's processes and FinSafe's sibling scopes in one budget.
    printf '1610612736\n' > /sys/fs/cgroup/finsafe-wanwu/memory.max
    ;;
esac
case "$(cat /proc/cmdline)" in
  *wanwu_ui=1*)
    # Wanwu supplies each task's provider and callback URL itself.
    # Do not retain the standalone test provider allowlist or forwarding gateway.
    printf '{}\n' > /tmp/ui-opencode.json
    model_config=/tmp/ui-opencode.json
    ;;
esac
docker run -d --name wanwu-wga-finsafe-full --network host --cgroupns host \
  --cgroup-parent=/finsafe-wanwu --memory "$sandbox_memory" \
  --security-opt seccomp=unconfined -e USER=root \
  -e OPENCODE_DISABLE_MODELS_FETCH=true -e OPENCODE_DISABLE_DEFAULT_PLUGINS=true \
  --mount type=bind,src=/sys/fs/cgroup/finsafe-wanwu,dst=/sys/fs/cgroup \
  --mount type=bind,src="$model_config",dst=/home/root/.config/opencode/opencode.json,readonly \
  --mount type=bind,src=/tests/nginx-server-without-auth.conf,dst=/opt/gem/nginx-server-without-auth.conf \
  --mount type=bind,src=/tests/nginx-server-with-auth.conf,dst=/opt/gem/nginx-server-with-auth.conf \
  --mount type=bind,src=/tests/nginx-server-port-proxy.conf.template,dst=/opt/gem/nginx-server-port-proxy.conf.template \
  wanwu-wga-finsafe:0.11.19
docker inspect wanwu-wga-finsafe-full --format 'SANDBOX_CGROUP_PARENT={{.HostConfig.CgroupParent}} CONTAINER_MEMORY_LIMIT={{.HostConfig.Memory}}'
printf 'SANDBOX_SHARED_MEMORY_LIMIT='
cat /sys/fs/cgroup/finsafe-wanwu/memory.max
python3 - <<'PY'
import json, subprocess, time, urllib.request
from pathlib import Path
# Full-system emulation can spend several minutes importing the AIO Python stack.
for attempt in range(360):
    if attempt in (15, 60, 120):
        print('SANDBOX_STARTUP_DIAGNOSTIC', attempt, flush=True)
        print(Path('/sys/fs/cgroup/finsafe-wanwu/memory.events').read_text(), flush=True)
        subprocess.run(['docker', 'exec', 'wanwu-wga-finsafe-full', 'sh', '-c',
                        'ps -eo pid,rss,comm; tail -n 8 /var/log/gem/opencode.log /var/log/gem/nginx-error.log /var/log/gem/nginx-access.log /var/log/gem/gem-server.log /var/log/gem/python-server.log /var/log/gem/mcp-hub.log'], timeout=30, check=False)
    try:
        health = json.load(urllib.request.urlopen('http://127.0.0.1:4096/global/health', timeout=2))
        assert urllib.request.urlopen('http://127.0.0.1:4097/health', timeout=2).read() == b'ok'
        urllib.request.urlopen('http://127.0.0.1:8080/v1/shell/exec', timeout=2)
    except urllib.error.HTTPError as error:
        if error.code in (405, 422) and 'health' in locals() and health.get('healthy'):
            print('PASS actual AIO and protected OpenCode are ready', flush=True)
            break
    except (OSError, AssertionError):
        pass
    else:
        if health.get('healthy'):
            print('PASS actual AIO and protected OpenCode are ready', flush=True)
            break
    time.sleep(1)
else:
    raise SystemExit('Actual AIO/OpenCode did not become ready')
PY
docker exec wanwu-wga-finsafe-full /opt/finsafe/finsafe --version
docker exec wanwu-wga-finsafe-full /root/.opencode/bin/opencode --version
docker exec -i wanwu-wga-finsafe-full python3 - <<'PY'
from pathlib import Path
import os
servers = {'/root/.opencode/bin/opencode': [], '/usr/local/bin/eino-agent': []}
for p in Path('/proc').iterdir():
    if not p.name.isdigit():
        continue
    try:
        executable = os.readlink(p / 'exe').removesuffix(' (deleted)')
        if executable not in servers:
            continue
        status = dict(line.split(':', 1) for line in (p / 'status').read_text().splitlines() if ':' in line)
        cgroup = (p / 'cgroup').read_text()
    except (FileNotFoundError, ProcessLookupError, PermissionError):
        continue
    assert status['NoNewPrivs'].strip() == '1'
    assert status['Seccomp'].strip() == '2'
    assert '/finsafe.slice/wrapper-self-confine-' in cgroup
    servers[executable].append(p.name)
assert all(servers.values()), 'Missing protected runtime: ' + str(servers)
p = Path('/var/lib/wanwu-unconfined-canary')
p.write_text('container root baseline')
assert p.read_text() == 'container root baseline'
p.unlink()
print('PASS actual image OpenCode and Eino cgroup/seccomp/NoNewPrivs; ordinary container root can write the test parent', servers, flush=True)
PY
case "$(cat /proc/cmdline)" in
  *wanwu_ui=1*) echo WANWU_UI_SANDBOX_READY ;;
  *)
    WANWU_LIVE_MODEL=1 WANWU_LIVE_MODEL_BASE_URL=http://10.0.2.2:8000/v1 \
      WANWU_LIVE_DOCKER_CONTAINER=wanwu-wga-finsafe-full /tests/wanwu-integration
    echo WANWU_FULL_IMAGE_LIVE_MODEL=PASS
    ;;
esac
docker ps --format '{{.Names}} {{.Status}}'
sync
echo 'Full image remains running; review ports are guest 8080 and 4096.'
while :; do sleep 60; done
