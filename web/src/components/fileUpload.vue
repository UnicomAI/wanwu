<template>
  <div>
    <el-upload
      class="upload-box"
      drag
      action=""
      :show-file-list="false"
      :auto-upload="false"
      :accept="accept"
      :limit="maxFileCount"
      :file-list="fileList"
      :on-change="uploadOnChange"
    >
      <div>
        <div>
          <img
            :src="require('@/assets/imgs/uploadImg.png')"
            class="upload-img"
          />
          <p class="click-text">
            {{ $t('common.fileUpload.uploadText') }}
            <span class="clickUpload">
              {{ $t('common.fileUpload.uploadClick') }}
            </span>
            <a
              class="clickUpload template"
              @click.stop="handleDownloadTemplate"
            >
              {{ $t('common.fileUpload.templateClick') }}
            </a>
          </p>
          <slot name="upload-tips" />
        </div>
      </div>
    </el-upload>
    <!-- 上传文件的列表 -->
    <div class="file-list" v-if="fileList.length > 0">
      <transition name="el-zoom-in-top">
        <ul class="document_lise">
          <li
            v-for="(file, index) in fileList"
            :key="index"
            class="document_lise_item"
          >
            <div style="padding: 8px 0" class="lise_item_box">
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
                    <i
                      class="el-icon-check check success"
                      v-if="file.progressStatus === 'success'"
                    ></i>
                    <i class="el-icon-close close fail" v-else></i>
                  </span>
                  <i
                    class="el-icon-loading"
                    v-else-if="file.percentage !== 100 && index === fileIndex"
                  ></i>
                </span>
                <span style="margin-left: 30px">
                  <i
                    class="el-icon-error error"
                    @click="handleRemove(file, index)"
                  ></i>
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
import uploadChunk from '@/mixins/uploadChunk';
import { delfile } from '@/api/chunkFile';
import { filterSize, authDownload } from '@/utils/util';

export default {
  props: {
    templateUrl: String,
    accept: String,
    maxSize: Number,
    maxFileCount: {
      type: Number,
      default: undefined,
    },
  },
  mixins: [uploadChunk],
  data() {
    return {
      // fileList: [],
    };
  },
  methods: {
    async handleDownloadTemplate() {
      if (!this.templateUrl) return;
      const fileName = this.templateUrl.split('/').pop();
      try {
        await authDownload(this.templateUrl, fileName);
      } catch (error) {
        this.$message.error(this.$t('common.fileUpload.downloadError'));
      }
    },
    uploadOnChange(file, fileList) {
      if (!fileList.length) return;
      // 验证文件大小，只有通过验证才继续上传
      if (!this.validateFileSize(file)) {
        return;
      }
      // 验证文件格式，只有通过验证才继续上传
      if (!this.validateFileFormat(file)) {
        return;
      }
      this.fileList = fileList;
      if (this.fileList.length > 0) {
        this.maxSizeBytes = 0;
        this.isExpire = true;
        this.startUpload();
      }
    },
    validateFileSize(file) {
      if (file.size === 0) {
        this.$message.warning(this.$t('common.fileUpload.fileSizeError'));
        return false;
      }
      // 文件大小限制处理，maxSize为可选属性
      if (this.maxSize) {
        const maxSizeBytes = this.maxSize * 1024 * 1024;
        if (file.size > maxSizeBytes) {
          this.$message.error(
            this.$t('common.fileUpload.fileSizeLimit') ||
              `文件大小不能超过${this.maxSize}MB`,
          );
          return false;
        }
      }
      return true;
    },
    validateFileFormat(file) {
      // 文件格式验证，accept为可选属性
      if (this.accept) {
        const fileName = file.name;
        const lastDotIndex = fileName.lastIndexOf('.');
        if (lastDotIndex === -1) {
          this.$message.error(this.$t('common.fileUpload.fileFormatError'));
          return false;
        }
        const fileExtension = fileName.slice(lastDotIndex).toLowerCase();

        const acceptFormats = this.accept
          .split(',')
          .map(format => format.trim().toLowerCase());

        if (!acceptFormats.includes(fileExtension)) {
          this.$message.error(this.$t('common.fileUpload.fileFormatError'));
          return false;
        }
      }
      return true;
    },
    uploadFile(chunkFileName, fileName, filePath) {
      this.$emit('uploadFile', chunkFileName, fileName, filePath);
    },
    clearFileList() {
      this.fileList = [];
    },
    filterSize,
    handleRemove(item, index) {
      const data = { fileList: [this.resList[index]['name']], isExpired: true };
      delfile(data).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.delete'));
        }
      });
      this.fileList = this.fileList.filter(files => files.name !== item.name);
      if (this.fileList.length === 0) {
        this.file = null;
      } else {
        this.fileIndex--;
      }
    },
  },
};
</script>
<style lang="scss" scoped>
.success {
  color: #67c23a;
}

.fail {
  color: #f56c6c;
}

.upload-box {
  .upload-img {
    width: 56px;
    height: 56px;
    margin-top: 10px;
  }

  .clickUpload,
  .template {
    color: $color;
    font-weight: bold;
  }

  .template {
    margin-left: 10px;
  }
}

.file-list {
  padding: 20px 0;

  .document_lise_item {
    cursor: pointer;
    padding: 0 10px;
    list-style: none;
    border-radius: 4px;
    border: 1px solid #7684fd;
    display: flex;
    align-items: center;
    margin-bottom: 10px;

    .lise_item_box {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .size {
        display: flex;
        flex: 1;
        align-items: center;

        .progress {
          width: 200px;
          margin-left: 30px;
        }

        img {
          width: 18px;
          height: 18px;
          margin-bottom: -3px;
        }

        .file-size {
          margin-left: 10px;
        }
      }
    }
  }

  .document_lise_item:hover {
    background: #eceefe;
  }
}
</style>
