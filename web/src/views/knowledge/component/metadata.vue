<template>
  <div>
    <el-button
      icon="el-icon-plus"
      type="primary"
      size="mini"
      @click="createMetaData"
      v-if="type !== 'create'"
    >Create</el-button>
    <div class="docMetaData" v-loading="loading">
      <div
        v-for="(item,index) in docMetaData"
        class="docItem"
        :key="'meta'+index"
      >
        <div class="docItem_data">
          <span class="docItem_data_label">
            <span class="label">Key:</span>
            <el-tooltip
              class="item"
              effect="dark"
              content="Can only contain lowercase letters, numbers and underscores, and must start with a lowercase letter"
              placement="top-start"
            >
              <span class="el-icon-question question" v-if="type === 'create'"></span>
            </el-tooltip>
          </span>
          <template v-if="type === 'create'">
            <el-input
              v-if="item.showEdit"
              v-model="item.metaKey"
              @blur="metakeyBlur(item,index)"
            ></el-input>
            <span v-else class="metaItemKey">{{item.metaKey}}</span>
          </template>
          <el-select
              v-else
              v-model="item.metaKey"
              placeholder="Please select"
              @change="keyChange($event,item)"
          >
              <el-option
              v-for="meta in keyOptions"
              :key="meta.metaKey"
              :label="meta.metaKey"
              :value="meta.metaKey"
              >
              </el-option>
          </el-select>
        </div>
        <el-divider direction="vertical"></el-divider>
        <div class="docItem_data">
          <span class="docItem_data_label label">type:</span>
          <el-select
            v-if="type === 'create'"
            v-model="item.metaValueType"
            placeholder="Please select"
            :disabled="Boolean(item.metaId)"
          >
            <el-option
              v-for="item in typeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            >
            </el-option>
          </el-select>
          <span v-else class="metaValueType">[{{item.metaValueType}}]</span>
        </div>
        <el-divider direction="vertical" v-if="type !== 'create'"></el-divider>
        <div class="docItem_data" v-if="type !== 'create'">
          <span class="docItem_data_label label">value:</span>
          <el-select
            v-model="item.metadataType"
            placeholder="Please select"
            style="margin-right:5px;"
            @change="valueChange(item)"
          >
            <el-option
              v-for="item in valueOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            >
            </el-option>
          </el-select>
          <el-input
            v-model="item.metaValue"
            v-if="item.metadataType ==='value' && item.metaValueType === 'string'"
            @blur="metaValueBlur(item)"
            placeholder="string"
          ></el-input>
          <el-input
            v-model="item.metaValue"
            v-if="item.metadataType ==='value'  && item.metaValueType === 'number'"
            @blur="metaValueBlur(item)"
            type="number"
            placeholder="number"
          ></el-input>
          <el-input
            v-model="item.metaRule"
            v-if="item.metadataType ==='regExp'"
            @blur="metaRuleBlur(item)"
            placeholder="regExp"
          ></el-input>
          <el-date-picker
            v-if="item.metaValueType === 'time' && item.metadataType==='value'"
            v-model="item.metaValue"
            align="right"
            format="yyyy-MM-dd HH:mm:ss"
            value-format="timestamp"
            type="datetime"
            placeholder="Select Date Time"
          >
          </el-date-picker>
        </div>
        <el-divider direction="vertical" v-if="type !== 'create'"></el-divider>
        <div class="docItem_data docItem_data_btn">
          <span
          v-if="type === 'create'"
          class="el-icon-edit-outline setBtn"
          @click="editMataItem(item)"
          ></span>
          <span
            class="el-icon-delete setBtn"
            @click="delMataItem(index,item)"
          ></span>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import {metaSelect,updateDocMeta} from "@/api/knowledge"
export default {
  props:['metaData','type','knowledgeId'],
  watch: {
    metaData:{
        handler(val) {
            if(val){
                this.docMetaData = val;
            }
        },
        deep: true,
        immediate: true,
    },
    docMetaData: {
      handler(val) {
        if (this.debounceTimer) clearTimeout(this.debounceTimer);
        const metaList = Array.isArray(val) ? val : []

        const payload = metaList.map(item => ({
          ...item,
          metaValue:
            item && item.metaValueType === 'time'
              ? String(item.metaValue)
              : (item && item.metaValue != null ? String(item.metaValue) : ''),
        }))

        this.debounceTimer = setTimeout(() => {
          this.$emit("updateMeata",payload);
        }, 500);
      },
      deep: true,
      immediate: true,
    },
  },
  data() {
    return {
      loading:false,
      debounceTimer: null,
      docMetaData: [],
      typeOptions: [
        {
          label: "String",
          value: "string",
        },
        {
          label: "Number",
          value: "number",
        },
        {
          label: "Time",
          value: "time",
        },
      ],
      valueOptions: [
        {
          value: "value",
          label: "ConfirmValue",
        },
        {
          value: "regExp",
          label: "Regular Expression",
        },
      ],
      keyOptions:[]
    };
  },
  created(){
    this.getList()
  },
  methods: {
    getList(){
      this.loading = true;
      metaSelect({knowledgeId:this.knowledgeId}).then(res =>{
          if(res.code === 0){
            this.loading = false;
            this.keyOptions = res.data.knowledgeMetaList || []
              if(this.type === 'create'){
                this.docMetaData = (res.data.knowledgeMetaList || []).map(item => ({
                  ...item,
                  metaValueType:item.metaValueType || 'string',
                  showEdit:false,
                  option: ''
                }));
              }
              
          }
      }).catch(() =>{
        this.loading = false;
      })
    },
    keyChange(val,item){
      item.metaValue = ''
      item.metadataType = 'value'
      const opt = Array.isArray(this.keyOptions)
        ? this.keyOptions.find(i => i.metaKey === val)
        : null
      item.metaValueType = opt ? opt.metaValueType : ''
    },
    createMetaData() {
      if(this.type === 'create' && this.docMetaData.length > 0  ){
        if (this.docMetaData.some(item => item.metaKey === '' || item.metaValueType === '')) {
            this.$message.error("Metadata management has unfilled required fields");
            return;
        }
      }else{
          if (this.docMetaData.length > 0 && !this.validateMetaData(this.docMetaData)) {
            return;
        }
      }

      this.docMetaData.push({
        metaId:"",
        metaKey: "",
        metaRule: "",
        metaValue: "",
        metaValueType: "string",
        showEdit:true,
        metadataType: "value",
        option:"add"
      });
    },
    validateMetaData() {
      const hasEmptyField = this.docMetaData.some((item) => {
        const isMetaKeyEmpty =
          !item.metaKey ||
          (typeof item.metaKey === "string" && item.metaKey.trim() === "");
        const isMetaRuleRequired = item.metadataType !== "value";
        const isMetaRuleEmpty =
          isMetaRuleRequired &&
          (!item.metaRule ||
            (typeof item.metaRule === "string" && item.metaRule.trim() === ""));
        return isMetaKeyEmpty || isMetaRuleEmpty;
      });
      if (hasEmptyField) {
        this.$message.error("Metadata management has unfilled required fields");
        return false;
      }
      return true;
    },
    editMataItem(item){
      item.showEdit = true;
      if(item.metaId){
        item.option = 'update';
      }
    },
    delMataItem(i,item) {
      if(item.metaId){
        item.option = 'delete'
        this.delMetaData(item)
      }else{
        this.docMetaData.splice(i, 1);
      }
    },
    delMetaData(item){
      const dataItem = [item]
      const data = {
        docId:'',
        knowledgeId:this.knowledgeId,
        metaDataList:dataItem.map(({metaId,option}) =>({
          metaId,
          option
        }))
      }
      updateDocMeta(data).then(res =>{
        if(res.code === 0){
            this.$message.success('Operation successful')
            this.getList();
        }
      })
    },
    valueChange(item) {
      item.metaValue = "";
      item.metaRule = "";
    },
    metakeyBlur(item,index) {
      const regex = /^[a-z][a-z0-9_]*$/;
      if (!item.metaKey || typeof item.metaKey !== 'string' || item.metaKey.trim() === '') {
        this.$message.warning('Please enter key value')
        return
      }
      if (!regex.test(item.metaKey)) {
        this.$message.warning("Please enter a valid key value");
        item.metaKey = '';
        return;
      }

      if(this.isFound()){
        this.$message.warning("Key value already exists");
        item.metaKey = ''
        return;
      }
      
      item.showEdit = false;
    },
    isFound(){
      const metaKeys = this.docMetaData.map(item => item.metaKey);
      const uniqueKeys = new Set(metaKeys);
      return uniqueKeys.size !== metaKeys.length;
    },
    metaValueBlur(item) {
      if (!item.metaValue) {
        this.$message.warning("Please enter value");
        return;
      }
    },
    metaRuleBlur(item) {
      if (!item.metaRule) {
        this.showWarning("Please enter regular expression value",item);
        return;
      }
      if (!this.isValidRegex(item.metaRule)) {
        this.showWarning("Please enter a valid regular expression",item);
        item.metaRule = "";
        return;
      }
    },
    showWarning(message,item){
        this.$message.warning(message);
        item.metaRule = "";
    },
    isValidRegex(str) {
        try {
            if (str.startsWith('/')) {
            if (!str.endsWith('/') && !str.match(/\/[a-z]*$/)) return false;
            const parts = str.slice(1).split('/');
            if (parts.length < 1) return false;
            new RegExp(parts[0], parts[1]);
            } else {
            new RegExp(str);
            }
            return true;
        } catch {
            return false;
        }
    },
  },
};
</script>
<style lang="scss" scoped>
.docMetaData {
  display:flex;
  gap:10px;
  flex-direction: column;
  .docItem {
    display: flex;
    align-items: center;
    border-radius: 8px;
    background: #f7f8fa;

    width: fit-content;
    .docItem_data {
      display: flex;
      align-items: center;
      padding:5px 10px;
      .metaItemKey{
        padding:0 15px;
      }
      .el-input,
      .el-select,
      .el-date-picker,
      .metaItemKey {
        min-width: 160px;
      }
      .label{
        min-width: fit-content;
      }
      .metaValueType{
        color:$color ;
      }
      .docItem_data_label {
        margin-right: 5px;
        display: flex;
        align-items: center;
        .question {
          color: #aaadcc;
          margin:2px 5px 0 2px;
          cursor: pointer;
        }
      }
      .setBtn {
        font-size: 16px;
        cursor: pointer;
        color: $btn_bg;
      }
    }
    .docItem_data_btn {
      display: flex;
      justify-content: center;
      .el-icon-delete {
        margin-left: 5px;
      }
    }
  }
}
</style>