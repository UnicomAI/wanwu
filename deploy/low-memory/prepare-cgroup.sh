#!/bin/sh
# Run on the Linux Docker daemon host, before starting the low-memory sandbox.
# Prints environment assignments only; redirect to .env.low-memory.
set -eu
[ "$(id -u)" = 0 ] || { echo 'Run as root on the Linux Docker daemon host.' >&2; exit 1; }
[ "$(uname -s)" = Linux ] || { echo 'A Linux Docker daemon host is required.' >&2; exit 1; }
[ -f /sys/fs/cgroup/cgroup.controllers ] || { echo 'cgroup v2 is required.' >&2; exit 1; }
driver=$(docker info --format '{{.CgroupDriver}}')
case "$driver" in
  cgroupfs)
    parent=/finsafe-wanwu
    root=/sys/fs/cgroup/finsafe-wanwu
    budget=$root
    mkdir -p "$root"
    printf '1610612736\n' > "$root/memory.max"
    printf '0\n' > "$root/memory.swap.max"
    ;;
  systemd)
    parent=finsafe-wanwu.slice
    budget=/sys/fs/cgroup/finsafe.slice/finsafe-wanwu.slice
    root=$budget/finsafe-wanwu-delegate.service
    # The dash places this slice under finsafe.slice, per systemd naming rules.
    mkdir -p /run/systemd/system
    unit=/run/systemd/system/finsafe-wanwu.slice
    delegate=/run/systemd/system/finsafe-wanwu-delegate.service
    if [ -e "$unit" ] || [ -e "$delegate" ]; then
      echo 'Low-memory systemd units already exist; reuse the generated env file.' >&2
      exit 1
    fi
    cat > "$unit" <<'UNIT'
[Unit]
Description=Wanwu sandbox and FinSafe shared memory budget
[Slice]
MemoryAccounting=yes
CPUAccounting=yes
TasksAccounting=yes
MemoryMax=1536M
MemorySwapMax=0
UNIT
    # Delegate a service subtree to FinSafe; do not manually manage slice children.
    cat > "$delegate" <<'UNIT'
[Unit]
Description=Delegated FinSafe subtree for Wanwu
[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/bin/true
Slice=finsafe-wanwu.slice
Delegate=yes
UNIT
    systemctl daemon-reload
    systemctl start finsafe-wanwu.slice
    systemctl start finsafe-wanwu-delegate.service
    ;;
  *) echo "Unsupported Docker cgroup driver: $driver" >&2; exit 1 ;;
esac
printf '+cpu +memory +pids\n' > "$root/cgroup.subtree_control"
mkdir -p "$root/finsafe.slice"
printf '+cpu +memory +pids\n' > "$root/finsafe.slice/cgroup.subtree_control"
[ "$(cat "$budget/memory.max")" = 1610612736 ]
printf 'FINSAFE_CGROUP_ROOT=%s\nFINSAFE_CGROUP_PARENT=%s\nFINSAFE_CGROUP_BUDGET_ROOT=%s\n' "$root" "$parent" "$budget"
