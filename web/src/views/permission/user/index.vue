<template>
  <div>
    <div class="table-wrap list-common wrap-fullheight">
      <div class="table-box">
        <search-input
          style="margin-right: 2px; margin-bottom: 20px"
          :placeholder="$t('user.form.user')"
          ref="searchInput"
          @handleSearch="getTableData"
        />
        <el-button v-if="!isSystem && isAdmin" style="margin-left: 13px" class="add-bt" size="mini" type="primary" @click="preInsert">
          <img src="@/assets/imgs/addUser.png" alt="" />
          <span>{{$t('user.button.create')}}</span>
        </el-button>
        <el-button v-if="!isSystem && isAdmin" class="add-bt invite-bt" size="mini" @click="handleInviteUser">
          <img src="@/assets/imgs/inviteUser.png" alt="" />
          <span>{{$t('user.button.invite')}}</span>
        </el-button>
        <el-table
          :data="tableData"
          :header-cell-style="{background: '#F9F9F9', color: '#999999'}"
          v-loading="loading"
          style="width: 100%"
        >
          <el-table-column prop="username" :label="$t('user.table.username')" align="left" />
          <el-table-column v-if="isSystem" :label="$t('user.table.company')" align="left">
            <template slot-scope="scope">
              <div>{{scope.row.company || '--'}}</div>
            </template>
          </el-table-column>
          <el-table-column v-if="!isSystem" :label="$t('user.table.role')" align="left">
            <template slot-scope="scope">
              <div
                v-if="scope.row.orgs && scope.row.orgs.length"
                v-for="(orgItem, orgIndex) in scope.row.orgs"
                :key="orgItem.org.id + orgIndex"
              >
                {{Array.isArray(orgItem.roles) ? (orgItem.roles.map(item => item.name).join(',') || '--') : '--' }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" :label="$t('user.table.createAt')" align="left" />
          <el-table-column v-if="isAdmin" align="left" :label="$t('user.table.status')">
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
          <el-table-column v-if="isAdmin" align="left" :label="$t('common.table.operation')" width="300">
            <template slot-scope="scope">
              <el-button class="operation" type="text" @click="preUpdate(scope.row)">{{$t('common.button.edit')}}</el-button>
              <el-button class="operation" type="text" @click="preDel(scope.row)">{{$t('common.button.delete')}}</el-button>
              <el-button type="text" @click="resetPsw(scope.row)">{{$t('user.table.resetPassword')}}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <Pagination class="pagination" ref="pagination" :listApi="listApi" @refreshData="refreshData" />
    </div>

    <el-dialog
      :title="isEdit ? $t('user.button.edit') : $t('user.button.create')"
      :visible.sync="dialogVisible"
      width="580px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <el-form :model="form" :rules="rules" ref="form" style="margin-top: -16px">
        <el-form-item :label="$t('user.table.username')" prop="username" >
          <el-input v-model="form.username" :disabled="isEdit" :placeholder="$t('common.hint.userName')" clearable />
        </el-form-item>
        <el-form-item :label="$t('user.dialog.password')" :prop="!isEdit ? 'password' : ''">
          <el-input
            v-model="form.password"
            type="password"
            :disabled="isEdit"
            :placeholder="isEdit ? '******' : $t('user.dialog.pwdPlaceholder')"
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('user.dialog.company')" prop="company">
          <el-input v-model="form.company" :placeholder="$t('common.input.placeholder')" maxlength="50" show-word-limit clearable />
        </el-form-item>
        <el-form-item :label="$t('user.dialog.phone')" prop="phone">
          <el-input v-model="form.phone" :placeholder="$t('common.input.placeholder')" clearable />
        </el-form-item>
        <el-form-item v-if="!isSystem" :label="$t('user.table.role')" prop="roleIds">
          <el-select
            v-model="form.roleIds"
            :placeholder="$t('common.select.placeholder')"
            :disabled="row.username === 'admin'"
            style="width: 540px"
            clearable
          >
            <el-option v-for="item in roleList" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isEdit" :label="$t('user.dialog.email')" prop="email">
          <el-input :disabled="isEdit" v-model="form.email" :placeholder="$t('common.noBindEmail')" clearable />
        </el-form-item>
        <el-form-item :label="$t('user.dialog.remark')" prop="remark" class="mark-textArea">
          <el-input type="textarea" :rows="3" v-model="form.remark" :placeholder="$t('common.input.placeholder')" maxlength="100" show-word-limit clearable />
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button size="small" @click="handleClose">{{$t('common.button.cancel')}}</el-button>
        <el-button size="small" type="primary" :loading="submitLoading" @click="handleSubmit">{{$t('common.button.confirm')}}</el-button>
      </span>
    </el-dialog>

    <!--邀Please用户-->
    <el-dialog
      v-if="!isSystem"
      :title="$t('user.button.invite')"
      :visible.sync="inviteVisible"
      width="580px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleInviteClose"
    >
      <el-form :model="inviteForm" :rules="inviteRules" ref="inviteForm" style="margin-top: -16px">
        <el-form-item :label="$t('user.inviteDialog.user')" prop="userId">
          <el-select
            filterable
            v-model="inviteForm.userId"
            :placeholder="$t('user.inviteDialog.searchPlaceholder')"
            style="width: 540px"
            :filter-method="searchInviteUserList"
            clearable
          >
            <el-option v-for="item in inviteUserList" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button size="small" @click="handleInviteClose">{{$t('common.button.cancel')}}</el-button>
        <el-button size="small" type="primary" :loading="submitLoading" @click="inviteUser">{{$t('common.button.confirm')}}</el-button>
      </span>
    </el-dialog>
    <reset-pwd ref="resetPwd" />
  </div>
</template>

<script>
  import Pagination from "@/components/pagination.vue"
  import resetPwd from '../components/resetPwd'
  import SearchInput from "@/components/searchInput.vue"
  import { urlEncrypt } from "@/utils/crypto"
  import { debounce } from 'throttle-debounce'
  import { fetchUserList, fetchInviteUser, fetchRoleList, inviteUser, createUser, editUser, deleteUser, changeUserStatus } from "@/api/permission/user"
  import { mapActions } from "vuex"
  import { checkPerm } from "@/router/permission"
  import { PERMS } from "@/router/constants"
  export default {
    components: { Pagination, resetPwd, SearchInput },
    data() {
      const checkPassword = (rule, value, callback) => {
        let reg = /^(?=.*[a-zA-Z])(?=.*\d)(?=.*[~!@#$%^&*()_+`\-={}:";'<>?,./]).{8,20}$/
        if (!reg.test(value)) {
          callback(new Error(this.$t('user.dialog.passwordError')))
        } else {
          return callback()
        }
      }
      const checkPhone = (rule, value, callback) => {
        let reg = /^1[3-9][0-9]{9}$/
        if (!reg.test(value)) {
          callback(new Error(this.$t('user.dialog.phoneError')))
        } else {
          return callback()
        }
      }
      return {
        isSystem: this.$store.state.user.permission.isSystem || false,
        isAdmin: this.$store.state.user.permission.isAdmin || false,
        listApi: fetchUserList,
        loading: false,
        isEdit: false,
        inviteLoading: false,
        inviteUserList: [],
        roleList: [],
        form: {
          username: '',
          password: '',
          company: '',
          phone: '',
          email: '',
          roleIds: '',
          remark: '',
        },
        inviteForm: {
          userId: '',
        },
        rules: {
          username: [
            { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
            { min: 2, max: 20, message: this.$t('common.hint.userNameLimit'), trigger: 'blur'},
            { pattern: /^(?!_)[a-zA-Z0-9_.\u4e00-\u9fa5]+$/, message: this.$t('common.hint.userName'), trigger: "blur"} // 结尾：(?!.*?_$)
          ],
          password: [
            { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
            { validator: checkPassword, trigger: "blur" }
          ],
          company: [
            { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
            { max: 50, message: this.$t('common.hint.companyLimit'), trigger: 'blur'},
          ],
          phone: [
            { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
            { validator: checkPhone, trigger: 'blur' }
          ],
          email: [
            // { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
            { pattern: /^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(.[a-zA-Z0-9_-]+)+$/, message: this.$t('common.hint.emailError'), trigger: "blur"}
          ],
          remark: [
            { max: 100, message: this.$t('common.hint.remarkLimit'), trigger: 'blur'},
          ]
        },
        inviteRules: {
          userId: [
            { required: true, message: this.$t('common.select.placeholder'), trigger: 'change' },
          ],
        },
        tableData: [],
        inviteVisible: false,
        dialogVisible: false,
        submitLoading: false,
        row: {}
      }
    },
    created() {
      this.getInviteUserList = debounce(500, async (name) => {
        const {data} = await fetchInviteUser({name})
        this.inviteUserList = data.select || []
      })
    },
    mounted() {
      this.getTableData()
      if (!this.isSystem) this.getRoleList()
    },
    methods: {
      ...mapActions('user', ['getPermissionInfo']),
      async getRoleList() {
        const {data} = await fetchRoleList()
        this.roleList = data.select || []
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
      searchInviteUserList(val) {
        if (val) this.getInviteUserList(val)
      },
      handleInviteUser() {
        this.inviteVisible = true
        this.getInviteUserList()
      },
      inviteUser() {
        this.$refs.inviteForm.validate(async (valid) => {
          if (!valid) return
          this.submitLoading = true
          try {
            const res = await inviteUser({...this.inviteForm})
            if (res.code === 0) {
              this.$message.success(this.$t('user.inviteDialog.success'))
              this.handleInviteClose()
              await this.getTableData()
            }
          } finally {
            this.submitLoading = false
          }
        })
      },
      handleInviteClose() {
        this.inviteVisible = false
        for (let key in this.inviteForm) {
          this.inviteForm[key] = ''
        }
        this.$refs.inviteForm.resetFields()
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
        this.dialogVisible = true
      },
      preUpdate(row) {
        this.row = row
        this.isEdit = true
        const curOrg = row.orgs ? (row.orgs[0] || {}) : {}
        this.setFormValue({
          ...row,
          roleIds: curOrg.roles && curOrg.roles[0] ? curOrg.roles[0].id : ''
        })

        this.dialogVisible = true
      },
      preDel(row) {
        this.$confirm(this.$t('user.confirm.delete'), this.$t('common.confirm.title'), {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning'
        }).then(async () => {
          let res = await deleteUser({userId: row.userId})
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'))
            await this.getTableData()
          }
        })
      },
      resetPsw(row) {
        this.$refs.resetPwd.openDialog(row)
      },
      changeStatus(row, val) {
        this.$confirm(val ? this.$t('user.switch.startHint') : this.$t('user.switch.stopHint'), this.$t('common.confirm.title'), {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning'
        }).then(async() => {
          let res = await changeUserStatus({userId: row.userId, status: val})
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
          params.password = urlEncrypt(this.form.password)
          params.nickname = this.form.username
          params.roleIds = params.roleIds ? [params.roleIds] : []
          if (this.isEdit) params.userId = this.row.userId

          try {
            const res = this.isEdit ? await editUser(params) : await createUser(params)
            if (res.code === 0) {
              this.$message.success(this.$t('common.message.success'))
              this.dialogVisible = false

              // IfModify of Yes当前User，ThenUpdatePermission
              const useInfo = this.$store.state.user.userInfo || {}
              if (useInfo.uid === this.row.userId) {
                await this.getPermissionInfo()
                if (checkPerm(PERMS.PERMISSION_USER)) {
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
    .invite-bt {
      margin-left: 15px;
      color: $color;
      border-color: $color;
      background: rgba(255, 255, 255, 0) !important;
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
