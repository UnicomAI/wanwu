<template>
  <div>
    <el-dialog
      :title="titleMap[type]"
      :visible.sync="dialogVisible"
      width="750"
      append-to-body
      :close-on-click-modal="false"
    >
      <div class="agentCategoryList" v-if="type === 'create'">
        <div
          v-for="agentCategoryItem in agentCategoryList"
          :key="agentCategoryItem.category"
          :class="[
            'agentCategoryItem',
            form.category === agentCategoryItem.category ? 'active' : '',
          ]"
          @click="form.category = agentCategoryItem.category"
        >
          <div class="itemImg">
            <img
              :src="require(`@/assets/imgs/${agentCategoryItem.img}`)"
              alt="agentCategoryItem.text"
            />
          </div>
          <div>
            <p class="agentCategoryItem_text">{{ agentCategoryItem.text }}</p>
            <h3 class="agentCategoryItem_desc">{{ agentCategoryItem.desc }}</h3>
          </div>
        </div>
      </div>
      <el-form ref="form" :model="form" label-width="120px" :rules="rules">
        <el-form-item
          :label="$t('agentDialog.agentLogo') + ':'"
          prop="avatar.path"
        >
          <el-upload
            class="logo-upload"
            action=""
            multiple
            :show-file-list="false"
            :auto-upload="false"
            :limit="2"
            accept=".png,.jpg,.jpeg"
            :file-list="logoFileList"
            :on-change="uploadOnChange"
          >
            <div class="echo-img">
              <img :src="getImageSrc()" />
              <p class="echo-img-tip" v-if="isLoading">
                {{ $t('common.fileUpload.imgUploading') }}
                <span class="el-icon-loading"></span>
              </p>
              <p class="echo-img-tip" v-else>
                {{ $t('common.fileUpload.clickUploadImg') }}
              </p>
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item :label="$t('agentDialog.agentName') + ':'" prop="name">
          <el-input
            :placeholder="$t('agentDialog.nameplaceholder')"
            v-model="form.name"
            maxlength="50"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('agentDialog.agentDesc') + ':'" prop="desc">
          <el-input
            type="textarea"
            :placeholder="$t('agentDialog.descplaceholder')"
            v-model="form.desc"
            show-word-limit
            maxlength="200"
          ></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">
          {{ $t('list.cancel') }}
        </el-button>
        <el-button type="primary" @click="doPublish">
          {{ $t('list.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { uploadAvatar } from '@/api/user';
import { createAgent, updateAgent } from '@/api/agent';
import { mapActions, mapGetters } from 'vuex';
import { MULTIPLE_AGENT, SINGLE_AGENT } from '@/views/agent/constants';
import { avatarSrc } from '@/utils/util';
export default {
  props: {
    type: {
      type: String,
      default: 'create',
    },
    editForm: {
      type: Object,
    },
  },
  data() {
    return {
      agentCategoryList: [
        {
          category: SINGLE_AGENT,
          img: 'agent_single.png',
          text: this.$t('agentDialog.singleAgent'),
          desc: this.$t('agentDialog.singleAgentDesc'),
        },
        {
          category: MULTIPLE_AGENT,
          img: 'agent_multiple.png',
          text: this.$t('agentDialog.multipleAgent'),
          desc: this.$t('agentDialog.multipleAgentDesc'),
        },
      ],
      isLoading: false,
      defaultLogo: '',
      logoFileList: [],
      imageUrl: '',
      dialogVisible: false,
      assistantId: '',
      form: {
        category: SINGLE_AGENT,
        name: '',
        desc: '',
        avatar: {
          key: '',
          path: '',
        },
      },
      rules: {
        name: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            pattern: this.$config.commonTextReg,
            message: this.$t('common.hint.text'),
            trigger: 'blur',
          },
          {
            min: 2,
            max: 50,
            message: this.$t('common.hint.textLimit'),
            trigger: 'blur',
          },
        ],
        desc: [
          {
            required: true,
            message: this.$t('agentDialog.descplaceholder'),
            trigger: 'blur',
          },
          {
            max: 200,
            message: this.$t('common.hint.descLimit'),
            trigger: 'blur',
          },
        ],
      },
      titleMap: {
        edit: this.$t('agentDialog.editApp'),
        create: this.$t('agentDialog.createApp'),
      },
    };
  },
  computed: {
    ...mapGetters('user', ['defaultIcons']),
  },
  watch: {
    defaultIcons: {
      handler(newVal) {
        if (this.type === 'create') {
          this.defaultLogo = newVal.agentIcon;
        }
      },
      deep: true,
      immediate: true,
    },
  },
  methods: {
    ...mapActions('app', ['setFromList']),
    getImageSrc() {
      if (this.imageUrl) return this.imageUrl;
      if (this.type === 'create') {
        return avatarSrc(this.defaultLogo);
      } else {
        return avatarSrc(this.form.avatar.path);
      }
    },
    openDialog() {
      if (this.type === 'edit' && this.editForm) {
        this.defaultLogo = '';
        const formInfo = structuredClone(this.editForm);
        this.form.name = formInfo.name || '';
        this.form.desc = formInfo.desc || '';
        this.form.avatar = formInfo.avatar || {};
        this.assistantId = formInfo.appId || formInfo.assistantId;
      } else {
        this.clearForm();
      }
      this.dialogVisible = true;
      this.$nextTick(() => {
        this.$refs['form'].clearValidate();
      });
    },
    clearForm() {
      this.form = {
        category: SINGLE_AGENT,
        name: '',
        desc: '',
        avatar: {
          key: '',
          path: '',
        },
      };
      this.assistantId = '';
      this.imageUrl = '';
    },
    uploadOnChange(file) {
      this.clearFile();
      this.logoFileList.push(file);
      this.imageUrl = URL.createObjectURL(file.raw);
      this.doLogoUpload();
    },
    clearFile() {
      this.form.avatar.path = '';
      this.logoFileList = [];
    },
    doLogoUpload() {
      let formData = new FormData();
      let config = { headers: { 'Content-Type': 'multipart/form-data' } };
      let file = this.logoFileList[0];
      formData.append('avatar', file.raw, file.name);
      this.isLoading = true;
      uploadAvatar(formData, config)
        .then(res => {
          if (res.code === 0) {
            this.form.avatar = res.data;
            this.isLoading = false;
          }
        })
        .catch(error => {
          this.clearFile();
        });
    },
    async doPublish() {
      let valid = false;
      await this.$refs.form.validate(vv => {
        if (vv) {
          valid = true;
        }
      });
      if (!valid) return;
      if (this.type === 'create') {
        this.createAgent();
      } else {
        this.editAgent();
      }
    },
    createAgent() {
      createAgent(this.form).then(res => {
        if (res.code === 0) {
          this.dialogVisible = false;
          const type = 'agent';
          const id = res.data.assistantId;
          this.$router.push({ path: `/agent/test?id=${id}` });
          this.setFromList(type);
        }
      });
    },
    editAgent() {
      const data = {
        ...this.form,
        assistantId: this.assistantId,
      };
      updateAgent(data).then(res => {
        if (res.code === 0) {
          this.dialogVisible = false;
          this.$emit('updateInfo');
        }
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.logo-upload {
  width: 100px;
  height: 100px;
  margin-top: 3px;
  ::v-deep {
    .el-upload {
      width: 100%;
      height: 100%;
      border-radius: 6px;
      border: 1px solid #dcdfe6;
      overflow: hidden;
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
        color: $color !important;
        font-size: 12px;
        line-height: 26px;
        z-index: 10;
      }
    }
  }
}

.agentCategoryList {
  display: flex;
  margin-bottom: 20px;
  gap: 15px;

  .agentCategoryItem {
    display: flex;
    align-items: center;
    cursor: pointer;
    border: 1px solid #ddd;
    padding: 10px;
    border-radius: 6px;
    gap: 15px;
    width: 50%;

    &.active {
      border-color: $color;
    }

    .itemImg {
      width: 45px;
      height: 45px;
      border: 1px solid #eeeded;
      border-radius: 8px;
      display: flex;
      justify-content: center;
      align-items: center;
      box-shadow: 0px 2px 4px -2px rgba(16, 24, 40, 0.06);

      img {
        width: 25px;
        height: fit-content;
      }
    }

    .agentCategoryItem_text {
      font-size: 14px;
      font-weight: 600;
      line-height: 1.8;
    }

    .agentCategoryItem_desc {
      line-height: 1.2;
      color: #b4b3b3;
      font-weight: unset;
    }
  }
}
</style>
