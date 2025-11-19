<template>
  <div class="page-wrapper mcp-management">
    <div class="common_bg">
      <div class="page-title">
        <img class="page-title-img" src="@/assets/imgs/mcp_menu.svg" alt=""/>
        <span class="page-title-name">{{ $t('menu.mcp') }}</span>
      </div>
      <div class="mcp-content-box mcp-third">
        <div class="mcp-main">
          <div class="mcp-content">
            <div class="mcp-card-box">
              <div class="card-search card-search-cust">
                <div>
              <span
                  v-for="item in typeList"
                  :key="item.key"
                  :class="['tab-span', {'is-active': typeRadio === item.key}]"
                  @click="changeTab(item.key)"
              >
                {{ item.name }}
              </span>
                </div>
                <search-input style="margin-right: 2px" :placeholder="$t('tool.square.searchHint')" ref="searchInput"
                              @handleSearch="doGetPublicMcpList"/>
              </div>

              <div class="card-loading-box" v-if="list.length">
                <div class="card-box" v-loading="loading">
                  <div
                      class="card"
                      v-for="(item, index) in list"
                      :key="index"
                      @click.stop="handleClick(item)"
                  >
                    <div class="card-title">
                      <img class="card-logo" :src="(item.avatar && item.avatar.path) ? avatarSrc(item.avatar.path) : defaultAvatar" alt=""/>
                      <div class="mcp_detailBox">
                        <span class="mcp_name">{{ item.name }}</span>
                        <span class="mcp_from">
                      <label>
                        {{ item.from }}
                      </label>
                    </span>
                      </div>
                    </div>
                    <div class="card-des">{{ item.desc }}</div>
                  </div>
                  <!--<p class="loading-tips" v-if="loading"><i class="el-icon-loading"></i></p>
                  <p class="loading-tips">No更多了</p>-->
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
  </div>
</template>

<script>
import { getPublicMcpList } from "@/api/mcp"
import SearchInput from "@/components/searchInput.vue"
import {avatarSrc} from "@/utils/util";
export default {
  components: { SearchInput },
  data() {
    return {
      defaultAvatar: require("@/assets/imgs/mcp_active.svg"),
      mcpSquareId: "",
      category: this.$t('tool.square.all'),
      list: [],
      loading:false,
      typeRadio: 'all',
      typeList: [
        {name: this.$t('tool.square.all'), key: 'all'},
        {name: this.$t('tool.square.gov'), key: 'gov'},
        {name: this.$t('tool.square.industry'), key: 'industry'},
        {name: this.$t('tool.square.edu'), key: 'edu'},
        {name: this.$t('tool.square.medical'), key: 'medical'},
        {name: this.$t('tool.square.data'), key: 'data'},
        {name: this.$t('tool.square.creator'), key: 'create'},
        {name: this.$t('tool.square.search'), key: 'search'},
      ]
    };
  },
  mounted() {
    this.doGetPublicMcpList()
  },
  methods: {
    avatarSrc,
    changeTab(key) {
      this.typeRadio = key
      this.$refs.searchInput.value = ''
      this.doGetPublicMcpList()
    },
    doGetPublicMcpList(){
      const searchInput = this.$refs.searchInput
      let params = {
        name: searchInput.value,
        category: this.typeRadio,
      }

      getPublicMcpList(params)
        .then((res) => {
          this.list = res.data.list || []
          this.loading = false
        })
        .catch(() => this.loading = false)
    },
    handleClick(val) {
      this.mcpSquareId = val.mcpSquareId;
      this.$router.push({path:`/mcp/detail/square?mcpSquareId=${val.mcpSquareId}`})
    },
  },
};
</script>

<style lang="scss">
.mcp-management {
  height: calc(100% - 50px);
  .common_bg{
    height: 100%;
  }
  .mcp-content-box{
    height: calc(100% - 145px);
  }
  .mcp-content {
    padding: 0 20px;
    width: 100%;
    height: 100%;
  }

  .mcp-third{
    min-height: 600px;
    .tab-span {
      display: inline-block;
      vertical-align: middle;
      padding: 6px 12px;
      border-radius: 6px;
      color: $color_title;
      cursor: pointer;
    }
    .tab-span.is-active {
      color: $color;
      background: #fff;
      font-weight: bold;
    }
    .mcp-main{
      display: flex;
      padding: 0 20px;
      height: 100%;
      .mcp-content{
        display: flex;
        width:100%;
        padding: 0;
        height: 100%;
        .mcp-menu{
          margin-top: 10px;
          margin-right: 20px;
          width: 90px;
          height: 450px;
          border: 1px solid $border_color; //#d0a7a7
          text-align: center;
          border-radius: 6px;
          color: #333;
          p{
            line-height: 28px;
            margin:10px 0;
          }
          .active{
            background: rgba(253, 231, 231, 1);
          }
        }
        .mcp-card-box{
          width: 100%;
          height: 100%;
          .input-with-select {
            width: 300px;
          }
          .card-loading-box{
            .card-box {
              display: flex;
              flex-wrap: wrap;
              margin: 6px -10px 0;
              align-content: flex-start;
              padding-bottom: 20px;
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

                  .action-icon {
                    display: block;
                  }
                }
                .card-title {
                  display: flex;
                  width: 100%;
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

              .loading-tips{
                height: 20px;
                color: #999;
                text-align: center;
                display: block;
                width: 100%;
                i{
                  font-size: 18px;
                }
              }
            }
          }
        }
      }
    }
    .card-logo{
      width: 50px;
      height: 50px;
      object-fit: cover;
    }
  }
  .card-search {
    text-align: right;
    padding: 10px 0;
  }
  .card-search-cust {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .empty{
    width: 200px;
    height: 100px;
    margin: 50px auto;
  }
}
</style>