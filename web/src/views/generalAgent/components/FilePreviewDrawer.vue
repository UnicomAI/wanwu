<template>
  <transition name="preview-slide">
    <div v-if="visible" class="preview-panel">
      <div class="preview-header">
        <div :title="fileName" class="preview-title">
          {{ fileName }}
        </div>
        <div class="preview-actions">
          <el-button size="small" @click="handleDownload">
            <i class="el-icon-download"></i>
            {{ $t('common.button.download') }}
          </el-button>
          <el-button
            size="small"
            v-if="
              previewUrl &&
              ['image', 'video', 'audio', 'pdf', 'html'].includes(previewType)
            "
            @click="openInNewTab"
          >
            <i class="el-icon-link"></i>
            {{ $t('generalAgent.filePreview.newWindow') }}
          </el-button>
          <button v-if="showClose" class="close-btn" @click="handleClose">
            <i class="el-icon-close"></i>
          </button>
        </div>
      </div>
      <div class="preview-body" ref="previewBody">
        <!-- 加载中 -->
        <div v-if="loading" class="preview-loading">
          <i class="el-icon-loading"></i>
          <span>{{ $t('generalAgent.filePreview.loading') }}</span>
        </div>

        <!-- 预览内容 -->
        <template v-else>
          <!-- 图片预览 -->
          <div v-if="previewType === 'image'" class="preview-image-wrapper">
            <img :src="previewUrl" class="preview-image" @error="handleError" />
          </div>

          <!-- 视频预览 -->
          <div
            v-else-if="previewType === 'video'"
            class="preview-video-wrapper"
          >
            <video :src="previewUrl" controls class="preview-video">
              {{ $t('common.fileUpload.videoTips') }}
            </video>
          </div>

          <!-- 音频预览 -->
          <div
            v-else-if="previewType === 'audio'"
            class="preview-audio-wrapper"
          >
            <div class="audio-cover">
              <i class="el-icon-headset"></i>
            </div>
            <audio :src="previewUrl" controls class="preview-audio">
              {{ $t('common.fileUpload.audioTips') }}
            </audio>
          </div>

          <!-- PDF 预览 -->
          <div v-else-if="previewType === 'pdf'" class="preview-pdf-wrapper">
            <iframe :src="previewUrl" class="preview-pdf"></iframe>
          </div>

          <!-- PPT 预览 -->
          <div v-else-if="previewType === 'ppt'" class="preview-ppt-wrapper">
            <ppt-preview
              :src="internalBlob"
              :file-name="fileName"
              @close="handleClose"
            />
          </div>

          <!-- HTML 预览 -->
          <div v-else-if="previewType === 'html'" class="preview-html-wrapper">
            <iframe
              :src="previewUrl"
              class="preview-html-frame"
              sandbox="allow-scripts allow-same-origin"
            ></iframe>
          </div>

          <!-- Excel 预览 -->
          <div
            v-else-if="previewType === 'excel' && previewExcelData"
            class="preview-excel-wrapper"
          >
            <div class="excel-tabs">
              <button
                v-for="(sheet, index) in previewExcelData"
                :key="sheet.name"
                :class="['excel-tab', { active: activeSheetIndex === index }]"
                @click="activeSheetIndex = index"
              >
                {{ sheet.name }}
              </button>
            </div>
            <div class="excel-table-container">
              <table
                class="excel-table"
                v-if="previewExcelData[activeSheetIndex]"
              >
                <tbody>
                  <tr
                    v-for="(row, rowIndex) in previewExcelData[activeSheetIndex]
                      .data"
                    :key="rowIndex"
                  >
                    <td
                      v-for="(cell, colIndex) in row"
                      :key="colIndex"
                      :class="{
                        'excel-header': rowIndex === 0,
                        'merged-cell': isMergedCell(
                          activeSheetIndex,
                          rowIndex,
                          colIndex,
                        ),
                      }"
                      :rowspan="
                        getRowspan(activeSheetIndex, rowIndex, colIndex)
                      "
                      :colspan="
                        getColspan(activeSheetIndex, rowIndex, colIndex)
                      "
                      v-show="
                        !isHiddenByMerge(activeSheetIndex, rowIndex, colIndex)
                      "
                    >
                      {{ cell }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Markdown 预览 -->
          <div
            v-else-if="previewType === 'markdown'"
            class="preview-markdown-wrapper"
          >
            <stream-markdown :content="previewContent" />
          </div>

          <!-- Word 预览 -->
          <div
            v-else-if="previewType === 'word'"
            class="preview-office-wrapper"
          >
            <vue-office-docx :src="internalBlob" @error="handleWordError" />
          </div>

          <!-- OFD 预览 -->
          <div v-else-if="previewType === 'ofd'" class="preview-ofd-wrapper">
            <ofd-preview
              :file-name="fileName"
              :src="internalBlob"
              @close="handleClose"
            />
          </div>

          <!-- 文本/代码预览 -->
          <div v-else-if="previewType === 'text'" class="preview-text-wrapper">
            <stream-markdown :content="fencedCode" />
          </div>

          <!-- 不支持的格式 -->
          <div v-else class="preview-unsupported">
            <i class="el-icon-document"></i>
            <p class="file-name">{{ fileName }}</p>
            <p class="notice-text">
              {{ $t('generalAgent.filePreview.unsupportedType') }}
            </p>
          </div>
        </template>
      </div>
    </div>
  </transition>
</template>

<script>
import PptPreview from './PptPreview.vue';
import OfdPreview from './OfdPreview.vue';
import StreamMarkdown from './StreamMarkdown.vue';
import VueOfficeDocx from '@vue-office/docx';
import '@vue-office/docx/lib/index.css';
import { resDownloadFile, getFileType, getMimeType } from '@/utils/util';
import * as XLSX from 'xlsx';

export default {
  name: 'FilePreviewDrawer',
  components: {
    PptPreview,
    OfdPreview,
    StreamMarkdown,
    VueOfficeDocx,
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    // 文件名（用于显示标题、判断类型等）
    fileName: {
      type: String,
      default: '',
    },
    // 只接收 blob 数据
    blob: {
      type: Blob,
      default: null,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    panelStyle: {
      type: Object,
      default: () => ({}),
    },
    // 是否显示关闭按钮
    showClose: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      activeSheetIndex: 0,
      // 内部预览状态
      previewType: '',
      previewUrl: '',
      previewContent: '',
      previewBlobUrl: '',
      previewExcelData: null,
      internalBlob: null,
    };
  },
  computed: {
    fileExt() {
      if (!this.fileName) return '';
      return this.fileName.split('.').pop().toLowerCase();
    },
    fencedCode() {
      if (!this.previewContent) return '';
      return '```' + this.fileExt + '\n' + this.previewContent + '\n```';
    },
  },
  watch: {
    visible(val) {
      if (val && this.blob) {
        // 打开且有 blob 数据时，处理预览
        this.processBlob();
      } else if (!val) {
        // 关闭时清理资源
        this.cleanupPreview();
      }
    },
    blob(newVal) {
      // blob 变化时重新处理
      if (newVal && this.visible) {
        this.processBlob();
      }
    },
  },
  methods: {
    // 处理 blob 数据
    async processBlob() {
      if (!this.blob || !this.fileName) {
        return;
      }

      // 重置状态
      this.previewType = '';
      this.previewUrl = '';
      this.previewContent = '';
      this.previewBlobUrl = '';
      this.previewExcelData = null;
      this.activeSheetIndex = 0;

      try {
        this.previewType = getFileType(this.fileName);
        // 根据文件扩展名自动设置正确的 MIME 类型，确保浏览器能正确预览而非下载
        const mimeType = getMimeType(this.fileExt);
        // 使用内部副本，避免直接修改 prop
        this.internalBlob = mimeType
          ? new Blob([this.blob], { type: mimeType })
          : this.blob;

        if (
          ['image', 'video', 'audio', 'pdf', 'html'].includes(this.previewType)
        ) {
          this.previewBlobUrl = URL.createObjectURL(this.internalBlob);
          this.previewUrl = this.previewBlobUrl;
        } else if (['markdown', 'text'].includes(this.previewType)) {
          this.previewContent = this.decodeTextBuffer(
            await this.internalBlob.arrayBuffer(),
          );
        } else if (this.previewType === 'excel') {
          const arrayBuffer = await this.internalBlob.arrayBuffer();
          // CSV 在中文 Windows 下常为 GBK 编码，按 UTF-8 解析会乱码；
          // xlsx/xls 内部已声明编码，走原路径即可
          const workbook =
            this.fileExt === 'csv'
              ? XLSX.read(this.decodeTextBuffer(arrayBuffer), {
                  type: 'string',
                })
              : XLSX.read(arrayBuffer, { type: 'array' });
          this.previewExcelData = workbook.SheetNames.map(sheetName => {
            const sheet = workbook.Sheets[sheetName];
            const jsonData = XLSX.utils.sheet_to_json(sheet, {
              header: 1,
              defval: '',
            });
            const merges = sheet['!merges'] || [];
            return {
              name: sheetName,
              data: jsonData,
              merges: merges.map(m => ({
                sr: m.s.r,
                sc: m.s.c,
                er: m.e.r,
                ec: m.e.c,
              })),
              colCount: Math.max(
                1,
                sheet['!ref']
                  ? XLSX.utils.decode_range(sheet['!ref']).e.c + 1
                  : 1,
              ),
            };
          });
        }
      } catch (error) {
        console.error('处理文件失败:', error);
        this.$message.error(this.$t('generalAgent.filePreview.processFailed'));
        this.previewType = 'unsupported';
      }
    },

    // 清理预览资源
    cleanupPreview() {
      if (this.previewBlobUrl && typeof this.previewBlobUrl === 'string') {
        URL.revokeObjectURL(this.previewBlobUrl);
        this.previewBlobUrl = '';
      }
    },

    handleClose() {
      this.$emit('update:visible', false);
      this.$emit('close');
    },

    handleError() {
      this.$message.error(this.$t('generalAgent.filePreview.previewFailed'));
    },

    openInNewTab() {
      if (this.previewUrl) {
        window.open(this.previewUrl, '_blank');
      }
    },

    // 文本类编码探测：UTF-8（含 BOM）优先，失败回退 GBK，兼容中文 Windows 导出
    decodeTextBuffer(arrayBuffer) {
      const bytes = new Uint8Array(arrayBuffer);
      let utf8 = '';
      try {
        utf8 = new TextDecoder('utf-8', { fatal: true }).decode(bytes);
      } catch (e) {
        return new TextDecoder('gbk').decode(bytes);
      }
      return utf8.replace(/^﻿/, '');
    },

    isMergedCell(sheetIndex, row, col) {
      if (!this.previewExcelData || !this.previewExcelData[sheetIndex])
        return false;
      const merges = this.previewExcelData[sheetIndex].merges || [];
      return merges.some(m => row === m.sr && col === m.sc);
    },

    isHiddenByMerge(sheetIndex, row, col) {
      if (!this.previewExcelData || !this.previewExcelData[sheetIndex])
        return false;
      const merges = this.previewExcelData[sheetIndex].merges || [];
      return merges.some(
        m =>
          row >= m.sr &&
          row <= m.er &&
          col >= m.sc &&
          col <= m.ec &&
          !(row === m.sr && col === m.sc),
      );
    },

    getRowspan(sheetIndex, row, col) {
      if (!this.previewExcelData || !this.previewExcelData[sheetIndex])
        return 1;
      const merges = this.previewExcelData[sheetIndex].merges || [];
      const merge = merges.find(m => row === m.sr && col === m.sc);
      return merge ? merge.er - merge.sr + 1 : 1;
    },

    getColspan(sheetIndex, row, col) {
      if (!this.previewExcelData || !this.previewExcelData[sheetIndex])
        return 1;
      const merges = this.previewExcelData[sheetIndex].merges || [];
      const merge = merges.find(m => row === m.sr && col === m.sc);
      return merge ? merge.ec - merge.sc + 1 : 1;
    },

    // 下载文件
    async handleDownload() {
      if (!this.fileName || !this.internalBlob) {
        return;
      }

      try {
        resDownloadFile(this.internalBlob, this.fileName);
        this.$message.success(
          this.$t('generalAgent.workspace.downloadSuccess'),
        );
      } catch (error) {
        console.error('下载文件失败:', error);
        this.$message.error(this.$t('generalAgent.workspace.downloadFailed'));
      }
    },

    handleWordError(error) {
      console.error('Word 文档渲染失败:', error);
      this.$message.error(
        this.$t('generalAgent.filePreview.wordPreviewFailed'),
      );
    },
  },
};
</script>

<style scoped>
.preview-panel {
  position: relative;
  height: 100%;
  width: 100%;
  background: #fff;
  border-left: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  z-index: 10;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
  flex-shrink: 0;
}

.preview-title {
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: 12px;
}

.preview-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.close-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  cursor: pointer;
  color: #606266;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f0f2f5;
  color: #10a37f;
}

.preview-body {
  flex: 1;
  overflow: auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

/* 覆盖 markdown.scss 中影响布局的伪元素样式 */
.preview-body ::v-deep .markdown-body::before,
.preview-body ::v-deep .markdown-body::after {
  display: none !important;
}

/* 所有预览容器统一设置 */
.preview-body > div {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

/* 滑入滑出动画 - 从左向右收起，从右向左展开 */
.preview-slide-enter-active,
.preview-slide-leave-active {
  transition: all 0.3s ease;
}

/* 进入时：从右边滑入（初始位置在 workspace 下方，不可见） */
.preview-slide-enter {
  transform: translateX(100%);
  opacity: 0;
}

/* 离开时：滑出到右边 */
.preview-slide-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

.preview-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.preview-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-height: 0;
  color: #909399;
  font-size: 14px;
}

.preview-loading i {
  font-size: 32px;
  margin-bottom: 12px;
  color: #10a37f;
}

.preview-image-wrapper {
  background: #f5f7fa;
  padding: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.preview-video-wrapper {
  background: #000;
  padding: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-video {
  max-width: 100%;
  max-height: 100%;
}

.preview-audio-wrapper {
  gap: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-excel-wrapper {
  display: flex;
  flex-direction: column;
  background: #fff;
}

.excel-tabs {
  display: flex;
  gap: 4px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
  overflow-x: auto;
}

.excel-tab {
  padding: 6px 16px;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.excel-tab:hover {
  border-color: #10a37f;
  color: #10a37f;
}

.excel-tab.active {
  background: #10a37f;
  border-color: #10a37f;
  color: #fff;
}

.excel-table-container {
  flex: 1;
  overflow: auto;
  min-height: 0;
}

.excel-table {
  border-collapse: collapse;
  width: max-content;
  min-width: 100%;
  font-size: 13px;
}

.excel-table td {
  border: 1px solid #e4e7ed;
  padding: 6px 10px;
  text-align: left;
  white-space: nowrap;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.excel-table td.excel-header {
  background: #f5f7fa;
  font-weight: 600;
  color: #303133;
  position: sticky;
  top: 0;
  z-index: 1;
}

.excel-table td.merged-cell {
  text-align: center;
  vertical-align: middle;
}

.audio-cover {
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
}

.audio-cover i {
  font-size: 48px;
}

.preview-audio {
  width: 100%;
  max-width: 500px;
}

.preview-pdf-wrapper,
.preview-ppt-wrapper,
.preview-ofd-wrapper,
.preview-html-wrapper {
  overflow: hidden;
  width: 100%;
}

.preview-pdf,
.preview-html-frame {
  width: 100%;
  height: 100%;
  border: none;
  flex: 1;
}

.preview-markdown-wrapper {
  overflow: auto;
  padding: 20px;
  background: #fff;
}

.preview-office-wrapper {
  flex: 1;
  min-height: 0;
  overflow: auto;
  background: #fff;
  width: 100%;

  ::v-deep .vue-office-docx {
    height: 100%;
    overflow-y: auto;
  }

  ::v-deep .docx-wrapper {
    background: #fff;
    padding: 0;
  }

  ::v-deep .docx-wrapper > section.docx {
    box-shadow: none;
    margin-bottom: 0;
    width: auto !important;
    min-width: 100% !important;
    max-width: 100% !important;
  }
}

.preview-unsupported {
  align-items: center;
  justify-content: center;
  color: #909399;
}

.preview-unsupported i {
  font-size: 48px;
  margin-bottom: 16px;
  color: #c0c4cc;
}

.preview-unsupported .file-name {
  font-size: 16px;
  color: #606266;
  margin-bottom: 8px;
}

.preview-unsupported .notice-text {
  font-size: 14px;
  margin-bottom: 16px;
}

.preview-toolbar {
  padding: 12px 20px;
  border-top: 1px solid #e4e7ed;
  background: #fafafa;
  display: flex;
  gap: 8px;
}
</style>

<style lang="scss" scoped>
// 代码文件预览样式
.preview-text-wrapper {
  background: #0d0d0d;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;

  ::v-deep .stream-markdown-container {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  ::v-deep .markdown-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;

    &::before,
    &::after {
      display: none !important;
    }

    pre.code-block {
      flex: 1;
      min-height: 0;
      margin: 0;
      display: flex;
      flex-direction: column;

      code {
        flex: 1;
        min-height: 0;
        display: flex;
        flex-direction: column;
      }

      .code-content {
        flex: 1;
        min-height: 0;
        overflow: auto;
      }
    }
  }
}

.preview-markdown-wrapper {
  ::v-deep .stream-markdown-container {
    height: 100%;
  }
}
</style>
