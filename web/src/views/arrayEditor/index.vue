<template>
  <div :ref="id" class="coder_editor" />
</template>

<script>
// 引入JavaScript支持
// import * as monaco from "monaco-editor/esm/vs/editor/editor.api";
import * as monaco from "monaco-editor";

export default {
  props: {
    id:{
        type:String,
        default:'arrayEditor'
    },
    value: {
      type: String,
      default: "",
    },
    language: {
      type: String,
      default: "",
    },
    theme: {
      type: String,
      default: "",
    },
    n:{
        type: Number,
         default: "",
    }
  },
  data() {
    return {
      monacoEditor: null, // 语言Edit器,
      monacoEditorConfig: {
        automaticLayout: true, // 自动Layout
        theme: "vs", // 官方自带三种主题vs, hc-black, or vs-dark
        tabSize: 0, // tab 缩进Length
        autoIndent: "None", // 控制Edit器在UserKey入、Paste、移动OR缩进行 when YesNo应自动调整缩进
        minimap: {
          enabled: false, // Close小地图
        },
        readOnly: false,
        lineNumbers: "on", // 隐藏控制行号
        autoClosingBrackets: true,
        formatOnPaste: true, //YesNoPaste自动Format
      },
    };
  },
  watch: {
    value(val) {
    //   this.monacoEditor.setValue(val);
    },
  },
  mounted() {
    this.$nextTick(()=>{
        this.init();
    })
    // method setWorkerUrl
  },
  methods: {
    init() {
      if (this.$refs[this.id]) {
        // InitEdit器，确保dom已经Render
        const config = Object.assign({}, this.monacoEditorConfig, {
          language: this.language,
          value: this.value,
        });
        this.monacoEditor = monaco.editor.create(
          this.$refs[this.id],
          config
        );
        //this.monacoEditor.editor.remeasureFonts();
        // Edit器BindEvent
        this.monacoEditorBindEvent();
      }
    },
    // DestroyEdit器
    monacoEditorDispose() {
      this.monacoEditor && this.monacoEditor.dispose();
    },
    // GetEdit器 of Value
    getCodeVal() {
      const content = this.monacoEditor && this.monacoEditor.getValue();
      if (!content) {
        this.$message.error("Cannot be empty, 提交Failed");
      }
      return content;
    },
    // Edit器Event
    monacoEditorBindEvent() {
      if (this.monacoEditor) {
        // 实 when GetEdit器 of Value
        this.monacoEditor.onDidChangeModelContent(() => {
          this.$emit("handleChange", this.monacoEditor.getValue(), this.n);
        });
      }
    },
  },
};
</script>

<style lang="scss">
.coder_editor {
  position: relative;
  width: 100%;
  height: 100%;
  .read {
    &::after {
      content: "";
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      left: 68px;
      z-index: 1;
    }
  }
}
</style>

