<template>
  <div
    class="tempSquare-detail page-wrapper"
    :style="isPublic ? `background: ${bgColor}; min-height: 100%` : ''"
  >
    <!-- 返回按钮，点击抛出 back 事件 -->
    <span class="back" @click="$emit('back')">
      <!-- 可通过 prop 灵活定义返回按钮文字 -->
      <slot name="back-text">
        {{ backText || $t('menu.back') }}
      </slot>
    </span>

    <div class="tempSquare-title">
      <div class="tempSquare-title-left">
        <img
          class="logo"
          v-if="detail.avatar && detail.avatar.path"
          :src="avatarSrc(detail.avatar.path)"
        />
        <div :class="['info', { fold: foldStatus }]">
          <p class="name">{{ detail.name }}</p>
          <p v-if="detail.desc && detail.desc.length > 260" class="desc">
            {{ foldStatus ? detail.desc : detail.desc.slice(0, 268) + '...' }}
            <span class="arrow" v-show="detail.desc.length > 260" @click="fold">
              {{
                foldStatus
                  ? $t('common.button.fold')
                  : $t('common.button.detail')
              }}
            </span>
          </p>
          <p v-else class="desc">{{ detail.desc }}</p>
        </div>
      </div>
      <div style="margin-left: 10px; flex-shrink: 0">
        <!-- 默认的下载按钮，点击时将当前 detail 抛出，由外部调用具体 API 或者下载逻辑 -->
        <el-button
          v-if="visibleDownload"
          type="primary"
          size="mini"
          @click="$emit('download', detail)"
        >
          {{ $t('tempSquare.download') }}
        </el-button>
        <!-- 预留插槽，如果是工作流页面可能需要增加一些像“克隆”一样的按钮 -->
        <slot name="extra-buttons"></slot>
      </div>
    </div>

    <div class="tempSquare-main">
      <div class="left-info">
        <div class="info-config" v-if="visibleVariableConfig">
          <div class="tabs">
            <div class="tab active">
              {{ $t('tempSquare.skills.apiKeyConfig.title') }}
            </div>
          </div>

          <ApiKeyTable
            :dataList="detail.variables || []"
            @create-variable="handleCreateVariable"
            @update-variable="handleUpdateVariable"
            @delete-variable="handleDeleteVariable"
          />
        </div>

        <div class="tabs">
          <div
            :class="['tab', { active: tabActive === 0 }]"
            @click="tabClick(0)"
          >
            {{ $t('square.info') }}
          </div>
          <div
            :class="['tab', { active: tabActive === 1 }]"
            @click="tabClick(1)"
            v-if="visibleHistory"
          >
            {{ $t('tempSquare.skills.historyVersion') }}
          </div>
        </div>

        <div style="padding-top: 10px">
          <template v-if="tabActive === 0">
            <slot name="info-header" />
            <div
              class="overview bg-border"
              v-if="detail.summary || detail.feature || detail.scenario"
            >
              <div class="overview-item" v-if="detail.summary">
                <div class="item-title">
                  <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                  <span>{{ $t('square.summary') }}</span>
                </div>
                <div class="item-desc" v-html="parseTxt(detail.summary)"></div>
              </div>
              <div class="overview-item" v-if="detail.feature">
                <div class="item-title">
                  <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                  <span>{{ $t('square.feature') }}</span>
                </div>
                <div class="item-desc" v-html="parseTxt(detail.feature)"></div>
              </div>
              <div class="overview-item" v-if="detail.scenario">
                <div class="item-title">
                  <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                  <span>{{ $t('square.scenario') }}</span>
                </div>
                <div class="item-desc">
                  <div v-html="parseTxt(detail.scenario)"></div>
                </div>
              </div>
            </div>
            <div class="overview bg-border" v-if="detail.note">
              <div class="overview-item">
                <div class="item-title">
                  <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                  <span>{{ $t('square.note') }}</span>
                </div>
                <div class="item-desc" v-html="parseTxt(detail.note)"></div>
              </div>
            </div>
            <div class="overview" v-if="detail.skillMarkdown">
              <div class="overview-item">
                <div class="item-desc">
                  <div class="tempSquare-markdown">
                    <MdRender :content="detail.skillMarkdown" />
                  </div>
                </div>
              </div>
            </div>
          </template>
          <template v-if="tabActive === 1 && visibleHistory">
            <div class="overview bg-border">
              <el-table :data="historyList" style="width: 100%">
                <el-table-column
                  prop="version"
                  :label="$t('tempSquare.skills.version')"
                />
                <el-table-column
                  prop="desc"
                  :label="$t('tempSquare.skills.versionDesc')"
                />
                <el-table-column
                  prop="updateTime"
                  :label="$t('tempSquare.skills.updateTime')"
                />
              </el-table>
            </div>
          </template>
        </div>
      </div>

      <div v-if="recommendList.length" class="right-recommend">
        <!-- 右侧标题也可以扩展成插槽，以支持更多定制化 -->
        <slot name="recommend-title">
          <p class="recommend-list-title">
            {{ $t('skillSpace.detail.otherSkill') }}
          </p>
        </slot>
        <div
          class="recommend-item"
          v-for="(item, i) in recommendList"
          :key="`${i}rc`"
          @click="$emit('click-recommend', item)"
        >
          <img
            class="logo"
            v-if="item.avatar && item.avatar.path"
            :src="avatarSrc(item.avatar.path)"
          />
          <p class="name">{{ item.name }}</p>
          <p class="intro">{{ item.desc }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
// 这里不导入任何 API 文件，仅仅作为一个单纯的展示组件 (Dump Component)
import { avatarSrc } from '@/utils/util';
import { parseTxtSafe } from '@/utils/sanitize';
import MdRender from '@/components/mdRender.vue';
import ApiKeyTable from '@/components/skills/ApiKeyTable.vue';

export default {
  name: 'SkillDetail',
  components: { MdRender, ApiKeyTable },
  props: {
    // 详情主数据
    detail: {
      type: Object,
      default: () => ({}),
    },
    // 右侧推荐列表数据
    recommendList: {
      type: Array,
      default: () => [],
    },
    // 是否为公开分享页面（控制一些背景等视觉样式）
    isPublic: {
      type: Boolean,
      default: false,
    },
    // 公开页面的背景由于有时需要自定义，所以提升为 prop
    bgColor: {
      type: String,
      default:
        'linear-gradient(1deg, rgb(247, 252, 255) 50%, rgb(233, 246, 254) 98%)',
    },
    // 返回按钮文本
    backText: {
      type: String,
      default: '',
    },
    // 是否显示变量配置区域
    visibleVariableConfig: {
      type: Boolean,
      default: true,
    },
    // 历史版本列表数据
    historyList: {
      type: Array,
      default: () => [],
    },
    // 是否显示历史版本
    visibleHistory: {
      type: Boolean,
      default: false,
    },
    // 是否显示下载
    visibleDownload: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      foldStatus: false,
      tabActive: 0,
    };
  },
  watch: {
    visibleHistory(val) {
      if (!val && this.tabActive === 1) {
        this.tabActive = 0;
      }
    },
  },
  // 可以在挂载时发起事件，通知父组件初始化数据
  mounted() {
    this.$emit('init');
  },
  methods: {
    avatarSrc,
    // 解析文本，遇到.换行等
    parseTxt: parseTxtSafe,
    tabClick(status) {
      this.tabActive = status;
    },
    fold() {
      this.foldStatus = !this.foldStatus;
    },
    handleCreateVariable(variable) {
      this.$emit('create-variable', {
        skillId: this.detail.skillId,
        variable,
      });
    },
    handleUpdateVariable(variable) {
      const { id, ...restVariable } = variable;
      this.$emit('update-variable', {
        id,
        variable: restVariable,
      });
    },
    handleDeleteVariable(variable) {
      this.$emit('delete-variable', {
        id: variable.id,
      });
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tabs.scss';
@import '@/style/squareDetail.scss';
.tempSquare-markdown {
  ::v-deep .code-header {
    padding: 0 0 5px 0;
  }
}

.info-config {
  margin-bottom: 10px;
}
</style>
