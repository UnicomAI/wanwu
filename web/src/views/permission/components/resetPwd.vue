<template>
  <div>
    <el-dialog
      :title="$t('resetPwd.title')"
      :visible.sync="dialogVisible"
      width="480px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <el-form :model="form" :rules="rules" ref="form">
        <el-form-item :label="$t('resetPwd.newPwd')" prop="newPassword">
          <el-input
            v-model="form.newPassword"
            type="password"
            :placeholder="$t('resetPwd.pwdPlaceholder')"
            clearable
          />
        </el-form-item>
        <el-form-item
          :label="$t('resetPwd.confirmPwd')"
          prop="newPasswordAgain"
        >
          <el-input
            v-model="form.newPasswordAgain"
            type="password"
            :placeholder="$t('resetPwd.confirmText')"
            clearable
          />
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button size="small" @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button size="small" type="primary" @click="handleSubmit">
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { rsaEncrypt } from '@/utils/crypto';
import { restUserPassword } from '@/api/user';
export default {
  props: {
    orgId: {
      type: String,
      default: '',
    },
  },
  data() {
    let checkPwd = (rule, value, callback) => {
      if (this.form.newPassword !== this.form.newPasswordAgain)
        callback(new Error(this.$t('resetPwd.differError')));
      callback();
    };
    let checkPassword = (rule, value, callback) => {
      let reg =
        /^(?=.*[a-zA-Z])(?=.*\d)(?=.*[~!@#$%^&*()_+`\-={}:";'<>?,./]).{8,20}$/;
      if (!reg.test(value)) {
        callback(new Error(this.$t('resetPwd.pwdError')));
      } else {
        return callback();
      }
    };
    let checkEmpty = (rule, value, callback, msg) => {
      if (value === '') {
        callback(new Error(msg));
      } else {
        return callback();
      }
    };
    return {
      row: {},
      form: {
        newPassword: '',
        newPasswordAgain: '',
      },
      rules: {
        newPassword: [
          {
            validator: (rule, value, callback) =>
              checkEmpty(
                rule,
                value,
                callback,
                this.$t('common.input.placeholder'),
              ),
            trigger: 'blur',
          },
          { validator: checkPassword, trigger: 'blur' },
        ],
        newPasswordAgain: [
          {
            validator: (rule, value, callback) =>
              checkEmpty(
                rule,
                value,
                callback,
                this.$t('common.input.placeholder'),
              ),
            trigger: 'blur',
          },
          { validator: checkPassword, trigger: 'blur' },
          { validator: checkPwd, trigger: 'blur' },
        ],
      },
      dialogVisible: false,
    };
  },
  created() {},
  methods: {
    openDialog(row) {
      this.row = row;
      this.dialogVisible = true;
      this.$nextTick(() => {
        this.$refs.form.resetFields();
      });
    },
    handleClose() {
      for (let key in this.form) {
        this.form[key] = '';
      }
      this.$refs.form.resetFields();
      this.dialogVisible = false;
    },
    handleSubmit() {
      this.$refs.form.validate(async valid => {
        if (!valid) return;
        const { cipher, keyId } = await rsaEncrypt(this.form.newPassword);
        let params = {
          userId: this.row.userId,
          cipher,
          keyId,
          ...(this.orgId && { orgId: this.orgId }),
        };
        let res = await restUserPassword(params);
        if (res.code === 0) {
          this.$message.success(this.$t('resetPwd.success'));
          this.dialogVisible = false;
        }
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.table-box {
  padding: 20px;
  .table-header {
    font-size: 16px;
    font-weight: bold;
    color: #555;
  }
  .add-bt {
    margin: 20px 0;
  }
}
</style>
