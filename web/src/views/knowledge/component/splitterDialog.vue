<template>
  <el-dialog
    :title="title"
    :visible.sync="dialogVisible"
    width="20%"
    :before-close="handleClose"
  >
    <div>
      <el-input
        :placeholder="placeholderText"
        suffix-icon="el-icon-search"
        @keyup.enter.native="addByEnterKey"
        v-model="tagName"
      ></el-input>
      <div
        class="add"
        @click="addTag"
      ><span class="el-icon-plus add-icon"></span>{{title}}</div>
      <div class="tag-box">
        <div
          v-for="(item,index) in tagList"
          :key="index"
          class="tag_item"
          @mouseenter="mouseEnter(item)"
          @mouseleave="mouseLeave(item)"
          @dblclick="handleDoubleClick(item)"
        >
          <el-checkbox
            v-model="item.checked"
            v-if="!item.showIpt"
          >{{item.splitterName.replace(/\n/g, '\\n')}}</el-checkbox>
          <el-input
            v-model="item.splitterName"
            v-if="item.showIpt"
            maxlength="50"
            @keydown.backspace.native="handleDelete(item,index)" 
            @blur="inputBlur(item)"
          ></el-input>
          <span
            class="el-icon-close del-icon"
            v-if="item.showDel && !item.showIpt && item.type !== 'preset'"
            @click="delTag(item,index)"
          ></span>
        </div>
      </div>
    </div>
    <span
      slot="footer"
      class="dialog-footer"
    >
      <el-button
        type="primary"
        @click="submitDialog"
      >Confirm</el-button>
    </span>
  </el-dialog>
</template>
<script>
export default {
  props:{
    title:{
        type:String,
        required:true,
        default:''
    },
    dataList:{
        type:Array,
        required:true,
        default:[]
    },
    placeholderText:{
        type:String,
        required:true,
        default:''
    }
  },
  watch:{
    dataList:{
        handler(val){
            if(val){
              this.tagList = val;
            }
        },
        deep: true,
        immediate:true
    },
    selectData:{
      handler(val){
        if(val){
          this.tagList = this.tagList.map(tag => ({
            ...tag,
            checked: val.some(item => item.splitterValue == tag.splitterValue)
          }));
        }
      }
    }
  },
  data() {
    return {
      dialogVisible: false,
      tagList:[],
      tagName: "",
      selectData:[]
    };
  },
  methods: {
    submitDialog() {
      const data = this.tagList.filter(item => item.checked);
      this.$emit('checkData',data);
      this.dialogVisible = false;
    },
    delTag(item){
      this.$emit('delItem',item)
    },
    showDiaglog(data=[]) {
      this.selectData = [];
      if(data.length){
        this.selectData = data;
      }
      this.dialogVisible = true;
    },
    handleClose() {
      this.dialogVisible = false;
    },
    mouseEnter(n) {
      n.showDel = true;
    },
    mouseLeave(n) {
      n.showDel = false;
    },
    handleDoubleClick(n) {
      n.showIpt = true;
    },
    inputBlur(n) {
      if(!n.splitterName) return;
      if (n.splitterId) {
        this.$emit('editItem',n)
      } else {
        this.$emit('createItem',n)
      }
    },
    handleDelete(n,i){
      if(n.splitterName === '' && !n.splitterId){
          this.tagList.splice(i,1)
      }
    },
    addTag() {
      const emptyTag = this.tagList.find(tag => !tag.splitterId && tag.splitterName === "");
      if(emptyTag) return;
      this.tagList.unshift({
        splitterName: "",
        checked: false,
        showDel: false,
        showIpt: true,
      });
    },
    addByEnterKey(e) {
      if (e.keyCode === 13) {
        this.$emit('relodData',this.tagName);
      }
    },
  },
};
</script>
<style lang="scss" scoped>
/deep/ {
  .el-dialog__body {
    padding: 5px 20px !important;
  }
  .add {
    margin-top: 10px;
    padding: 10px 0;
    cursor: pointer;
    .add-icon {
      margin-right: 5px;
    }
  }
  .tag-box {
    max-height: 300px;
    overflow-y: scroll;
  }
  .tag_item {
    cursor: pointer;
    background: $color_opacity;
    padding: 5px;
    margin: 10px 0;
    border-radius: 4px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    .del-icon {
      color: $color;
      cursor: pointer;
      font-size: 16px;
    }
  }
}
</style>