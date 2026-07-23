<template>
  <div class="admin-avatar" :style="avatarStyle">
    <img class="admin-avatar__image" :src="avatarUrl" alt="" />
  </div>
</template>

<script>
import { avatarSrc } from '@/utils/util';

export default {
  name: 'AdminAvatar',
  props: {
    path: {
      type: String,
      default: '',
    },
    localPath: {
      type: String,
      default: '',
    },
    size: {
      type: [Number, String],
      default: 26,
    },
  },
  computed: {
    avatarUrl() {
      return (
        this.localPath ||
        avatarSrc(this.path, require('@/assets/imgs/avatar_default.png'))
      );
    },
    avatarStyle() {
      const size = Number(this.size) || 26;
      return {
        width: `${size}px`,
        height: `${size}px`,
        flexBasis: `${size}px`,
      };
    },
  },
};
</script>

<style lang="scss" scoped>
.admin-avatar {
  display: inline-flex;
  flex: 0 0 26px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 4px;
  background: #f2f4f7;
}

.admin-avatar__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
