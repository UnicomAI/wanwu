<template>
  <el-dialog
    :visible.sync="dialogVisible"
    width="60%"
    :before-close="handleClose"
  >
  <template #title>
    <span class="dialog-title">Metadata Management</span>
    <span>[ MetadatakeyIs Empty when ，去<span class="link" @click="goCreate">CreateMetadata</span> ]</span>
  </template>
    <div>
        <el-table
        :data="filteredTableData"
        style="width: 100%;margin-top: 20px">
          <el-table-column
              prop="metaKey"
              align="center">
            <template #header>
              <div style="display: inline-flex; align-items: center;">
                <span>Key</span>
              </div>
            </template>
            <template #default="{ row }">
              <el-select
                v-model="row.metaKey"
                placeholder="Please selectkey"
                @change="keyChange($event,row)"
                :disabled="Boolean(row.metaId)"
            >
                <el-option
                v-for="item in keyOptions"
                :key="item.metaKey"
                :label="item.metaKey"
                :value="item.metaKey"
                >
                </el-option>
              </el-select>
            </template>
          </el-table-column>
          <el-table-column
              prop="metaValueType"
              label="Type"
              align="center">
            <template #default="{ row }">
              <span class="metaValueType">[{{row.metaValueType}}]</span>
            </template>
          </el-table-column>
          <el-table-column
              prop="metaValue"
              label="Value"
              align="center"
              min-width="90">
            <template #default="{ row }">
              <el-input
                  v-if="row.metaValueType === 'string' || row.metaValueType === ''"
                  v-model="row.metaValue"
                  @blur="handleBlur(row)"
                  @change="handleValueChange(row)"
                  clearable
                  :disabled="!row.editable"
                  placeholder="Please enterContent"
              />
              <el-input
                  v-if="row.metaValueType === 'number'"
                  v-model="row.metaValue"
                  @blur="handleBlur(row)"
                  clearable
                  :disabled="!row.editable"
                  type="number"
                  placeholder="Please enternumber"
                  @change="handleValueChange(row)"
              />
              <el-date-picker
                  v-if="row.metaValueType === 'time'"
                  v-model="row.metaValue"
                  @blur="handleBlur(row)"
                  @change="handleValueChange(row)"
                  clearable
                  :disabled="!row.editable"
                  align="right"
                  format="yyyy-MM-dd HH:mm:ss"
                  value-format="timestamp"
                  type="datetime"
                  placeholder="Please select date and time"
                  style="width:100%;"
              />
            </template>
          </el-table-column>
          <el-table-column
              label="Operation"
              align="center"
          >
            <template #default="{ row }">
              <i class="el-icon-edit-outline table-opera-icon"
                 style="margin-right: 10px;"
                 @click="editItem(row)"/>
              <i class="el-icon-delete table-opera-icon"
                 @click="delItem(row)"/>
            </template>
          </el-table-column>
        </el-table>
    </div>
    <span
      slot="footer"
      class="dialog-footer"
    >
      <el-button @click="dialogVisible = false">Cancel</el-button>
      <el-button @click="addItem" type="primary">+ CreateMetadata</el-button>
      <el-button
        type="primary"
        @click="submitDialog('submit')"
        :disabled="isDisabled"
      >Confirm</el-button>
    </span>
  </el-dialog>
</template>
<script>
import { updateDocMeta,metaSelect} from "@/api/knowledge";
export default {
  props:['knowledgeId','name'],
  computed: {
    filteredTableData() {
      return this.tableData.filter(item => item.option !== "delete");
    }
  },
  data() {
    return {
      dialogVisible: false,
      tableData: [],
      docId: "",
      keyOptions:[],
      isDisabled:false
    };
  },
  watch: {
    tableData: {
      deep: true,
      immediate: true,
      handler(val) {
        const hasEmptyField = Array.isArray(val) && val.some(item => {
        if (item.option === 'delete') return false;
        const isMetaKeyEmpty = !item.metaKey || 
          (typeof item.metaKey === 'string' && item.metaKey.trim() === '');
        
        const isMetaValueEmpty = item.metaValue === null || 
          item.metaValue === undefined || 
          (typeof item.metaValue === 'string' && item.metaValue.trim() === '') ||
          item.metaValue === '';
        
        return isMetaKeyEmpty || isMetaValueEmpty;
      });
      
        const hasValidData = Array.isArray(val) && val.length > 0 && 
          val.some(item => item.option !== 'delete');
        
        this.isDisabled = !hasValidData || hasEmptyField;
      }
    }
  },
  created(){
    this.getList()
  },
  methods: {
    getList(){
        metaSelect({knowledgeId:this.knowledgeId}).then(res =>{
            if(res.code === 0){
                this.keyOptions = res.data.knowledgeMetaList || []
            }
        }).catch(() =>{})
    },
    keyChange(val,row){
      row.metaValueType = this.keyOptions.filter(i => i.metaKey === val).map(e => e.metaValueType)[0];
    },
    goCreate(){
      this.dialogVisible = false;
      this.$router.push({path:`/knowledge/doclist/${this.knowledgeId}`,query:{name:this.name}})
    },
    submitDialog(type) {
      this.isDisabled = true;
      this.tableData.forEach(i => {
        delete i.editable
        delete i.created
        i.metaValue = i.metaValue.toString()
      });
      const metaData = this.tableData.filter(item => item.option !== '');
      if(type === 'submit' && !metaData.length){
         this.dialogVisible = false;
         return false;
      }
      const data = {
        metaDataList:metaData,
        docId:this.docId,
        knowledgeId:this.knowledgeId
      }
      updateDocMeta(data).then(res =>{
        if(res.code === 0){
            this.$message.success('OperationSuccess');
            this.$emit('updateData')
            this.isDisabled = true;
            if(type === 'submit'){
              this.dialogVisible = false;
            }
        }
      }).catch(() =>{
        this.isDisabled = true;
      })
      
    },
    addItem(){
      this.tableData.push({
        metaKey: '',
        metaValueType: '',
        metaValue: '',
        metaRule: '',
        metaId: '',
        option: 'add',
        editable: true,
        created: true,
      });
    },
    editItem(n){
        n.editable = !n.editable;
    },
    handleValueChange(n){
      if(Boolean(n.metaId)){
        n.option = 'update'
      }
    },
    handleBlur(n){
      if (n.metaKey && n.metaValueType && n.metaValue) n.editable = false;
    },
    delItem(index){
      index.option = "delete";
      this.submitDialog('delete')
    },
    showDiglog(data,id) {
      this.dialogVisible = true;
      this.docId = id;
      this.tableData = data.map(item => ({
          ...item,
          option: '',
          editable: false,
          created: false,
        }));
    },
    handleClose() {
      this.dialogVisible = false;
    }
  },
};
</script>
<style lang="scss" scoped>
/deep/ {
  .el-dialog__body {
    padding: 5px 20px !important;
  }
}
.table-opera-icon{
  font-size: 18px;
  cursor: pointer;
  color:$color;
}
.metaValueType,.link{
  color:$color;
}
.dialog-title{
  font-weight:bold;
  line-height:24px;
  font-size: 18px;
  color: $color_title;
  margin-right:10px;
}
.question{
  color: #aaadcc;
  margin-left: 5px;
}
.link{
  cursor: pointer;
}
</style>