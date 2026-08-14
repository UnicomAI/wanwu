<template>
  <div class="message-center-menu-icon">
    <svg-icon class="icon" :icon-class="iconClass" />
    <i v-if="hasUnread" class="message-center-menu-icon__dot" />
  </div>
</template>

<script>
export default {
  name: 'MessageCenterMenuIcon',
  props: {
    iconClass: {
      type: String,
      required: true,
    },
    hasUnread: {
      type: Boolean,
      default: false,
    },
  },
};
</script>

<style lang="scss" scoped>
.message-center-menu-icon {
  position: relative;
  width: auto;
  height: auto;
  color: #77869e;
  font-size: 16px;
  display: inline-flex !important;
  align-items: center;
  vertical-align: unset !important;

  .icon {
    font-size: 16px;
  }

  &__dot {
    position: absolute;
    top: -3px;
    right: -4px;
    width: 6px;
    height: 6px;
    border: 1px solid #fff;
    border-radius: 50%;
    background: #f56c6c;
  }
}
</style>
