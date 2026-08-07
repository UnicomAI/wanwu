<template>
  <div>
    <el-popover
      ref="versionPopover"
      placement="bottom"
      width="400"
      trigger="click"
      popper-class="version-popover"
      @hide="onPopoverHide"
    >
      <VersionTimeLine
        ref="versionTimeLine"
        :appId="appId"
        :appType="appType"
        v-on="$listeners"
      />

      <i
        slot="reference"
        :class="['version-popover-trigger', iconClass]"
        :style="iconStyle"
      />
    </el-popover>
  </div>
</template>

<script>
import VersionTimeLine from '@/components/versionTimeLine';
export default {
  name: 'VersionPopover',
  components: {
    VersionTimeLine,
  },
  props: {
    appId: {
      type: String,
      required: true,
    },
    appType: {
      type: String,
      required: true,
    },
    iconStyle: {
      type: Object,
      default: () => ({
        margin: '13px 12px',
        fontSize: '30px',
        color: '#5983ff',
        cursor: 'pointer',
      }),
    },
    iconClass: {
      type: String,
      default: 'el-icon-time',
    },
  },
  methods: {
    onPopoverHide() {
      this.$nextTick(() => {
        const popover = document.querySelector('.version-popover');
        if (popover && popover.contains(document.activeElement)) {
          document.activeElement.blur();
        }
      });
    },
  },
};
</script>

<style scoped>
.version-popover-trigger {
  pointer-events: auto !important;
}
</style>
