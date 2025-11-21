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

        // Prefer modern Clipboard API
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(text);
        } else {
          // Fallback: Create input and use execCommand
          const input = document.createElement('input');
          input.value = text;
          input.setAttribute('readonly', '');
          input.style.cssText = 'position: absolute; left: -9999px;';
          document.body.appendChild(input);
          input.select();
          document.execCommand('copy');
          document.body.removeChild(input);
        }

        this.$message.success(this.$t('common.copy.success'));
      } catch (err) {
        console.error('Copy failed:', err);
        this.$message.error('Copy failed, please copy manually');
      }
    }
  }
}
</script>

<style scoped>

</style>