<template>
    <div>
        <el-dialog
            title="API Key"
            :visible.sync="dialogVisible"
            width="50%"
            :before-close="handleClose">
            <el-table
                :data="tableData"
                style="width: 100%">
                <el-table-column label="API Key"  prop="apiKey"  width="300" />
                <el-table-column label="Create Time"  prop="createdAt"  />
                <el-table-column label="Operation" width="200">
                <template slot-scope="scope">
                    <el-button
                    size="mini"
                    @click="handleCopy(scope.row) && copycb()">Copy</el-button>
                    <el-button
                    size="mini"
                    @click="handleDelete(scope.row)">Delete</el-button>
                </template>
                </el-table-column>
            </el-table>
            <span slot="footer" class="dialog-footer">
                <el-button @click="dialogVisible = false">Cancel</el-button>
                <el-button type="primary" @click="handleCreate">创 建</el-button>
            </span>
        </el-dialog>
    </div>
</template>
<script>
import {createApiKey,delApiKey,getApiKeyList} from "@/api/appspace";
export default {
     props: {
        appType: {
        type: String,
        required: true
        },
        appId: {
        type: String,
        required: true
        },
    },
    data(){
        return{
            tableData:[],
            dialogVisible:false
        }
    },
    created(){
        this.getTableData()
    },
    methods:{
        showDialog(){
            this.dialogVisible = true;
        },
        handleClose(){
            this.dialogVisible = false;
        },
        handleCopy(row){
            let text = row.apiKey;
            var textareaEl = document.createElement('textarea');
            textareaEl.setAttribute('readonly', 'readonly');
            textareaEl.value = text;
            document.body.appendChild(textareaEl);
            textareaEl.select();
            var res = document.execCommand('copy');
            document.body.removeChild(textareaEl);
            return res;
        },
        copycb(){
            this.$message.success('Content copied to clipboard')
        },
        handleCreate(){
            const data = {appId:this.appId,appType:this.appType}
            createApiKey(data).then(res =>{
                if(res.code === 0){
                    this.tableData.push(res.data);
                }
            })
        },
        getTableData(){
            const data = {appId:this.appId,appType:this.appType}
            getApiKeyList(data).then(res =>{
                if(res.code === 0){
                    this.tableData = res.data || []
                }
            })
        },
        handleDelete(row){
            this.$confirm('ConfirmtoDeletecurrentAPIkey吗？',this.$t('knowledgeManage.tip'),
                {
                confirmButtonText:'Confirm',
                cancelButtonText: 'Delete',
                type: "warning"
                }
            )
            .then(() => {
                delApiKey({apiId:row.apiId}).then(res =>{
                    if(res.code === 0){
                        this.$message.success('DeleteSuccess')
                        this.getTableData()
                    }
                })
            })
            .catch((error) => {
                this.getTableData()
            });
        }
    }
}
</script>