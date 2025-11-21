<template>
    <div>
        <div class="pwd">
            <el-form label-width="100px" :model="form" :rules="rules" ref="form" class="form">
                <el-form-item :label="$t('resetPwd.oldPwd')" prop="oldPassword">
                    <el-input v-model="form.oldPassword" type="password" :placeholder="$t('common.input.placeholder') + $t('resetPwd.oldPwd')" clearable></el-input>
                </el-form-item>
                <el-form-item :label="$t('resetPwd.newPwd')" prop="newPassword">
                    <el-input v-model="form.newPassword" type="password" :placeholder="$t('resetPwd.pwdPlaceholder')" clearable></el-input>
                </el-form-item>
                <el-form-item :label="$t('resetPwd.confirmPwd')" prop="newPasswordAgain">
                    <el-input v-model="form.newPasswordAgain" type="password" :placeholder="$t('resetPwd.confirmText')" clearable></el-input>
                </el-form-item>
            </el-form>
            <div class="form-footer">
                <el-button size="small" @click="cancel">{{$t('common.button.cancel')}}</el-button>
                <el-button size="small" type="primary" @click="doSubmit">{{$t('common.button.confirm')}}</el-button>
            </div>
        </div>

        <!--ModifySuccess弹框-->
        <el-dialog
            title=""
            :visible.sync="dialogVisible"
            width="450px"
            append-to-body
            :before-close="reLogin"
            :close-on-click-modal="false"
            :show-close="false"
        >
            <p style="text-align: center"><i class="el-icon-circle-check" style="font-size: 42px;color: #5a9600;"></i></p>
            <p style="margin: 20px 0 30px 0;text-align: center">{{$t('resetPwd.dialog.success')}}&nbsp;<span style="color: #5a9600;font-weight: bold">{{jumpTimer}}</span>&nbsp;{{$t('resetPwd.dialog.jumpText')}}</p>
        </el-dialog>
    </div>
</template>

<script>
    import { urlEncrypt } from "@/utils/crypto";
    import {restOwnPassword} from '@/api/user'

    export default {
        data(){
            var checkPwd = (rule, value, callback) => {
                if (this.form.newPassword != this.form.newPasswordAgain) callback(new Error(this.$t('resetPwd.differError')))
                callback();
            };
            var checkPassword = (rule, value, callback) => {
                var reg = /^(?=.*[a-zA-Z])(?=.*\d)(?=.*[~!@#$%^&*()_+`\-={}:";'<>?,./]).{8,20}$/;
                if (!reg.test(value)) {
                    callback(new Error(this.$t('resetPwd.pwdError')))
                } else {
                    return callback();
                }
            };
            var checkEmpty = (rule, value, callback,msg) => {
                if (value === '') {
                    callback(new Error(msg));
                }else {
                    return callback();
                }
            };
            return{
                userId: this.$store.state.user.userInfo.uid,
                form:{
                    oldPassword:'',
                    newPassword:'',
                    newPasswordAgain:''
                },
                rules: {
                    oldPassword: [
                        { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
                        { validator: (rule, value, callback) => checkEmpty(rule, value, callback, this.$t('common.input.placeholder')), trigger: "blur" }
                    ],
                    newPassword: [
                        { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
                        { validator: (rule, value, callback) => checkEmpty(rule, value, callback, this.$t('common.input.placeholder')), trigger: "blur" },
                        { validator: checkPassword, trigger: "blur" }
                    ],
                    newPasswordAgain: [
                        { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
                        { validator: (rule, value, callback) => checkEmpty(rule, value, callback, this.$t('common.input.placeholder')), trigger: "blur" },
                        { validator: checkPassword, trigger: "blur" },
                        { validator: checkPwd, trigger: "blur" },
                    ]
                },
                t:null,
                jumpTimer: 3,
                dialogVisible:false
            }
        },
        created(){

        },
        beforeDestroy(){
          this.clearT()
        },
        methods:{
            cancel(){
                this.$emit('handleClose')
            },
            clearT(){
              this.t && clearInterval(this.t)
            },
            doSubmit(){
                this.$refs.form.validate(async (valid) => {
                    if (!valid) return;
                    let params = {
                        userId:this.userId,
                        oldPassword: urlEncrypt(this.form.oldPassword),
                        newPassword: urlEncrypt(this.form.newPassword)
                    }
                    let res = await restOwnPassword(params)
                    if(res.code === 0){
                        this.dialogVisible = true
                        // this.$t('common.message.success')
                        this.t = setInterval(()=>{
                            this.jumpTimer --;
                            if(this.jumpTimer === 0){
                                this.clearT()
                                this.reLogin()
                            }
                        },1000)
                    }
                })
            },
            reLogin(){
                window.localStorage.removeItem('access_cert')
                window.location.href = window.location.origin + this.$basePath + '/aibase/login'
            }
        }
    }
</script>

<style lang="scss" scoped>
.routerview-container{
    background-color: #fff;
}
.pwd {
    margin: 10px auto 0;
    background-color: #fff;
    border-radius: 4px;
    .form-header{
        line-height: 90px;
        font-size: 18px;
        font-weight: bold;
        text-align: center;
        color: #333;
        border-bottom: 1px solid #f9f9f9;
        .el-icon-back{
            position: absolute;
            left:20px;
            top:30px;
            font-size: 20px;
        }
    }
    .form{
        margin-top: -30px;
    }
    .form-footer{
        text-align: right;
    }
}
</style>
