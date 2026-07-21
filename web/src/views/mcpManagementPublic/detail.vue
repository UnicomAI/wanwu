<template>
  <div class="mcp-detail page-wrapper" id="timeScroll">
    <span class="back" @click="back">
      {{
        $t('menu.back') + (isFromSquare ? $t('menu.mcp') : $t('menu.resource'))
      }}
    </span>
    <div class="mcp-title">
      <img
        class="logo"
        :src="
          detail.avatar && detail.avatar.path
            ? avatarSrc(detail.avatar.path)
            : defaultAvatar
        "
        alt=""
      />
      <div :class="['info', { fold: foldStatus }]">
        <p class="name">{{ detail.name }}</p>
        <p v-if="detail.desc && detail.desc.length > 260" class="desc">
          {{ foldStatus ? detail.desc : detail.desc.slice(0, 268) + '...' }}
          <span class="arrow" v-show="detail.desc.length > 260" @click="fold">
            {{
              foldStatus ? $t('common.button.fold') : $t('common.button.detail')
            }}
          </span>
        </p>
        <p v-else class="desc">{{ detail.desc }}</p>
      </div>
    </div>
    <div class="mcp-main">
      <div class="left-info">
        <!-- tabs -->
        <div class="tabs">
          <div
            v-if="mcpSquareId"
            :class="['tab', { active: tabActive === 0 }]"
            @click="tabClick(0)"
          >
            {{ $t('square.info') }}
          </div>
          <div style="display: inline-block">
            <div
              :class="['tab', { active: tabActive === 1 }]"
              @click="tabClick(1)"
            >
              {{ transportLabel }}
            </div>
          </div>
        </div>

        <div v-if="tabActive === 0">
          <div class="overview bg-border">
            <div class="overview-item">
              <div class="item-title">
                <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                <span>{{ $t('square.summary') }}</span>
              </div>
              <div class="item-desc" v-html="parseTxt(detail.summary)"></div>
            </div>
          </div>
          <div class="overview bg-border">
            <div class="overview-item">
              <div class="item-title">
                <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                <span>{{ $t('square.feature') }}</span>
              </div>
              <div class="item-desc" v-html="parseTxt(detail.feature)"></div>
            </div>
          </div>
          <div class="overview bg-border">
            <div class="overview-item">
              <div class="item-title">
                <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                <span>{{ $t('square.scenario') }}</span>
              </div>
              <div class="item-desc">
                <div v-html="parseTxt(detail.scenario)"></div>
              </div>
            </div>
          </div>
          <div class="overview bg-border">
            <div class="overview-item">
              <div class="item-title">
                <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                <span>{{ $t('square.manual') }}</span>
              </div>
              <div class="item-desc" v-html="parseTxt(detail.manual)"></div>
            </div>
          </div>
          <div class="overview bg-border">
            <div class="overview-item">
              <div class="item-title">
                <img src="@/assets/imgs/detail_title_icon.png" alt="" />
                <span>{{ $t('square.detail') }}</span>
              </div>
              <div class="item-desc">
                <div class="mcp-markdown">
                  <MdRender :content="detail.detail" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="tool" style="padding: 0 5px" v-if="tabActive === 1">
          <div class="tool-item">
            <p class="title">{{ urlLabel }}</p>
            <div class="sse-url" style="display: flex">
              <div class="sse-url__input">{{ displayUrl }}</div>
              <el-button
                v-if="isFromSquare"
                class="sse-url__bt"
                type="primary"
                :disabled="detail.hasCustom"
                @click="preSendToCustomize"
              >
                {{ $t('tool.square.sendButton') }}
              </el-button>
            </div>
            <p class="see-url__hint">
              <i class="el-icon-info"></i>
              {{
                isFromSquare
                  ? $t('tool.square.sendHint1')
                  : $t('tool.square.sendHint2')
              }}
            </p>
          </div>
          <div class="tool-item" v-if="tools && tools.length">
            <p class="title">{{ $t('tool.square.tool.info') }}</p>
            <div class="tool-intro">
              <el-collapse class="mcp-el-collapse">
                <el-collapse-item
                  v-for="(n, i) in tools"
                  :key="n.name + i"
                  :title="n.name"
                  :name="i"
                >
                  <div class="desc">
                    {{ $t('tool.square.tool.desc') }}
                    <span v-html="parseTxt(n.description)"></span>
                  </div>
                  <div class="params">
                    <p>{{ $t('tool.square.tool.params') }}</p>
                    <div
                      class="params-table"
                      v-for="(m, j) in n.params"
                      :key="m.name + j"
                    >
                      <div class="tr">
                        <div class="td">{{ m.name }}</div>
                        <div class="td color">{{ m.type }}</div>
                        <div class="td color">{{ m.requiredBadge }}</div>
                      </div>
                      <p
                        class="params-desc"
                        v-html="parseTxt(m.description)"
                      ></p>
                    </div>
                  </div>
                </el-collapse-item>
              </el-collapse>
            </div>
          </div>
          <!--v0.6.1 隐藏安装说明-->
          <!--<div class="tool-item bottom-install-intro">
            <p class="title">{{ $t('tool.square.tool.setup') }}</p>
            <div>
              <div class="install-intro-item">
                <p class="install-intro-title">
                  {{ $t('tool.square.tool.cursor.title') }}
                </p>
                <p>{{ $t('tool.square.tool.cursor.step1') }}</p>
                <p>{{ $t('tool.square.tool.cursor.step2') }}</p>
                <p>{{ $t('tool.square.tool.cursor.step3') }}</p>
                <p>{{ $t('tool.square.tool.cursor.step4') }}</p>
              </div>
              <div class="install-intro-item">
                <p class="install-intro-title">
                  {{ $t('tool.square.tool.claude.title') }}
                </p>
                <p>{{ $t('tool.square.tool.claude.step1') }}</p>
                <p>{{ $t('tool.square.tool.claude.step2') }}</p>
                <p>{{ $t('tool.square.tool.claude.step3') }}</p>
              </div>
            </div>
          </div>-->
        </div>
      </div>

      <div class="right-recommend">
        <p class="recommend-list-title">
          {{ $t('tool.square.tool.other') }}
        </p>
        <div
          class="recommend-item"
          v-for="(item, i) in recommendList"
          :key="`${i}rc`"
          @click="handleClick(item)"
        >
          <img
            class="logo"
            :src="
              item.avatar && item.avatar.path
                ? avatarSrc(item.avatar.path)
                : defaultAvatar
            "
            alt=""
          />
          <p class="name">{{ item.name }}</p>
          <p class="intro">{{ item.desc }}</p>
        </div>
      </div>
    </div>

    <sendDialog
      ref="dialog"
      :dialogVisible="dialogVisible"
      :detail="detail"
      @handleClose="handleClose"
      @getIsCanSendStatus="getIsCanSendStatus"
    />
  </div>
</template>
<script>
import sendDialog from './sendDialog';
import {
  getRecommendsList,
  getPublicMcpInfo,
  getDetail,
  getTools,
} from '@/api/mcp';
import { avatarSrc, formatTools } from '@/utils/util';
import { parseTxtSafe } from '@/utils/sanitize';
import MdRender from '@/components/mdRender.vue';

export default {
  name: 'McpServiceCustomDetail',
  props: {
    type: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      defaultAvatar: require('@/assets/imgs/mcp_active.svg'),
      isFromSquare: true,
      mcpSquareId: '',
      mcpId: '',
      detail: {},
      tools: [],
      foldStatus: false,
      tabActive: 0,
      recommendList: [],
      dialogVisible: false,
    };
  },
  computed: {
    transportLabel() {
      return this.detail.transport === 'streamable'
        ? 'Streamable HTTP URL及工具'
        : this.$t('tool.square.sseUrl');
    },
    urlLabel() {
      return this.detail.transport === 'streamable'
        ? 'Streamable HTTP:'
        : 'SSE URL:';
    },
    displayUrl() {
      return this.detail.transport === 'streamable'
        ? this.detail.streamableUrl
        : this.detail.sseUrl;
    },
  },
  watch: {
    $route: {
      handler() {
        this.initData();
      },
      // 深度观察监听
      deep: true,
    },
  },
  mounted() {
    this.initData();
    this.getRecommendList();
  },
  methods: {
    avatarSrc,
    initData() {
      this.mcpSquareId = this.$route.query.mcpSquareId;
      this.mcpId = this.$route.query.mcpId;
      this.isFromSquare = this.type === 'square';
      this.tabActive = 0;
      this.getDetailData();

      //滚动到顶部
      const main = document.querySelector('.el-main > .page-container');
      if (main) main.scrollTop = 0;
    },
    getDetailData() {
      if (this.isFromSquare) {
        getPublicMcpInfo({ mcpSquareId: this.mcpSquareId }).then(res => {
          this.detail = res.data || {};
          this.tools = formatTools(res.data.tools);
        });
      } else {
        if (!this.mcpSquareId) this.tabActive = 1;
        getDetail({ mcpId: this.mcpId }).then(res => {
          this.detail = res.data || {};
        });
        this.getToolsList();
      }
    },
    getToolsList() {
      getTools({
        mcpId: this.mcpId,
      }).then(res => {
        this.tools = formatTools(res.data.tools);
      });
    },
    getIsCanSendStatus() {
      getPublicMcpInfo({ mcpSquareId: this.mcpSquareId }).then(res => {
        this.detail.hasCustom = res.data.hasCustom;
      });
    },
    getRecommendList() {
      const params = {
        mcpSquareId: this.mcpSquareId,
      };
      getRecommendsList(params).then(res => {
        this.recommendList = res.data.list;
      });
    },
    handleClick(val) {
      this.$router.push(`/mcp/detail/square?mcpSquareId=${val.mcpSquareId}`);
    },
    // 解析文本，遇到.换行等
    parseTxt: parseTxtSafe,
    tabClick(status) {
      this.tabActive = status;
    },
    fold() {
      this.foldStatus = !this.foldStatus;
    },
    preSendToCustomize() {
      this.dialogVisible = true;
      this.$refs.dialog.ruleForm.serverUrl = this.detail.sseUrl;
    },
    handleClose() {
      this.dialogVisible = false;
    },
    back() {
      if (this.isFromSquare) this.$router.push({ path: '/mcp' });
      else this.$router.push({ path: '/mcpService?mcp=integrate' });
    },
  },
  components: {
    sendDialog,
    MdRender,
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/tabs.scss';
@import '@/style/squareDetail.scss';
</style>
