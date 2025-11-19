<template>
  <el-dialog
    :title="title"
    :visible.sync="dialogVisible"
    width="20%"
    :before-close="handleClose"
  >
    <div>
      <el-input
        v-if="type !== 'section'"
        placeholder="Search标签"
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
          class="tag_item"
          @mouseenter="mouseEnter(item)"
          @mouseleave="mouseLeave(item)"
          @dblclick="handleDoubleClick(item)"
        >
          <el-checkbox
            v-model="item.selected"
            v-if="!item.showIpt && type !== 'section'"
          >{{item.tagName}}</el-checkbox>
          <span v-if="!item.showIpt && type === 'section'">{{item.tagName}}</span>
          <el-input
            v-model="item.tagName"
            v-if="item.showIpt"
            maxlength="50"
            placeholder="按回车(enter)键Confirm"
            @keydown.backspace.native="handleDelete(item,index)" 
            @keyup.enter.native="inputBlur(item)"
           
          ></el-input>
          <span
            class="el-icon-close del-icon"
            v-if="item.showDel && !item.showIpt"
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
      >确 定</el-button>
    </span>
  </el-dialog>
</template>
<script>
import { delTag, tagList, createTag, editTag, bindTag,bindTagCount,updateDocTag} from "@/api/knowledge";
export default {
  props:['type','title','currentList'],
  data() {
    return {
      dialogVisible: false,
      tagList: [],
      tagName: "",
      knowledgeId: ""
    };
  },
  watch:{
    currentList:{
      handler(val){
        if(val.length){
          this.tagList = val
        }else{
          this.tagList = []
        }
      },
      deep:true
    }
  },
  methods: {
    submitDialog() {
      if(this.type === 'section'){
        this.$emit('sendList',this.tagList)
      }else{
         this.bindTag()
      }
    },
    bindTag(){
      const ids = this.tagList.filter((item) => item.selected).map((item) => item.tagId);
      bindTag({ knowledgeId: this.knowledgeId, tagIdList: ids }).then((res) => {
        if (res.code === 0) {
          this.$emit("relodaData");
        }
      });
      this.dialogVisible = false;
    },
    getList() {
      tagList({ knowledgeId: this.knowledgeId, tagName: this.tagName }).then(
        (res) => {
          if (res.code === 0) {
            this.tagList = res.data.knowledgeTagList.map((item) => ({
              ...item,
              showDel: false,
              showIpt: false,
            }));
          }
        }
      );
    },
    bindTagCount(tagId){
      return bindTagCount({tagId}).then(res =>{
        if(res.code === 0){
          const tagBindCount = res.data.tagBindCount;
          return tagBindCount > 0;
        }
        return 'unknow'
      }).catch(() =>{
        return 'unknow'
      })
    },
    delTag(item,index){
      if(this.type !== 'section'){
          this.delTag_item(item)
      }else{
        this.tagList.splice(index,1)
      }
    },
    async delTag_item(item) {
      const isBind = await this.bindTagCount(item.tagId);
      if(isBind == 'unknow') return
      await this.$confirm(
        `Delete标签${item.tagName}`,
        item.selected && isBind ? "标签正在使用中，YesNoDelete？" : "Confirm要Delete当前标签？",
        {
          confirmButtonText: "确定",
          cancelButtonText: "Cancel",
          type: "warning",
        }
      )
        .then(async() => {
          const res = await delTag({ tagId: item.tagId })
            if (res.code === 0) {
                this.getList();
            }
        })
        .catch((error) => {
            this.getList();
        });
    },
    showDiaglog(id='') {
      this.dialogVisible = true;
      if(this.type !== 'section'){
        this.knowledgeId = id;
        this.getList();
      }
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
      if(!n.tagName) return;

      if(this.type === 'section'){
        n.showIpt = false;
        return
      }

      if (n.tagId) {
        this.edit_tag(n);
      } else {
        this.add_Tag(n);
      }
    },
    handleDelete(n,i){
      if(n.tagName === '' && !n.tagId){
          this.tagList.splice(i,1)
      }
    },
    add_Tag(n) {
      createTag({ tagName: n.tagName }).then((res) => {
        if (res.code === 0) {
          n.showIpt = false;
          this.getList();
        }
      });
    },
    edit_tag(n) {
      editTag({ tagId: n.tagId, tagName: n.tagName }).then((res) => {
        if (res.code === 0) {
          n.showIpt = false;
          this.getList();
        }
      });
    },
    addTag() {
      const emptyTag = this.tagList.find(tag => !tag.tagId && tag.tagName === "");
      if(emptyTag) return;
      if(this.type === 'section' && this.tagList.length > 10){
        this.$message.warning('最多可Create10个关键词')
        return;
      }
      this.tagList.unshift({
        tagName: "",
        checked: false,
        showDel: false,
        showIpt: true,
      });
    },
    addByEnterKey(e) {
      if (e.keyCode === 13) {
        this.getList();
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