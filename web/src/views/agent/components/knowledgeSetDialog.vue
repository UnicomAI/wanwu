<template>
    <div>
        <el-dialog
        title="Recall Parameter Configuration"
        :visible.sync="dialogVisible"
        width="50%"
        :before-close="handleClose">
        <span v-if="dialogVisible">
           <searchConfig ref='searchConfig' @sendConfigInfo="sendConfigInfo" :setType="'agent'" :config="knowledgeConfig"/>
        </span>
        <span slot="footer" class="dialog-footer">
            <el-button @click="dialogVisible = false">Cancel</el-button>
            <el-button type="primary" @click="submit">Confirm</el-button>
        </span>
        </el-dialog>
    </div>
</template>
<script>
import searchConfig from '@/components/searchConfig.vue';
export default {
    components:{
        searchConfig
    },
    data(){
        return{
            dialogVisible:false,
            knowledgeConfig:{}
        }
    },
    methods:{
        sendConfigInfo(data){
            this.knowledgeConfig = { ...data.knowledgeMatchParams };
        },
        showDialog(row){
            this.dialogVisible = true;
            this.knowledgeConfig = row || {};
        },
        handleClose(){
            this.dialogVisible = false;
        },
        submit(){
            // VerifyModel Selection
            const { matchType, priorityMatch, rerankModelId } = this.knowledgeConfig;
            const needRerankModel = matchType === 'vector' || 
                                   matchType === 'text' || 
                                   (matchType === 'mix' && priorityMatch === 0);
            
            if (needRerankModel && !rerankModelId) {
                this.$message.error('Please selectModel');
                return;
            }
            
            if(matchType === 'mix' && priorityMatch === 1){
                this.knowledgeConfig.rerankModelId = '';
            }
            this.dialogVisible = false;
            this.$emit('setKnowledgeSet',this.knowledgeConfig)
        }
    }
}
</script>
<style lang="scss" scoped>
/deep/{
    .el-input-number--small{
        line-height: 28px!important;
    }
}
.question{
    cursor: pointer;
    color:#ccc;
}
</style>