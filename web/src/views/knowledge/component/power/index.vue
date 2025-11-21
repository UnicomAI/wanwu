<template>
  <div class="power-management">
    <el-dialog
      :visible.sync="dialogVisible"
      width="60%"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      class="power-management-dialog"
      @close="handleDialogClose"
    >
      <div slot="title" class="custom-dialog-title">
        <div class="title-content">
          <i class="el-icon-s-custom title-icon"></i>
          <span class="title-text">{{ dialogTitle }}</span>
          <span class="title-tip" v-if="currentView === 'transfer'">[ 转移after您 of Permission将变is'Edit' ]</span>
        </div>
        <div class="title-subtitle" v-if="knowledgeName">
          <span class="knowledge-name">[ {{ knowledgeName }} ]</span>
        </div>
      </div>
        <div class="list-header" v-if="currentView === 'list'">
         <el-input placeholder="EnterMemberNameSearch" v-model="userName" style="width:200px;margin-right:10px;"></el-input>
           <el-button
            type="primary"
            size="small"
            @click="confirmSearch"
          >Confirm</el-button>
          <el-button
            type="primary"
            size="small"
            icon="el-icon-plus"
            @click="showCreate"
            :disabled="[0, 10].includes(permissionType)"
          >Add</el-button>
        </div>
        <PowerList ref="powerList" v-if="currentView === 'list'" @transfer="showTransfer" :knowledgeId="knowledgeId" :permissionType="permissionType" />
        <PowerCreate ref="powerCreate" v-if="currentView === 'create'" :knowledgeId="knowledgeId" />
        <PowerCreate ref="powerTransfer" v-if="currentView === 'transfer'" :transfer-mode="true" :knowledgeId="knowledgeId" />
      <div
        slot="footer"
        class="dialog-footer"
      >
        <el-button
          v-if="currentView === 'create' || currentView === 'transfer'"
          @click="showList"
        >Back</el-button>
        <el-button
          v-if="currentView === 'create'"
          type="primary"
          @click="handleConfirm"
        >Confirm</el-button>
        <el-button
          v-if="currentView === 'transfer'"
          type="primary"
          @click="handleTransferConfirm"
        >ConfirmTransfer</el-button>
        <el-button
          v-if="currentView === 'list'"
          @click="handleDialogClose"
        >Close</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import PowerList from "./list.vue";
import PowerCreate from "./create.vue";
import { transferUserPower,addUserPower } from "@/api/knowledge";
export default {
  name: "PowerManagement",
  inject: ['reloadKnowledgeData'],
  components: {
    PowerList,
    PowerCreate,
  },
  data() {
    return {
      currentView: "list",
      dialogVisible: false,
      knowledgeId:'',
      knowledgeName:'',
      permissionType:'',
      currentTransferUser: null,
      userName:''
    };
  },
  computed: {
    dialogTitle() {
      if (this.currentView === "list") {
        return "Permission管理";
      } else if (this.currentView === "create") {
        return "AddPermission";
      } else if (this.currentView === "transfer") {
        return "TransferPermission";
      }
      return "Permission管理";
    },
  },
  mounted() {
    this.knowledgeId = this.$route.query.knowledgeId;
  },
  methods: {
    confirmSearch(){
      this.$refs.powerList.getFilterResult(this.userName)
    },
    showDialog() {
      this.currentView = "list";
      this.dialogVisible = true;
      this.$nextTick(() => {
        if (this.$refs.powerList) {
            this.$refs.powerList.getUserPower()
        }
    })
    },
    showCreate() {
      this.currentView = "create";
    },

    showTransfer(row) {
      this.currentView = "transfer";
      this.currentTransferUser = row;
    },

    showList() {
      this.currentView = "list";
      this.$nextTick(() => {
        this.refreshList();
      });

    },
    handleConfirm() {
      const createData = this.$refs.powerCreate.getResults();
      const userData = this.handleData(createData);
      if (userData && userData.length > 0) {
        addUserPower({knowledgeId:this.knowledgeId,knowledgeUserList:userData}).then(res => {
          if(res.code === 0){
            this.$message.success("AddSuccess");
            this.showList();
            this.refreshList();
          }
        }).catch(() => {})
      }else{
        this.$message.error("Please selectuser");
      }
    },
    handleData(createData){
      if (createData.node.length > 0) {
        var userList = [];
        createData.node.forEach(function(group) {
          group.users.forEach(function(user) {
            userList.push({
              userId: user.id,
              orgId: user.orgId,
              permissionType:createData.selectedPermission
            });
          });
        });
        return userList;
      }
      return [];
    },
    handleDialogClose() {
      this.dialogVisible = false;
    },

    // ConfirmTransferPermission
    handleTransferConfirm() {
      const data = this.$refs.powerTransfer.getTransferData();
      const params = {
        ...data,
        permissionId: this.currentTransferUser.permissionId
      }
      if (data.knowledgeUser && !Array.isArray(data.knowledgeUser)) {
        transferUserPower(params).then(res => {
          if(res.code === 0){
            this.$message.success("TransferSuccess");
            this.showList();
            this.dialogVisible = false;
             if (this.reloadKnowledgeData) {
              this.reloadKnowledgeData();  // RefreshList
            }
          }
        }).catch(() => {})
      }else{
        this.$message.error("Please selectuser");
      }
    },
    refreshList() {
      if (this.$refs.powerList) {
        this.$refs.powerList.getUserPower()
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.power-management {
  .list-header {
    display: flex;
    justify-content:flex-start;
    align-items: center;
    .header-left {
      .page-title {
        font-size: 18px;
        font-weight: 600;
        color: #303133;
      }
    }

    .header-right {
      .el-button {
        border-radius: 4px;
      }
    }
  }
}

.power-management-dialog {
  /deep/ .el-dialog {
    border-radius: 8px;
  }

  /deep/ .el-dialog__header {
    padding: 20px 20px 10px 20px;
    border-bottom: 1px solid #e4e7ed;
  }

  /deep/ .el-dialog__body {
    padding: 20px;
    max-height: 70vh;
    overflow-y: auto;
  }

  /deep/ .el-dialog__footer {
    padding: 10px 20px 20px 20px;
    text-align: right;
    border-top: 1px solid #e4e7ed;
  }
}

.custom-dialog-title {
  display: flex;
  align-items: center;
  
  .title-content {
    display: flex;
    align-items: center;
    
    .title-icon {
      font-size: 20px;
      color: #409eff;
      margin-right: 8px;
    }
    
    .title-text {
      font-size: 18px;
      font-weight: 600;
      color: #303133;
    }
    .title-tip {
      margin-left:5px;
      font-size: 12px;
      color: $color;
    }
  }
  
  .title-subtitle {
    margin-left: 5px;
    
    .knowledge-name {
      font-size: 14px;
      color: #606266;
      padding: 4px 8px;
    }
  }
}

.dialog-footer {
  .el-button {
    padding: 8px 20px;
    border-radius: 4px;

    &.el-button--primary {
      background-color: #f56c6c;
      border-color: #f56c6c;

      &:hover {
        background-color: #f78989;
        border-color: #f78989;
      }
    }
  }
}
</style>