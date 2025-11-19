<template>
    <div>
      <el-dialog
        top="10vh"
        title="AddSensitive Word"
        :close-on-click-modal="false"
        :visible.sync="dialogVisible"
        width="50%"
        :before-close="handleClose"
        >
        <el-form
            :model="ruleForm"
            ref="ruleForm"
            label-width="140px"
            class="demo-ruleForm"
            :rules="rules"
            @submit.native.prevent
        >
            <el-form-item class="itemCenter">
              <el-radio-group v-model="ruleForm.importType">
                <el-radio-button :label="'single'">单条Add</el-radio-button>
                <el-radio-button :label="'file'">BatchUpload</el-radio-button>
              </el-radio-group>  
            </el-form-item>
            <el-form-item
            label="Sensitive Word"
            prop="word"
            v-if="ruleForm.importType === 'single'"
            >
            <el-input
                v-model="ruleForm.word"
                placeholder="您可Add一个词"
            ></el-input>
            </el-form-item>
            <el-form-item
            label="Sensitive WordType"
            prop="sensitiveType"
            v-if="ruleForm.importType === 'single'"
            >
            <el-select v-model="ruleForm.sensitiveType" placeholder="Please select" style="width:100%;">
                <el-option
                v-for="item in sensitiveTypeOptions"
                :key="item.value"
                :label="item.name"
                :value="item.value">
                </el-option>
            </el-select>
            </el-form-item>
            <el-form-item
            label="BatchUpload"
            prop="fileName"
            v-if="ruleForm.importType === 'file'"
            >
            <el-upload
                class="upload-box"
                drag
                action=""
                :show-file-list="false"
                :auto-upload="false"
                accept=".xlsx"
                :file-list="fileList"
                :on-change="uploadOnChange"
              >
              <div>
                <div>
                    <img :src="require('@/assets/imgs/uploadImg.png')" class="upload-img" />
                    <p class="click-text">
                        将File拖到此处，OR
                        <span class="clickUpload">点击Upload</span>
                        <a class="clickUpload template" :href="`/user/api/v1/static/docs/sensitive.xlsx`" download @click.stop>模版Download</a>
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
            </el-form-item>
        </el-form>
        <span
            slot="footer"
            class="dialog-footer"
        >
            <el-button 
                @click="handleClose()">
                {{$t('common.confirm.cancel')}}
            </el-button>
            <el-button
                type="primary"
                @click="submitForm('ruleForm')"
            >{{$t('common.confirm.confirm')}}</el-button>
        </span>
        </el-dialog>
    </div>
</template>
<script>
import uploadChunk from "@/mixins/uploadChunk";
import { delfile } from "@/api/chunkFile";
import { uploadSensitiveWord } from "@/api/safety";
export default {
    mixins: [uploadChunk],
    data(){
        return{
            sensitiveTypeOptions:[
                {
                    value:'Political',
                    name:'Political'
                },
                {
                    value:'Revile',
                    name:'Abusive'
                },
                {
                    value:'Pornography',
                    name:'Pornographic'
                },
                {
                    value:'ViolentTerror',
                    name:'Violent/Terrorist'
                },
                {
                    value:'Illegal',
                    name:'Prohibited'
                },
                {
                    value:'InformationSecurity',
                    name:'Information Security'
                },
                {
                    value:'Other',
                    name:'Other'
                }
            ],
            title:"新建词表",
            dialogVisible:false,
            ruleForm:{
                importType:'single',
                word:'',
                sensitiveType:'',
                fileName:'',
                tableId:''
            },
            fileList:[],
            rules: {
                word: [{ required: true, message:'Please enterSensitive Word', trigger: "blur" }],
                sensitiveType:[{ required: true, message: 'Please输选择Sensitive WordType', trigger: "blur" }],
                fileName:[{ required: true, message: 'PleaseUploadFile', trigger: "blur" }]
            }
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
        },
        uploadFile(chunkFileName){
            this.ruleForm.fileName = chunkFileName;
        },
        handleClose(){
            this.dialogVisible = false;
            this.ruleForm.tableId = '';
            this.clearform()
        },
        clearform(){
            this.fileList = []
            this.$refs.ruleForm.resetFields()
            this.$refs.ruleForm.clearValidate()
        },
        submitForm(formName){
            this.$refs[formName].validate((valid) =>{
                if(valid){
                   uploadSensitiveWord(this.ruleForm).then(res =>{
                    if(res.code == 0){
                        this.$message.success('OperationSuccess')
                        this.$emit('reload')
                        this.dialogVisible = false;
                    }
                   }).catch(err =>{
                     
                   })
                }else{
                    return false;
                }
            })
        },
        showDialog(tableId){
            this.dialogVisible = true;
            this.ruleForm.tableId = tableId;
            this.clearform();
        }
    }
}
</script>
<style lang="scss" scoped>
.itemCenter{
    display:flex;
    justify-content: center;
    /deep/.el-form-item__content{
        margin-left: 0 !important;
    }
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