<template>
  <div class="page-wrapper">
    <div class="page-title">
      <img class="page-title-img" src="@/assets/imgs/safety.svg" alt="" />
      <span class="page-title-name">Safety Guardrail</span>
      <p class="page-tips">Supported user Custom Sensitive Word Table，Configuration Sensitive Word，Real-time interception of high-risk input and output content to ensure content safety and compliance. Create App when associated Configuration。</p>
    </div>
    <div style="padding: 0 20px 20px 20px;">
      <safetyList :appData="knowledgeData" @editItem="showCreate" @reloadData="getTableData" ref="knowledgeList" v-loading="tableLoading" />
      <createSafety ref="createSafety" @reloadData="getTableData" />
    </div>
  </div>
</template>
<script>
import { getSensitiveList } from "@/api/safety";
import safetyList from './component/safetyList.vue';
import createSafety from './component/create.vue';
export default {
    components: { safetyList,createSafety },
    data(){
       return{
        knowledgeData:[],
        tableLoading:false
       } 
    },
    mounted(){
      this.getTableData();
    },
    methods:{
        getTableData(){
            this.tableLoading = true
            getSensitiveList().then(res => {
                this.knowledgeData = res.data.list || [];
                this.tableLoading = false
            }).catch((error) =>{
                this.tableLoading = false
                this.$message.error(error)
            })
        },
        showCreate(row){
            this.$refs.createSafety.showDialog(row)
        }
    }
}
</script>
<style lang="scss" scoped>
.search-box {
 display:flex;
 justify-content:space-between;
}

/deep/{
  .el-loading-mask{
    background: none !important;
  }
}
.page-tips{
  color:#888888;
  padding-top:15px;
  font-weight:normal;
}
</style>
