<template>
  <div class="detail-ul">
    <div class="avatar">
      <el-upload
          class="avatar-uploader"
          action=""
          name="files"
          :show-file-list="false"
          :multiple="false"
          :http-request="handleUploadAvatar"
          :on-error="handleUploadError"
          accept=".png,.jpg,.jpeg"
      >
        <div class="echo-img">
          <img :src="form.avatar || defaultAvatar" alt=""/>
          <p class="echo-img-tip" v-if="isLoading">
            Image uploading
            <span class="el-icon-loading"></span>
          </p>
          <p class="echo-img-tip" v-else>Click to upload image</p>
        </div>
      </el-upload>
      <div class="row">
        <label>{{$t('userInfo.username')}}</label>
        <span>{{form.username}}</span>
      </div>
    </div>
    <div class="row">
      <label>{{$t('userInfo.password')}}</label>
      <span class="pwd-span">{{form.password || '--'}}</span>
      <el-button type="primary" size="mini" style="margin-left: 30px;" @click="showPwd">Modify密码</el-button>
    </div>
    <div class="row"><label>{{$t('userInfo.company')}}</label><span>{{form.company || '--'}}</span></div>
    <div class="row"><label>{{$t('userInfo.phone')}}</label><span>{{form.phone || '--'}}</span></div>
    <div class="row"><label>{{$t('userInfo.email')}}</label><span>{{form.email || '--'}}</span></div>
    <div class="row"><label>{{$t('userInfo.remark')}}</label><span>{{form.remark || '--'}}</span></div>
    <el-dialog
      :title="$t('resetPwd.title')"
      :visible.sync="pwdVisible"
      width="600px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <Pwd @handleClose="handleClose" />
    </el-dialog>
  </div>
</template>

<script>
import Pwd from "../pwd/index.vue"
import {uploadAvatar, restAvatar} from "@/api/user";
import {avatarSrc} from "@/utils/util";
export default {
  components: {Pwd},
  data(){
    return{
      defaultAvatar:require("@/assets/imgs/avatar_default.png"),
      isLoading: false,
      form:{
        avatar:'',
        userId:'',
        username:'',
        password: '',
        company:'',
        phone:'',
        email:'',
        remark: ''
      },
      pwdVisible: false
    }
  },
  watch: {
    $route: {
      handler () {
        this.justifyShowPwd()
      },
      deep: true
    },
  },
  mounted() {
    this.justifyShowPwd()
  },
  methods:{
    handleUploadAvatar(data) {
      if (data.file) {
        this.isLoading = true
        const formData = new FormData()
        const config = {headers: {"Content-Type": "multipart/form-data"}}
        formData.append('avatar', data.file)
        this.isLoading = true;
        uploadAvatar(formData, config).then(res => {
          if (res.code === 0) {
            const avatar = avatarSrc(res.data.path)
            restAvatar({avatar: res.data}).then(res => {
              if (res.code === 0) {
                this.form.avatar = avatar
                this.$forceUpdate()
              }
            })
          }
        }).finally(() => this.isLoading = false)
      }
    },
    handleUploadError() {
      this.$message.error(this.$t('common.message.uploadError'))
    },
    justifyShowPwd() {
      const {showPwd} = this.$route.query || {}
      if (showPwd === '1') {
        this.showPwd()
      }
    },
    setData(data){
      const {avatar, userId, username, company, phone, email, remark} = data || {}
      this.form = {userId, username, company, phone, email, remark, password: '***'}
      this.form.avatar = avatarSrc(avatar.path)
    },
    showPwd() {
      this.pwdVisible = true
    },
    handleClose() {
      this.pwdVisible = false
    }
  }
}
</script>

<style lang="scss" scoped>
.detail-ul{
  width: 500px;
  margin: 26px auto;

  .row{
    display: flex;
    justify-content: space-between;
    flex-wrap: wrap;
    align-items: center;
    line-height: 26px;
    padding: 14px 0;
    label{
      color: $color_title;
      font-size: 15px;
      font-weight: bold;
      width: 50px;
    }
    span{
      color: #425466;
      flex: 1;
      //background: #fff;
      border-radius: 4px;
      margin-top: 5px;
      padding: 3px 20px;
      box-shadow: 0 0 15px 0 rgba(89,104,178,0.06), 0 15px 20px 0 rgba(89,104,178,0.06);
    }
    .pwd-span {
      width: 306px;
    }
  }
  .avatar {
    display: flex;
    align-items: center;
    gap: 10px;

    .avatar-uploader {
      width: 128px;
      height: 128px;
      flex-shrink: 0;
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
          position:relative;
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

    .row {
      flex: 1;
    }
  }
}
</style>
