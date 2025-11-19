<template>
  <el-button v-bind="$attrs" @click="handleCopy" class="copy-icon">
    <i v-if="showIcon" class="el-icon-document-copy"></i>
    {{ $t('common.button.copy') }}
  </el-button>
</template>

<script>
export default {
  name: "CopyIcon",
  inheritAttrs: false,
  props: {
    text: {
      type: String,
      required: true
    },
    showIcon: {
      type: Boolean,
      default: true
    },
  },
  methods: {
    async handleCopy() {
      try {
        const text = this.text;

        // 优FirstUse现代 Clipboard API
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(text);
        } else {
          // Downgrade方案：Create input 并Use execCommand
          const input = document.createElement('input');
          input.value = text;
          input.setAttribute('readonly', '');
          input.style.cssText = 'position: absolute; left: -9999px;';
          document.body.appendChild(input);
          input.select();
          document.execCommand('copy');
          document.body.removeChild(input);
        }

        this.$message.success('已Copy到剪贴板');
      } catch (err) {
        console.error('Copy failed:', err);
        this.$message.error('Copy failed，Please手动Copy');
      }
    }
  }
}
</script>

<style scoped>

</style>