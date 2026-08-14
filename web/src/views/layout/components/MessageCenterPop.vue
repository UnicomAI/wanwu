<template>
  <section class="message-center-pop">
    <header class="message-center-pop__header">
      <div class="message-center-pop__title">
        {{ $t('messageCenter.noUnreadMessages') }}({{ totalUnread }})
      </div>
      <el-popconfirm
        :title="$t('messageCenter.markAllAsReadConfirm')"
        @confirm="handleMarkAllAsRead"
      >
        <el-tooltip
          slot="reference"
          :content="$t('messageCenter.markAllAsRead')"
          placement="top"
        >
          <span class="message-center-pop__clear-tooltip">
            <button
              type="button"
              class="message-center-pop__clear"
              :disabled="!unreadMessages.length"
            >
              <svg-icon class="icon" icon-class="chatClear" />
            </button>
          </span>
        </el-tooltip>
      </el-popconfirm>
      <button
        type="button"
        class="message-center-pop__entry"
        @click="handleToMessageCenter"
      >
        {{ $t('messageCenter.goToMessageCenter') }}
      </button>
    </header>

    <div class="message-center-pop__divider" />

    <div
      v-if="loading || unreadMessages.length"
      v-loading="loading"
      :class="[
        'message-center-pop__list',
        {
          'message-center-pop__list--loading':
            loading && !unreadMessages.length,
        },
      ]"
    >
      <template v-if="unreadMessages.length">
        <button
          v-for="message in unreadMessages"
          :key="message.id"
          type="button"
          class="message-center-pop__item"
          @click="$emit('select', message)"
        >
          <div class="message-center-pop__item-main">
            <div class="message-center-pop__item-title">
              {{ message.title }}
            </div>
            <div class="message-center-pop__item-content">
              {{ formatMessageText(message.msgText) }}
            </div>
          </div>
          <time class="message-center-pop__item-time">
            {{ message.updateAt }}
          </time>
        </button>
        <div class="message-center-pop__more">
          <div class="message-center-pop__more-divider" />
          <p>{{ $t('messageCenter.moreMsgTips') }}</p>
        </div>
      </template>
    </div>
    <div v-else class="message-center-pop__empty">
      <img src="@/assets/imgs/empty.png" />
      <p>{{ $t('messageCenter.noUnread') }}</p>
    </div>
  </section>
</template>

<script>
import svgIcon from '@/components/svgIcon.vue';
export default {
  components: { svgIcon },
  name: 'MessageCenterPop',
  props: {
    unreadMessages: {
      type: Array,
      default: () => [],
    },
    totalUnread: {
      type: Number,
      default: 0,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  methods: {
    formatMessageText(msgText) {
      return String(msgText || '').replace(/\$\$\{([\s\S]*?)\}\$\$/g, '$1');
    },
    handleMarkAllAsRead() {
      this.$emit('clear');
    },
    handleToMessageCenter() {
      this.$router.push({
        path: '/messageCenter',
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.message-center-pop {
  color: #303133;

  &__header {
    display: flex;
    align-items: center;
    height: 26px;
    line-height: 26px;
  }

  &__title {
    font-size: 14px;
    font-weight: 500;
  }

  &__clear,
  &__entry {
    padding: 0;
    border: 0;
    background: transparent;
    color: #606266;
    font-size: 13px;
    line-height: 26px;
    cursor: pointer;
  }

  &__clear {
    margin-left: 8px;
    color: $color;

    &:disabled {
      color: #c0c4cc;
      cursor: not-allowed;
    }
  }

  &__entry {
    margin-left: auto;
    color: #303133;

    &:hover {
      color: $color;
    }
  }

  &__divider {
    height: 1px;
    margin: 8px -12px 0;
    background: #ebeef5;
  }

  &__list {
    max-height: 190px;
    overflow-y: auto;
  }

  &__list--loading {
    min-height: 180px;
  }

  &__item {
    display: flex;
    width: 100%;
    padding: 14px 4px;
    border: 0;
    border-bottom: 1px solid #f2f4f7;
    background: transparent;
    color: inherit;
    text-align: left;
    cursor: pointer;

    &:hover {
      background: #f7f9fc;
    }
  }

  &__item-main {
    min-width: 0;
    flex: 1;
  }

  &__item-title,
  &__item-content,
  &__item-time {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__item-title {
    color: #303133;
    font-size: 14px;
    font-weight: 500;
    line-height: 20px;
  }

  &__item-content {
    margin-top: 4px;
    color: #909399;
    font-size: 13px;
    line-height: 18px;
  }

  &__item-time {
    width: 142px;
    margin-left: 12px;
    color: #909399;
    font-size: 13px;
    line-height: 20px;
    text-align: right;
  }

  &__more {
    padding: 12px 0 8px;
    color: #909399;
    font-size: 13px;
    text-align: center;

    p {
      margin: 8px 0 0;
    }
  }

  &__more-divider {
    height: 1px;
    background: #ebeef5;
  }

  &__status,
  &__empty {
    display: flex;
    min-height: 180px;
    align-items: center;
    justify-content: center;
    color: #909399;
    font-size: 13px;
  }

  &__empty {
    flex-direction: column;

    img {
      width: 96px;
      height: 96px;
      object-fit: contain;
    }

    p {
      margin: 8px 0 0;
    }
  }
}
</style>
