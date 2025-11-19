<template>
  <div class="mcp-detail" id="timeScroll">
    <span class="back" @click="back">{{$t('menu.back') + (isFromSquare ? $t('menu.mcp') : $t('menu.tool'))}}</span>
    <div class="mcp-title">
      <img class="logo" :src="(detail.avatar && detail.avatar.path) ? avatarSrc(detail.avatar.path) : defaultAvatar" alt=""/>
      <div :class="['info',{fold:foldStatus}]">
        <p class="name">{{detail.name}}</p>
        <p v-if="detail.desc && detail.desc.length > 260" class="desc">
          {{foldStatus ? detail.desc : detail.desc.slice(0,268) + '...'}}
          <span class="arrow" v-show="detail.desc.length > 260" @click="fold">
            {{foldStatus ? $t('common.button.fold') : $t('common.button.detail')}}
          </span>
        </p>
        <p v-else class="desc">{{detail.desc}}</p>
      </div>
    </div>
    <div class="mcp-main">
      <div class="left-info">
        <!-- tabs -->
        <div class="mcp-tabs">
          <div v-if="mcpSquareId" :class="['mcp-tab',{ 'active': tabActive === 0 }]" @click="tabClick(0)">{{ $t('tool.square.info') }}</div>
          <div style="display: inline-block">
            <div :class="['mcp-tab',{ 'active': tabActive === 1 }]" @click="tabClick(1)">{{ $t('tool.square.sseUrl') }}</div>
          </div>
        </div>

        <div v-if="tabActive === 0">
          <div class="overview bg-border" >
            <div class="overview-item">
              <div class="item-title">• &nbsp;{{ $t('tool.square.summary') }}</div>
              <div class="item-desc" v-html="parseTxt(detail.summary)"></div>
            </div>
            <div class="overview-item">
              <div class="item-title">• &nbsp;{{ $t('tool.square.feature') }}</div>
              <div class="item-desc" v-html="parseTxt(detail.feature)"></div>
            </div>
            <div class="overview-item">
              <div class="item-title">• &nbsp;{{ $t('tool.square.scenario') }}</div>
              <div class="item-desc" >
                <div v-html="parseTxt(detail.scenario)"></div>
              </div>
            </div>
          </div>
          <div class="overview bg-border" >
            <div class="overview-item">
              <div class="item-title">• &nbsp;{{ $t('tool.square.manual') }}</div>
              <div class="item-desc" v-html="parseTxt(detail.manual)"></div>
            </div>
          </div>
          <div class="overview bg-border" >
            <div class="overview-item">
              <div class="item-title">• &nbsp;{{ $t('tool.square.detail') }}</div>
              <div class="item-desc">
                <div class="readme-content markdown-body mcp-markdown" v-html="md.render(detail.detail || '')"></div>
              </div>
            </div>
          </div>
        </div>
        <!--<div class="install bg-border" v-if="tabActive === 2">
            &lt;!&ndash;copy from https://mcpmarket.cn/&ndash;&gt;
            <div class="login-required-message" style="text-align: center;background-color: #a7535305; padding: 40px 20px; border-radius: 8px; margin: 20px 0;">
                <i class="fas fa-lock el-icon-lock" style="font-size: 48px; color: #D33A3A; margin-bottom: 20px; display: block;"></i>
                <h3 style="margin-bottom: 15px; color: #333;font-size: 20px;">需要登录</h3>
                <p style="margin-bottom: 25px; color: #666; line-height: 40px;">
                    要GetSSE URLAndConfigurationMCP服务器，PleaseFirst登录您 of 账号。如果No账号，您可以快速注册一个。
                </p>
                <div style="display: flex; justify-content: center; gap: 15px;">
                    <a href="https://mcpmarket.cn/auth/login?next=%2Fserver%2F67ff4974764487b6b9e11c21" style="display: inline-block; padding: 10px 20px; background-color: #D33A3A; color: white; text-decoration: none; border-radius: 6px; font-weight: 500;">
                        登录
                    </a>
                    <a href="https://mcpmarket.cn/auth/login?next=%2Fserver%2F67ff4974764487b6b9e11c21" style="display: inline-block; padding: 10px 20px; background-color: white; color: #D33A3A; text-decoration: none; border-radius: 6px; border: 1px solid #D33A3A; font-weight: 500;">
                        使用社交账号登录
                    </a>
                </div>
            </div>
        </div>-->

        <div class="tool bg-border" v-if="tabActive === 1">
          <div class="tool-item ">
            <p class="title">SSE URL:</p>
            <div class="sse-url" style="display: flex">
              <div class="tool-item-bg sse-url__input">{{detail.sseUrl}}</div>
              <el-button v-if="isFromSquare" class="sse-url__bt" type="primary" :disabled="detail.hasCustom" @click="preSendToCustomize">{{ $t('tool.square.sendButton') }}</el-button>
            </div>
            <p style="line-height: 40px;color: #666;">
              {{isFromSquare ? $t('tool.square.sendHint1') : $t('tool.square.sendHint2')}}
            </p>
          </div>
          <div class="tool-item" v-if="tools && tools.length">
            <p class="title">{{ $t('tool.square.tool.info') }}</p>
            <div class="tool-item-bg tool-intro">
              <el-collapse class="mcp-el-collapse">
                <el-collapse-item v-for="(n,i) in tools" :key="n.name + i" :title="n.name" :name="i">
                  <div class="desc">{{ $t('tool.square.tool.desc') }}<span v-html="parseTxt(n.description)"></span></div>
                  <div class="params">
                    <p>{{ $t('tool.square.tool.params') }}</p>
                    <div class="params-table" v-for="(m, j) in n.params" :key="m.name + j">
                      <div class="tr">
                        <div class="td">{{m.name}}</div>
                        <div class="td color">{{m.type}}</div>
                        <div class="td color">{{m.requiredBadge}}</div>
                      </div>
                      <p class="params-desc" v-html="parseTxt(m.description)"></p>
                    </div>
                  </div>
                </el-collapse-item>
              </el-collapse>
            </div>
          </div>
          <!--<div class="tool-item">
              <p class="title">MCP服务器Configuration:</p>
              <div class="tool-item-bg service-config"></div>
          </div>-->
          <div class="tool-item">
            <p class="title">{{ $t('tool.square.tool.setup') }}</p>
            <div class="tool-item-bg">
              <div class="install-intro-item">
                <p class="install-intro-title">{{ $t('tool.square.tool.cursor.title') }}</p>
                <p>{{ $t('tool.square.tool.cursor.step1') }}</p>
                <p>{{ $t('tool.square.tool.cursor.step2') }}</p>
                <p>{{ $t('tool.square.tool.cursor.step3') }}</p>
                <p>{{ $t('tool.square.tool.cursor.step4') }}</p>
              </div>
              <div class="install-intro-item">
                <p class="install-intro-title">{{ $t('tool.square.tool.claude.title') }}</p>
                <p>{{ $t('tool.square.tool.claude.step1') }}</p>
                <p>{{ $t('tool.square.tool.claude.step2') }}</p>
                <p>{{ $t('tool.square.tool.claude.step3') }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="right-recommend">
        <p style="margin: 20px 0;color: #333;">{{ $t('tool.square.tool.other') }}</p>
        <div class="recommend-item" v-for="(item ,i) in recommendList" :key="`${i}rc`" @click="handleClick(item)">
          <img class="logo" :src="(item.avatar && item.avatar.path) ? avatarSrc(item.avatar.path) : defaultAvatar" alt=""/>
          <p class="name">{{item.name}}</p>
          <p class="intro">{{item.desc}}</p>
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
import sendDialog from './sendDialog'
import { md } from '@/mixins/marksown-it'
import { getRecommendsList, getPublicMcpInfo, getDetail, getTools } from "@/api/mcp"
import {avatarSrc, formatTools} from "@/utils/util"

export default {
  props: {
    type: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      defaultAvatar: require("@/assets/imgs/mcp_active.svg"),
      md:md,
      isFromSquare: true,
      mcpSquareId:'',
      mcpId: '',
      detail: {},
      tools: [],
      foldStatus:false,
      tabActive:0,
      recommendList: [],
      dialogVisible: false,
    };
  },
  watch: {
    $route: {
      handler() {
        this.initData()
      },
      // 深度ObserveListen
      deep: true
    }
  },
  mounted() {
    this.initData()
    this.getRecommendList()
  },
  methods: {
    avatarSrc,
    initData(){
      this.mcpSquareId = this.$route.query.mcpSquareId
      this.mcpId = this.$route.query.mcpId
      this.isFromSquare = this.type === 'square'
      this.tabActive = 0
      this.getDetailData()

      //滚动到顶部
      const main = document.querySelector(".el-main > .page-container")
      if (main) main.scrollTop = 0
    },
    getDetailData(){
      if (this.isFromSquare) {
        getPublicMcpInfo({mcpSquareId:this.mcpSquareId}).then((res) => {
          this.detail = res.data || {}
          this.tools = formatTools(res.data.tools)
        })
      } else {
        if (!this.mcpSquareId) this.tabActive = 1
        getDetail({mcpId:this.mcpId}).then((res) => {
          this.detail = res.data || {}
        })
        this.getToolsList()
      }
    },
    getToolsList(){
      getTools({
        mcpId: this.mcpId,
      }).then((res) => {
        this.tools = formatTools(res.data.tools)
      })
    },
    getIsCanSendStatus(){
      getPublicMcpInfo({mcpSquareId:this.mcpSquareId})
        .then((res) => {
          this.detail.hasCustom = res.data.hasCustom
        })
    },
    getRecommendList() {
      const params = {
        mcpSquareId: this.mcpSquareId
      }
      getRecommendsList(params).then((res) => {
        this.recommendList = res.data.list
      })
    },
    handleClick(val){
      this.$router.push(`/mcp/detail/square?mcpSquareId=${val.mcpSquareId}`)
    },
    // ParseText，遇到.换行Etc
    parseTxt(txt){
      if (!txt) return ''
      const text = txt.replaceAll('\n\t','<br/>&nbsp;').replaceAll('\n','<br/>').replaceAll('\t', '   &nbsp;')
      return text
    },
    tabClick(status){
      this.tabActive = status
    },
    fold(){
      this.foldStatus = !this.foldStatus
    },
    preSendToCustomize(){
      this.dialogVisible = true
      this.$refs.dialog.ruleForm.serverUrl = this.detail.sseUrl
    },
    handleClose(){
      this.dialogVisible = false
    },
    back() {
      if (this.isFromSquare) this.$router.push({path: '/mcp'})
      else this.$router.push({path: '/tool?type=mcp&mcp=integrate'})
    },
  },
  components: {
    sendDialog
  },
};
</script>
<style lang="scss">
@import "@/style/markdown.scss";
.markdown-body{
  font-family: 'Microsoft YaHei', Arial, sans-serif;
}
.mcp-detail{
  padding: 20px;
  overflow: auto;
  margin: 20px;
  .back {
    color: $color;
    cursor: pointer;
  }
  .mcp-title{
    padding: 20px 0;
    display: flex;
    border-bottom: 1px solid #bfbfbf;
    .logo{
      width: 54px;
      height: 54px;
      object-fit: cover;
    }
    .info{
      position: relative;
      width: 1240px;
      margin-left: 15px;
      .name{
        font-size: 16px;
        color: #5D5D5D;
        font-weight: bold;
      }
      .desc{
        margin-top: 10px;
        line-height: 22px;
        color: #9F9F9F;
        word-break: break-all;
      }
      .arrow{
        position: absolute;
        display: block;
        right: 0;
        bottom: -5px;
        cursor: pointer;
        color: $color;
        margin-left: 10px;
        font-size: 13px;
      }
    }
    .fold{
      height: auto;
    }
  }
  .mcp-main{
    display: flex;
    margin: 10px 0 0 0;
    .left-info{
      width: calc(100% - 420px);
      margin-right: 20px;
      .mcp-tabs{
        margin: 20px 0 0 0;
        .mcp-tab{
          display: inline-block;
          vertical-align: middle;
          width: 160px;
          height: 40px;
          border-bottom: 1px solid #333;
          line-height: 40px;
          text-align: center;
          cursor: pointer;
        }
        .active{
          background: #333;
          color: #fff;
          font-weight: bold;
        }
      }
      .overview{
        .overview-item{
          display: flex;
          padding: 15px 0;
          border-bottom: 1px solid #eee;
          line-height: 24px;
          .item-title{
            width: 80px;
            color: $color;
            font-weight: bold;
          }
          .item-desc{
            width: calc(100% - 100px);
            margin-left: 10px;
            flex:1;
            color: #333;
          }

        }
        .overview-item:last-child{
          border-bottom: none;
        }
      }
      .tool{
        .tool-item{
          padding: 20px 0;
          border-bottom: 1px solid #eee;
          .title{
            font-weight: bold;
            line-height: 46px;
          }
          .tool-item-bg{
            background: inherit;
            background-color: rgba(249, 249, 249, 1);
            border: none;
            border-radius: 10px;
            padding: 20px;
          }
        }
        .tool-item:last-child{
          border-bottom: none;
        }
        .sse-url{
          .sse-url__input{
            flex:1;
            margin-right: 20px;
            padding: 12px;
            color: $color;
          }
          .sse-url__bt{
            width: 120px;
          }
        }
        .install-intro-item{
          p{
            line-height: 26px;
            color: #333;
          }
          .install-intro-title{
            color: $color;
            margin-top: 10px;
            font-weight: bold;
          }
        }
      }
    }
    .right-recommend{
      width: 400px;
      overflow-y: auto;
      border-left:1px solid #eee;
      padding: 20px;
      max-height: 900px;
      .recommend-item{
        position: relative;
        border: 1px solid $border_color;
        background: $color_opacity;
        margin-bottom: 15px;
        border-radius: 10px;
        padding: 20px 20px 20px 80px;
        text-align: left;
        cursor: pointer;
        .logo{
          width: 46px;
          height: 46px;
          object-fit: cover;
          position: absolute;
          left:20px;
          border: 1px solid #fff;
          border-radius: 4px;
        }
        .name{
          color: #5D5D5D;
          font-weight: bold;
        }
        .intro{
          height: 36px;
          color: #5D5D5D;
          margin-top: 8px;
          font-size: 13px;
          overflow: hidden;
        }
      }
    }
  }
  .bg-border{
    margin-top: 20px;
    /*min-height: calc(100vh - 360px);*/
    background-color: rgba(255, 255, 255, 1);
    box-sizing: border-box;
    /*border:1px solid rgba(208, 167, 167, 1);*/
    border-radius: 10px;
    padding: 10px 20px;
    box-shadow: 2px 2px 15px $color_opacity;
  }
  .overview-item .item-desc{
    line-height: 28px;
  }
}

.mcp-el-collapse.el-collapse {
  border: none;
}
.mcp-el-collapse .el-collapse-item {
  margin: 10px 0;
  border: none;

  .el-collapse-item__header {
    border: none;
    color: $color;
    font-weight: bold;
    padding: 0 20px;
  }

  .el-collapse-item__wrap {
    padding: 0 20px;
    border: none;
  }

  .desc {
    background: $color_opacity;
    padding: 10px 15px;
    border-radius: 6px;
    border: 1px solid $border_color;
  }

  .params {
    margin-top: 12px;

    .params-table {
      border-radius: 6px;
      border: 1px solid #ddd;
      padding: 10px 12px;
      background-color: #fff;
      margin-top: 6px;

      .tr {
        display: flex;

        .td {
          padding: 0 30px 0 0;
        }

        .color {
          color: $color;
        }
      }

      .params-desc {
        margin-top: 4px;
        color: #999;
      }
    }
  }
}
.mcp-markdown {
  /deep/.code-header{
    /*height: 0!important;*/
    padding: 0 0 5px 0;
  }
}
.el-button.is-disabled {
  background: #f9f9f9 !important;
}
</style>