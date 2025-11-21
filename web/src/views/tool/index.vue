<template>
  <div class="page-wrapper mcp-management">
    <div class="common_bg">
      <div class="page-title">
        <img class="page-title-img" src="@/assets/imgs/tool.svg" alt=""/>
        <span class="page-title-name">{{ $t('menu.tool') }}</span>
      </div>
      <!-- tabs -->
      <div class="mcp-tabs">
        <div :class="['mcp-tab',{ 'active': ![tool, prompt].includes(tabActive) }]" @click="tabClick(mcp)">{{ $t('tool.mcp') }}</div>
        <div :class="['mcp-tab',{ 'active': tabActive === tool }]" @click="tabClick(tool)">{{ $t('tool.tool') }}</div>
        <div :class="['mcp-tab',{ 'active': tabActive === prompt }]" @click="tabClick(prompt)">{{ $t('tool.prompt.title') }}</div>
      </div>

      <mcp ref="mcp" v-if="![tool, prompt].includes(tabActive)"/>
      <tool ref="tool" v-if="tabActive === tool"/>
      <prompt ref="prompt" v-if="tabActive === prompt"/>
    </div>
  </div>
</template>
<script>
import mcp from './mcp'
import tool from './tool'
import prompt from './prompt'
import { MCP, TOOL, PROMPT } from './constants'

export default {
  data() {
    return {
      tabActive: MCP,
      mcp: MCP,
      tool: TOOL,
      prompt: PROMPT,
    };
  },
  watch: {
    $route: {
      handler() {
        const {type} = this.$route.query || {}
        this.tabActive = type
      },
      // 深度ObserveListen
      deep: true
    }
  },
  mounted() {
    const {type} = this.$route.query || {}
    this.tabActive = type
  },
  methods: {
    tabClick(type) {
      this.tabActive = type
    },
  },
  components: {
    mcp,
    tool,
    prompt
  },
};
</script>
<style lang="scss">
.mcp-management {
  height: calc(100% - 50px);

  .common_bg {
    height: 100%;
  }

  .title {
    font-size: 20px;
    margin: 0;
    padding: 0 0 20px 0;
    text-align: center;

    .svg-icon {
      width: 1.6em;
      height: 1.6em;
      color: $color;
      vertical-align: -0.25em;
    }
  }

  .mcp-tabs {
    margin: 20px;

    .mcp-tab {
      display: inline-block;
      vertical-align: middle;
      width: 160px;
      height: 40px;
      border-bottom: 1px solid #333;
      line-height: 40px;
      text-align: center;
      cursor: pointer;
    }

    .active {
      background: #333;
      color: #fff;
      font-weight: bold;
    }
  }

  .mcp-content-box {
    height: calc(100% - 145px);
  }

  .mcp-content {
    padding: 0 20px;
    width: 100%;
    height: 100%;
  }

  .el-tabs__nav-wrap {
    text-align: center;
  }

  .el-tabs__nav-scroll {
    display: inline-block;
  }

  .el-tabs__nav-wrap::after {
    display: none;
  }

  .card-box {
    display: flex;
    flex-wrap: wrap;
    margin: 6px -10px 0;
    /*overflow: auto;*/
    .card {
      position: relative;
      padding: 20px 16px;
      border-radius: 12px;
      height: fit-content;
      background: #fff;
      display: flex;
      flex-direction: column;
      align-items: center;
      width: calc((100% / 4) - 20px);
      margin: 0 10px 20px;
      box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
      border: 1px solid rgba(0, 0, 0, 0);

      &:hover {
        cursor: pointer;
        box-shadow: 0 2px 8px #171a220d, 0 4px 16px #0000000f;
        border: 1px solid $border_color;
      }

      .card-title {
        display: flex;
        width: 100%;
        height: 58px;
        border-bottom: 1px solid #ddd;
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
          padding: 3px 0;

          .mcp_name {
            display: block;
            font-size: 15px;
            font-weight: 700;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
            color: #5d5d5d;
          }

          .mcp_from {
            label {
              padding: 3px 7px;
              font-size: 12px;
              color: #84868c;
              background: #f2f5f9;
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
        -webkit-line-clamp: 3;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        font-size: 13px;
        height: 55px;
        word-wrap: break-word;
      }
    }

    .card-item-create {
      background: $color_opacity;
      box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
      border: 1px solid rgba(56, 75, 247, 0.47);

      .app-card-create {
        width: 100%;
        height: 100%;
        min-height: 125px;
        text-align: center;
        display: flex;
        align-items: center;
        justify-content: center;

        .create-img-wrap {
          display: inline-block;
          vertical-align: middle;
          margin-right: 10px;
          position: relative;

          .create-img {
            width: 40px;
            height: 40px;
            border-radius: 8px;
            background: $color;
            padding: 10px;
          }

          .create-filter {
            width: 40px;
            height: 8px;
            background: rgba(2, 81, 252, 0.3);
            filter: blur(5px);
            position: absolute;
            bottom: -6px;
          }

          .create-type {
            width: 30px;
            position: absolute;
            background: rgba(171, 198, 255, 0.5);
            backdrop-filter: blur(6.55px);
            border-radius: 5px;
            padding: 6px;
            top: -10px;
            left: -12px;
          }
        }

        span {
          display: inline-block;
          vertical-align: middle;
          font-size: 16px;
          color: $color_title;
          font-weight: bold;
        }
      }
    }
  }

  .no-list {
    display: flex;
    justify-content: center;
    align-items: center;
    height: calc(100vh - 330px);
    min-height: 200px;
    font-size: 30px;
    // color: #ddd;
    text-align: center;

    i {
      font-size: 50px;
      color: $color;
      cursor: pointer;
    }

    span {
      padding-top: 20px;
      display: block;
    }
  }

  .card-search {
    text-align: right;
    padding: 10px 0;
  }

  .el-tabs__content {
    max-width: 1500px;
    margin: 0 auto;
  }

  .card-search-cust {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .card-search-des {
      color: #585a73;
      font-size: 12px;

      .el-button {
        padding: 5px 12px;

        span {
          font-size: 12px;
        }
      }
    }

    .radio-box {
      margin: 10px 0;
      padding: 0;
    }
  }

  .el-radio__input.is-checked .el-radio__inner {
    border-color: $color;
    background: $color;
  }
}
</style>