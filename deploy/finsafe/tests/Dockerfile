FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive LOG_DIR=/var/log/gem USER=root \
    OPENCODE_DISABLE_MODELS_FETCH=true OPENCODE_DISABLE_DEFAULT_PLUGINS=true
RUN apt-get update && apt-get install -y --no-install-recommends \
      bubblewrap ca-certificates curl git iproute2 libstdc++6 python3 supervisor && \
    rm -rf /var/lib/apt/lists/*
COPY opencode.conf /opt/gem/supervisord/opencode.conf
COPY opencode.yaml /etc/finsafe/opencode.yaml
COPY tests/supervisord.conf /etc/supervisor/supervisord.conf
RUN mkdir -p /var/log/gem /root/.opencode/bin /opt/finsafe \
      /home/root/workspace /home/root/.config/opencode /home/root/.cache/opencode \
      /home/root/.local/share/opencode /home/root/.local/state/opencode /home/root/.npm
ENTRYPOINT ["/usr/bin/supervisord", "-n", "-c", "/etc/supervisor/supervisord.conf"]
