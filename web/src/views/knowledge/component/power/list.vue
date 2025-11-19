<template>
  <div class="power-list-container">
    <div class="table-content">
      <el-table
        :data="tableData.filter(data => !name || data.userName.toLowerCase().includes(name.toLowerCase()))"
        style="width: 100%"
        class="power-table"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266' }"
        border
        v-loading="loading"
      >
        <el-table-column prop="userName" label="成员" width="200">
          <template slot-scope="scope">
            <div class="name-cell">
              <span class="name-text">{{ scope.row.userName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="orgName" label="组织" width="200">
          <template slot-scope="scope">
            <div class="org-cell">
              <span class="org-text">{{ scope.row.orgName || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="permissionType" label="权限">
          <template slot-scope="scope">
            <div class="type-cell">
              <span v-if="!scope.row.editing" class="type-text">{{ powerType[scope.row.permissionType] }}</span>
              <el-select 
                v-else 
                v-model="scope.row.permissionType" 
                size="small" 
                class="permission-select"
              >
                <el-option label="可读" :value="0"></el-option>
                <el-option label="可Edit" :value="10"></el-option>
                <el-option label="管理员" :value="20"></el-option>
              </el-select>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Operation" width="180" align="center">
          <template slot-scope="scope">
            <div class="action-buttons">
              <!-- System管理员权限：只Show转让Button -->
              <template v-if="scope.row.transfer && !scope.row.editing">
                <el-button
                  type="text"
                  size="small"
                  icon="el-icon-s-promotion"
                  @click="handleTransfer(scope.row)"
                  class="action-btn transfer-btn"
                >
                  转让
                </el-button>
              </template>
              <!-- 非管理员权限：ShowEditAndDeleteButton -->
              <template v-if="scope.row.editing">
                <el-button
                  type="text"
                  size="small"
                  icon="el-icon-check"
                  @click="handleSave(scope.row)"
                  class="action-btn save-btn"
                >
                  Save
                </el-button>
                <el-button
                  type="text"
                  size="small"
                  icon="el-icon-close"
                  @click="handleCancel(scope.row)"
                  class="action-btn cancel-btn"
                >
                  Cancel
                </el-button>
              </template>
              <template v-if="showEdit(scope.row)">
                <el-button
                  type="text"
                  size="small"
                  icon="el-icon-edit"
                  @click="handleEdit(scope.row)"
                  class="action-btn edit-btn"
                >
                  Edit
                </el-button>
                <el-button
                  type="text"
                  size="small"
                  icon="el-icon-delete"
                  @click="handleDelete(scope.row)"
                  class="action-btn delete-btn"
                >
                  Delete
                </el-button>
              </template>
              <span v-else-if="showInfo(scope.row)" class="noPower">--</span>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script>
import { getUserPower,editUserPower,delUserPower } from "@/api/knowledge";
import { POWER_TYPE } from "@/views/knowledge/config";
export default {
  name: 'PowerList',
  props: {
    knowledgeId: {
      type: String,
      default: ''
    },
    permissionType:{
      type:Number,
      default:0
    }
  },
  data() {
    return {
      powerType: POWER_TYPE,
      tableData: [],
      name:'',
      loading:false
    }
  },
  methods: {
    showEdit(row){
      if (row.editing) return false;
      return (
        !this.permissionType === 0 ||
        !this.permissionType === 10 ||
        (this.permissionType === 20 && row.permissionType === 0) ||
        (this.permissionType === 20 && row.permissionType === 10) ||
        (this.permissionType === 30 && row.permissionType === 0) ||
        (this.permissionType === 30 && row.permissionType === 10) ||
        (this.permissionType === 30 && row.permissionType === 20)
      );
    },
    showInfo(row){
      if (row.editing) return false;
      return (
        row.permissionType === 0 ||
        row.permissionType === 10 ||
        (this.permissionType === 0 && !row.transfer) ||
        (this.permissionType === 20 && !row.transfer) ||
        (this.permissionType === 20 && row.permissionType === 20)||
        (this.permissionType === 10 && row.permissionType === 30) ||
        (this.permissionType === 10 && row.permissionType === 20)
      );
    },
    getFilterResult(name) {
      this.name = name;
    },
    getUserPower() {
      this.loading = true;
      getUserPower({knowledgeId:this.knowledgeId}).then(res => {
        if(res.code === 0){
          this.loading = false;
          var list = res.data.knowledgeUserInfoList || [];
          this.tableData = list.map(function(item) {
            item.editing = false;
            return item;
          });
        }
      }).catch(() => {
        this.loading = false;
      })
    },
    handleEdit(row) {
      row.editing = true
      row.originalType = row.permissionType // Save原始Value
    },
    handleSave(row) {
      // SaveEdit
      row.editing = false
      row.originalType = row.permissionType
      const knowledgeUser = {
          orgId:row.orgId,
          userId:row.userId,
          permissionType:row.permissionType,
          permissionId:row.permissionId
        }
      editUserPower({knowledgeId:this.knowledgeId,knowledgeUser:knowledgeUser}).then(res => {
        if(res.code === 0){
          this.$message.success('权限ModifySuccess')
          this.getUserPower()
        }
      }).catch(() => {})
    },
    handleCancel(row) {
      row.permissionType = row.originalType
      row.editing = false
    },
    handleTransfer(row) {
      this.$confirm('确定要转让管理员权限吗？转让后您将失去管理员权限。', '转让Confirm', {
        confirmButtonText: '确定转让',
        cancelButtonText: 'Cancel',
        type: 'warning'
      }).then(() => {
        this.$emit('transfer', row)
      }).catch(() => {
        this.$message.info('已Cancel转让')
      })
    },
    handleDelete(row) {
      this.$confirm('确定要Delete这条Data吗？', 'Tip', {
        confirmButtonText: '确定',
        cancelButtonText: 'Cancel',
        type: 'warning'
      }).then(() => {
        delUserPower({knowledgeId:this.knowledgeId,permissionId:row.permissionId}).then(res => {
          if(res.code === 0){
            this.$message.success('DeleteSuccess')
            this.getUserPower()
          }
        }).catch(() => {})
      }).catch(() => {
        this.$message.info('已CancelDelete')
      })
    }
  }
}
</script>

<style lang="scss" scoped>
.power-list-container {
  padding-top: 15px;
  background: #fff;
  border-radius: 4px;
  
  .table-content {
    .power-table {
      border: 1px solid #e4e7ed;
      border-radius: 4px;
      .noPower{
        color:#ccc;
        font-size:12px;
      }
      /deep/ .el-table__header {
        th {
          background-color: #f5f7fa;
          color: #606266;
          font-weight: 500;
          border-bottom: 1px solid #e4e7ed;
          text-align: center;
        }
      }
      
      /deep/ .el-table__body {
        tr {
          &:hover {
            background-color: #f5f7fa;
          }
        }
        
        td {
          border-bottom: 1px solid #e4e7ed;
          padding: 12px 0;
        }
      }
      
      /deep/ .el-table__empty-block {
        background-color: #fafafa;
      }
    }
    
    .name-cell, .org-cell, .type-cell {
      display: flex;
      align-items: center;
      justify-content: center;
      
      .name-text, .org-text, .type-text {
        color: #606266;
        font-size: 14px;
      }
      
      .permission-select {
        width: 100%;
      }
    }
    
    .action-buttons {
      display: flex;
      justify-content: center;
      align-items: center;
      gap: 8px;
      
      .action-btn {
        padding: 4px 8px;
        border-radius: 4px;
        transition: all 0.3s;
        
        &.edit-btn {
          color: $btn_bg;
          
          &:hover {
            color: #5a6cff;
            background-color: #f0f2ff;
          }
        }
        
        &.transfer-btn {
          color: #e6a23c;
          
          &:hover {
            color: #ebb563;
            background-color: #fdf6ec;
          }
        }
        
        &.save-btn {
          color: #67c23a;
          
          &:hover {
            color: #85ce61;
            background-color: #f0f9ff;
          }
        }
        
        &.cancel-btn {
          color: #909399;
          
          &:hover {
            color: #a6a9ad;
            background-color: #f5f7fa;
          }
        }
        
        &.delete-btn {
          color: #f56c6c;
          
          &:hover {
            color: #f78989;
            background-color: #fef0f0;
          }
        }
      }
    }
  }
}
</style>