<template>
    <el-dialog
    title="Visual Settings"
    :visible.sync="dialogVisible"
    width="40%"
    :before-close="handleClose">
        <el-form :model="ruleForm" ref="ruleForm"  class="demo-ruleForm" label-width="120px" label-position="left">
            <el-form-item 
            label="imageUploadlimit" 
            prop="visionConfig.picNum"
            :rules="[{ required: true, message: 'Please选SettingimageUploadlimit', trigger: 'blur'}]"
            >
                <el-slider
                    v-model="ruleForm.visionConfig.picNum"
                    :min="1"
                    :max="ruleForm.visionConfig.maxPicNum"
                    :step="1"
                    show-input>
                </el-slider>
            </el-form-item>
        </el-form>
        <span slot="footer" class="dialog-footer">
            <el-button type="primary" @click="submit('ruleForm')">Confirm</el-button>
        </span>
    </el-dialog>
</template>
<script>
import { sensitiveSelect } from "@/api/safety";
export default {
    data(){
        return{
            dialogVisible:false,
            ruleForm:{
                visionConfig:{
                    picNum:3,
                    maxPicNum:6
                }
            }
        }
    },
    methods:{
        handleClose(){
            this.dialogVisible = false;
        },
        showDialog(visionConfig=null){
            this.dialogVisible = true;
            this.$nextTick(() =>{
                const form = this.$refs.ruleForm;
                 if (form) {
                    form.clearValidate();
                    if (!visionConfig) form.resetFields();
                }
                this.ruleForm.visionConfig = visionConfig;  
            })
        },
        submit(formName){
            this.$refs[formName].validate((valid) => {
                if (valid) {
                    this.$emit('sendVisual',this.ruleForm.visionConfig)
                    this.dialogVisible = false;
                } else {
                    return false;
                }
            });
        }
    }
}
</script>
<style lang="scss" scoped>
/deep/.el-input-number--small {
    line-height: 28px !important;
}
.goSafety{
    margin-left:10px;
    color: #6977F9;
    background: #ECEEFE;
    padding:6px 15px;
    border-radius:4px;
    cursor:pointer;
}
</style>