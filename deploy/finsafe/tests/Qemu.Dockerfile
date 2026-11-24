# Optional test host on ARM machines; guest executes the actual Linux x64 binaries.
FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends \
      qemu-system-x86 cpio gzip python3 e2fsprogs ca-certificates && rm -rf /var/lib/apt/lists/*
