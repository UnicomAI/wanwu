#!/bin/sh
# Optional full x86_64 kernel fixture for ARM Docker hosts. Test-only PID 1.
set -eu
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
export HOME=/root USER=root LOG_DIR=/var/log/gem
export OPENCODE_DISABLE_MODELS_FETCH=true OPENCODE_DISABLE_DEFAULT_PLUGINS=true
finish() {
  result=$?
  if [ "$result" -ne 0 ]; then
    cat /var/log/gem/supervisord.log /var/log/gem/opencode.log /var/log/gem/fixture-services.log || true
  fi
  echo "WANWU_VM_EXIT=$result"
  sync
  python3 -c 'import ctypes; ctypes.CDLL(None).reboot(0x4321fedc)'
}
trap finish EXIT
mount -t proc proc /proc
mount -t sysfs sysfs /sys
mount -t devtmpfs devtmpfs /dev
exec </dev/console >/dev/console 2>&1
mkdir -p /dev/pts /run /sys/fs/cgroup /sys/kernel/security
mount -t devpts devpts /dev/pts
mount -t tmpfs tmpfs /run
mount -t securityfs securityfs /sys/kernel/security
mount -t cgroup2 cgroup2 /sys/fs/cgroup
ip link set lo up
printf '+cpu +memory +pids\n' > /sys/fs/cgroup/cgroup.subtree_control
mkdir /sys/fs/cgroup/finsafe.slice
printf '+cpu +memory +pids\n' > /sys/fs/cgroup/finsafe.slice/cgroup.subtree_control
uname -a
cat /etc/os-release
/opt/finsafe/finsafe --version
/root/.opencode/bin/opencode --version
timeout 30 /opt/finsafe/finsafe probe || echo "Probe did not finish successfully; actual launch checks follow."
cp /tests/model-config.json /home/root/.config/opencode/opencode.json

echo START_BASELINE
cp /tests/opencode-baseline.conf /opt/gem/supervisord/opencode.conf
supervisord -c /etc/supervisor/supervisord.conf
timeout 240 python3 /tests/compatibility.py
supervisorctl shutdown
while [ -f /tmp/wanwu-supervisor.pid ]; do sleep 1; done

echo START_CONFINED
cp /tests/opencode-confined.conf /opt/gem/supervisord/opencode.conf
supervisord -c /etc/supervisor/supervisord.conf
timeout 240 python3 /tests/compatibility.py --confined
echo WANWU_COMPATIBILITY_X64=PASS

if [ -x /tests/wanwu-integration ]; then
  echo START_WANWU_INTEGRATION
  python3 /tests/services.py > /var/log/gem/fixture-services.log 2>&1 &
  attempt=0
  until grep -q FIXTURE_SERVICES_READY /var/log/gem/fixture-services.log; do
    attempt=$((attempt + 1))
    [ "$attempt" -lt 30 ]
    sleep 1
  done
  timeout 360 /tests/wanwu-integration
  echo WANWU_INTEGRATION_X64=PASS
fi
