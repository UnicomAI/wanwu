<template>
  <div class="page-wrapper">
    <div class="page-title">
      <img class="page-title-img" src="@/assets/imgs/safety.svg" alt="" />
      <span class="page-title-name">安全护栏</span>
      <p class="page-tips">支持用户CustomSensitive Word Table，Configuration行业Sensitive Word，实 when 拦截高风险Content of EnterAnd输出，保障Content安全合规。可在CreateApp when 关联Configuration。</p>
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
