<template>
  <el-dialog
    :visible.sync="dialogVisible"
    title=""
    width="460px"
    append-to-body
    :close-on-click-modal="false"
    :before-close="() => handleClose(true)"
    :show-close="true"
    custom-class="channel-scan-dialog"
  >
    <!-- 扫码中 -->
    <div v-if="status === 'scanning'" class="scan-content">
      <div class="scan-header">
        <img class="scan-icon" :src="iconImg" alt="" />
        <span class="scan-title">{{ scanTitle }}</span>
      </div>
      <div ref="qrCodeRef" class="qr-code-wrap">
        <div class="qr-code-inner"></div>
      </div>
      <p class="scan-hint">
        {{ $t('channel.scanDialog.hint', { time: expireTime }) }}
      </p>
      <div class="polling-status">
        <i class="el-icon-loading polling-icon"></i>
        <span>{{ $t('channel.scanDialog.waiting') }}</span>
      </div>
      <div class="dialog-actions">
        <el-button plain @click="handleRefreshQrCode" :loading="refreshLoading">
          <i class="el-icon-refresh-right"></i>
          {{ $t('channel.scanDialog.refresh') }}
        </el-button>
        <el-button type="default" @click="handleCancel">
          {{ $t('common.button.cancel') }}
        </el-button>
      </div>
    </div>

    <!-- 过期/失败 -->
    <div
      v-else-if="status === 'expired' || status === 'failed'"
      class="result-content"
    >
      <i class="result-icon fail-icon el-icon-close"></i>
      <p class="result-msg">
        {{
          errorMsg ||
          (status === 'expired'
            ? this.$t('channel.scanDialog.expiredMsg')
            : $t('channel.scanDialog.failedMsg'))
        }}
      </p>
      <div class="dialog-actions" style="margin-top: 25px">
        <el-button type="primary" @click="handleReloadQrCode">
          {{ $t('channel.scanDialog.retry') }}
        </el-button>
      </div>
    </div>

    <!-- 成功 -->
    <div v-else-if="status === 'success'" class="result-content">
      <i class="result-icon success-icon el-icon-circle-check"></i>
      <p class="result-title">{{ successTitle }}</p>
      <div class="dialog-actions" style="margin-top: 25px">
        <el-button type="primary" @click="handleDone">
          {{ $t('channel.scanDialog.done') }}
        </el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script>
import QRCode from 'qrcodejs2';
import { fetchScanQrCode, pollScanStatus, finishPollScan } from '@/api/channel';
import { WECHAT } from '../constants';

export default {
  name: 'ChannelScanDialog',
  data() {
    return {
      dialogVisible: false,
      status: 'scanning', // scanning | expired | failed | success
      channelType: WECHAT,
      taskId: '',
      refreshLoading: false,
      qrInstance: null,
      pollTimer: null,
      pollInterval: 2000,
      countdownTimer: null,
      errorMsg: '',
      expireTime: 300,
    };
  },
  computed: {
    iconImg() {
      return this.channelType === WECHAT
        ? require('@/assets/imgs/wechat.png')
        : require('@/assets/imgs/dingtalk.png');
    },
    scanTitle() {
      const key =
        this.channelType === WECHAT ? 'wechatScanTitle' : 'dingtalkScanTitle';
      return this.$t(`channel.scanDialog.${key}`);
    },
    successTitle() {
      const key =
        this.channelType === WECHAT ? 'wechatSuccess' : 'dingtalkSuccess';
      return this.$t(`channel.scanDialog.${key}`);
    },
  },
  beforeDestroy() {
    this.stopPoll();
    this.stopCountdown();
  },
  methods: {
    /**
     * 打开扫码弹窗
     * @param {string} channelType - wechat | dingtalk
     */
    open(channelType) {
      this.channelType = channelType || WECHAT;

      this.status = 'scanning';
      this.dialogVisible = true;

      this.$nextTick(() => {
        this.doFetchAndScan();
      });
    },

    /** 获取二维码并开始扫描流程（打开/刷新/重试共用） */
    async doFetchAndScan() {
      this.refreshLoading = true;

      try {
        const res = await fetchScanQrCode(this.channelType);
        const { qrUrl, sessionId, expireTime } = res.data || res;

        this.qrUrl = qrUrl || '';
        this.taskId = sessionId || '';
        this.expireTime = expireTime || 0;

        // 切换回扫码状态并渲染
        this.status = 'scanning';
        this.renderQRCode(this.qrUrl);
        this.startCountdown();
        this.startPoll();
      } catch (e) {
        this.setStatus('failed');
      } finally {
        this.refreshLoading = false;
      }
    },

    renderQRCode(url) {
      if (this.qrInstance) {
        this.$refs.qrCodeRef && (this.$refs.qrCodeRef.innerHTML = '');
        this.qrInstance = null;
      }
      if (!url || !this.$refs.qrCodeRef) return;
      this.qrInstance = new QRCode(this.$refs.qrCodeRef, {
        text: url,
        width: 180,
        height: 180,
        colorDark: '#000000',
        colorLight: '#ffffff',
        correctLevel: QRCode.CorrectLevel.M,
      });
    },

    startCountdown() {
      this.stopCountdown();
      this.countdownTimer = setInterval(() => {
        if (this.expireTime > 0) {
          this.expireTime--;
        } else {
          this.stopCountdown();
          if (this.status === 'scanning') {
            this.setStatus('expired');
          }
        }
      }, 1000);
    },

    stopCountdown() {
      if (this.countdownTimer) {
        clearInterval(this.countdownTimer);
        this.countdownTimer = null;
      }
    },

    startPoll() {
      this.stopPoll();
      if (!this.taskId) return;
      this.pollTimer = setInterval(async () => {
        try {
          const res = await pollScanStatus(this.channelType, this.taskId);
          const { status, error } = res.data || res;
          this.errorMsg = error || '';

          if (status === 'success') {
            this.setStatus('success');
            this.$emit('success', res.data || res);
          } else if (status === 'expired') {
            this.setStatus('expired');
          } else if (status === 'error') {
            this.setStatus('failed');
          }
        } catch (e) {
          this.setStatus('failed');
        }
      }, this.pollInterval);
    },

    async finishPollScan() {
      if (!this.taskId) return;
      await finishPollScan(this.channelType, this.taskId);
    },

    stopPoll() {
      if (this.pollTimer) {
        clearInterval(this.pollTimer);
        this.pollTimer = null;
      }
    },

    setStatus(newStatus) {
      this.status = newStatus;
      this.stopPoll();
      this.stopCountdown();
    },

    handleRefreshQrCode() {
      this.doFetchAndScan();
    },

    handleReloadQrCode() {
      this.status = 'scanning';
      this.$nextTick(() => {
        this.doFetchAndScan();
      });
    },

    handleDone() {
      this.handleClose();
    },

    handleCancel() {
      this.handleClose(true);
    },

    handleClose(cancel) {
      this.dialogVisible = false;
      this.cleanup();
      if (cancel) this.finishPollScan();
    },

    cleanup() {
      this.stopPoll();
      this.stopCountdown();
      if (this.qrInstance) {
        this.$refs.qrCodeRef && (this.$refs.qrCodeRef.innerHTML = '');
        this.qrInstance = null;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.scan-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 0 10px;
}
.scan-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 22px;
}
.scan-icon {
  width: 26px;
  height: 26px;
  border-radius: 50%;
}
.scan-title {
  font-size: 17px;
  font-weight: 600;
  color: #303133;
}
.qr-code-wrap {
  width: 200px;
  height: 200px;
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  .qr-code-inner {
    ::v-deep img,
    ::v-deep canvas {
      width: 100% !important;
      height: 100% !important;
    }
  }
}
.scan-hint {
  font-size: 14px;
  color: #606266;
  margin-top: 18px;
}
.polling-status {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  font-size: 13px;
  color: #67c23a;
  .polling-icon {
    animation: rotating 1.5s linear infinite;
    font-size: 15px;
  }
}

.result-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 50px 0 30px;
}
.result-icon {
  font-size: 52px;
  margin-bottom: 10px;
  border-radius: 50%;
  padding: 8px;
}
.fail-icon {
  font-size: 26px;
  color: #f56c6c;
  background: #fef0f0;
  border: 2px solid #fde2e2;
}
.success-icon {
  color: #67c23a;
}
.result-msg {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}
.result-title {
  font-size: 18px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 8px;
}
.result-desc {
  font-size: 13px;
  color: #909399;
  line-height: 1.6;
  margin-bottom: 28px;
  max-width: 320px;
  text-align: center;
}
.dialog-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>

<style lang="scss">
.channel-scan-dialog {
  border-radius: 12px;
  overflow: hidden;
  .el-dialog__header {
    padding: 18px 24px 0;
    font-size: 16px;
    font-weight: 700;
  }
  .el-dialog__body {
    padding: 0 30px 28px;
  }
}
</style>
