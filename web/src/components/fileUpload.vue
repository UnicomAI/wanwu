<template>
    <div>
        <el-upload
            class="upload-box"
            drag
            action=""
            :show-file-list="false"
            :auto-upload="false"
            :accept="accept"
            :file-list="fileList"
            :on-change="uploadOnChange"
            >
            <div>
            <div>
                <img :src="require('@/assets/imgs/uploadImg.png')" class="upload-img" />
                <p class="click-text">
                    将File拖到此处，OR
                    <span class="clickUpload">点击Upload</span>
                    <a class="clickUpload template" :href="templateUrl" download @click.stop>模版Download</a>
                </p>
            </div>
            </div>
        </el-upload>
        <!-- UploadFile of List -->
        <div class="file-list" v-if="fileList.length > 0">
                <transition name="el-zoom-in-top">
                <ul class="document_lise">
                <li
                    v-for="(file, index) in fileList"
                    :key="index"
                    class="document_lise_item"
                >
                    <div style="padding:8px 0;" class="lise_item_box">
                    <span class="size">
                        <img :src="require('@/assets/imgs/fileicon.png')" />
                        {{ file.name }}
                        <span class="file-size">
                        {{ filterSize(file.size) }}
                        </span>
                        <el-progress 
                            :percentage="file.percentage" 
                            v-if="file.percentage !== 100"
                            :status="file.progressStatus"
                            max="100"
                            class="progress"
                        ></el-progress>
                    </span>
                    <span class="handleBtn">
                        <span>
                        <span v-if="file.percentage === 100">
                            <i class="el-icon-check check success" v-if="file.progressStatus === 'success'"></i>
                            <i class="el-icon-close close fail" v-else></i>
                        </span>
                        <i class="el-icon-loading" v-else-if="file.percentage !== 100 && index === fileIndex"></i>
                        </span>
                        <span style="margin-left:30px;">
                        <i class="el-icon-error error" @click="handleRemove(file,index)"></i>
                        </span>
                    </span>
                    </div>
                </li>
                </ul>
            </transition>
        </div>
    </div>
</template>
<script>
import uploadChunk from "@/mixins/uploadChunk";
import { delfile } from "@/api/chunkFile";
export default {
    props:['templateUrl','accept'],
    mixins: [uploadChunk],
    data(){
        return{
            fileList:[],
        }
    },
    methods:{
        uploadOnChange(file, fileList){
            if (!fileList.length) return;
            this.fileList = fileList;
            if(this.fileList.length > 0){
                this.maxSizeBytes = 0;
                this.isExpire = true;
                this.startUpload();
            }
        },
        uploadFile(chunkFileName){
            this.$emit('uploadFile',chunkFileName)
        },
        clearFileList(){
            this.fileList = [];
        },
        filterSize(size) {
            if (!size) return "";
            var num = 1024.0; //byte
            if (size < num) return size + "B";
            if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + "KB"; //kb
            if (size < Math.pow(num, 3))
                return (size / Math.pow(num, 2)).toFixed(2) + "MB"; //M
            if (size < Math.pow(num, 4))
                return (size / Math.pow(num, 3)).toFixed(2) + "G"; //G
            return (size / Math.pow(num, 4)).toFixed(2) + "T"; //T
        },
        handleRemove(item,index){
           const data = {fileList:[this.resList[index]['name']],isExpired:true}
            delfile(data).then(res =>{
                if(res.code === 0){
                this.$message.success('DeleteSuccess')
                }
            })
            this.fileList = this.fileList.filter((files) => files.name !== item.name);
            if(this.fileList.length === 0){
                this.file = null
            }else{
                this.fileIndex--
            }
        }
    }
}
</script>
<style lang="scss" scoped>
.success{
    color: #67C23A;
}
.fail{
    color: #F56C6C;
}
.upload-box{
    .upload-img{
        width:56px;
        height:56px;
        margin-top: 10px;
    }
    .clickUpload,.template{
       color: $color;
       font-weight: bold;
    }
    .template{
        margin-left:10px;
    }
}
.file-list{
  padding: 20px 0;
  .document_lise_item{
    cursor: pointer;
    padding:0 10px;
    list-style: none;
    border-radius:4px;
    border:1px solid #7684fd;
    display:flex;
    align-items:center;
    margin-bottom:10px;
    .lise_item_box{
      width:100%;
      display:flex;
      align-items:center;
      justify-content:space-between;
      .size{
          display:flex;
          flex:1;
          align-items:center;
          .progress{
            width:200px;
            margin-left:30px;
          }
          img{
            width: 18px;
            height:18px;
            margin-bottom:-3px;
          }
          .file-size{
            margin-left:10px;
          }
      }
    }
  }
  .document_lise_item:hover{
    background: #ECEEFE;
  }
}
</style>