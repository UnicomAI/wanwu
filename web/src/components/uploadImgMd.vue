<template>
  <div
    class="rich-textarea-wrapper"
    :class="{
      'read-only': isReadOnly,
      'has-placeholder': showPlaceholder,
    }"
  >
    <div
      class="rich-textarea"
      ref="editorRef"
      :contenteditable="!isReadOnly"
      @compositionstart="onCompositionStart"
      @compositionend="onCompositionEnd"
      @focus="onFocus"
      @blur="onBlur"
      @input="onInput"
      @click="$refs.imagePreview.handleImageClick($event)"
      v-html="localContent"
    ></div>

    <!-- placeholder 文字 -->
    <div
      v-if="showPlaceholder && !isReadOnly"
      class="placeholder-text"
      @click="focusEditor"
    >
      {{ placeholder }}
    </div>

    <!-- 只在非只读模式下显示上传按钮 -->
    <el-upload
      v-if="!isReadOnly"
      action=""
      :multiple="false"
      :accept="acceptType"
      :auto-upload="false"
      :show-file-list="false"
      :on-change="uploadOnChange"
    >
      <i class="el-icon-picture-outline"></i>
    </el-upload>

    <!-- 图片预览组件 -->
    <ImagePreview ref="imagePreview" />
  </div>
</template>

<script>
import { Md2Img, Img2Md } from '@/utils/util';
import { POWER_TYPE_EDIT, POWER_TYPE_READ } from '@/views/knowledge/constants';
import { getDocLimit } from '@/api/knowledge';
import { uploadFileMD } from '@/api/chunkFile';
import ImagePreview from './ImagePreview.vue';

export default {
  components: {
    ImagePreview,
  },
  props: {
    value: {
      type: String,
      default: '',
    },
    permissionType: {
      type: Number,
      default: POWER_TYPE_EDIT,
    },
    knowledgeId: {
      type: String,
      default: '',
    },
    placeholder: {
      type: String,
      default() {
        return this.$t('common.input.placeholder');
      },
    },
  },
  data() {
    return {
      localValue: this.value, // value是MD格式
      localContent: Md2Img(this.value), // content是HTML格式
      savedRange: null,
      savedCharacterOffset: 0,
      isComposing: false, // 用于处理中文输入法等复杂输入场景
      acceptType: '.png,.jpg,.jpeg',
      maxSize: 3,
    };
  },
  computed: {
    isReadOnly() {
      return [POWER_TYPE_READ].includes(this.permissionType);
    },
    currentValue() {
      return this.value;
    },
    showPlaceholder() {
      // 当没有内容且不是只读模式时显示placeholder
      return !this.localValue && !this.isReadOnly;
    },
  },
  watch: {
    knowledgeId: {
      handler(newVal) {
        if (newVal) {
          getDocLimit({ knowledgeId: newVal }).then(res => {
            if (res.code === 0) {
              this.acceptType =
                '.' +
                res.data.uploadLimitList
                  .find(item => item.acceptType === 'image')
                  .flatMap(item => item.extList || [])
                  .join(',.');
              this.maxSize = res.data.uploadLimitList.find(
                item => item.fileType === 'image',
              ).maxSize;
            }
          });
        }
      },
    },
    currentValue(newVal) {
      if (newVal !== this.localValue) {
        this.localValue = newVal;
        this.localContent = Md2Img(newVal);
        this.$refs.editorRef.innerHTML = this.localContent;
      }
    },
  },
  methods: {
    focusEditor() {
      if (this.isReadOnly) return;
      this.$refs.editorRef.focus();
    },
    uploadOnChange(file) {
      if (file) {
        if (file.size / 1024 / 1024 > this.maxSize) {
          this.$message.error(
            this.$t('knowledgeManage.multiKnowledgeDatabase.imageSizeLimit', {
              maxSize: this.maxSize,
            }),
          );
          return;
        }
        const formData = new FormData();
        const config = { headers: { 'Content-Type': 'multipart/form-data' } };
        formData.append('files', file.raw);
        formData.append('markdown', true);
        uploadFileMD(formData, config).then(res => {
          if (res.code === 0) {
            this.insertImage(res.data.fileList[0].fileUrl);
          }
        });
      }
    },

    onCompositionStart() {
      // 开始输入法输入，标记状态
      this.isComposing = true;
    },

    onCompositionEnd(event) {
      // 输入法输入结束
      if (this.isComposing) {
        this.isComposing = false;
        // 部分浏览器/输入法下，合成结束前的最终 input 带 isComposing 被 onInput 跳过，这里补处理一次
        this.onInput(event);
      }
    },

    onFocus() {
      if (this.isReadOnly) return;
      this.saveCaretPosition();
    },

    onBlur() {
      if (this.isReadOnly) return;
      this.saveCaretPosition();
    },

    onInput(event) {
      if (this.isReadOnly) return;
      // 对于复杂输入（如中文输入法），我们不立即保存，而是等待输入完成
      if (this.isComposing) return;

      const currentHTML = event.target.innerHTML;
      this.localValue = Img2Md(currentHTML);
      // 保存光标位置，因为内容变化后光标位置可能改变
      this.saveCaretPosition();
      this.$emit('input', this.localValue);
    },

    getCharacterOffsetFromRange(range) {
      if (!range) return null;

      const preSelectionRange = range.cloneRange();
      preSelectionRange.selectNodeContents(this.$refs.editorRef);
      preSelectionRange.setEnd(range.startContainer, range.startOffset);

      // 获取选区内的HTML内容
      const container = preSelectionRange.cloneContents();
      let offset = 0;

      // 递归遍历所有节点，包括嵌套的节点
      function traverseNodes(node) {
        if (node.nodeType === Node.TEXT_NODE) {
          // 文本节点：直接计算文本长度
          offset += node.textContent.length;
        } else if (node.nodeType === Node.ELEMENT_NODE) {
          // 元素节点：+1（表示一个节点）
          offset += 1;
          // 递归处理子节点
          for (const childNode of node.childNodes) {
            traverseNodes(childNode);
          }
        }
      }

      // 遍历容器内的所有节点
      for (const childNode of container.childNodes) {
        traverseNodes(childNode);
      }

      return offset;
    },

    restoreCaretPositionByOffset(targetOffset) {
      if (!targetOffset || targetOffset < 0) return;

      const editor = this.$refs.editorRef;
      if (!editor) return;

      const range = document.createRange();
      const selection = getSelection();
      let currentOffset = 0;

      // 递归遍历所有节点，包括嵌套的节点
      function traverseNodes(node) {
        if (node.nodeType === Node.TEXT_NODE) {
          const textLength = node.textContent.length;
          if (currentOffset + textLength >= targetOffset) {
            // 光标在当前文本节点内
            const positionInText = targetOffset - currentOffset;
            range.setStart(node, positionInText);
            range.collapse(true);
            selection.removeAllRanges();
            selection.addRange(range);
            this.savedRange = range.cloneRange();
            return true;
          }
          currentOffset += textLength;
        } else if (node.nodeType === Node.ELEMENT_NODE) {
          // 元素节点：+1（表示一个节点）
          if (currentOffset + 1 >= targetOffset) {
            // 光标在元素节点后面
            range.setStartAfter(node);
            range.collapse(true);
            selection.removeAllRanges();
            selection.addRange(range);
            this.savedRange = range.cloneRange();
            return true;
          }
          currentOffset += 1;
          // 递归处理子节点
          for (const childNode of node.childNodes) {
            if (traverseNodes.call(this, childNode)) {
              return true;
            }
          }
        }
        return false;
      }

      // 遍历编辑器的所有子节点
      for (const childNode of editor.childNodes) {
        if (traverseNodes.call(this, childNode)) {
          return;
        }
      }

      // 如果没找到精确位置，将光标移到末尾
      range.selectNodeContents(editor);
      range.collapse(false);
      selection.removeAllRanges();
      selection.addRange(range);
      this.savedRange = range.cloneRange();
    },

    saveCaretPosition() {
      const selection = getSelection();
      if (selection && selection.rangeCount > 0) {
        this.savedRange = selection.getRangeAt(0).cloneRange();
        this.savedCharacterOffset = this.getCharacterOffsetFromRange(
          this.savedRange,
        );
      }
    },

    insertImage(mdImageLink) {
      // 使用之前保存的光标位置，如果没有则添加到末尾
      if (this.savedRange) {
        const range = this.savedRange;

        // 删除当前选中内容
        range.deleteContents();

        // 创建一个临时节点包含 Markdown 链接
        const textNode = document.createTextNode(mdImageLink);
        range.insertNode(textNode);

        // 更新光标位置
        const newRange = document.createRange();
        newRange.setStartAfter(textNode);
        newRange.collapse(true);

        this.savedRange = newRange.cloneRange();

        // 触发内容更新
        const newHTML = this.$refs.editorRef.innerHTML;
        this.localContent = Md2Img(newHTML, false);
        this.localValue = Img2Md(newHTML, false);
        this.$emit('input', this.localValue);

        if (this.savedCharacterOffset !== null) {
          this.$nextTick(() => {
            const newOffset = this.savedCharacterOffset + 1;
            this.restoreCaretPositionByOffset(newOffset);
          });
        }
      } else {
        // 如果没有保存的光标位置，直接添加到末尾
        this.localValue += mdImageLink;
        this.localContent = Img2Md(this.localValue, false);
        this.$emit('input', this.localValue);
      }
    },
  },
};
</script>

<style scoped>
.rich-textarea-wrapper {
  min-height: 40px;
  max-height: 150px;
  padding: 8px 12px;
  text-align: start;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  color: #606266;
  background-color: #fff;
  font-family: inherit;
  font-size: inherit;
  line-height: 1.5;
  outline: none;
  resize: vertical;
  overflow-y: auto;
  position: relative;

  &.read-only {
    background-color: #f5f7fa;
    border-color: #e4e7ed;
    color: #c0c4cc;
    cursor: not-allowed;
  }

  ::v-deep img {
    width: auto;
    max-height: 115px;
    cursor: zoom-in;
  }

  .rich-textarea {
    outline: none;
    min-height: 24px;
  }

  &:hover,
  &:focus {
    border: 1px solid var(--color);
  }
}

.placeholder-text {
  position: absolute;
  top: 8px;
  left: 12px;
  color: #c0c4cc;
  pointer-events: none;
  font-style: italic;
  user-select: none;
}

.rich-textarea-wrapper.has-placeholder:hover .placeholder-text,
.rich-textarea-wrapper.has-placeholder:focus .placeholder-text {
  color: #a8abb2;
}
</style>
