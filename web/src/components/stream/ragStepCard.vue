<template>
  <div
    :class="['rag-step-card', { error: status === 'error' }]"
    :data-rag-step-type="type"
  >
    <div
      :class="[
        'rag-step-header',
        { streaming: status === 'running', error: status === 'error' },
      ]"
      @click="toggleExpand"
    >
      <div class="header-left">
        <i :class="isExpanded ? 'el-icon-arrow-up' : 'el-icon-arrow-down'"></i>
        <el-tooltip
          :disabled="status !== 'error' || !errorMessage"
          effect="dark"
          placement="top"
        >
          <template #content>
            <div class="rag-step-error-tooltip">{{ errorMessage }}</div>
          </template>
          <div
            :class="[
              'step-icon-wrapper',
              { 'session-error-icon': status === 'error' },
            ]"
          >
            <i :class="['step-icon', iconClass]"></i>
          </div>
        </el-tooltip>
        <span class="step-title">{{ title }}</span>
      </div>
      <div class="header-right">
        <span class="step-time">{{ displayDuration }}</span>
        <div v-if="status === 'running'" class="status-badge running">
          <span class="status-dot"></span>
          <span class="status-text">{{ $t('rag.step.running') }}</span>
        </div>
        <div v-else-if="status !== 'error'" class="status-badge completed">
          <span class="status-dot"></span>
          <span class="status-text">{{ $t('rag.step.completed') }}</span>
        </div>
      </div>
    </div>
    <el-collapse-transition>
      <div v-show="isExpanded" class="rag-step-body">
        <div class="rag-step-content rag-thinking-body">
          <slot></slot>
        </div>
      </div>
    </el-collapse-transition>
  </div>
</template>

<script>
/**
 * RagStepCard — RAG 问答过程卡片
 *
 * 视觉复刻 views/generalAgent/components/ActivityBlock.vue，零跨目录依赖，
 * 专门承载 RAG 的"知识库检索"与"思考过程"两步。
 *
 * 流式中（status=running）自动展开 + 内部计时；
 * 完成（status 变 done）后自动折叠、停止计时、展示 props.duration。
 */
export default {
  name: 'RagStepCard',
  props: {
    type: {
      type: String,
      required: true, // 'qa_search' | 'knowledge_search' | 'thinking'
    },
    status: {
      type: String,
      default: 'running', // 'running' | 'done'
    },
    startAt: {
      type: Number,
      default: 0,
    },
    duration: {
      type: String,
      default: '', // 完成态最终耗时，父组件算好（e.g. '2.982s'）
    },
    errorMessage: {
      type: String,
      default: '',
    },
    defaultExpanded: {
      type: Boolean,
      default: true,
    },
    // 外部驱动收起：thinking 卡片由父组件在"正式回答开始流式"时置 true
    shouldCollapse: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      isExpanded: this.defaultExpanded,
      liveTimerMs: 0,
      timerHandle: null,
    };
  },
  computed: {
    title() {
      if (this.type === 'qa_search') {
        return this.$t('rag.step.qaSearch');
      }
      if (this.type === 'knowledge_search') {
        return this.$t('rag.step.knowledgeSearch');
      }
      if (this.type === 'thinking') {
        return this.$t('rag.step.thinking');
      }
      return this.type;
    },
    iconClass() {
      // 使用 element-ui 内置 icon，避免新增图片资源
      if (this.status === 'error') return 'el-icon-warning-outline';
      if (this.type === 'qa_search') return 'el-icon-chat-line-square';
      if (this.type === 'knowledge_search') return 'el-icon-document';
      return 'el-icon-cpu';
    },
    // 展示耗时：running 时用内部计时器，done 时用父组件传入的 duration
    displayDuration() {
      if (this.status === 'running') {
        return `${(this.liveTimerMs / 1000).toFixed(3)}s`;
      }
      return this.duration || '';
    },
  },
  watch: {
    status: {
      immediate: true,
      handler(val) {
        if (val === 'running') {
          this.isExpanded = true;
          this.startLiveTimer();
        } else if (val === 'error') {
          this.stopLiveTimer();
          this.isExpanded = false;
        } else {
          // done：停止计时；qa_search / knowledge_search 立即收起，thinking 等 shouldCollapse 信号
          this.stopLiveTimer();
          if (this.type !== 'thinking') {
            this.isExpanded = false;
          }
        }
      },
    },
    shouldCollapse(val) {
      if (val) this.isExpanded = false;
    },
  },
  beforeDestroy() {
    this.stopLiveTimer();
  },
  methods: {
    toggleExpand() {
      this.isExpanded = !this.isExpanded;
    },
    startLiveTimer() {
      this.stopLiveTimer();
      const base = this.startAt || Date.now();
      this.liveTimerMs = Date.now() - base;
      this.timerHandle = setInterval(() => {
        this.liveTimerMs = Date.now() - base;
      }, 100);
    },
    stopLiveTimer() {
      if (this.timerHandle) {
        clearInterval(this.timerHandle);
        this.timerHandle = null;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.rag-step-card {
  margin-bottom: 12px;
}

.rag-step-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-radius: 12px;
  border: 1px solid rgba(99, 102, 241, 0.2);
  background: linear-gradient(
    135deg,
    rgba(99, 102, 241, 0.05) 0%,
    #fafafa 100%
  );
  cursor: pointer;
  user-select: none;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);

  &:hover {
    background: linear-gradient(180deg, #f5f7f9 0%, #f3f4f6 100%);
  }

  &.error {
    border-color: rgba(245, 108, 108, 0.35);
    background: linear-gradient(
      135deg,
      rgba(245, 108, 108, 0.08) 0%,
      #fafafa 100%
    );
    box-shadow: 0 4px 12px rgba(245, 108, 108, 0.1);

    .step-title,
    .step-icon {
      color: #d93025;
    }

    .step-icon {
      animation: none;
    }
  }
  &.streaming {
    border-color: rgba(99, 102, 241, 0.4);
    box-shadow: 0 4px 12px rgba(99, 102, 241, 0.12);

    .step-title {
      color: #4f46e5;
      font-weight: 600;
    }

    .step-icon {
      animation: rag-step-spin 1.2s linear infinite;
      color: #4f46e5;
    }
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1;
    min-width: 0;

    i.el-icon-arrow-down,
    i.el-icon-arrow-up {
      color: #9ca3af;
      font-size: 12px;
      transition: transform 0.2s ease;
    }

    .step-icon-wrapper {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 22px;
      height: 22px;

      &.session-error-icon {
        border-radius: 50%;
        background: rgba(245, 108, 108, 0.1);
        border: 1px solid rgba(245, 108, 108, 0.22);
        color: #d93025;

        .step-icon {
          color: inherit;
          font-size: 13px;
          font-weight: bold;
        }
      }
      .step-icon {
        font-size: 16px;
        color: #6b7280;
      }
    }

    .step-title {
      font-size: 14px;
      font-weight: 600;
      color: #1f2937;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;

    .step-time {
      font-size: 13px;
      color: #6b7280;
      font-variant-numeric: tabular-nums;
      background: rgba(0, 0, 0, 0.04);
      padding: 3px 8px;
      border-radius: 6px;
    }

    .status-badge {
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 3px 10px;
      border-radius: 12px;
      font-size: 12px;
      font-weight: 500;

      .status-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
      }

      &.running {
        background: rgba(99, 102, 241, 0.1);
        border: 1px solid rgba(99, 102, 241, 0.2);

        .status-dot {
          background: #6366f1;
          animation: rag-step-pulse 1.5s infinite;
          box-shadow: 0 0 6px rgba(99, 102, 241, 0.5);
        }

        .status-text {
          color: #4f46e5;
        }
      }

      &.completed {
        background: rgba(16, 163, 127, 0.1);
        border: 1px solid rgba(16, 163, 127, 0.2);

        .status-dot {
          background: #10a37f;
          box-shadow: 0 0 4px rgba(16, 163, 127, 0.4);
        }

        .status-text {
          color: #0d8a6a;
        }
      }
    }
  }
}

.rag-step-card.error .rag-step-body {
  background: linear-gradient(
    180deg,
    rgba(245, 108, 108, 0.08) 0%,
    rgba(245, 108, 108, 0.03) 100%
  );
  border-color: rgba(245, 108, 108, 0.22);
}
.rag-step-body {
  position: relative;
  margin: 10px 0 4px;
  padding: 14px 16px;
  background: linear-gradient(
    180deg,
    rgba(99, 102, 241, 0.07) 0%,
    rgba(99, 102, 241, 0.03) 100%
  );
  border: 1px solid rgba(99, 102, 241, 0.16);
  border-radius: 10px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.6);
}

.rag-step-content {
  font-size: 13px;
  color: #334155; // slate-700：在淡紫底上有真实对比，不再洗白
  line-height: 1.75;
  word-break: break-word;
}

// 覆盖全局 .chunk_stable / .chunk_active 可能带来的错位，确保卡片内排版受局部控制
.rag-thinking-body {
  ::v-deep {
    .chunk_stable,
    .chunk_active {
      margin: 0;
      padding: 0;
    }

    > *:first-child {
      margin-top: 0;
    }
    > *:last-child {
      margin-bottom: 0;
    }

    p {
      margin: 6px 0;
    }

    ul,
    ol {
      padding-left: 20px;
      margin: 6px 0;
    }

    li {
      margin: 2px 0;
    }

    strong {
      color: #1e293b; // slate-800，在淡紫底上做真实强调
      font-weight: 600;
    }

    code {
      background: rgba(99, 102, 241, 0.12);
      color: #4338ca; // indigo-700，提一档与新背景拉开
      padding: 1px 5px;
      border-radius: 4px;
      font-size: 12px;
    }

    // 图片尺寸约束：LLM 原样嵌入的知识库图片不能撑破卡片
    img {
      max-width: 100%;
      max-height: 240px;
      height: auto;
      object-fit: contain;
      display: block;
      border-radius: 6px;
      margin: 6px 0;
    }

    // 方案 A：中和 LLM 原文里常见的视觉噪音
    // - 中文斜体（<em>）观感差，强制不斜
    // - 上标/下标（<sup>/<sub>）保持基线，避免 "上下文^1^" 把 1 缩成上标
    em {
      font-style: normal;
    }
    sup,
    sub {
      vertical-align: baseline;
      font-size: inherit;
      line-height: inherit;
    }
  }
}

@keyframes rag-step-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes rag-step-pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
