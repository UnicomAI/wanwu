<template>
  <div>
    <div class="table-wrap list-common wrap-fullheight">
      <div class="table-box">
        <search-input
          style="margin-bottom: 20px"
          :placeholder="$t('role.form.role')"
          ref="searchInput"
          @handleSearch="getTableData"
        />
        <el-button v-if="isAdmin" class="add-bt" size="mini" type="primary" @click="preInsert">
          <img src="@/assets/imgs/addRole.png" alt="" />
          <span>{{$t('role.button.create')}}</span>
        </el-button>
        <el-table
          :data="tableData"
          :header-cell-style="{background: '#F9F9F9', color: '#999999'}"
          v-loading="loading"
          style="width: 100%"
        >
          <el-table-column prop="name" :label="$t('role.table.name')" align="left" />
          <el-table-column prop="creator.name" :label="$t('role.table.creator')" align="left" />
          <el-table-column prop="createdAt" :label="$t('role.table.createAt')" align="left" />
          <el-table-column v-if="isAdmin" align="left" :label="$t('role.table.status')">
            <template slot-scope="scope">
              <div style="height: 26px">
                <el-switch
                  @change="(val)=>{changeStatus(scope.row,val)}"
                  style="display: block; height: 22px; line-height: 22px"
                  v-model="scope.row.status"
                  :active-text="$t('common.switch.start')"
                  :inactive-text="$t('common.switch.stop')"
                />
              </div>
            </template>
          </el-table-column>
          <el-table-column v-if="isAdmin" align="left" :label="$t('common.table.operation')" width="240">
            <template slot-scope="scope">
              <el-button class="operation" type="text" @click="preUpdate(scope.row)">{{$t('common.button.edit')}}</el-button>
              <el-button type="text" @click="preDel(scope.row)">{{$t('common.button.delete')}}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <Pagination class="pagination" ref="pagination" :listApi="listApi" @refreshData="refreshData" />
    </div>

    <el-dialog
      :title="isEdit ? $t('role.button.edit') : $t('role.button.create')"
      :visible.sync="dialogVisible"
      width="580px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <el-form :model="form" :rules="rules" ref="form" style="margin-top: -16px">
        <el-form-item :label="$t('role.table.name')" prop="name" >
          <el-input v-model="form.name" :placeholder="$t('common.hint.roleName')" clearable />
        </el-form-item>
        <el-form-item :label="$t('role.dialog.perm')" prop="permissions">
          <select-tree
            ref="permTree" 
            :data-list="permList"
            :default-value="defaultPermValue"
            :tree-key-map="{value: 'perm'}"
            :disabled="row.isAdmin"
            @handleChange="changeTree"
          />
        </el-form-item>
        <el-form-item :label="$t('role.dialog.remark')" prop="remark" class="mark-textArea">
          <el-input type="textarea" :rows="3" v-model="form.remark" :placeholder="$t('common.input.placeholder')" maxlength="100" show-word-limit clearable />
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button size="small" @click="handleClose">{{$t('common.button.cancel')}}</el-button>
        <el-button size="small" type="primary" :loading="submitLoading" @click="handleSubmit">{{$t('common.button.confirm')}}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import Pagination from "@/components/pagination.vue"
import SearchInput from "@/components/searchInput.vue"
import SelectTree from "../components/selectTree.vue"
import { fetchPermTree, createRole, editRole, deleteRole, fetchRoleList, changeRoleStatus } from "@/api/permission/role"
import { mapActions } from "vuex"
import { checkPerm } from "@/router/permission"
import { PERMS } from "@/router/constants"
export default {
  components: { Pagination, SearchInput, SelectTree },
  data(){
    return {
      isAdmin: this.$store.state.user.permission.isAdmin || false,
      listApi: fetchRoleList,
      loading: false,
      isEdit: false,
      permList: [],
      defaultPermValue: [],
      form: {
        name: '',
        remark: '',
        permissions: [],
      },
      rules: {
        name: [
          { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
          { min: 1, max: 64, message: this.$t('common.hint.roleNameLimit'), trigger: 'blur'},
          { pattern: /^[a-zA-Z0-9_.\u4e00-\u9fa5]+$/, message: this.$t('common.hint.roleName'), trigger: "blur"}
        ],
        remark: [
          { max: 100, message: this.$t('common.hint.remarkLimit'), trigger: 'blur'},
        ]
      },
      tableData: [],
      dialogVisible: false,
      submitLoading: false,
      row: {}
    }
  },
  created() {},
  mounted() {
    this.getTableData()
    this.getPermTree()
  },
  methods:{
    ...mapActions('user', ['getPermissionInfo']),
    changeTree (perms) {
      this.form.permissions = perms
    },
    async getPermTree() {
      const {data} = await fetchPermTree()
      this.permList = data.routes || []
    },
    async getTableData() {
      const searchInput = this.$refs.searchInput
      const searchInfo = {...searchInput.value && {name: searchInput.value}}
      this.loading = true
      try {
        this.tableData = await this.$refs.pagination.getTableData(searchInfo)
      } finally {
        this.loading = false
      }
    },
    // Get从分页Component传递 of  data
    refreshData(data) {
      this.tableData = data
    },
    setFormValue(row) {
      const obj = {...this.form}
      for (let key in obj) {
        obj[key] = row ? row[key] : (Array.isArray(obj[key]) ? [] : '')
      }
      this.form = obj
    },
    handleClose() {
      this.$refs.form.resetFields()
      this.dialogVisible = false
    },
    preInsert() {
      this.isEdit = false
      this.row = {}
      this.setFormValue()

      this.defaultPermValue = []
      this.dialogVisible = true
    },
    preUpdate(row) {
      this.row = row
      this.isEdit = true
      // Process一级PermissionBack问题
      const perms = row.permissions || []
      const permKeys = perms.map(item => item.perm)
      const permissions = perms.map((item) => {
        if (permKeys.some(key => key.includes(`${item.perm}.`))) {
          return null
        } else {
          return item
        }
      }).filter(item => item) || []
      this.setFormValue({...row, permissions: permissions.map(item => item.perm)})

      this.defaultPermValue =  permissions || []
      this.dialogVisible = true
    },
    preDel(row) {
      this.$confirm(this.$t('role.confirm.delete'), this.$t('common.confirm.title'), {
        confirmButtonText: this.$t('common.confirm.confirm'),
        cancelButtonText: this.$t('common.confirm.cancel'),
        type: 'warning'
      }).then(async () => {
        let res = await deleteRole({roleId: row.roleId})
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'))
          await this.getTableData()
        }
      })
    },
    changeStatus(row, val) {
      this.$confirm(val ? this.$t('role.switch.startHint') : this.$t('role.switch.stopHint'), this.$t('common.confirm.title'), {
        confirmButtonText: this.$t('common.confirm.confirm'),
        cancelButtonText: this.$t('common.confirm.cancel'),
        type: 'warning'
      }).then(async() => {
        let res = await changeRoleStatus({roleId: row.roleId, status:val})
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'))
          await this.getTableData()
        }
      }).catch(() => {
        this.getTableData()
      })
    },
    handleSubmit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return

        this.submitLoading = true
        const params = { ...this.form }
        if (this.isEdit) params.roleId = this.row.roleId

        try {
          const res = this.isEdit ? await editRole(params) : await createRole(params)
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'))
            this.dialogVisible = false

            // IfcurrentUserHas这Role，ThenUpdatePermission
            const permission = this.$store.state.user.permission || {}
            const roles = permission.roles ? permission.roles.map(item => item.id) : []
            if (roles.includes(this.row.roleId)) {
              await this.getPermissionInfo()
              if (checkPerm(PERMS.PERMISSION_ROLE)) {
                await this.getTableData()
                return
              }
              window.location.reload()
              return
            }
            await this.getTableData()
          }
        } finally {
          this.submitLoading = false
        }
      })
    }
  }
}
</script>

<style lang="scss" scoped>
.routerview-container {
  top:0;
}
.table-box {
  margin-top: -20px;
  text-align: right;
  .table-header {
    font-size: 16px;
    font-weight: bold;
    color: #555;
  }
  .add-bt {
    margin: 0 0 20px;
    img {
      width: 16px;
      margin-right: 5px;
      display: inline-block;
      vertical-align: middle;
    }
    span {
      display: inline-block;
      vertical-align: middle;
    }
  }
  /deep/ .el-switch__label * {
    font-size: 13px;
  }
}
.mark-textArea /deep/ {
  .el-textarea__inner {
    font-family: inherit;
    font-size: inherit;
  }
}
/deep/ .operation.el-button--text.el-button {
  padding: 3px 10px 3px 0;
  border-right: 1px solid #EAEAEA !important;
}
</style>
