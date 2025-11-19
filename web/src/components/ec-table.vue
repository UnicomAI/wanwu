<template>
  <div class="container my-pagination" style="height:100%;" id="table-container" v-loading="loading">
    <el-table 
    v-bind="$attrs" 
    :data="tableData" 
    :row-key="rowKey" 
    :key="toggleIndex" 
    ref="ecTable" 
    :max-height="_maxHeight"
    :border="border" 
    :stripe="stripe" 
    :size="size"
    :summary-method="remoteSummary?remoteSummaryMethod:summaryMethod"
    :span-method="spanMethod" 
    style="width: 100%"
    :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
    class="common-table"
    >
      <slot></slot>
      <el-table-column min-width="1"></el-table-column>
    </el-table>
    <div class="ecTable-pagination" v-if="!hidePagination">
      <el-pagination
        class="pagination table-pagination"
        :current-page="pageNo"
        :page-size="page_size"
        :page-sizes="[5, 10, 20, 50]"
        :style="{ float: 'right', padding: '20px' }"
        :total="total"
        @current-change="handleCurrentChange"
        @size-change="handleSizeChange"
        layout="total, sizes, prev, pager, next, jumper"
      >
      </el-pagination>
    </div>
  </div>
</template>

<script>
export default {
  components: {},
  name: "ec-table",
  props: {
    listApi: "",//tableInterface
    resKey:{ type:String,default:''},//判断YesNoYes正常ListBack
    data: { type: Array, default: () => [] },//DefaultData
    height: { type: [String,Number], default: "100%" },//Table高度
    rowKey: { type: String, default: "" },//行Data of key，唯一标识符
    border: { type: Boolean, default: false },//YesNo带Has纵向边框
    stripe: { type: Boolean, default: false },//YesNoShow斑马纹
    size: { type: String, default: "default" },//table尺寸Setting
    remoteSummary: { type: Boolean, default: false },//YesNoShowCustom合计计算Method
    summaryMethod: { type: Function, default: null },//Custom合计计算Method
    spanMethod:{type:Function,default:null},//Custom合并单元格
    hidePagination: { type: Boolean, default: false },//隐藏分页
    pageSize:{ type: Number, default: 10 },
    searchInfo:{type:Object,default(){return {}}},//QueryInformation
    isHandelpage:{type:Boolean,default:false},//YesNo前端手动分页
  },
  data() {
    return {
        total: 0,
        pageNo:1,
        page_size:5,
        toggleIndex: 0,
        tableData: [],
        search_info: {},
        loading:false
    };
  },
  computed:{
    _height() {
				return Number(this.height)?Number(this.height)+'px':this.height
			},
    _maxHeight(){
      //计算table自动撑开 of Max高度
      this.$nextTick(() =>{
        const c_height =  document.getElementById('table-container').clientHeight;
        if(this.hidePagination){
          return c_height;
        }else{
          return c_height - 50;
        }
      })
    },
  },
  watch:{
    data(){
      this.total = this.data.length;
      if(this.isHandelpage === true){//前端手动分页
        const start = (this.pageNo - 1) * this.page_size;
        const end = start + this.page_size;
        this.tableData = this.data.slice(start,end);
      }else{
        this.tableData = this.data;
      }
    },
    pageSize:{
      handler(val){
        this.page_size = val
      },
      immediate:true
    }
  },
  mounted() { 
    if(this.listApi){
      if(this.resKey === ''){
        this.getTableData(this.searchInfo)
      }
    }else{
      // if(this.isHandelpage === true){
      //   if(this.data.length > 0){
      //     const start = (this.pageNo - 1) * this.page_size;
      //     this.tableData = this.data.slice(start,this.page_size);
      //   }
      // }else{
      //   this.tableData = this.data;
      //   this.total = this.tableData.length;
      // }
    }
  },
  methods: {
    async getTableData(searchInfo) {
        this.loading = true;
        this.search_info = searchInfo
        var reqData = {
            pageNo:this.pageNo,
            pageSize:this.page_size,
            ...this.search_info
        }
        //YesNoHas分页
        if(this.hidePagination){
          delete reqData['pageNo']
          delete reqData['pageSize']
        }
        //RequestInterface
        try {
          var table = await this.listApi(reqData);
        }catch(error){
          this.loading = false
          return false
        }
      let dataObj = undefined
      if(table.code == 0){
        if(this.resKey !== ''){
          dataObj = table.data
          this.total = table.data.datasetCount * 1;
          this.$emit('refreshData',dataObj)
          this.loading = false
        }else{
          let list = table.data.list;
          if (table.data == null || table.data.list == null) list = [];
          this.loading = false;
          this.tableData = list;
          this.total = table.data.total * 1 || 0;
          this.pageNo = table.data.pageNo * 1 || 0;
          this.page_size = table.data.pageSize * 1 || 0;
          this.$emit('refreshData',this.tableData)
        }
      }else{
        this.loading = false;
        this.$message.error(table.message||table.msg)
      }
    },
    updateTableData(data,total=''){
      this.tableData = data
      if(total !== ''){
        this.total = total
      }
    },
    mergeComon(id, rowIndex) { // 合并单元格
      const idName = this.tableData[rowIndex][id]
      if (rowIndex > 0) {
        if (this.tableData[rowIndex][id] != this.tableData[rowIndex - 1][id]) {
          let i = rowIndex; let num = 0
          while (i < this.tableData.length) {
            if (this.tableData[i][id] === idName) {
              i++
              num++
            } else {
              i = this.tableData.length
            }
          }
          return {
            rowspan: num,
            colspan: 1
          }
        } else {
          return {
            rowspan: 0,
            colspan: 1
          }
        }
      } else {
        let i = rowIndex; let num = 0
        while (i < this.tableData.length) {
          if (this.tableData[i][id] === idName) {
            i++
            num++
          } else {
            i = this.tableData.length
          }
        }
        return {
          rowspan: num,
          colspan: 1
        }
      }
    },
    remoteSummaryMethod(param){
      const {columns} = param
      const sums = []
      columns.forEach((column, index) => {
        if(index === 0) {
          sums[index] = this.$t('common.table.total')
          return
        }
        const values =  this.summary[column.property]
        if(values){
          sums[index] = values
        }else{
          sums[index] = ''
        }
      })
      return sums
    },
    reload(params, page=1){//RefreshList
      this.pageNo = page;
      this.search_info = params || {}
      this.$refs.ecTable.clearSelection();
      this.getTableData(this.searchInfo)
    },
    //Insert行 unshiftRow
		unshiftRow(row){
				this.tableData.unshift(row)
		},
    //Insert行 pushRow
    pushRow(row){
      this.tableData.push(row)
    },
    //根据key覆盖Data
    updateKey(row, rowKey=this.rowKey){
      this.tableData.filter(item => item[rowKey]===row[rowKey] ).forEach(item => {
        Object.assign(item, row)
      })
    },
    //根据index覆盖Data
    updateIndex(row, index){
      Object.assign(this.tableData[index], row)
    },
    //根据indexDelete
    removeIndex(index){
      this.tableData.splice(index, 1)
    },
    //根据indexBatch Delete
    removeIndexes(indexes=[]){
      indexes.forEach(index => {
        this.tableData.splice(index, 1)
      })
    },
    //根据keyDelete
    removeKey(key, rowKey=this.rowKey){
      this.tableData.splice(this.tableData.findIndex(item => item[rowKey]===key), 1)
    },
    //根据keysBatch Delete
    removeKeys(keys=[], rowKey=this.rowKey){
      keys.forEach(key => {
        this.tableData.splice(this.tableData.findIndex(item => item[rowKey]===key), 1)
      })
    },
    //Clear勾选
    clearSelection(){
      this.$refs.ecTable.clearSelection()
    },
    toggleRowSelection(row, selected){
      this.$refs.ecTable.toggleRowSelection(row, selected)
    },
    toggleAllSelection(){
      this.$refs.ecTable.toggleAllSelection()
    },
    setCurrentRow(row){//高亮选中某一行
      this.$refs.scTable.setCurrentRow(row)
    },
    doLayout(){//对table进行重新Layout
      this.$refs.scTable.doLayout()
    },
    handleSizeChange(val) {//分页DataModify点击
      this.page_size = val;
      if(this.isHandelpage=== true){//前端手动分页，Modify分页页数
        const start = (this.pageNo - 1) * this.page_size;
        const end = start + this.page_size;
        this.tableData = this.data.slice(start,end);
      }else{
        this.getTableData(this.searchInfo);
      }
    },
    handleCurrentChange(val) {//分页点击
      this.pageNo = val;
      if(this.isHandelpage=== true){//前端手动分页
        const start = (this.pageNo - 1) * this.page_size;
        const end = start + this.page_size;
        this.tableData = this.data.slice(start,end);
      }else{
        this.getTableData(this.searchInfo);
      }
    },
  }
};
</script>

<style lang="scss" scoped>
/deep/{
  .el-loading-spinner .path { stroke: #e60001 !important;}
  /* 隐藏Scrollbar但保留滚动功能 */
  .el-table--scrollable-y .el-table__body-wrapper::-webkit-scrollbar {
    display: none !important;
  }
  .el-table--scrollable-y .el-table__body-wrapper {
    -ms-overflow-style: none !important; 
    scrollbar-width: none !important;
  }
}
.container {
  .con_table_pagination {
    margin-top: 5px;
  }
}
.common-table /deep/ tbody {
    td,tr{
      background-color: #fff!important;
    }
}
.my-pagination /deep/ {
  .ecTable-pagination{
    height: 50px;
    // display:flex;
    // align-items: center;
    // justify-content:flex-end;
    // padding:0 15px;
  }
  .el-pagination.is-background .el-pager li:not(.disabled).active {
    background-color: #e60001;
    color: #fff;
    border-radius: 120px;
  }
  .el-pagination .el-select .el-input .el-input__inner {
    padding-right: 25px;
    // border-radius: 91px;
    width: 109px;
    border-color: #cccccc;
  }
  .el-pagination .el-select .el-input .el-input__inner {
    padding-right: 25px;
    // border-radius: 91px;
    width: 109px;
    border-color: #cccccc;
  }
  .el-pager li:hover{
    color: #e60001;
  }
  .el-pager li.active {
    color: #e60001;
    cursor: default;
    // border-radius: 30px;
    border: 1px solid #e60001;
  }
  .el-pagination__editor.el-input .el-input__inner {
    height: 28px;
    background: #ffffff;
    // border-radius: 15px;
    border: 1px solid #cccccc;
  }
  .el-pagination.is-background .el-pager li:not(.disabled):hover {
    color: #e60001;
  }
}
</style>
