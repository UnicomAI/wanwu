# Build with docker-compose.finsafe.yaml; keep the original AIO entrypoint.
ARG WANWU_WGA_SANDBOX_IMAGE=crpi-6pj79y7ddzdpexs8.cn-hangzhou.personal.cr.aliyuncs.com/wanwulite/wga-sandbox:v0.6.5-df77eb71-ontology-0.1.6
FROM --platform=$BUILDPLATFORM ubuntu:24.04 AS finsafe-runtime

# Public Linux release, pinned together with its published SHA-256.
ARG FINSAFE_VERSION=0.11.19
ARG FINSAFE_SHA256=c6ac5c709cd34a2c1168534beb969da1813968159c0444f4271224411753a394
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl zstd && \
    rm -rf /var/lib/apt/lists/*
RUN set -eu; \
    curl --fail --location --retry 3 \
      "https://github.com/finogeeks/finsafe/releases/download/v${FINSAFE_VERSION}/finshield-${FINSAFE_VERSION}-linux-x64.tar.zst" \
      --output /tmp/finsafe.tar.zst; \
    printf '%s  /tmp/finsafe.tar.zst\n' "$FINSAFE_SHA256" | sha256sum --check --strict; \
    mkdir -p /opt/finsafe; \
    tar --zstd -xf /tmp/finsafe.tar.zst --strip-components=1 -C /opt/finsafe; \
    rm /tmp/finsafe.tar.zst

FROM ${WANWU_WGA_SANDBOX_IMAGE}
USER root
ARG TARGETARCH
RUN if [ "$TARGETARCH" != amd64 ]; then \
      echo 'This integration uses the official Linux x64 runtime; build for linux/amd64.' >&2; \
      exit 1; \
    fi
RUN apt-get update && apt-get install -y --no-install-recommends bubblewrap ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=finsafe-runtime /opt/finsafe/ /opt/finsafe/
COPY opencode.conf /opt/gem/supervisord/opencode.conf
COPY opencode.yaml /etc/finsafe/opencode.yaml
COPY eino-agent.conf /opt/gem/supervisord/eino-agent.conf
COPY eino-agent.yaml /etc/finsafe/eino-agent.yaml
RUN mkdir -p /home/root/workspace /home/root/.config/opencode \
      /home/root/.cache/opencode /home/root/.local/share/opencode \
      /home/root/.local/state/opencode /home/root/.npm && \
    /opt/finsafe/finsafe --version
HEALTHCHECK --interval=15s --timeout=5s --start-period=60s --retries=3 \
    CMD curl -fsS http://127.0.0.1:4096/global/health >/dev/null && curl -fsS http://127.0.0.1:4097/health >/dev/null && nc -z localhost 8080
