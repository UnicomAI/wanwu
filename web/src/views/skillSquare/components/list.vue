<template>
  <div class="tempSquare-management">
    <div class="tempSquare-content-box tempSquare-third">
      <div class="tempSquare-main">
        <div class="tempSquare-content">
          <div class="tempSquare-card-box">
            <div
              class="card-loading-box scroll-card-container"
              v-if="list && list.length"
            >
              <div class="card-box scroll-card-pr" v-loading="loading">
                <skill-card
                  v-for="(item, index) in list"
                  :key="index"
                  :info="item"
                  :type="4"
                  @download="handleDownload"
                  @click="handleCardClick"
                >
                  <template v-slot:operations="{ info }">
                    <div v-if="!isMineType" class="skill-card-operations">
                      <el-tooltip
                        v-if="showShared"
                        :content="
                          info.isShared
                            ? $t('skillSpace.isShared')
                            : $t('skillSpace.toResource')
                        "
                        placement="top"
                      >
                        <div class="card-btn">
                          <span>
                            {{ formatCount(info.acquiredCount, 2, false) }}
                          </span>
                          <i
                            :class="[
                              'el-icon-s-promotion',
                              { 'is-disabled': info.isShared },
                            ]"
                            @click.stop="handleSendToResource(info)"
                          ></i>
                        </div>
                      </el-tooltip>

                      <el-tooltip
                        v-if="!info.isClosedSource"
                        :content="$t('tempSquare.download')"
                        placement="top"
                      >
                        <div class="card-btn">
                          <span>
                            {{ formatCount(info.downloadCount, 2, false) }}
                          </span>
                          <i
                            class="el-icon-download"
                            @click.stop="handleDownload(info)"
                          ></i>
                        </div>
                      </el-tooltip>
                    </div>
                    <!-- 我发布的--仅展示数量 -->
                    <div class="skill-card-operations" v-else>
                      <el-tooltip
                        v-if="showShared"
                        :content="$t('skillSpace.sharedCount')"
                        placement="top"
                      >
                        <div class="card-btn">
                          <span>
                            {{ formatCount(info.acquiredCount, 2, false) }}
                          </span>
                          <i
                            :class="['el-icon-s-promotion']"
                            @click.stop="() => {}"
                          ></i>
                        </div>
                      </el-tooltip>

                      <el-tooltip
                        :content="$t('skillSpace.downloadCount')"
                        placement="top"
                      >
                        <div class="card-btn">
                          <span>
                            {{ formatCount(info.downloadCount, 2, false) }}
                          </span>
                          <i
                            class="el-icon-download"
                            @click.stop="() => {}"
                          ></i>
                        </div>
                      </el-tooltip>
                    </div>
                  </template>
                </skill-card>
                <div
                  v-if="isBuiltinType"
                  class="card card-item-more"
                  @click="handleLinkMore()"
                >
                  <div class="card-content">
                    <span>{{ $t('skillSpace.list.moreText') }}</span>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="empty">
              <el-empty :description="$t('common.noData')"></el-empty>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import SkillCard from '@/views/templateSquare/skills/card.vue';
import { formatCount } from '@/utils/util';

export default {
  components: { SkillCard },
  props: {
    type: {
      type: [String, Number],
      default: '',
    },
    list: {
      type: Array,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    },
    showShared: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      basePath: this.$basePath,
      templateUrl: '',
    };
  },
  computed: {
    isBuiltinType() {
      return this.type === 'builtin';
    },
    isMineType() {
      return this.type === 'mine';
    },
  },
  methods: {
    formatCount,
    handleDownload(info) {
      this.$emit('download', info);
    },
    handleSendToResource(info) {
      if (info.isShared) return;
      this.$emit('send-to-resource', info);
    },
    handleLinkMore() {
      this.$emit('link-more');
    },
    handleCardClick(info) {
      this.$emit('card-click', info);
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tempSquare.scss';
.scroll-card-container {
  max-height: calc(100vh - 120px);
}
.tempSquare-management {
  .tempSquare-main {
    padding: 0 !important;
  }
  .card-search-cust {
    justify-content: flex-start;
    margin-top: 10px;
  }

  .card-item-more {
    display: flex;
    height: auto !important;
    justify-content: center;
    align-items: center;
    min-height: 140px;
    .card-content {
      font-size: 16px;
      font-weight: 500;
      color: #5d5d5d;
      &:hover {
        color: $color;
      }
    }
  }

  .skill-card-operations {
    display: flex;
    align-items: center;
    gap: 10px;
    i {
      cursor: pointer;
      &.is-disabled {
        cursor: not-allowed;
        color: #c0c4cc;
      }
    }
    .card-btn {
      display: inline-flex;
      gap: 2px;
      align-items: center;
    }
  }
}
</style>
