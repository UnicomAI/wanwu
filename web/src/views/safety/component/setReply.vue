<template>
    <div>
      <el-dialog
        top="10vh"
        title="ReplySetting"
        :close-on-click-modal="false"
        :visible.sync="dialogVisible"
        width="50%"
        :before-close="handleClose"
        >
        <el-form
            :model="ruleForm"
            ref="ruleForm"
            label-width="140px"
            class="demo-ruleForm"
            @submit.native.prevent
        >
            <el-form-item
            label="ReplySetting"
            prop="reply"
             :rules="[{ required: true, message: 'Please enterReplySetting', trigger: 'blur' },]"
            >
            <el-input 
            v-model="ruleForm.reply" 
            type="textarea"
            :rows="4"
            maxlength="100"
            show-word-limit
            placeholder="不超 100字"></el-input>
            <p class="tips">If不Setting,默认Reply：您当前RequestOR者历史Information中Contains敏感Content，根据相关安全要Request，该轮对话终止，Please清除历史InformationOR者重新开启新 of 对话。</p>
            </el-form-item>
        </el-form>
        <span
            slot="footer"
            class="dialog-footer"
        >
            <el-button 
                @click="handleClose()">
                {{$t('common.confirm.cancel')}}
            </el-button>
            <el-button
                type="primary"
                @click="submitForm('ruleForm')"
            >{{$t('common.confirm.confirm')}}</el-button>
        </span>
        </el-dialog>
    </div>
</template>
<script>
import { setReply,getReplay } from "@/api/safety";
export default {
    data(){
        return{
            ruleForm:{
                reply:'',
                tableId:''
            },
            dialogVisible:false
        }
    },
    methods:{
        getReplayInfo(tableId){
            getReplay({tableId}).then(res => {
                if(res.code === 0){
                    this.ruleForm.reply = res.data.reply || '';
                }
            }).catch(err => {
                console.log(err)
            })
        },
        showDialog(tableId){
            this.dialogVisible = true;
            this.ruleForm.tableId = tableId;
            this.getReplayInfo(this.ruleForm.tableId)
            this.clear();
        },
        handleClose(){
            this.clear()
            this.dialogVisible = false;
        },
        clear(){
            this.$nextTick(() =>{
                this.$refs.ruleForm.resetFields();
                this.$refs.ruleForm.clearValidate();
            })
        },
        submitForm(formName){
            this.$refs[formName].validate((valid) =>{
                if(valid){
                    setReply(this.ruleForm).then(res =>{
                        if(res.code === 0){
                            this.$message.success('SettingSuccess')
                            this.dialogVisible = false;
                        }
                    })
                }else{
                    return false;
                }
            })
        },
    }
}
</script>
<style lang="scss" scoped>
.tips{color:#888888;line-height:1.5;margin-top: 10px;}
</style>