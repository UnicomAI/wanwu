<template>
    <div>
      <el-dialog
        top="10vh"
        :title="title"
        :close-on-click-modal="false"
        :visible.sync="dialogVisible"
        width="40%"
        :before-close="handleClose"
        >
        <el-form
            :model="ruleForm"
            ref="ruleForm"
            label-width="140px"
            class="demo-ruleForm"
            :rules="rules"
            @submit.native.prevent
        >
            <el-form-item
            label="Sensitive Word TableName"
            prop="tableName"
            >
            <el-input
                v-model="ruleForm.tableName"
                placeholder="Please enter table name; can contain Chinese characters, letters, or numbers"
                maxlength="15"
                show-word-limit
            ></el-input>
            </el-form-item>
            <el-form-item
            label="tableRemark"
            prop="remark"
            >
            <el-input 
            v-model="ruleForm.remark" 
            type="textarea"
            :rows="4"
            placeholder="Enter a remark, e.g. describe the app scenario, channel, or any important notes"></el-input>
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
import { createSensitive,editSensitive } from "@/api/safety";
export default {
    data(){
        var checkName = (rule, value, callback) => {
        const reg = /^[\u4e00-\u9fa5a-zA-Z0-9]+$/;
        if (!reg.test(value)) {
            callback(
            new Error(
                'Please enter a table name; it can contain Chinese characters, letters, or numbers'
            )
            );
        } else {
            return callback();
        }
        };
        return{
            title:"New word table",
            dialogVisible:false,
            ruleForm:{
                tableName:'',
                remark:'',
            },
            rules: {
                tableName: [
                { required: true, message:'Please enterSensitive Word TableName', trigger: "blur" },
                { validator: checkName, trigger: "blur" },
                ],
                remark:[{ required: true, message: 'Please entertableRemark', trigger: "blur" }]
            },
            tableId:''
        }
    },
    methods:{
        handleClose(){
            this.dialogVisible = false;
            this.clearform()
        },
        clearform(){
            this.tableId = ''
            this.$refs.ruleForm.resetFields()
            this.$refs.ruleForm.clearValidate()
        },
        submitForm(formName){
            this.$refs[formName].validate((valid) =>{
                if(valid){
                    if(this.tableId !== ''){
                      this.editSensitive()
                    }else{
                      this.createSensitive()
                    }
                }else{
                    return false;
                }
            })
        },
        createSensitive(){
            createSensitive(this.ruleForm).then(res =>{
                if(res.code === 0){
                    this.$message.success('CreateSuccess');
                    this.$emit('reloadData')
                    this.dialogVisible = false;
                    this.$router.push({path:`/safety/wordList/${res.data.tableId}`});
                }
            }).catch((error) =>{
                this.$message.error(error)
            })
        },
        editSensitive(){
            const data = {
                ...this.ruleForm,
                'tableId':this.tableId
            }
            editSensitive(data).then(res =>{
                if(res.code === 0){
                    this.$message.success('EditSuccess');
                    this.$emit('reloadData')
                    this.clearform();
                    this.dialogVisible = false;
                }
            }).catch((error) =>{
                this.$message.error(error)
            })
        },
        showDialog(row=null){
            this.dialogVisible = true;
            if (row) {
                this.title = 'Edit word table';
                this.tableId = row.tableId;
                this.ruleForm = {
                    tableName:row.tableName,
                    remark:row.remark
                }
            }else{
                this.title = 'New word table';
                this.ruleForm = {
                    tableName:'',
                    remark:''
                }
            }
        }
    }
}
</script>