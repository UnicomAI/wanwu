<template>
  <div>
    <el-dialog
      :title="titleMap[type]"
      :visible.sync="dialogVisible"
      width="750"
      append-to-body
      :close-on-click-modal="false"
    >
      <el-form ref="form" :model="form" label-width="120px" :rules="rules">
         <el-form-item :label="$t('ragDiglog.agentLogo')+ ':'" prop="avatar.path">
            <el-upload
                class="logo-upload"
                action=""
                multiple
                :show-file-list="false"
                :auto-upload="false"
                :limit="2"
                accept="image/*"
                :file-list="logoFileList"
                :on-change="uploadOnChange"
                >
            <div class="echo-img">
              
                <img :src="getImageSrc()" /> 
                <p class="echo-img-tip" v-if="isLoading">
                  图片Upload中
                  <span class="el-icon-loading"></span>
                </p>
                <p class="echo-img-tip" v-else>点击Upload图片</p>
            </div>
            </el-upload>
        </el-form-item>
        <el-form-item :label="$t('ragDiglog.agentName')+':'" prop="name">
          <el-input
            :placeholder="$t('ragDiglog.nameplaceholder')"
            v-model="form.name"
            maxlength="30"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('ragDiglog.agentDesc')+':'" prop="desc">
          <el-input
            type="textarea"
            :placeholder="$t('ragDiglog.descplaceholder')"
            v-model="form.desc"
            show-word-limit
            maxlength="600"
          ></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">{{$t('list.cancel')}}</el-button>
        <el-button type="primary" @click="doPublish">{{$t('list.confirm')}}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { uploadAvatar} from "@/api/user";
import { createRag,updateRag} from "@/api/rag";
import {mapActions,mapGetters} from 'vuex';
export default {
  props: {
    type: {
      type: String,
      default: "create",
    },
    editForm: {
      type: Object,
    },
  },
  data() {
    return {
      isLoading:false,
      defaultLogo:'',
      logoFileList:[],
      imageUrl:'',
      dialogVisible: false,
      ragId:'',
      form: {
        name: "",
        desc: "",
        avatar:{
          key:'',
          path:''
        },
      },
      rules: {
        name: [
          { required: true, message: this.$t('ragDiglog.nameRules'), trigger: "blur" },
          {
            validator: (rule, value, callback) => {
              if (/^[A-Za-z0-9.\u4e00-\u9fa5_-]+$/.test(value)) {
                callback();
              } else {
                callback(
                  new Error(
                    this.$t('ragDiglog.nameplaceholder')
                  )
                );
              }
            },
            trigger: "change",
          },
          {
            max:30,message:this.$t('ragDiglog.pluginNameRules'),trigger: "blur"
          }
        ],
        desc: [
          { required: true, message: 'Please enter文本Description', trigger: "blur" },
          { max:600, message:'文本Description限制600字符以内',trigger: "blur"}
        ],
      },
      titleMap: {
        edit: this.$t('ragDiglog.editApp'),
        create:this.$t('ragDiglog.createApp')
      },
    };
  },
  computed: {
    ...mapGetters('user', ['defaultIcons'])
  },
  watch:{
    defaultIcons:{
      handler(newVal) {
        if(this.type === 'create'){
          this.defaultLogo = newVal.ragIcon;
        }
      },
      deep: true,
      immediate: true
    }
  },
  methods: {
    ...mapActions("app", ["setFromList"]),
    getImageSrc(){
      if (this.imageUrl) return this.imageUrl;
      if (this.type === 'create') {
        return this.defaultLogo ? `/user/api/${this.defaultLogo}` : '';
      } else {
        return this.form.avatar.path ? `/user/api/${this.form.avatar.path}` : '';
      }
    },
    openDialog() {
      if (this.type === "edit" && this.editForm) {
        this.defaultLogo = ''
        const formInfo =JSON.parse(JSON.stringify(this.editForm));
        this.form.name = formInfo.name;
        this.form.desc = formInfo.desc;
        this.form.avatar = formInfo.avatar
        this.ragId = formInfo.appId
      } else {
        this.clearForm();
      }
      this.dialogVisible = true;
      this.$nextTick(()=>{
          this.$refs['form'].clearValidate()
      })
    },
    clearForm() {
      this.form = {
        name: "",
        desc: "",
        avatar:{
          key:'',
          path:''
        }
      };
      this.ragId = '';
      this.imageUrl = ''
    },
    uploadOnChange(file){
      this.clearFile();
      this.logoFileList.push(file);
      this.imageUrl = URL.createObjectURL(file.raw);
      this.doLogoUpload();
    },
    clearFile() {
      this.form.avatar.path = "";
      this.logoFileList = [];
    },
    doLogoUpload(){
      var formData = new FormData();
      var config = { headers: { "Content-Type": "multipart/form-data" } };
      var file = this.logoFileList[0];
      formData.append("avatar", file.raw, file.name);
      this.isLoading = true;
      uploadAvatar(formData,config).then(res => {
        if(res.code === 0){
            this.form.avatar = res.data
            this.isLoading = false;
        }
      }).catch((error) =>{
        this.clearFile();
      })
    },
    async doPublish() {
      let valid = false;
      await this.$refs.form.validate((vv) => {
        if (vv) {
          valid = true;
        }
      });
      if (!valid) return;
      if(this.type === 'create'){
        this.createRag();
      }else{
        this.editRag();
      }
    },
    createRag(){
        createRag(this.form).then(res =>{
          if(res.code === 0){
            this.dialogVisible = false;
            const type = 'rag';
            const id = res.data.ragId;
            this.$router.push({path:`/rag/test?id=${id}`})
            this.setFromList(type)
          }
        })
    },
    editRag(){
      const data = {
        ...this.form,
        ragId:this.ragId
      }
      updateRag(data).then(res =>{
        if(res.code === 0){
          this.dialogVisible = false;
          this.$emit('updateInfo')
        }
      })
    },
  },
};
</script>

<style lang="scss" scoped>
.logo-upload {
    width: 100px;
    height: 100px;
    margin-top: 3px;
    /deep/ {
      .el-upload {
        width: 100%;
        height: 100%;
        border-radius:6px;
        border:1px solid #DCDFE6;
        overflow:hidden;
      }
      .echo-img {
        width: 100%;
        height: 100%;
        position: relative;
        img {
          object-fit: cover;
          height: 100%;
        }
        .echo-img-tip {
          position: absolute;
          width: 100%;
          bottom: 0;
          background: $color_opacity;
          color: $color  !important;
          font-size: 12px;
          line-height: 26px;
          z-index: 10;
        }
      }
    }
  }
</style>
