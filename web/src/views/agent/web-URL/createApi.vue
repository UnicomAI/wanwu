<template>
  <div>
    <div class="header">
      <div class="header-api">
        <el-tag
          effect="plain"
          class="root-url"
        >API Base URL</el-tag>
        {{apiURL}}
      </div>
      <el-button
        size="small"
        @click="openApiDialog"
        class="apikeyBtn"
      >
        <img :src="require('@/assets/imgs/apikey.png')" />
        API Key
      </el-button>
      <div class="show-doc" @click="jumpApiDoc">
        viewAPIDocument
      </div>
    </div>
    <el-table
      :data="tableData"
      style="width: 100%"
    >
      <el-table-column
        label="API Key"
        prop="apiKey"
        width="300"
      >
      <template slot-scope="scope">
        <span>{{scope.row.apiKey.slice(0,6) + '******'}}</span>
      </template>
      </el-table-column>
      <el-table-column
        label="Create Time"
        prop="createdAt"
      />
      <el-table-column
        label="Operation"
        width="200"
      >
        <template slot-scope="scope">
          <el-button
            size="mini"
            @click="handleCopy(scope.row) && copycb()"
          >Copy</el-button>
          <el-button
            size="mini"
            @click="handleDelete(scope.row)"
          >Delete</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
<script>
import { getApiKeyRoot,createApiKey,delApiKey,getApiKeyList } from "@/api/appspace";
export default {
  props: {
    appType: {
      type: String,
      required: true,
    },
    appId: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      apiURL: "",
      tableData: [],
      dialogVisible:false
    };
  },
  created() {
    this.getTableData();
    this.apiKeyRootUrl();
  },
  methods: {
    jumpApiDoc() {
      const {data} = this.$store.state.user.commonInfo || {}
      const {linkList} = data || {}
      const apiDocLink = linkList[`api-${this.appType}`]
      if (apiDocLink) window.open(apiDocLink)
    },
    handleClose(){
      this.dialogVisible = false;
    },
    apiKeyRootUrl() {
      const data = { appId: this.appId, appType:this.appType };
      getApiKeyRoot(data).then((res) => {
        if (res.code === 0) {
          this.apiURL = res.data || "";
        }
      });
    },
    openApiDialog(){
      this.handleCreate();
    },
    handleCopy(row) {
      let text = row.apiKey;
      var textareaEl = document.createElement("textarea");
      textareaEl.setAttribute("readonly", "readonly");
      textareaEl.value = text;
      document.body.appendChild(textareaEl);
      textareaEl.select();
      var res = document.execCommand("copy");
      document.body.removeChild(textareaEl);
      return res;
    },
    copycb() {
      this.$message.success("Content copied to clipboard");
    },
    handleCreate() {
      const data = { appId: this.appId, appType: this.appType };
      createApiKey(data).then((res) => {
        if (res.code === 0) {
          this.getTableData()
        }
      });
    },
    getTableData() {
      const data = { appId: this.appId, appType: this.appType };
      getApiKeyList(data).then((res) => {
        if (res.code === 0) {
          this.tableData = res.data || [];
        }
      });
    },
    handleDelete(row) {
      this.$confirm(
        "ConfirmtoDeletecurrentAPIkey吗？",
        this.$t("knowledgeManage.tip"),
        {
          confirmButtonText: "Confirm",
          cancelButtonText: "Cancel",
          type: "warning",
        }
      )
        .then(() => {
          delApiKey({ apiId: row.apiId }).then((res) => {
            if (res.code === 0) {
              this.$message.success("DeleteSuccess");
              this.getTableData();
            }
          });
        })
        .catch((error) => {
          this.getTableData();
        });
    },
  },
};
</script>
<style lang="scss" scoped>
  .header{
    width:100%;
    display:flex;
    justify-content:flex-start;
    align-items:flex-start;
    height:60px;
    .show-doc {
      margin-left: 20px;
      line-height: 60px;
      color: $color;
      text-decoration: underline;
      cursor: pointer;
    }
    .header-api {
      padding: 6px 10px;
      box-shadow: 1px 2px 2px #ddd;
      background-color: #fff;
      border-radius: 6px;
      .root-url {
        background-color: #eceefe;
        color: $color;
        border: none;
      }
    }
    .apikeyBtn{
      margin-left:10px;
      border:1px solid $btn_bg;
      padding:12px;
      color: $btn_bg;
      display:flex;
      align-items:center;
    }
  }

</style>