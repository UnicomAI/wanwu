<template>
  <div class="mcp-content-box customize">
    <div class="mcp-content">
      <div class="card-search card-search-cust">
        <div>
          <p class="card-search-des" style="display: flex; align-items: center">
            <span>{{$t('tool.integrate.slogan')}}</span>
            <LinkIcon type="mcp" />
          </p>
        </div>
        <div>
          <search-input :placeholder="$t('tool.integrate.search')" ref="searchInput" @handleSearch="fetchList" />
          <el-button size="mini" type="primary" @click="handleAddMCP">Import</el-button>
        </div>
      </div>

      <div class="card-box">
        <div class="card card-item-create">
          <div class="app-card-create" @click="handleAddMCP">
            <div class="create-img-wrap">
              <img class="create-type" src="@/assets/imgs/create_mcp.svg" alt="" />
              <img class="create-img" src="@/assets/imgs/create_icon.png" alt="" />
              <div class="create-filter"></div>
            </div>
            <span>{{$t('tool.integrate.create')}}</span>
          </div>
        </div>
        <div
          v-if="list && list.length"
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
            <el-dropdown
              placement="bottom">
              <span class="el-dropdown-link">
                <i class="el-icon-more"
                   @click.stop/>
              </span>
              <el-dropdown-menu slot="dropdown"  style="margin-top: -10px">
                <el-dropdown-item
                  @click.native="handleEdit(item)">
                  {{$t('common.button.edit')}}
                </el-dropdown-item>
                <el-dropdown-item
                  @click.native="handleDelete(item)">
                  {{$t('common.button.delete')}}
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
          <div class="card-des">{{ item.desc }}</div>
        </div>
      </div>

      <!--<div class="no-list" v-if="list.length === 0 && is">
        <div>
          <i class="el-icon-circle-plus-outline" @click="handleAddMCP"></i>
          <span>Add你 of 第一个MCP Server</span>
        </div>
      </div>-->
      <el-empty class="noData" v-if="!(list && list.length)" :description="$t('common.noData')"></el-empty>
    </div>
    <addDialog :dialogVisible="addOpen" :title="addTitle" :initialData="dialogParams" @handleFetch="fetchList()" @handleClose="handleClose"></addDialog>
  </div>
</template>
<script>
import addDialog from "./addDialog.vue";
import SearchInput from "@/components/searchInput.vue"
import { getList, setDelete } from "@/api/mcp";
import LinkIcon from "@/components/linkIcon.vue"
import {avatarSrc} from "@/utils/util";
export default {
  components: { LinkIcon, SearchInput, addDialog },
  data() {
    return {
      defaultAvatar: require("@/assets/imgs/mcp_active.svg"),
      addOpen: false, // CustomAddmcpSwitch
      addTitle: "",
      dialogParams: {
        name: "",
        from: "",
        sseUrl: "",
        desc: "",
      }, // AddCustommcpParameter
      list: [],
    };
  },
  mounted() {
    this.fetchList()
  },
  methods: {
    avatarSrc,
    fetchList() {
      const searchInput = this.$refs.searchInput
      const params = {
        name: searchInput.value,
      }
      getList(params)
        .then((res) => {
          this.list = res.data.list || []
        })
    },
    handleClick(val) {
      // mcpSquareId HasValue MCP广场, NoThenCustom
      this.$router.push({path: `/tool/detail/custom?mcpId=${val.mcpId}&mcpSquareId=${val.mcpSquareId}`})
    },
    handleAddMCP() {
      this.addOpen = true;
      this.addTitle = this.$t('tool.integrate.addTitle')
    },
    handleEdit(item) {
      this.addOpen = true;
      this.addTitle = this.$t('tool.integrate.editTitle')
      this.dialogParams = {
        ...item
      };
    },
    handleClose() {
      this.addOpen = false
      this.dialogParams = {
        name: "",
        from: "",
        sseUrl: "",
        desc: "",
      };
    },
    handleDelete(item) {
      this.$confirm(
        this.$t('tool.integrate.deleteHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          dangerouslyUseHTMLString: true,
          type: "warning",
          center: true,
        }
      ).then(async () => {
        setDelete({
          mcpId: item.mcpId,
        }).then((res) => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delInfo'))
            this.fetchList()
          } else {
            this.$message.error( res.msg || this.$t('common.info.delInfoErr'))
          }
        })
      })
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
</style>