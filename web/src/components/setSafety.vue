<template>
    <el-dialog
    title="AddSensitive Word"
    :visible.sync="dialogVisible"
    width="50%"
    :before-close="handleClose">
        <el-form :model="ruleForm" ref="ruleForm"  class="demo-ruleForm">
            <el-form-item 
            label="Sensitive Word Table" 
            prop="tables"
            :rules="[{ required: true, message: 'Please selectSensitive Word Table', trigger: 'blur'}]"
            >
                <el-select 
                    v-model="ruleForm.tables" 
                    placeholder="Please select" 
                    @visible-change="visibleChange" 
                    style="width:65%;" 
                    multiple 
                    value-key="tableId" 
                    filterable
                    clearable>
                    <el-option
                    v-for="item in safetyOptions"
                    :key="item.tableId"
                    :label="item.tableName"
                    :value="item">
                    </el-option>
                </el-select>
                <span @click="goCreate" class="goSafety"><span class="el-icon-d-arrow-right"></span>CreateSensitive Word</span>
            </el-form-item>
        </el-form>
        <span slot="footer" class="dialog-footer">
            <el-button type="primary" @click="submit('ruleForm')">确 定</el-button>
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
                tables:[]
            },
            safetyOptions:[]
        }
    },
    created(){
        this.getList();
    },
    methods:{
        handleClose(){
            this.dialogVisible = false;
        },
        visibleChange(val){
            if(val){
                this.getList();
            }
        },
        showDialog(row=null){
            this.dialogVisible = true;
            this.$nextTick(() =>{
                const form = this.$refs.ruleForm;
                 if (form) {
                    form.clearValidate();
                    if (!row) form.resetFields();
                }
                this.ruleForm.tables = row ? 
                this.safetyOptions.filter(item => 
                    row.some(i => i.tableId === item.tableId)
                ) : [];
            })
        },
        getList(){
            sensitiveSelect().then(res =>{
                if(res.code === 0){
                    this.safetyOptions = res.data.list || []
                }
            })
        },
        submit(formName){
            this.$refs[formName].validate((valid) => {
                if (valid) {
                    this.$emit('sendSafety',this.ruleForm.tables)
                    this.dialogVisible = false;
                } else {
                    return false;
                }
            });
        },
        goCreate(){
            this.$router.push({path:'/safety'})
        }
    }
}
</script>
<style lang="scss" scoped>
.goSafety{
    margin-left:10px;
    color: #6977F9;
    background: #ECEEFE;
    padding:6px 15px;
    border-radius:4px;
    cursor:pointer;
}
</style>