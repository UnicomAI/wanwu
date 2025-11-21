<template>
  <div class="mcp-content-box customize">
    <div class="mcp-content">
      <div class="card-search card-search-cust">
        <div>
          <p class="card-search-des" style="display: flex; align-items: center">
            <span>{{ $t('menu.app.builtIn') }}</span>
          </p>
        </div>
        <div>
          <search-input :placeholder="$t('tool.builtIn.search')" ref="searchInput" @handleSearch="handleSearch" />
        </div>
      </div>

      <div class="card-box">
        <div
          v-if="list && list.length"
          class="card"
          v-for="(item, index) in list"
          :key="index"
          @click.stop="handleClick(item)"
        >
          <div class="card-title">
            <img class="card-logo" v-if="item.avatar && item.avatar.path" :src="avatarSrc(item.avatar.path)" />
            <div class="mcp_detailBox" style="width: 100%">
              <span class="mcp_name">{{ item.name }}</span>
              <span class="mcp_from tool_tag">
                <label
                  style="font-size: 11px"
                  v-if="item.tags && item.tags.length"
                  v-for="it in item.tags"
                  :key="it"
                >
                  {{ it }}
                </label>
              </span>
            </div>
          </div>
          <div class="card-des">{{ item.desc }}</div>
        </div>
      </div>
      <el-empty class="noData" v-if="!(list && list.length)" :description="$t('common.noData')"></el-empty>
    </div>
  </div>
</template>
<script>
import SearchInput from "@/components/searchInput.vue"
import { getBuiltInList } from "@/api/mcp";
import {avatarSrc} from "@/utils/util";
export default {
  components: { SearchInput },
  data() {
    return {
      list: [],
    };
  },
  mounted() {
    this.fetchList()
  },
  methods: {
    avatarSrc,
    handleSearch() {
      this.fetchList()
    },
    fetchList(cb) {
      const searchInput = this.$refs.searchInput
      const params = {
        name: searchInput.value,
      }
      getBuiltInList(params)
        .then((res) => {
          this.list = res.data.list || []
          cb && cb(this.list)
        })
        .catch(() => {})
    },
    handleClick(val) {
      // 内置ToolDetails
      this.$router.push({path: `/tool/detail/builtIn?toolSquareId=${val.toolSquareId}`})
    },
  },
};
</script>
<style lang="scss">
.card-search-cust {
  text-align: left !important;

  .radio-box {
    margin: 20px 0 0 0 !important;
  }
}
.card-logo{
  width: 50px;
  height: 50px;
  object-fit: cover;
}
.mcp-content-box .noData {
  width: 100%;
  text-align: center;
  margin-top: -60px;
  /deep/ .el-empty__description p {
    color: #B3B1BC;
  }
}
.tool_tag {
  height: 22px;
  label {
    display: inline-block !important;
    width: auto !important;
    margin-right: 5px;
    
  }
}
</style>