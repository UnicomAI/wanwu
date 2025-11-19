<template>
  <div class="session rl">
    <div
      class="history-box showScroll"
      id="timeScroll"
      v-loading="loading"
      ref="timeScroll"
    >
      <div
        v-for="(n,i) in session_data.history"
        :key="`${i}sdhs`"
      >
        <!--问题-->
        <div
          v-if="n.query"
          class="session-question"
        >
          <div :class="['session-item','rl']">
            <img
              class="logo"
              :src="userAvatarSrc"
            />
            <div class="answer-content">
              <div class="answer-content-query">
                <span
                  class="session-setting-id"
                  v-if="$route.params && $route.params.id && (type && type !=='webChat')"
                >AgentID: {{$route.params.id}}</span>
                <el-popover
                  placement="bottom-start"
                  trigger="hover"
                  :visible-arrow="false"
                  popper-class="query-copy-popover"
                  content=""
                >
                  <p
                    class="query-copy"
                    @click="queryCopy(n.query)"
                    style="cursor: pointer"
                  ><i class="el-icon-s-order"></i>&nbsp;{{$t('agent.copyToInput')}}</p>
                  <span
                    slot="reference"
                    class="answer-text"
                    style="display:inline-block;margin-top:5px;"
                  >{{n.query}}</span>
                </el-popover>
                <div
                  class="echo-doc-box"
                  v-if="hasFiles(n)"
                >
                <el-button v-show="canScroll(i,n.showScrollBtn)" icon="el-icon-arrow-left " @click="prev($event,i)" circle class="scroll-btn left" size="mini" type="primary"></el-button>
                 <div class="imgList" :ref="`imgList-${i}`">
                 <div v-for="(file,j) in n.fileList" :key="`${j}sdsl`" class="docInfo-img-container">
                      <img v-if="hasImgs(n,file)"
                        :src="file.fileUrl"
                        class="docIcon imgIcon" />
                  <div v-else class="docInfo-container">
                  <img
                    :src="require('@/assets/imgs/fileicon.png')"
                    class="docIcon"
                    style="width:30px!important;"
                  />
                  <div class="docInfo">
                    <p class="docInfo_name">FileName：{{file.name}}</p>
                    <p class="docInfo_size">File大小:{{getFileSizeDisplay(file.size)}}</p>
                  </div>
                  </div>
                 </div>
                 </div>
                 <el-button v-show="canScroll(i,n.showScrollBtn)" icon="el-icon-arrow-right" @click="next($event,i)" circle class="scroll-btn right" size="mini" type="primary"></el-button>
                </div>
              </div>

              <!-- <div v-if="n.requestFileUrls && n.requestFileUrls.length" class="file-path query-file">
                <el-image :src='n.requestFileUrls[0]' :preview-src-list="[n.requestFileUrls[0]]"></el-image>
              </div> -->
            </div>
          </div>
        </div>
        <!--loading-->
        <div
          v-if="n.responseLoading"
          class="session-answer"
        >
          <div class="session-answer-wrapper">
            <img
              class="logo"
              :src="'/user/api/'+ defaultUrl"
            />
            <div class="answer-content"><i class="el-icon-loading"></i></div>
          </div>
        </div>
        <!--pending-->
        <div
          v-if="n.pendingResponse"
          class="session-answer"
        >
          <div class="session-answer-wrapper">
            <img
              class="logo"
              :src="'/user/api/'+ defaultUrl"
            />
            <div
              class="answer-content"
              style="padding:10px;color:#E6A23C;"
            >{{n.pendingResponse}}</div>
          </div>
        </div>

        <!-- 回答故障  code:7-->
        <div
          class="session-error"
          v-if="n.error"
        ><i class="el-icon-warning"></i>&nbsp;{{n.response}}</div>

        <!--回答 文字+图片-->
        <div
          v-if="(n.response && !n.error)"
          class="session-answer"
          :id="'message-container'+i"
        >
          <div
            v-if="[0,1,2,3,4,6,20,21,10].includes(n.qa_type)"
            class="session-answer-wrapper"
          >
            <img
              class="logo"
              :src="'/user/api/'+ defaultUrl"
            />
            <div
              class="session-wrap"
              style="width:calc(100% - 30px);"
            >
              <div
                v-if="showDSBtn(n.response)"
                class="deepseek"
                @click="toggle($event,i)"
              >
                <template v-if="n.qa_type === 20">
                  <img
                    :src="require('@/assets/imgs/tool-icon.png')"
                    class="think_icon"
                  />{{n.toolText}}
                </template>
                <template v-else>
                  <img
                    :src="require('@/assets/imgs/think-icon.png')"
                    class="think_icon"
                  />{{n.thinkText}}
                </template>
                <i v-bind:class="{'el-icon-arrow-down': !n.isOpen,'el-icon-arrow-up': n.isOpen}"></i>
              </div>
              <div
                class="answer-content"
                v-bind:class="{'ds-res':showDSBtn(n.response)}"
                v-html="showDSBtn(n.response)?replaceHTML(n.response,n):n.response"
              ></div>
            </div>
          </div>
          <div
            v-else
            class="session-answer-wrapper"
          >
            <img
              class="logo"
              :src="'/user/api/'+ defaultUrl"
            />
            <div
              v-if="n.code === 7"
              class="answer-content session-error"
            ><i class="el-icon-warning"></i> &nbsp;{{n.response}}</div>
            <div
              v-else
              class="answer-content"
              v-html="(n.response)"
            ></div>
          </div>
          <!--File-->
          <div
            v-if="n.gen_file_url_list && n.gen_file_url_list.length"
            class="file-path response-file"
          >
            <el-image
              v-for="(g,k) in n.gen_file_url_list"
              :key="k"
              :src='g'
              :preview-src-list="[g]"
            ></el-image>
          </div>
          <!--出处-->
          <div
            v-if="n.searchList && n.searchList.length && n.finish === 1"
            class="search-list"
          >
            <div
              v-for="(m,j) in n.searchList"
              :key="`${j}sdsl`"
              class="search-list-item"
            >
              <div
                class="serach-list-item"
                v-if="showSearchList(j,n.qa_type,n.citations)"
              >
                <span @click="collapseClick(n,m,j)"><i :class="['',m.collapse?'el-icon-caret-bottom':'el-icon-caret-right']"></i>出处：</span>
                <a
                  v-if="m.link"
                  :href="m.link"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="link"
                >{{m.link}}</a>
                <span v-if="m.title">
                  <sub
                    class="subTag"
                    :data-parents-index="i"
                    :data-collapse="m.collapse?'true':'false'"
                    v-if="n.qa_type === 1"
                  >{{j + 1}}</sub> {{m.title}}
                </span>
                <!-- <span @click="goPreview($event,m)" class="search-doc">查看全文</span> -->
              </div>
              <el-collapse-transition>
                <div
                  v-show="m.collapse?true:false"
                  class="snippet"
                >
                  <p v-html="m.snippet"></p>
                </div>
              </el-collapse-transition>
            </div>
          </div>
          <!--loading-->
          <div
            v-if="n.finish === 0 && sessionStatus == 0 && i === session_data.history.length - 1"
            class="text-loading"
          >
            <div></div>
            <div></div>
            <div></div>
          </div>
          <!--停止Generate 重新Generate 点赞   session code Yes0 when 不可Operation-->
          <div class="answer-operation">
            <div class="opera-left">
              <span
                v-if="i===(session_data.history.length-1)"
                class="restart"
                @click="refresh"
              >
                <!-- <i class="el-icon-refresh" @click="refresh">&nbsp;{{$t('agent.refresh')}}</i> -->
                <img :src="require('@/assets/imgs/refresh-icon.png')" />
                <!-- &nbsp;{{$t('agent.refresh')}} -->
              </span>
            </div>
            <div
              class="opera-right"
              style="flex:0;"
              @click="()=>{copy(n.oriResponse) && copycb()}"
            >
              <img :src="require('@/assets/imgs/copy-icon.png')" />
              <!-- <i class="el-icon-copy-document copy-icon" style="padding: 0 6px;margin: 0;" :title="$t('agent.clickCopy')" @click="()=>{copy(n.oriResponse) && copycb()}"></i> -->
            </div>
          </div>
        </div>

        <!-- 回答 仅图片-->
        <div
          v-if="!n.response && n.gen_file_url_list && n.gen_file_url_list.length"
          class="session-answer"
        >
          <div class="session-answer-wrapper">
            <img
              class="logo"
              :src="'/user/api/'+ defaultUrl"
            />
            <div class="answer-content">
              <div
                v-if="n.gen_file_url_list && n.gen_file_url_list.length"
                class="file-path response-file no-response"
              >
                <el-image
                  v-for="(g,k) in n.gen_file_url_list"
                  :key="k"
                  :src='g'
                  :preview-src-list="[g]"
                ></el-image>
              </div>
            </div>
          </div>
          <!--仅图片 when 只有 重新Generate-->
          <div class="answer-operation">
            <div class="opera-left">
              <span
                v-if="i===(session_data.history.length-1)"
                class="restart"
              ><i
                  class="el-icon-refresh"
                  @click="refresh"
                >&nbsp;{{$t('agent.refresh')}}</i></span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import smoothscroll from "smoothscroll-polyfill";
import { md } from "@/mixins/marksown-it";
import { marked } from "marked";
var highlight = require("highlight.js");
import "highlight.js/styles/atom-one-dark.css";
import commonMixin from "@/mixins/common";
import { mapGetters } from "vuex";

marked.setOptions({
  renderer: new marked.Renderer(),
  gfm: true,
  tables: true,
  breaks: false,
  pedantic: false,
  sanitize: false,
  smartLists: true,
  smartypants: false,
  highlight: function (code) {
    return highlight.highlightAuto(code).value;
  },
});

export default {
  mixins: [commonMixin],
  props: ["sessionStatus", "defaultUrl", "type"],
  data() {
    return {
      md: md,
      autoScroll: true,
      scrollTimeout: null,
      isDs:
        [
          "txt2txt-002-001",
          "txt2txt-002-002",
          "txt2txt-002-004",
          "txt2txt-002-005",
          "txt2txt-002-006",
          "txt2txt-002-007",
          "txt2txt-002-008",
        ].indexOf(this.$route.params.id) != -1,
      loading: false,
      marked: marked,
      session_data: {
        tool: "",
        searchList: [],
        history: [],
        response: "",
      },
      basePath: this.$basePath,
      current_data: [],
      //标注相关
      c: null,
      ctx: null,
      canvasShow: false,
      cv: null,
      currImg: {
        url: "",
        width: 0, // 原始宽高
        height: 0,
        w: 0, // Compress后 of 宽高
        h: 358,
        roteX: 0, // Compress后 of 比例
        roteY: 0,
      },
      imgConfig: ["jpeg", "PNG", "png", "JPG", "jpg", "bmp", "webp"],
      audioConfig: ["mp3", "wav"],
      fileScrollStateMap: {},
      resizeTimer: null,
    };
  },
  computed: {
    ...mapGetters('user', ['userAvatar']),
    userAvatarSrc(){
      return this.userAvatar 
      ? '/user/api/' + this.userAvatar 
      : require('@/assets/imgs/robot-icon.png');
    }
  },
  watch: {
    sessionStatus: {
      handler(val, oldVal) {},
      immediate: true,
    },
    'session_data.history':{
      handler(){
        this.$nextTick(() => {
          this.updateAllFileScrollStates();
        });
      },
      deep:true
    }
  },
  mounted() {
    this.setupScrollListener();
    smoothscroll.polyfill();
    document.addEventListener('click', this.handleCitationClick);
    window.addEventListener('resize', this.handleWindowResize);
    this.updateAllFileScrollStates();
  },
  beforeDestroy() {
    if(this.handleCitationClick) {
      document.removeEventListener('click', this.handleCitationClick);
    }
    const container = document.getElementById("timeScroll");
    if (container) {
      container.removeEventListener("scroll", this.handleScroll);
    }
    clearTimeout(this.scrollTimeout);

    window.removeEventListener('resize', this.handleWindowResize);
    if (this.resizeTimer) {
      clearTimeout(this.resizeTimer);
    }
    
    // RemoveImageErrorEventListen器
    if (this.imageErrorHandler) {
      document.body.removeEventListener("error", this.imageErrorHandler, true);
    }
  },
  methods: {
    updateAllFileScrollStates() {
      this.session_data.history.forEach((item, index) => {
        if (item.fileList && item.fileList.length > 0) {
          this.$nextTick(() => {
            this.checkFileScrollState(index);
          });
        }
      });
    },
    checkFileScrollState(index) {
      const refKey = `imgList-${index}`;
      const containerArray = this.$refs[refKey];
      if (containerArray && containerArray.length > 0) {
        const container = containerArray[0];
        const canScroll = container.scrollWidth > container.clientWidth;
        if (this.session_data.history[index]) {
          this.$set(this.session_data.history[index], 'showScrollBtn', canScroll);
        }
        this.$set(this.fileScrollStateMap, index, canScroll);
      }
    },
    handleWindowResize() {
      if (this.resizeTimer) {
        clearTimeout(this.resizeTimer);
      }
      this.resizeTimer = setTimeout(() => {
        this.updateAllFileScrollStates();
      }, 200);
    },
    canScroll(i,showScrollBtn) {
      if (showScrollBtn !== null && showScrollBtn !== undefined) {
        return showScrollBtn;
      }
      // NoThen从 fileScrollStateMap 中Get
      return this.fileScrollStateMap[i] || false;
    },
    prev(e,i){
      e.stopPropagation()
      const refKey = `imgList-${i}`;
      const containerArray = this.$refs[refKey];
      if (containerArray && containerArray.length > 0) {
        const container = containerArray[0];
        container.scrollBy({
          left: -200,
          behavior: "smooth",
        });
      }
    },
    next(e,i){
      e.stopPropagation()
      const refKey = `imgList-${i}`;
      const containerArray = this.$refs[refKey];
      if (containerArray && containerArray.length > 0) {
        const container = containerArray[0];
        container.scrollBy({
          left: 200,
          behavior: "smooth",
        });
      }
    },
    hasFiles(n){
       return n.fileList && n.fileList.length > 0;
    },
    hasImgs(n,file){
      if (!n.fileList || n.fileList.length === 0 || !file || !file.name) {
        return false;
      }
      let type = file.name.split('.').pop().toLowerCase();
      return this.imgConfig.map(t => t.toLowerCase()).includes(type);
    },
    handleCitationClick(e) {
      this.$handleCitationClick(e, {
        sessionStatus: this.sessionStatus,
        sessionData: this.session_data,
        citationSelector: '.citation',
        scrollElementId: 'timeScroll',
        onToggleCollapse: (item, collapse) => {
          this.$set(item, 'collapse', collapse);
        }
      });
    },
    showSearchList(j, qa_type, citations) {
      return qa_type === 1 ? citations.includes(j + 1) : true;
    },
    setCitations(index) {
      let citation = `#message-container${index} .citation`;
      const allCitations = document.querySelectorAll(citation);
      const citationsSet = new Set();
      allCitations.forEach((element) => {
        const text = element.textContent.trim();
        if (text) {
          citationsSet.add(Number(text));
        }
      });

      return Array.from(citationsSet);
    },
    goPreview(event, item) {
      event.stopPropagation(); // 阻止Event冒泡
      let { meta_data } = item;
      let { file_name, download_link, page_num, row_num, sheet_name } =
        meta_data;
      var index = file_name.lastIndexOf(".");
      var ext = file_name.substr(index + 1);
      let openUrl = "";
      let fileUrl = encodeURIComponent(download_link);
      const fileType = ["docx", "doc", "txt", "pdf", "xlsx"];
      if (fileType.includes(ext)) {
        switch (ext) {
          case "docx" || "doc":
            openUrl = `${window.location.origin}/doc?fileUrl=` + fileUrl;
            break;
          case "txt":
            openUrl = `${window.location.origin}/txtView?fileUrl=` + fileUrl;
            break;
          case "pdf":
            if (page_num.length > 0) {
              openUrl =
                `${window.location.origin}/pdfView?fileUrl=` +
                fileUrl +
                "&page=" +
                page_num[0];
            }
            break;
          case "xlsx":
            openUrl =
              `${window.location.origin}/jsExcel?url=` +
              fileUrl +
              "&rownum=" +
              row_num +
              "&sheetName=" +
              sheet_name;
            break;
          default:
            this.$message.warning("暂Not Supported此格式查看");
        }
      }
      if (openUrl !== "") {
        window.open(openUrl, "_blank","noopener,noreferrer");
      } else {
        this.$message.warning("暂Not Supported此格式查看");
      }
    },
    setupScrollListener() {
      const container = document.getElementById("timeScroll");
      container.addEventListener("scroll", this.handleScroll);
    },
    handleScroll(e) {
      const container = document.getElementById("timeScroll");
      const { scrollTop, clientHeight, scrollHeight } = container;
      // 检测YesNo接近Bottom（5px容差）
      const nearBottom = scrollHeight - (scrollTop + clientHeight) < 5;
      // User手动滚动 when Cancel自动置底
      if (!nearBottom) {
        this.autoScroll = false;
      }
      // Clear之前 of 定 when 器
      clearTimeout(this.scrollTimeout);
      // Setting新 of 定 when 器检测滚动Stop
      this.scrollTimeout = setTimeout(() => {
        // IfStop when 接近Bottom，Restore自动置底
        if (nearBottom) {
          this.autoScroll = true;
          this.scrollBottom();
        }
      }, 500); // 500ms内No新滚动视为Stop
    },
    replaceHTML(data, n) {
      let _data = data;
      const thinkStart = /<think>/i;
      const thinkEnd = /<\/think>/i;
      const toolStart = /<tool>/i;
      const toolEnd = /<\/tool>/i;

      // Process think Tag
      if (thinkEnd.test(data)) {
        n.thinkText = "已深度思考";
        if (!thinkStart.test(data)) {
          data = "<think>\n" + data;
        }
      }

      // AddProcess tool Tag
      if (toolEnd.test(data)) {
        n.toolText = "已使用Tool"; // NeedAdd对应 of 翻译
        if (!toolStart.test(data)) {
          data = "<tool>\n" + data;
        }
      }
      // 统一Replace为 section Tag
      return data
        .replace(/think>/gi, "section>")
        .replace(/tool>/gi, "section>");
    },
    showDSBtn(data) {
      // const pattern = /<\/?think>/;
      const pattern = /<(think|tool)(\s[^>]*)?>|<\/(think|tool)>/;
      const matches = data.match(pattern);
      if (!matches) {
        return false;
      }
      return true;
    },
    toggle(event, index) {
      const name = event.target.className;
      if (
        name === "deepseek" ||
        name === "el-icon-arrow-up" ||
        name === "el-icon-arrow-down"
      ) {
        this.session_data.history[index].isOpen =
          !this.session_data.history[index].isOpen;
        this.$set(
          this.session_data.history,
          index,
          this.session_data.history[index]
        );
        let elm = null;
        if (name === "el-icon-arrow-up" || name === "el-icon-arrow-down") {
          elm = event.target.parentNode.parentNode
            .getElementsByClassName("answer-content")[0]
            .getElementsByTagName("section")[0];
        } else {
          elm = event.target.parentNode
            .getElementsByClassName("answer-content")[0]
            .getElementsByTagName("section")[0];
        }
        if (!Boolean(this.session_data.history[index].isOpen)) {
          elm.className = "hideDs";
        } else {
          elm.className = "";
        }
      }
    },
    queryCopy(text) {
      this.$emit("queryCopy", text);
    },
    copy(text) {
      text = text.replaceAll("<br/>", "\n");
      var textareaEl = document.createElement("textarea");
      textareaEl.setAttribute("readonly", "readonly"); // 防止手机上弹出软Key盘
      textareaEl.value = text;
      document.body.appendChild(textareaEl);
      textareaEl.select();
      var res = document.execCommand("copy");
      document.body.removeChild(textareaEl);
      return res;
    },
    copycb() {
      this.$message.success(this.$t("agent.copyTips"));
    },
    collapseClick(n, m, j) {
      if (!m.collapse) {
        this.$set(n.searchList, j, { ...m, collapse: true });
      } else {
        this.$set(n.searchList, j, { ...m, collapse: false });
      }
      //this.scrollBottom()
    },
    doLoading() {
      this.loading = true;
    },
    scrollBottom() {
      this.loading = false;
      if (!this.autoScroll) return;
      this.$nextTick(() => {
        document.getElementById("timeScroll").scrollTop =
          document.getElementById("timeScroll").scrollHeight;
      });
    },
    codeScrollBottom() {
      this.$nextTick(() => {
        this.loading = false;
        document.getElementsByTagName("code").scrollTop =
          document.getElementsByTagName("code").scrollHeight;
      });
    },
    pushHistory(data) {
      this.session_data.history.push(data);
      this.scrollBottom();
    },
    replaceLastData(index, data) {
      if (!data.response) {
        data.response = "无响应Data";
      }
      this.$set(this.session_data.history, index, data);
      this.scrollBottom();
      this.codeScrollBottom(); //codeContent置底
      if (data.finish === 1) {
        const setCitations = this.setCitations(index);
        this.$set(this.session_data.history[index], "citations", setCitations);
      }
    },
    getFileSizeDisplay(fileSize) {
      if (!fileSize || typeof fileSize !== "number" || isNaN(fileSize)) {
        return "...";
      }
      return fileSize > 1024
        ? `${(fileSize / (1024 * 1024)).toFixed(2)} MB`
        : `${fileSize} bytes`;
    },
    //websocket Replace全部Data
    replaceData(data) {
      this.session_data = data;
      this.scrollBottom();
    },
    //http 只Replacehistory
    replaceHistory(data) {
      this.session_data.history = data;
      this.scrollBottom();
      //this.loadAllImg()
    },
    replaceHistoryWithImg(data) {
      this.session_data.history = data;
      this.$nextTick(() => {
        this.preTagging(data[0].annotation);
      });
    },
    clearData() {
      this.session_data = {
        tool: "",
        searchList: [],
        history: [],
        response: "",
      };
    },
    loadAllImg() {
      this.session_data.history.forEach((n, i) => {
        n.gen_file_url_list.forEach((m, j) => {
          setTimeout(() => {
            this.$set(this.session_data.history[i].gen_file_url_list, j, {
              ...m,
              loadedUrl: m.url,
              loading: false,
            });
          }, 2000);
        });
      });
    },
    gropdownClick() {
      this.$emit("clearHistory");
    },
    getList() {
      return JSON.parse(
        JSON.stringify(
          this.session_data.history.filter((item) => {
            delete item.operation;
            return item;
          })
        )
      );
      // return JSON.parse(JSON.stringify(this.session_data.history.filter((item)=>{ delete item.operation ; return !item.pending})))
    },
    getAllList() {
      return JSON.parse(JSON.stringify(this.session_data.history));
    },
    stopLoading() {
      this.session_data.history = this.session_data.history.filter((item) => {
        return !item.pending;
      });
    },
    stopPending() {
      this.session_data.history = this.session_data.history.filter((item) => {
        if (item.pending) {
          item.responseLoading = false;
          item.pendingResponse = "This response has been terminated";
        }
        return item;
      });
    },
    refresh() {
      if (this.sessionStatus === 0) {
        return;
      }
      this.$emit("refresh");
    },
    preZan(index, item) {
      if (this.sessionStatus === 0) {
        return;
      }
      this.$set(this.session_data.history, index, { ...item, evaluate: 1 });
    },
    preCai(index, item) {
      if (this.sessionStatus === 0) {
        return;
      }
      this.$set(this.session_data.history, index, { ...item, evaluate: 2 });
    },
    //=================标注相关===============
    initCanvasUtil() {
      this.canvasShow = true;
      this.$nextTick(() => {
        // 开始画图 canvas, 2d, 宽高，形状
        this.cv &&
          this.cv.destroy() &&
          this.cv.clearPre() &&
          this.cv.clearLabels() &&
          (this.cv = null);
        this.cv = new CanvasUtil(this);
      });
    },
    preTagging(response) {
      // canvas大小Reset
      this.currImg = {
        url: "",
        width: 0,
        height: 0,
        w: 0,
        h: 358,
        roteX: 0,
        roteY: 0,
        dx: 0,
        dy: 0,
      };
      // Image原始宽高
      var image = new Image();
      image.src = response.annotationImg;
      image.onload = () => {
        this.currImg.width = image.width;
        this.currImg.height = image.height;
        //if (!this.c) {
        this.c = document.getElementById("mycanvas");
        this.ctx = this.c.getContext("2d");
        //}
        this.resizeCanvas();
        this.initCanvasUtil();

        this.$nextTick(() => {
          this.echoLabels(response);
        });
      };
    },
    echoLabels(response) {
      this.cv.echoLabels(response);
    },
    resizeCanvas() {
      this.currImg.w = 0;
      this.currImg.h = 358;
      this.currImg.dx = 0;
      this.currImg.dy = 0;
      this.currImg.roteX = 0;
      this.currImg.roteY = 0;

      let currImg = this.currImg;
      let contain = document.getElementById("mycantain");
      if (currImg.width > contain.offsetWidth) {
        // 宽度Greater Than容器
        this.currImg.roteX = currImg.width / contain.offsetWidth;
        currImg.w = contain.offsetWidth;
        currImg.h = (currImg.height * contain.offsetWidth) / currImg.width;
        // Compress后高度Greater Thancantain
        if (currImg.h > contain.offsetHeight) {
          currImg.h = contain.offsetHeight;
          currImg.w = (currImg.width * currImg.h) / currImg.height;
          currImg.roteX = currImg.width / currImg.w;
          currImg.dx = (contain.offsetWidth - currImg.w) / 2;
        } else {
          currImg.roteY = currImg.height / currImg.h;
          currImg.dy = (contain.offsetHeight - currImg.h) / 2;
        }
      } else {
        // 高度Compress比例
        currImg.roteY = currImg.height / currImg.h;
        // Compress后宽度
        currImg.w = (currImg.width * currImg.h) / currImg.height;
        currImg.roteX = currImg.width / currImg.w;
        currImg.dx = (contain.offsetWidth - currImg.w) / 2;
      }

      this.canvasShow = true;
      this.c.width = currImg.w;
      this.c.height = currImg.h;
      this.$nextTick(() => {
        this.cv && this.cv.resizeCurrImg(currImg);
      });
    },
    listenerImg() {
      //捕获ImageLoad Error
      this.imageErrorHandler = (e) => {
        if (e.target.tagName === "IMG") {
          this.handleImageError(e.target);
        }
      };
      document.body.addEventListener("error", this.imageErrorHandler, true);
    },
    handleImageError(img) {
      // 防止重复Process
      if (img.classList.contains("failed")) {
        return;
      }
      img.classList.add("failed");

      // SettingImage为不可见，避免闪烁
      img.style.visibility = "hidden";
      img.style.display = "none";
    },
  },
};
</script>

<style scoped lang="scss">
.serach-list-item {
  .link:hover {
    color: $color !important;
  }
  .search-doc {
    margin-left: 10px;
    cursor: pointer;
    color: $color !important;
  }
  .subTag {
    display: inline-flex;
    color: $color;
    border-radius: 50%;
    width: 18px;
    height: 18px;
    border: 1px solid $color;
    line-height: 18px;
    vertical-align: middle;
    margin-left: 2px;
    justify-content: center;
    align-items: center;
    font-size: 14px;
    overflow: hidden;
    white-space: nowrap;
    margin-bottom: 2px;
    transform: scale(0.8);
  }
}

/deep/ {
  pre {
    white-space: pre-wrap !important;
    min-height: 50px;
    word-wrap: break-word;
    resize: vertical;
    .hljs {
      max-height: 300px !important;
      white-space: pre-wrap !important;
      min-height: 50px;
      word-wrap: break-word;
      resize: vertical;
    }
    code {
      display: block;
      white-space: pre-wrap;
      word-break: break-all;
      scroll-behavior: smooth;
    }
  }
  .el-loading-mask {
    background: none !important;
  }
  .answer-content {
    width: 100%;
    img {
      width: 80% !important;
    }
    section li,li {
      list-style-position: inside !important; /* 将标记Symbol放在Content框内 */
    }
   
    .citation {
      display: inline-flex;
      color: $color;
      border-radius: 50%;
      width: 18px;
      height: 18px;
      border: 1px solid $color;
      cursor: pointer;
      line-height: 18px;
      vertical-align: middle;
      margin-left: 5px;
      justify-content: center;
      align-items: center;
      font-size: 14px;
      overflow: hidden;
      white-space: nowrap;
      margin-bottom: 2px;
      transform: scale(0.8);
    }
  }
  .search-list {
    img {
      width: 80% !important;
    }
  }
}
.more {
  color: $color;
}
.session {
  word-break: break-all;
  height: 100%;
  overflow-y: auto;
  .session-item {
    min-height: 80px;
    display: flex;
    // justify-content:flex-end;
    padding: 20px;
    line-height: 28px;
    img {
      width: 30px;
      height: 30px;
      object-fit: cover;
    }
    .logo {
      border-radius: 6px;
    }
    .answer-content {
      padding: 0 10px 10px 15px;
      position: relative;
      color: #333;
      .answer-content-query {
        display: flex;
        flex-wrap: wrap;
        flex-direction: column;
        align-items: flex-end;
        width: 100%;
        .answer-text {
          background: #7288fa;
          color: #fff;
          padding: 8px 10px 8px 20px;
          border-radius: 10px 0 10px 10px;
          margin:0!important;
          line-height:1.5;
        }
        .session-setting-id {
          color: rgba(98, 98, 98, 0.5);
          font-size: 12px;
          margin-top: -8px;
        }
        .echo-doc-box {
          margin-top: 10px;
          width: 100%;
          max-width: 100%;
          display: flex;
          gap:8px;
          justify-content: space-between;
          align-items: center;
          position: relative;
          .scroll-btn{
            position:absolute;
            top:50%;
            transform: translateY(-15px);
            &.left{
                left:5px;
            }
            &.right{
                right:5px;
            }
          }
          .imgList{
            width:100%;
            gap: 10px;
            overflow-x:hidden;
            scroll-behavior: smooth;
            display:flex;
            flex-wrap: nowrap;
            flex-direction: row-reverse;
          }
          .docInfo-container{
            display: flex;
            align-items: center;
            background: #fff;
            border: 1px solid rgb(235, 236, 238);
            padding: 5px 10px 5px 5px;
            border-radius: 5px;
          }
          .docInfo-img-container{
            flex-shrink: 0;  /* 防止Image被Compress */
            width: auto;  /* OR固定宽度 */
            p{
              text-align: center;
              color: $color;
              font-size: 12px;
            }
          }
          .docIcon {
            width: 30px;
            height: 30px;
          }
          .imgIcon {
            width:auto!important;
            height:70px!important;
            display:block;
            border-radius:6px;
          }
          .docInfo {
            .docInfo_name {
              color: #333;
            }
            .docInfo_size {
              color: #bbbbbb;
            }
          }
        }
      }
      li {
        display: revert !important;
      }
    }
  }
  .session-answer {
    // background-color: #eceefe;
    border-radius: 10px;
    .answer-annotation {
      line-height: 0 !important;
      .annotation-img {
        width: 460px;
        object-fit: contain;
        height: 358px;
      }
      .tagging-canvas {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        margin: auto;
      }
    }

    .no-response {
      margin: 15px 0;
    }
    /*出处*/
    .search-list {
      padding: 10px 20px 3px 54px;
      .search-list-item {
        margin-bottom: 5px;
        line-height: 22px;
        p:nth-child(1) {
          white-space: normal;
        }
        a,
        span {
          color: #666;
          cursor: pointer;
          white-space: normal;
          overflow-wrap: break-word;
        }
        a {
          text-decoration: underline;
        }
        a:hover {
          color: deepskyblue;
        }
        .snippet {
          padding: 5px 14px;
        }
      }
    }
    /*Operation*/
    .answer-operation {
      display: flex;
      // justify-content: space-between;
      align-items: center;
      padding: 5px 20px 15px 63px;
      color: #777;
      .opera-left {
        // flex: 8;
        .restart {
          cursor: pointer;
          img {
            width: 20px;
            height: 20px;
            padding: 2px;
          }
        }
      }
      .opera-right {
        // flex: 1;
        cursor: pointer;
        display: inline-flex;
        padding-left: 10px;
        img {
          width: 20px;
          height: 20px;
          padding: 2px;
        }
        .split-icon {
          background: rgba(195, 197, 217, 0.65);
          height: 22px;
          margin: 0 10px;
          width: 1px;
        }
        .copy-icon {
          font-size: 17px;
          padding: 3px 6px;
          margin: 0 15px;
          cursor: pointer;
        }
        .copy-icon:hover {
          color: #33a4df;
        }
      }
    }
  }

  /*Image*/
  .file-path {
    .el-image {
      height: 200px !important;
      background-color: #f9f9f9;
      /deep/.el-image__inner,
      img {
        width: 100%;
        height: 100%;
        object-fit: contain;
      }
    }
    audio {
      width: 300px !important;
    }
  }
  .query-file {
    padding: 10px 0;
  }
  .response-file {
    margin: 0 0 0 66px;
    width: 400px;
    font-size: 0;
    .img {
      display: inline-block;
      width: 200px;
      height: 200px;
      img {
        width: 100%;
        height: 100%;
      }
    }
  }

  .session-error {
    background-color: #fef0f0;
    border-color: #fde2e2;
    color: #f56c6c !important;
    margin-top: 10px;
    padding: 10px;
    border-radius: 4px;
    .el-icon-warning {
      font-size: 16px;
    }
  }

  .history-box {
    height: calc(100% - 46px);
    overflow-y: auto;
    padding: 20px;
  }
  /*Delete历史...*/
  .session-setting {
    position: relative;
    height: 36px;
    .right-setting {
      position: absolute;
      right: 10px;
      top: -5px;
      color: #ff2324;
      font-size: 16px;
      cursor: pointer;
      /deep/ {
        .el-dropdown-menu {
          width: 100px;
        }
        .el-dropdown-menu__item {
          padding: 0 15px !important;
        }
      }
    }
  }

  .think_icon {
    width: 12px !important;
    height: 12px !important;
    margin-right: 3px;
  }
  .ds-res {
    /deep/ section {
      color: #8b8b8b;
      position: relative;
      font-size: 12px;
      * {
        font-size: 12px;
      }
    }
    /deep/ section::before {
      content: "";
      position: absolute;
      height: 100%;
      width: 1px;
      background: #ddd;
      left: -8px;
    }
    /deep/.hideDs {
      display: none;
    }
  }

  .deepseek {
    font-size: 13px;
    color: #8b8b8b;
    font-weight: bold;
    margin:0 0 10px 6px;
    cursor: pointer;
  }
}

/* 仅By样式调整位置：
   问题在右侧（Content在右、头像在最右），答案在左侧（默认） */
.session-question {
  .session-item {
    flex-direction: row-reverse;
    margin-left: auto;
    width: auto;
  }
}
.session-answer {
  .session-answer-wrapper {
    display: flex;
    align-items: flex-start;
    gap: 10px; /* AvatarAndContent之间10px距离 */
    padding: 20px 20px 0 20px;
    min-height: 80px;
    background: none; /* 确保外层容器No背景色 */
    
    .logo {
      width: 30px;
      height: 30px;
      border-radius: 6px;
      object-fit: cover;
      flex-shrink: 0; /* 防止Avatar被Compress */
      background: none; /* AvatarNo背景色 */
    }
    
    .answer-content {
      flex: 1;
      background-color: #eceefe; /* 只HasContentArea域Has背景色 */
      border-radius: 0 10px 10px 10px;
      padding: 20px;
      line-height:1.6;
    }
  }
}

/* ImageLoadWhen Failed of Style */
img.failed {
  position: relative;
  border: 2px dashed #ff6b6b;
  background-color: #fff5f5;
  opacity: 0.5;
}

img.failed::after {
  content: "图片加载Failed";
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: #ff6b6b;
  font-size: 12px;
  background: rgba(255, 255, 255, 0.9);
  padding: 4px 8px;
  border-radius: 4px;
  white-space: nowrap;
}

.text-loading,
.text-loading > div {
  position: relative;
  box-sizing: border-box;
}

.text-loading {
  display: block;
  font-size: 0;
  color: #c8c8c8;
}

.text-loading.la-dark {
  color: #e8e8e8;
}

.text-loading > div {
  display: inline-block;
  float: none;
  background-color: currentColor;
  border: 0 solid currentColor;
}

.text-loading {
  width: 54px;
  height: 18px;
  margin: 6px 0 0 55px;
}

.text-loading > div {
  width: 8px;
  height: 8px;
  margin: 4px;
  border-radius: 100%;
  animation: ball-beat 0.7s -0.15s infinite linear;
}

.text-loading > div:nth-child(2n-1) {
  animation-delay: -0.5s;
}

@keyframes ball-beat {
  50% {
    opacity: 0.2;
    transform: scale(0.75);
  }

  100% {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
