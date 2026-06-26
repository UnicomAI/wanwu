<template>
  <div class="card" @click.stop="handleClick">
    <div class="card-title">
      <img
        class="common-card-logo"
        v-if="info.avatar && info.avatar.path"
        :src="avatarSrc(info.avatar.path)"
      />
      <div class="mcp_detailBox">
        <span class="mcp_name" :title="info.name">
          {{ info.name }}
          <i
            v-if="visibleVersionInfo && info.isPublished"
            class="el-icon-success published-icon"
          />
          <el-tag
            v-if="info.isClosedSource"
            size="mini"
            type="warning"
            class="closed-source-tag"
          >
            {{ $t('tempSquare.skills.closedSource') }}
          </el-tag>
        </span>
        <span class="mcp_from from_tag">
          <label :title="`${$t('tempSquare.author')}：${info.author || ''}`">
            {{ $t('tempSquare.author') }}：{{ info.author }}
          </label>
          <label
            v-if="visibleVersionInfo && info.isPublished"
            class="version-label"
            :title="`${info.version}`"
          >
            {{ info.version }}
          </label>
        </span>
      </div>
    </div>
    <div class="card-des">{{ info.desc }}</div>
    <div class="card-bottom">
      <div class="card-bottom-left"></div>
      <div class="card-bottom-right">
        <div v-if="visibleVersionInfo && info.isPublished" class="publishType">
          <span v-if="info.publishType === 'private'" class="publishType-tag">
            <span class="el-icon-lock"></span>
            {{ $t('appSpace.private') }}
          </span>
          <span v-else class="publishType-tag">
            <span class="el-icon-unlock"></span>
            {{ $t('appSpace.public') }}
          </span>
        </div>
        <slot
          name="operations"
          v-bind="{
            info,
            type,
            downloadTemplate,
            sendToResource,
            handleDelete,
          }"
        >
          <el-tooltip
            v-if="type === 4"
            :content="$t('tempSquare.skills.save')"
            placement="top"
          >
            <i class="el-icon-collection" @click.stop="sendToResource"></i>
          </el-tooltip>

          <!-- 自定义类型显示更多操作 -->
          <el-tooltip
            v-if="type === 3"
            :content="$t('tempSquare.download')"
            placement="top"
          >
            <i class="el-icon-download" @click.stop="downloadTemplate"></i>
          </el-tooltip>

          <el-dropdown v-if="[2, 3].includes(type)" placement="bottom">
            <span class="el-dropdown-link">
              <i class="el-icon-more" @click.stop />
            </span>
            <el-dropdown-menu slot="dropdown" style="margin-top: -10px">
              <el-dropdown-item
                v-if="type === 2 && info.isPublished"
                @click.native="handlePublishConfig"
              >
                {{ $t('appSpace.publishSet') }}
              </el-dropdown-item>
              <el-dropdown-item @click.native="handleDelete">
                {{ $t('common.button.delete') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </slot>
      </div>
    </div>
  </div>
</template>

<script>
import { SKILLADDED, SKILLCUSTOM, SKILLBUILTIN } from '../constants';
import { avatarSrc } from '@/utils/util';

export default {
  name: 'SkillCard',
  props: {
    info: {
      type: Object,
      default: () => ({}),
    },
    // 1:内置，2:自定义，3:我添加, 4:广场skill
    type: {
      type: Number,
      default: 1,
    },
    visibleVersionInfo: {
      type: Boolean,
      default: false,
    },
  },
  methods: {
    avatarSrc,
    handleClick() {
      const hasClickListener = this.$listeners && this.$listeners.click;
      if (hasClickListener) {
        this.$emit('click', this.info);
        return;
      }
      if (this.type === 2) {
        // 我创建的skill直接进入工作台编辑
        const { skillId } = this.info || {};
        const path = '/skill/workshop';
        this.$router.push({
          path,
          query: {
            chatType: 'skill',
            customSkillId: skillId,
          },
        });
      } else {
        const path = '/skill/detail';
        const typeMap = {
          1: SKILLBUILTIN,
          3: SKILLADDED,
        };
        const type = typeMap[this.type];
        this.$router.push({
          path,
          query: { templateSquareId: this.info.skillId, type },
        });
      }
    },
    downloadTemplate() {
      this.$emit('download', this.info);
    },
    handleDelete() {
      this.$emit('delete', this.info);
    },
    handlePublishConfig() {
      this.$emit('publishConfig', this.info);
    },
    // 发送到资源库
    sendToResource() {
      this.$emit('sendToResource', this.info);
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tempSquare.scss';

.card {
  position: relative;
  padding: 20px 16px;
  border-radius: 12px;
  height: fit-content;
  background: #fff url('@/assets/imgs/card_bg.png');
  background-size: 100% 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: calc((100% / 4) - 20px);
  margin: 0 10px 20px;
  box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(0, 0, 0, 0);
  &:hover {
    cursor: pointer;
    box-shadow:
      0 2px 8px #171a220d,
      0 4px 16px #0000000f;
    border: 1px solid $border_color;

    .action-icon {
      display: block;
    }
  }
  .card-title {
    display: flex;
    width: 100%;
    padding-bottom: 7px;
    .svg-icon {
      width: 50px;
      height: 50px;
    }
    .mcp_detailBox {
      width: calc(100% - 70px);
      margin-left: 10px;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      padding: 0 0 3px 0;
      .mcp_name {
        min-height: 22px;
        display: block;
        font-size: 15px;
        font-weight: 700;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
        color: $create_card_text_color;
        line-height: 1;
      }
      .mcp_from {
        label {
          padding: 3px 7px;
          font-size: 12px;
          color: $tag_color;
          background: $tag_bg;
          border-radius: 3px;
          display: block;
          height: 22px;
          width: 100%;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }
    }

    margin-bottom: 13px;
  }
  .card-des {
    width: 100%;
    display: -webkit-box;
    text-overflow: ellipsis;
    color: #5d5d5d;
    font-weight: 400;
    overflow: hidden;
    /*! autoprefixer: ignore next */
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    font-size: 13px;
    height: auto;
    word-wrap: break-word;
  }
  .card-bottom {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 14px;
    margin-bottom: -6px;
    .card-bottom-left {
      color: #888;
    }
    .card-bottom-right {
      i {
        margin-left: 5px;
        cursor: pointer;
      }
    }
  }
}

.card-bottom-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.published-icon {
  color: #67c23a;
  margin-left: 4px;
  font-size: 14px;
  vertical-align: middle;
}

.author-label {
  flex-shrink: 0;
}

.version-label {
  margin-left: 8px;
}

.publishType {
  &-tag {
    display: inline-flex;
    align-items: center;
    background: #f0f2f5;
    color: #909399;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 12px;
    line-height: 1.5;

    .el-icon-lock,
    .el-icon-unlock {
      margin-right: 4px;
    }
  }
}

.card-title .mcp_detailBox .mcp_from {
  display: flex !important;
  align-items: center;
  label {
    width: auto !important;
    max-width: 50%;
    flex-shrink: 1;
  }
}

.closed-source-tag {
  margin-left: 6px;
  vertical-align: middle;
}
</style>
