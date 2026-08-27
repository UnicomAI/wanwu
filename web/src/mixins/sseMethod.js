import { fetchEventSource } from '../sse/index.js';
import { store } from '@/store/index';
import Print from '../utils/printPlus2.js';
import {
  parseSub,
  convertLatexSyntax,
  parseSubConversation,
} from '@/utils/util.js';
import { mapActions, mapGetters } from 'vuex';
import { i18n } from '@/lang';
import StreamProcessor from '@/utils/streamProcessor.js';
import { guid, getXClientId, formatReqUrl } from '@/utils/util';

var originalFetch = window.fetch;

import { md } from './markdown-it';
import $ from './jquery.min.js';
import { OPENURL_API, USER_API } from '@/utils/requestConstants';
import { AGENT_MESSAGE_CONFIG } from '@/components/stream/constants';
import { processToolResultBlocks } from '@/utils/toolResultProcessor.js';
import { cancelAgentStream, cancelOpenurlAgentStream } from '@/api/agent';
import { cancelRagStream } from '@/api/rag';
import { cancelModelExperienceStream } from '@/api/modelExprience';

const AGENT_API_URL = `${USER_API}/assistant/stream`;
const AGENT_CONNECT_API_URL = `${USER_API}/assistant/stream/connect`;
const AGENT_DRAFT_CONNECT_API_URL = `${USER_API}/assistant/draft/stream/connect`;
const RAG_API_URL = `${USER_API}/rag/chat`;
const RAG_CONNECT_API_URL = `${USER_API}/rag/stream/connect`;
const EXPRIENCE_API_URL = `${USER_API}/model/experience/llm`;
const EXPRIENCE_CONNECT_API_URL = `${USER_API}/model/experience/llm/connect`;

export default {
  data() {
    return {
      isTestChat: false,
      defaultUrl: '/v1/static/icon/model-default-icon.png', // 模型体验默认URL
      inputVal: '',
      eventSource: null,
      ctrlAbort: null,
      sseParams: {},
      sseResponse: {},
      echo: true,
      conversationId: '', //会话id
      chatList: [],
      reminderList: [],
      queryFilePath: '',
      stopBtShow: false,
      origin: window.location.origin,
      reconnectCount: 0,
      isEnd: true,
      sseApi: AGENT_API_URL,
      rag_sseApi: RAG_API_URL,
      exprience_sseApi: EXPRIENCE_API_URL,
      lastIndex: 0,
      query: '',
      isStoped: false,
      access_token: '',
      runResponse: '',
      fileList: [], // 文件列表
      fileInfoList: [], // 上传后的文件信息（fileId, fileName, fileUrl等）
      instanceSessionStatus: -1,
      sessionComRef: null,
      activeAgentStreamParams: {
        assistantId: '',
        conversationId: '',
        draft: false,
      },
      _subConversionsMap: null, // 子会话存储 Map
      _subConversionProcessors: null, // 子会话处理器 Map
      _subMainProcessorsMap: null, // 子会话内部正文片段处理器 Map (Key: subId_order)
      responseFiles: [], // 用于存储 SSE 返回的附件文件列表

      // ---- 推理内容（reasoning_content）流处理相关 ----
      _isInReasoning: false, // 客户端侧：推理打字动画是否仍在进行中
      _reasoningSSEDone: false, // 服务端侧：SSE 是否已停止推送 reasoning_content
      _pendingOutputQueue: [], // 正文缓冲队列：推理动画完毕前暂存所有正文帧
      _reasoningPrint: null, // 推理专用打字机实例（Print）
    };
  },
  created() {
    if (!this.isExplorePage()) {
      this.rag_sseApi = `${RAG_API_URL}/draft`;
    }
    const vuex = JSON.parse(localStorage.getItem('access_cert'));
    if (vuex) {
      this.access_token = vuex.user.token;
    }
  },
  mounted() {
    //this.addVisibilitychangeEvent()
  },
  beforeDestroy() {
    this.setStoreSessionStatus(-1);
    this.stopEventSource();
    this._print && this._print.stop();
  },
  computed: {
    ...mapGetters('app', ['sessionStatus']),
    ...mapGetters('user', ['token', 'userInfo']),
  },
  methods: {
    ...mapActions('app', ['setStoreSessionStatus']),
    isExplorePage() {
      return this.$route.path.includes('/explore/');
    },
    newFetch(url, options) {
      // 可以调用原始的 fetch 函数
      if (this.isStoped) {
        return;
      }
      return originalFetch(url, options)
        .then(response => {
          // 可以在这里修改响应或者添加额外的处理
          let query = this.query;

          if (response.status != 200) {
            let me = this;
            try {
              const stream = response.body;

              const reader = stream.getReader();
              const decoder = new TextDecoder('utf-8');

              function readStream() {
                reader
                  .read()
                  .then(({ done, value }) => {
                    if (done) {
                      console.log('Stream complete');
                      reader.releaseLock();
                      return;
                    }

                    // Decode and process each chunk of data.
                    const decodedValue = decoder.decode(value, {
                      stream: true,
                    });

                    if (decodedValue) {
                      let msg = JSON.parse(decodedValue).msg;
                      me.setStoreSessionStatus(-1);
                      var fillData = {
                        query: query,
                        qa_type: 0,
                        finish: 1,
                        response: msg, //非代码文本使用自定义转换规则，不使用markdown,(markdown渲染会导致卡顿或样式丢失)
                        oriResponse: '',
                        searchList: [], //过滤包含yunyingshang文件的出处
                      };
                      this.runResponse = msg;
                    }
                    readStream();

                    // Continue reading the stream.
                  })
                  .catch(err => {
                    console.error('Reading stream failed1:', err);
                  });
              }

              readStream();
              me.isStoped = true;
            } catch (e) {
              console.error('Reading stream failed:', e);
            }
          }

          return response;
        })
        .catch(err => {
          this.$message.warning(i18n.t('sse.connectError'));
          this.isEnd = true;
          this.setStoreSessionStatus(-1);
          this.runDisabled = false;
        });
    },
    ...mapActions('app', ['setStoreSessionStatus']),
    queryCopy(text) {
      this.setPrompt(text);
    },
    /*过滤掉markdown中自定义的行号*/
    getContentInBraces(shtml) {
      let temp = document.createElement('div');
      temp.setAttribute('id', 'temp');
      temp.innerHTML = shtml;
      document.body.appendChild(temp);
      $(temp).find('.line-num').remove();
      return temp.innerText;
    },
    // 填充开场白
    setProloguePrompt(val) {
      // this.$refs['editable'].setPrompt(val)
      const editable =
        this.$refs.editable || (this.getEditableRef && this.getEditableRef());
      if (editable) {
        editable.setPrompt(val);
      }
      this.preSend();
    },
    //获取上传的文件
    getFileIdList() {
      const editable =
        this.$refs.editable || (this.getEditableRef && this.getEditableRef());
      let list = editable.getFileIdList();
      let fileIds = [];
      this.queryFilePath = '';
      if (list.length) {
        fileIds = list.map(n => {
          return n.fileId;
        });
        this.queryFilePath = list[0].url;
      }
      return fileIds.join(',');
    },
    mouseEnter(n) {
      n.hover = true;
    },
    mouseLeave(n) {
      n.hover = false;
    },
    setSessionStatus(status) {
      // this.setStoreSessionStatus(status)
      if (this.fieldId) {
        this.instanceSessionStatus = status;
      } else {
        this.setStoreSessionStatus(status);
      }
    },
    getCurrentSessionStatus() {
      return this.fieldId ? this.instanceSessionStatus : this.sessionStatus;
    },
    setSseParams(data) {
      // this.sseParams = data

      this.sseParams = data ? Object.assign({}, data) : {};
      if (data && data.sessionComRef) {
        this.sessionComRef = data.sessionComRef;
      }
    },
    // 转换会话类型
    convertConversionType(type) {
      const _map = Object.values(AGENT_MESSAGE_CONFIG).reduce((acc, item) => {
        acc[item.EVENT_TYPE] = item.CONVERSATION_TYPE;
        return acc;
      }, {});
      return _map[type];
    },

    fetchEventSource(url, params, options = {}) {
      const {
        onopen,
        onmessage,
        onclose = () => {
          console.log('===> eventSource onClose');
          this.setStoreSessionStatus(-1); //关闭后改变状态
          this.sseOnCloseCallBack();
        },
        onerror = e => {
          console.log(i18n.t('sse.connectError'));
          if (e.readyState === EventSource.CLOSED) {
            console.log('connection is closed');
          } else {
            console.warn('Error occured', e);
          }
          this.stopEventSource(); //前端主动关闭连接
          this.setStoreSessionStatus(-1); //关闭后改变状态
        },
        headers,
        signal,
        ...rest
      } = options;
      this.ctrlAbort = new AbortController();
      return new fetchEventSource(this.origin + formatReqUrl(url), {
        method: 'POST',
        headers: headers || {
          'Content-Type': 'application/json',
          Authorization: 'Bearer ' + this.token,
          'x-user-id': this.userInfo.uid,
          'x-org-id': this.userInfo.orgId,
        },
        signal: signal || this.ctrlAbort.signal,
        body: JSON.stringify(params),
        openWhenHidden: true,
        onopen: onopen,
        onmessage: onmessage,
        onclose: onclose,
        onerror: onerror,
        rest,
      });
    },

    /**
     * 初始化推理内容流处理器
     * 创建独立的 reasoningProcessor 和 _reasoningPrint，并初始化缓冲状态
     * @param {Object} options
     * @param {number} options.lastIndex - 当前会话索引
     * @param {Object} options.md - markdown-it 实例
     * @param {Function} options.parseSub - 引用解析函数
     * @param {Function} options.convertLatexSyntax - LaTeX 转换函数
     * @returns {StreamProcessor} 推理专用流处理器实例
     */
    _initReasoningStream({ lastIndex, md, parseSub, convertLatexSyntax }) {
      // 初始化状态标志位
      this._isInReasoning = false;
      this._reasoningSSEDone = false;
      this._pendingOutputQueue = [];

      const reasoningProcessor = new StreamProcessor({
        lastIndex,
        md: md,
        parseSub,
        convertLatexSyntax,
      });

      // 思考打字机：onPrintEnd 在队列暂时为空时就会触发（可能多次）
      // 只有服务端也确认完成推理推送后，这次清空才是真正结束
      this._reasoningPrint = new Print({
        onPrintEnd: () => {
          if (this._reasoningSSEDone) {
            this._flushPendingOutput();
          }
        },
      });

      return reasoningProcessor;
    },

    /**
     * 清空正文缓冲队列，将所有暂存的正文内容送入正文打字机
     * 由 _reasoningPrint.onPrintEnd 或检测到打字机空载时主动调用
     */
    _flushPendingOutput() {
      this._isInReasoning = false;
      if (this._pendingOutputQueue && this._pendingOutputQueue.length) {
        this._pendingOutputQueue.forEach(item => {
          this._print.print(item.sentence, item.commonData, item.cb);
        });
        this._pendingOutputQueue = [];
      }
    },

    /**
     * 对每个 SSE 消息帧进行推理/正文路由分发
     * 根据 _isInReasoning 状态决定正文数据走直通路径还是缓冲队列
     * @param {Object} options
     * @param {string} options.reasoning - 当前帧的推理内容
     * @param {string} options.output - 当前帧的正文内容
     * @param {number} options.finish - 当前帧的完成状态
     * @param {Object} options.commonData - 当前帧的公共数据
     * @param {Function} options.doRenderReasoning - 推理内容打字机回调
     * @param {Function} options.doRenderMain - 正文内容打字机回调
     */
    _dispatchReasoningOrOutput({
      reasoning,
      output,
      finish,
      commonData,
      doRenderReasoning,
      doRenderMain,
    }) {
      // 推理帧：首次出现时激活缓冲路径
      if (reasoning) {
        if (!this._isInReasoning) {
          this._isInReasoning = true;
        }
        this._reasoningPrint.print(
          { response: reasoning, finish },
          commonData,
          doRenderReasoning,
        );
      }

      // 正文帧（含 finish 结束帧）
      if (output || (!reasoning && [1, 2].includes(finish))) {
        const mainSentence = { response: output || '', finish };

        // 首次收到 output，服务端侧推理结束：立即冻结思考打字机，
        // 把未动画完的 reasoning 残余文本一次性灌入处理器（保证再展开
        // 看到的是完整静态内容），然后直通正文。这样用户不会再看到
        // 折叠卡片里继续滴字，也不会等思考打字机慢慢打完才出正文。
        if (output && !this._reasoningSSEDone) {
          this._reasoningSSEDone = true;
          if (this._isInReasoning && this._reasoningPrint) {
            const rp = this._reasoningPrint;
            let remaining = '';
            const curSent = rp.sentenceArr[rp.sIndex];
            if (curSent) {
              const curText = curSent.response || '';
              const typedIdx =
                rp.looper && typeof rp.looper.index === 'number'
                  ? Math.min(rp.looper.index, curText.length)
                  : 0;
              remaining += curText.slice(typedIdx);
            }
            for (let i = rp.sIndex + 1; i < rp.sentenceArr.length; i++) {
              remaining +=
                (rp.sentenceArr[i] && rp.sentenceArr[i].response) || '';
            }
            if (remaining) {
              doRenderReasoning(
                { world: remaining, finish: 0, isEnd: false },
                null,
              );
            }
            rp.stop();
            this._flushPendingOutput();
          }
        }

        if (this._isInReasoning) {
          // 思考动画还未完毕：正文进缓冲等待
          this._pendingOutputQueue.push({
            sentence: mainSentence,
            commonData,
            cb: doRenderMain,
          });
        } else {
          // 无思考内容或思考已完毕：直通送入正文打字机
          this._print.print(mainSentence, commonData, doRenderMain);
        }
      }
    },

    doragSend() {
      this.stopBtShow = true;
      this.isStoped = false;
      let _history = this.$refs['session-com'].getList();
      this.sendRagEventSource(this.inputVal, '', _history.length);
    },
    getRagStreamHeaders() {
      return {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + this.token,
        'x-user-id': this.userInfo.uid,
        'x-org-id': this.userInfo.orgId,
        'X-Client-ID': getXClientId(),
      };
    },
    /**
     * 连接 RAG 保持会话的全量重放流。
     * 调用方须已将进行中问答占位写入 lastIndex，避免重放内容重复追加。
     */
    connectRagEventSource({
      ragId,
      conversationId,
      prompt = '',
      fileList = [],
      lastIndex,
    } = {}) {
      if (!ragId || !conversationId || typeof lastIndex !== 'number') return;
      this.stopBtShow = true;
      this.isStoped = false;
      this.echo = false;
      this.fileList = fileList || [];
      this.setSseParams({ ragId, conversationId });
      this.sendRagEventSource(prompt, '', lastIndex, {
        streamApi: RAG_CONNECT_API_URL,
        streamData: { ragId, conversationId },
        clearInput: false,
        skipPushHistory: true,
        reconnect: true,
      });
    },
    sendEventStream(prompt, msgStr, lastIndex) {
      let sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) {
        console.warn('[sseMethod] session-com ref missing');
        return;
      }
      if (this.getCurrentSessionStatus() === 0) {
        this.$message.warning(i18n.t('sse.incompleteError'));
        return;
      }

      this.sseResponse = {};
      this.setStoreSessionStatus(0);
      this.clearInput();
      this._isInReasoning = false;

      let params = {
        query: prompt,
        pending: true,
        responseLoading: true,
        requestFileUrls: [],
        fileList: this.fileList,
        pendingResponse: '',
      };
      sessionCom.pushHistory(params);

      // 初始化流处理器
      const processor = new StreamProcessor({
        lastIndex,
        md,
        parseSub,
        convertLatexSyntax,
      });
      // 初始化推理流（reasoningProcessor 及相关缓冲状态由 _initReasoningStream 统一管理）
      const reasoningProcessor = this._initReasoningStream({
        lastIndex,
        md,
        parseSub,
        convertLatexSyntax,
      });

      this._print = new Print({
        onPrintEnd: () => {
          this.onMainPrintEnd && this.onMainPrintEnd();
        },
      });
      let history_list = sessionCom.getSessionData();
      const history =
        history_list['history'].length > 1
          ? history_list['history'][history_list['history'].length - 2][
              'history'
            ]
          : [];

      this.eventSource = this.fetchEventSource(
        this.rag_sseApi,
        { ...this.sseParams, history: history },
        {
          onopen: async e => {
            if (e.status !== 200) {
              try {
                const errorData = await e.json();
                let commonData = {
                  ...this.sseParams,
                  query: prompt,
                };
                let fillData = {
                  ...commonData,
                  response: errorData.msg,
                };
                sessionCom.replaceLastData(lastIndex, fillData);
              } catch (e) {
                const text = await e.text();
                this.$message.error(text || i18n.t('sse.error'));
              }

              this.stopEventSource();
              this.setStoreSessionStatus(-1);
            }
          },
          onmessage: e => {
            if (e && e.data) {
              let data;
              try {
                data = JSON.parse(e.data);
              } catch (error) {
                return; // 如果解析失败，直接返回，不处理这条消息
              }

              this.sseResponse = data;
              let commonData = {
                ...this.sseResponse,
                ...this.sseParams,
                query: prompt,
                fileList: this.fileList,
                response: '',
                filepath: data.file_url || '',
                requestFileUrls: '',
                gen_file_url_list: [],
                searchList:
                  data.data && data.data.searchList ? data.data.searchList : [],
                thinkText: i18n.t('sse.thinkingText'),
                isOpen: true,
                citations: [],
              };

              if (data.code === 0 || data.code === 1) {
                //finish 0：进行中  1：关闭   2:敏感词关闭
                const reasoning =
                  data.data && data.data.reasoning_content
                    ? data.data.reasoning_content
                    : '';
                const output =
                  data.data && data.data.output ? data.data.output : '';

                const doRender = (worldObj, search_list, field) => {
                  this.setStoreSessionStatus(0);
                  if (field === 'main') {
                    processor.updateSearchList(search_list);
                    processor.append(worldObj.world);
                  } else {
                    reasoningProcessor.updateSearchList(search_list);
                    reasoningProcessor.append(worldObj.world);
                  }

                  const renderResult = processor.getRenderResult();
                  const reasoningRenderResult =
                    reasoningProcessor.getRenderResult();

                  let fillData = {
                    ...commonData,
                    ...renderResult,
                    activeReasoning: reasoningRenderResult.activeResponse || '',
                    stableReasoningChunks:
                      reasoningRenderResult.stableChunks || [],
                    finish: worldObj.finish,
                    searchList: search_list
                      ? search_list.map(n => ({
                          ...n,
                          snippet: n.snippet ? md.render(n.snippet) : '',
                        }))
                      : commonData.searchList,
                  };

                  if (worldObj.finish === 2) {
                    fillData.response = this.$t('sse.sensitiveTips');
                    sessionCom.replaceLastData(lastIndex, fillData);
                    this.$nextTick(() => sessionCom.scrollBottom());
                    this.setStoreSessionStatus(-1);
                  } else {
                    sessionCom.replaceLastData(lastIndex, fillData);
                  }

                  if (worldObj.isEnd && worldObj.finish === 1) {
                    this.setStoreSessionStatus(-1);
                  }
                };

                this._dispatchReasoningOrOutput({
                  reasoning,
                  output,
                  finish: data.finish,
                  commonData,
                  doRenderReasoning: (worldObj, search_list) =>
                    doRender(worldObj, search_list, 'reasoning'),
                  doRenderMain: (worldObj, search_list) =>
                    doRender(worldObj, search_list, 'main'),
                });
              } else if (data.code === 7 || data.code === -1) {
                this.setStoreSessionStatus(-1);
                sessionCom.replaceLastData(lastIndex, {
                  ...commonData,
                  response: data.message,
                  error: true,
                });
              }
            }
          },
        },
      );
    },
    /**
     * sendRagEventSource — RAG 问答流式方法（AG-UI 协议版）
     *
     * 请求格式与原 sendEventStream 完全相同（{ ragId, question, fileInfo, history }），
     * 响应格式从旧 RAG JSON 改为 AG-UI 事件流，事件类型：
     *   RUN_STARTED / CUSTOM(rag_search_list) /
     *   REASONING_MESSAGE_START|CONTENT|END /
     *   TEXT_MESSAGE_START|CONTENT|END /
     *   RUN_FINISHED / RUN_ERROR
     *
     * RUN_ERROR 事件的 data.code 与后端 internal/bff-service/service/rag_chat.go 的
     * Rag 错误码常量一一对应，下表新增码时双端须同步修改。
     */
    sendRagEventSource(prompt, msgStr, lastIndex, options = {}) {
      // 与 internal/bff-service/service/rag_chat.go 的 EventNameRagSearchList 对应
      const CUSTOM_EVENT_SEARCH_LIST = 'rag_search_list';
      // 与 EventNameRagKnowledgeStart 对应：后端通知即将进入知识库检索，前端据此来创建"知识库检索"卡片
      const CUSTOM_EVENT_KNOWLEDGE_START = 'rag_knowledge_start';
      const CUSTOM_EVENT_KNOWLEDGE_END = 'rag_knowledge_end';
      // 与 EventNameRagQAStart 对应：后端通知即将进入问答库检索，前端据此来创建"问答库检索"卡片
      const CUSTOM_EVENT_QA_START = 'rag_qa_start';
      const CUSTOM_EVENT_QA_END = 'rag_qa_end';
      const CUSTOM_EVENT_QA_ERROR = 'rag_qa_error';
      // 与 EventNameRagQASearchList 对应：问答库检索结果（含命中/未命中，未命中 value=[]）
      const CUSTOM_EVENT_QA_SEARCH_LIST = 'rag_qa_search_list';
      // RAG RUN_ERROR code → vue-i18n key 映射表
      // 与 internal/bff-service/service/rag_chat.go 的 RagErrCode* 常量对应
      const RAG_ERROR_CODE_I18N = {
        sensitive_block: 'sse.sensitiveTips',
        upstream_error: 'sse.error',
        unknown_error: 'sse.error',
      };
      let sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) {
        console.warn('[sseMethod] session-com ref missing');
        return;
      }
      if (this.getCurrentSessionStatus() === 0) {
        this.$message.warning(i18n.t('sse.incompleteError'));
        return;
      }

      this.sseResponse = {};
      this.setStoreSessionStatus(0);
      if (options.clearInput !== false) {
        this.clearInput();
      }
      this._isInReasoning = false;

      // 正常提问创建占位；断线重连复用调用方预先补出的占位。
      if (!options.skipPushHistory) {
        sessionCom.pushHistory({
          query: prompt,
          pending: true,
          responseLoading: true,
          requestFileUrls: [],
          fileList: this.fileList,
          pendingResponse: '',
        });
      }

      // 初始化流处理器（主文本 + 推理）
      const processor = new StreamProcessor({
        lastIndex,
        md,
        parseSub,
        convertLatexSyntax,
      });
      const reasoningProcessor = this._initReasoningStream({
        lastIndex,
        md,
        parseSub,
        convertLatexSyntax,
      });
      this._print = new Print({
        onPrintEnd: () => {
          this.onMainPrintEnd && this.onMainPrintEnd();
        },
      });

      // 贯穿整个流的 searchList（KB 检索结果，由 rag_search_list 更新，初始为空）
      let currentSearchList = [];
      // 问答库检索结果（由 rag_qa_search_list 更新，即使未命中也会下发空数组）
      let currentQASearchList = [];

      // 是否有任何文字/推理内容到达（用于 RUN_FINISHED 兜底 setStoreSessionStatus）
      let streamHasContent = false;

      /**
       * ragSteps — 过程卡片数据，供 streamMessageField.vue 的 RagStepCard 渲染。
       * 三种步骤类型均为 lazy 创建（SSE 启动时不创建任何卡片）：
       *   - qa_search：收到 CUSTOM(rag_qa_start) 才创建；CUSTOM(rag_qa_search_list) 关闭
       *                （命中/未命中都会发，未命中 value=[]）。用户未配置问答库时后端不发 qa_start，卡片不出现。
       *   - knowledge_search：收到 CUSTOM(rag_knowledge_start) 才创建；CUSTOM(rag_search_list) 关闭
       *                未命中由首个 CONTENT 兜底关闭。用户未配置知识库 / QA 命中时后端不发该事件，卡片不出现。
       *   - thinking：首个 REASONING_MESSAGE_CONTENT 到达才创建；
       *                REASONING_MESSAGE_END 关闭（或 TEXT_MESSAGE_CONTENT 兜底）。
       * RUN_FINISHED / RUN_ERROR 时兜底关闭所有还 running 的步骤。
       * 每次修改后用 [...ragSteps] 触发 Vue 响应式（避免同引用 push 不更新）。
       */
      const ragSteps = [];
      const findStep = type => ragSteps.find(s => s.type === type);
      const resolveEventTimestamp = timestamp =>
        timestamp != null && Number.isFinite(Number(timestamp))
          ? Number(timestamp)
          : Date.now();
      const createStep = (type, startAt) => {
        const step = {
          type,
          status: 'running',
          startAt: resolveEventTimestamp(startAt),
          endAt: 0,
          duration: '',
        };
        ragSteps.push(step);
        return step;
      };
      const closeStep = (step, endAt) => {
        if (!step || !['running', 'error'].includes(step.status)) return;
        const isError = step.status === 'error';
        step.status = isError ? 'error' : 'done';
        step.endAt = resolveEventTimestamp(endAt);
        step.duration = `${((step.endAt - step.startAt) / 1000).toFixed(3)}s`;
      };
      // 未命中兜底：首个 CONTENT 到达时关闭所有检索卡片（qa_search / knowledge_search）
      const ensureSearchStepClosed = endAt => {
        ['qa_search', 'knowledge_search'].forEach(t => {
          const s = findStep(t);
          if (s && s.status === 'running') closeStep(s, endAt);
        });
      };
      // 错误/结束兜底：任何还 running 的步骤都关闭
      const closeAllRunning = endAt => {
        ragSteps.forEach(s => {
          if (['running', 'error'].includes(s.status)) closeStep(s, endAt);
        });
      };
      // knowledge_search 步骤改为懒创建：等后端 CUSTOM(rag_knowledge_start) 明确告知
      // "即将进入知识库检索"再建卡片。问答库命中场景后端不发该事件，卡片不出现。

      // 公共数据基础（response/searchList 由各事件处理器动态填充）
      const commonData = {
        ...this.sseParams,
        query: prompt,
        fileList: this.fileList,
        response: '',
        requestFileUrls: '',
        gen_file_url_list: [],
        searchList: [],
        thinkText: i18n.t('sse.thinkingText'),
        isOpen: true,
        citations: [],
      };

      // 初始渲染：进入 loading 态（ragSteps 初始为空，等后端信号决定是否建检索卡片）
      // finish:0 必须显式传：否则 replaceLastData 会把空 response 兜底成"无响应数据"
      sessionCom.replaceLastData(lastIndex, {
        ...commonData,
        responseLoading: true,
        finish: 0,
        ragSteps: [...ragSteps],
      });

      /**
       * doRender — 统一渲染回调，供 _dispatchReasoningOrOutput 内部调用
       * field: 'main' | 'reasoning'
       * _sl 参数来自 Print 回调，此处忽略，改用 currentSearchList 闭包值
       */
      const doRender = (worldObj, _sl, field) => {
        // 用户已主动停止：不要再把状态拉回 0，也不再继续渲染，
        // 否则停止按钮会被 reasoning 打字机的残余帧反复拉回显示。
        if (this.getCurrentSessionStatus() === -1) return;
        this.setStoreSessionStatus(0);
        if (field === 'main') {
          processor.updateSearchList(currentSearchList);
          processor.append(worldObj.world);
        } else {
          reasoningProcessor.updateSearchList(currentSearchList);
          reasoningProcessor.append(worldObj.world);
        }

        const renderResult = processor.getRenderResult();
        const reasoningRenderResult = reasoningProcessor.getRenderResult();

        const fillData = {
          ...commonData,
          ...renderResult,
          activeReasoning: reasoningRenderResult.activeResponse || '',
          stableReasoningChunks: reasoningRenderResult.stableChunks || [],
          finish: worldObj.finish,
          searchList: currentSearchList,
          qaSearchList: currentQASearchList,
          ragSteps: [...ragSteps],
        };

        sessionCom.replaceLastData(lastIndex, fillData);
        this.$nextTick(() => sessionCom.scrollBottom());

        if (worldObj.isEnd && worldObj.finish === 1) {
          this.setStoreSessionStatus(-1);
        }
      };

      this.eventSource = this.fetchEventSource(
        options.streamApi || this.rag_sseApi,
        options.streamData || this.sseParams,
        {
          headers: this.getRagStreamHeaders(),
          onopen: async e => {
            if (e.status !== 200) {
              // 克隆一份响应：e.json() 会消费 body，catch 里再对 e.text() 会抛
              // "body stream already read"；_e 是抛出的 Error，不是 Response，
              // 原来的 `_e.text()` 必然 TypeError。
              const errClone = e.clone();
              try {
                const errorData = await e.json();
                if (options.reconnect) {
                  const history = sessionCom.getSessionData().history || [];
                  if (history.length - 1 === lastIndex) {
                    sessionCom.removeLastHistory();
                    this.echo = history.length === 0;
                  }
                } else {
                  sessionCom.replaceLastData(lastIndex, {
                    ...commonData,
                    response: errorData.msg,
                  });
                }
              } catch (_e) {
                let text = '';
                try {
                  text = await errClone.text();
                } catch (_e2) {
                  // 兜底：body 读不出来就用 i18n 通用错误
                }
                if (!options.reconnect) {
                  this.$message.error(text || i18n.t('sse.error'));
                }
              }
              this.stopEventSource();
              this.setStoreSessionStatus(-1);
            }
          },
          onmessage: e => {
            if (!e?.data) return;
            let data;
            try {
              data = JSON.parse(e.data);
            } catch (_err) {
              return;
            }

            switch (data.type) {
              // ── 运行生命周期 ─────────────────────────────────────
              case 'RUN_STARTED':
                // 流已建立，状态已在 setStoreSessionStatus(0) 时设置，无需额外操作
                break;

              case 'RUN_FINISHED': {
                // 兜底：关闭所有还在 running 的过程卡片
                closeAllRunning(data.timestamp);

                // Fast-forward：后端已声明运行结束，把两个打字机队列里的剩余内容
                // 一次性灌进 processor，并停止动画。
                // 不这样做会导致：SSE 瞬间收完几万字，打字机按 ~30 char/s 慢慢播，
                //   - reasoning 卡在"还在打字"状态 → 用户看到思考卡片不动
                //   - output 被压在 pendingQueue，等 reasoning 打字机 drain 后才开始
                //   - 整体感知"后端明明返回完了，前端还在慢吞吞"
                // 该操作不影响正常流式体验（流未结束前不会走到这里）。
                //
                // 注：临时读 Print 私有字段 sentenceArr/sIndex/looper.index
                //     属于技术债（见 review P1-4），后续推动 Print 暴露 flush()
                //     公有方法后应重构。
                const drainPrint = printInstance => {
                  if (!printInstance) return '';
                  let remaining = '';
                  const curSent = printInstance.sentenceArr
                    ? printInstance.sentenceArr[printInstance.sIndex]
                    : null;
                  if (curSent) {
                    const curText = curSent.response || '';
                    const typedIdx =
                      printInstance.looper &&
                      typeof printInstance.looper.index === 'number'
                        ? Math.min(printInstance.looper.index, curText.length)
                        : 0;
                    remaining += curText.slice(typedIdx);
                  }
                  if (printInstance.sentenceArr) {
                    for (
                      let i = printInstance.sIndex + 1;
                      i < printInstance.sentenceArr.length;
                      i++
                    ) {
                      remaining +=
                        (printInstance.sentenceArr[i] &&
                          printInstance.sentenceArr[i].response) ||
                        '';
                    }
                  }
                  printInstance.stop();
                  return remaining;
                };

                // 1) reasoning 打字机剩余内容 → reasoningProcessor
                const remainingReasoning = drainPrint(this._reasoningPrint);
                if (remainingReasoning) {
                  reasoningProcessor.updateSearchList(currentSearchList);
                  reasoningProcessor.append(remainingReasoning);
                }

                // 2) pendingOutputQueue 里堆积的 output（等 reasoning 打完才准备播的）
                //    直接拼成整段文本喂给主 processor，跳过打字机
                if (
                  this._pendingOutputQueue &&
                  this._pendingOutputQueue.length
                ) {
                  let pendingText = '';
                  this._pendingOutputQueue.forEach(item => {
                    pendingText +=
                      (item.sentence && item.sentence.response) || '';
                  });
                  this._pendingOutputQueue = [];
                  if (pendingText) {
                    processor.updateSearchList(currentSearchList);
                    processor.append(pendingText);
                  }
                }
                this._isInReasoning = false;
                this._reasoningSSEDone = true;

                // 3) 主打字机剩余内容 → processor
                const remainingMain = drainPrint(this._print);
                if (remainingMain) {
                  processor.updateSearchList(currentSearchList);
                  processor.append(remainingMain);
                }

                // 4) 最终渲染 + 收尾
                this.setStoreSessionStatus(-1);
                const renderResult = processor.getRenderResult();
                const reasoningRenderResult =
                  reasoningProcessor.getRenderResult();
                sessionCom.replaceLastData(lastIndex, {
                  ...commonData,
                  ...renderResult,
                  activeReasoning: reasoningRenderResult.activeResponse || '',
                  stableReasoningChunks:
                    reasoningRenderResult.stableChunks || [],
                  searchList: currentSearchList,
                  qaSearchList: currentQASearchList,
                  ragSteps: [...ragSteps],
                  finish: 1,
                });
                this.$nextTick(() => sessionCom.scrollBottom());
                break;
              }

              case 'RUN_ERROR': {
                this.setStoreSessionStatus(-1);
                closeAllRunning(data.timestamp);
                // response: 面向用户的短文案（走 i18n 错误码表），作为错误卡片标题；
                // errorDetail: 后端 data.message 原文（含上游具体原因），作为副标题展示，
                //   便于用户/排查人员看到真实原因，而不只是"未知错误"四个字。
                const i18nKey = data.code && RAG_ERROR_CODE_I18N[data.code];
                const errText = i18nKey ? i18n.t(i18nKey) : i18n.t('sse.error');
                sessionCom.replaceLastData(lastIndex, {
                  ...commonData,
                  response: '',
                  errResponse: errText,
                  errorDetail: data.message || '',
                  error: true,
                  ragSteps: [...ragSteps],
                });
                this.stopEventSource();
                break;
              }

              // ── CUSTOM 事件（rag_qa_start/end、rag_knowledge_start/end 及检索结果）──
              case 'CUSTOM':
                if (data.name === CUSTOM_EVENT_QA_START) {
                  // 后端通知即将进入问答库检索：懒创建卡片
                  if (!findStep('qa_search'))
                    createStep('qa_search', data.timestamp);
                  sessionCom.replaceLastData(lastIndex, {
                    ...commonData,
                    responseLoading: true,
                    finish: 0,
                    searchList: currentSearchList,
                    qaSearchList: currentQASearchList,
                    ragSteps: [...ragSteps],
                  });
                } else if (data.name === CUSTOM_EVENT_QA_SEARCH_LIST) {
                  // 问答库检索完成（命中或未命中都发；未命中 value=[]）
                  currentQASearchList = (data.value || []).map(n => {
                    // QA 条目没有 snippet，用 question+answer 合成
                    const qaSnippet =
                      n.question || n.answer
                        ? `**Q:** ${n.question || ''}\n\n**A:** ${n.answer || ''}`
                        : '';
                    const raw = n.snippet || qaSnippet;
                    return {
                      ...n,
                      // QA 卡片优先显示知识库名（user_kb_name），不要用泛称 title="问答库"
                      title: n.user_kb_name || n.title || '',
                      snippet: raw ? md.render(raw) : '',
                    };
                  });
                  // 检索结果只更新内容；阶段结束时间由 rag_qa_end 提供。
                  sessionCom.replaceLastData(lastIndex, {
                    ...commonData,
                    responseLoading: true,
                    finish: 0,
                    searchList: currentSearchList,
                    qaSearchList: currentQASearchList,
                    ragSteps: [...ragSteps],
                  });
                  this.$nextTick(() => sessionCom.scrollBottom());
                } else if (data.name === CUSTOM_EVENT_QA_ERROR) {
                  // 问答库错误不结束步骤；最终耗时仍以 rag_qa_end 的 timestamp 为准。
                  const qaStep =
                    findStep('qa_search') ||
                    createStep('qa_search', data.timestamp);
                  qaStep.status = 'error';
                  qaStep.errorMessage =
                    typeof data.value === 'string'
                      ? data.value
                      : data.value == null
                        ? ''
                        : JSON.stringify(data.value);
                  sessionCom.replaceLastData(lastIndex, {
                    ...commonData,
                    responseLoading: true,
                    finish: 0,
                    searchList: currentSearchList,
                    qaSearchList: currentQASearchList,
                    ragSteps: [...ragSteps],
                  });
                  this.$nextTick(() => sessionCom.scrollBottom());
                } else if (data.name === CUSTOM_EVENT_QA_END) {
                  closeStep(findStep('qa_search'), data.timestamp);
                  sessionCom.replaceLastData(lastIndex, {
                    ...commonData,
                    responseLoading: true,
                    finish: 0,
                    searchList: currentSearchList,
                    qaSearchList: currentQASearchList,
                    ragSteps: [...ragSteps],
                  });
                } else if (data.name === CUSTOM_EVENT_KNOWLEDGE_START) {
                  // 后端通知即将进入知识库检索：懒创建卡片（幂等，重复帧不重复建）
                  if (!findStep('knowledge_search'))
                    createStep('knowledge_search', data.timestamp);
                  sessionCom.replaceLastData(lastIndex, {
                    ...commonData,
                    responseLoading: true,
                    finish: 0,
                    searchList: currentSearchList,
                    qaSearchList: currentQASearchList,
                    ragSteps: [...ragSteps],
                  });
                } else if (data.name === CUSTOM_EVENT_KNOWLEDGE_END) {
                  closeStep(findStep('knowledge_search'), data.timestamp);
                  sessionCom.replaceLastData(lastIndex, {
                    ...commonData,
                    responseLoading: true,
                    finish: 0,
                    searchList: currentSearchList,
                    qaSearchList: currentQASearchList,
                    ragSteps: [...ragSteps],
                  });
                } else if (data.name === CUSTOM_EVENT_SEARCH_LIST) {
                  currentSearchList = (data.value || []).map(n => ({
                    ...n,
                    snippet: n.snippet ? md.render(n.snippet) : '',
                  }));
                  // 命中结果到达：关闭 knowledge_search 步骤（若存在）
                  // 检索结果只更新内容；阶段结束时间由 rag_knowledge_end 提供。
                  // 在流式文字开始之前先把引用来源渲染到 UI
                  sessionCom.replaceLastData(lastIndex, {
                    ...commonData,
                    responseLoading: true,
                    finish: 0,
                    searchList: currentSearchList,
                    qaSearchList: currentQASearchList,
                    ragSteps: [...ragSteps],
                  });
                  this.$nextTick(() => sessionCom.scrollBottom());
                }
                break;

              // ── 推理内容（reasoning）────────────────────────────
              case 'REASONING_MESSAGE_START':
                this._isInReasoning = true;
                break;

              case 'REASONING_MESSAGE_CONTENT': {
                streamHasContent = true;
                const reasoning = data.delta || '';
                if (!reasoning) break;
                // 未命中兜底：没有 CUSTOM 也要关闭检索卡片
                ensureSearchStepClosed(data.timestamp);
                // 首个 reasoning_content 到达才创建思考卡片（有些模型无推理过程）
                if (!findStep('thinking')) createStep('thinking');
                this._dispatchReasoningOrOutput({
                  reasoning,
                  output: '',
                  finish: 0,
                  commonData,
                  doRenderReasoning: (wo, sl) => doRender(wo, sl, 'reasoning'),
                  doRenderMain: (wo, sl) => doRender(wo, sl, 'main'),
                });
                break;
              }

              case 'REASONING_MESSAGE_END':
                // 关闭思考卡片（如已创建）
                closeStep(findStep('thinking'));
                // 服务端推理阶段结束；若打字机已空载则立即触发 output 队列 flush
                if (!this._reasoningSSEDone) {
                  this._reasoningSSEDone = true;
                  if (
                    this._isInReasoning &&
                    this._reasoningPrint &&
                    this._reasoningPrint.sIndex >=
                      this._reasoningPrint.sentenceArr.length &&
                    this._reasoningPrint.printStatus === 0
                  ) {
                    this._flushPendingOutput();
                  }
                }
                break;

              // ── 正文内容（text output）───────────────────────────
              case 'TEXT_MESSAGE_START':
                if (data.detailId) {
                  commonData.detailId = data.detailId;
                }
                break;

              case 'TEXT_MESSAGE_CONTENT': {
                if (data.detailId) {
                  commonData.detailId = data.detailId;
                }
                streamHasContent = true;
                const output = data.delta || '';
                if (!output) break;
                // 未命中兜底：没有 CUSTOM、也没有 reasoning，正文到达也要关闭检索卡片
                ensureSearchStepClosed(data.timestamp);
                // 非推理模型不会发 REASONING_MESSAGE_END；正文到达即视为思考阶段结束
                closeStep(findStep('thinking'));
                this._dispatchReasoningOrOutput({
                  reasoning: '',
                  output,
                  finish: 0,
                  commonData,
                  doRenderReasoning: (wo, sl) => doRender(wo, sl, 'reasoning'),
                  doRenderMain: (wo, sl) => doRender(wo, sl, 'main'),
                });
                break;
              }

              case 'TEXT_MESSAGE_END':
                if (data.detailId) {
                  commonData.detailId = data.detailId;
                }
                // 发送 finish=1 信号给 Print；Print 动画结束时调用 doRender，
                // doRender 在 worldObj.isEnd && worldObj.finish===1 时调用 setStoreSessionStatus(-1)
                this._dispatchReasoningOrOutput({
                  reasoning: '',
                  output: '',
                  finish: 1,
                  commonData,
                  doRenderReasoning: (wo, sl) => doRender(wo, sl, 'reasoning'),
                  doRenderMain: (wo, sl) => doRender(wo, sl, 'main'),
                });
                break;

              default:
                break;
            }
          },
        },
      );
    },

    getAgentStreamHeaders() {
      return {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + this.token,
        'x-user-id': this.userInfo.uid,
        'x-org-id': this.userInfo.orgId,
        'X-Client-ID': getXClientId(),
      };
    },
    setActiveAgentStreamParams(params = {}) {
      const assistantId =
        params.assistantId || this.activeAgentStreamParams.assistantId || '';
      const conversationId =
        params.conversationId ||
        this.activeAgentStreamParams.conversationId ||
        '';
      const draft =
        params.draft !== undefined
          ? params.draft
          : this.activeAgentStreamParams.draft || false;

      this.activeAgentStreamParams = {
        assistantId,
        conversationId,
        draft,
      };
    },
    clearActiveAgentStreamParams() {
      this.activeAgentStreamParams = {
        assistantId: '',
        conversationId: '',
        draft: false,
      };
    },
    async cancelCurrentAgentStream() {
      if (!['agentChat', 'webChat'].includes(this.type)) return;

      const { assistantId, conversationId, draft } =
        this.activeAgentStreamParams;

      if (!assistantId || (!draft && !conversationId)) return;

      const params = {
        assistantId,
      };

      if (draft) {
        params.draft = true;
      }

      if (conversationId) {
        params.conversationId = conversationId;
      }

      const config =
        this.type === 'webChat'
          ? {
              headers: {
                'X-Client-ID': getXClientId(),
              },
              isOpenUrl: true,
            }
          : {};
      try {
        this.type === 'webChat'
          ? await cancelOpenurlAgentStream(
              params.assistantId,
              {
                conversationId: params.conversationId,
              },
              config,
            )
          : await cancelAgentStream(params, config);
      } catch (error) {
        console.error(error);
      }
    },
    async cancelCurrentRagStream() {
      const ragId = this.sseParams?.ragId;
      if (!ragId || !['chat', 'test'].includes(this.chatType)) return;

      const data = {
        ragId,
        draft: this.chatType === 'test',
      };
      if (this.chatType === 'chat' && this.sseParams?.conversationId) {
        data.conversationId = this.sseParams.conversationId;
      }
      try {
        await cancelRagStream(data);
      } catch (error) {
        console.warn('[rag chat] cancel stream failed', error);
      }
    },
    async cancelCurrentModelExperienceStream() {
      if (!this.sessionId || !this.apiParams) return;

      try {
        await cancelModelExperienceStream({
          sessionId: this.apiParams.sessionId,
        });
      } catch (error) {
        console.error(error);
      }
    },
    doSend(params) {
      this.stopBtShow = true;
      this.isStoped = false;
      let _history = this.$refs['session-com'].getList();
      this.sendEventSource(this.inputVal, '', _history.length);
    },
    sendEventSource(prompt, msgStr, lastIndex, options = {}) {
      console.log('####  sendEventSource', new Date().getTime());
      let sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) {
        console.warn('[sseMethod] session-com ref missing');
        return;
      }
      if (this.getCurrentSessionStatus() === 0) {
        this.$message.warning(i18n.t('sse.incompleteError'));
        return;
      }

      this.sseResponse = {};
      this.setStoreSessionStatus(0);
      if (options.clearInput !== false) {
        this.clearInput();
      }

      let params = options.pushHistoryData || {
        query: prompt,
        pending: true,
        responseLoading: true,
        requestFileUrls: this.queryFilePath ? [this.queryFilePath] : [],
        fileList: this.fileList,
        // 保留本次请求的附件元数据，SSE 业务错误后重试时需要从历史记录恢复。
        fileInfo: this.sseParams.fileInfo || [],
        pendingResponse: '',
      };
      if (!options.skipPushHistory) {
        sessionCom.pushHistory(params);
      }

      this._print = new Print({
        onPrintEnd: () => {
          this.onMainPrintEnd && this.onMainPrintEnd();
        },
      });

      let data = null;
      let headers = null;
      //判断是是不是openurl对话
      if (options.streamApi && options.streamData) {
        this.sseApi = options.streamApi;
        data = options.streamData;
        headers = options.headers || this.getAgentStreamHeaders();
      } else if (this.type === 'agentChat') {
        if (!this.isExplorePage()) {
          this.sseApi = `${AGENT_API_URL}/draft`;
        } else {
          this.sseApi = AGENT_API_URL;
        }
        data = {
          ...this.sseParams,
          prompt,
          systemPrompt: this.sseParams.systemPrompt, //提示词对比参数
        };
        headers = this.getAgentStreamHeaders();
      } else {
        this.sseApi = `${OPENURL_API}/agent/${this.sseParams.assistantId}/stream`;
        data = {
          conversationId: this.sseParams.conversationId,
          fileInfo: this.sseParams.fileInfo,
          prompt,
        };
        headers = {
          'X-Client-ID': getXClientId(),
        };
      }

      this._subConversionsMap = new Map(); // 子会话数据Map
      this._subConversionProcessors = new Map(); // 子会话处理器
      this._subMainProcessorsMap = new Map(); // 子会话内部正文片段处理器 (Key: subId_order)
      this._mainProcessors = new Map(); // 每个 order 的主处理器

      if (['agentChat', 'webChat'].includes(this.type)) {
        const isDraftStream =
          this.type === 'agentChat' &&
          (this.chatType === 'test' ||
            this.sseApi === `${AGENT_API_URL}/draft`);

        this.setActiveAgentStreamParams({
          assistantId: data.assistantId || this.sseParams.assistantId,
          conversationId: data.conversationId || this.sseParams.conversationId,
          draft: isDraftStream,
        });
      }

      let hasStreamMessage = false;

      this.eventSource = this.fetchEventSource(this.sseApi, data, {
        headers,
        ...(this.type === 'webChat' &&
          !options.streamApi && { isOpenUrl: true }),
        onopen: async e => {
          console.log('已建立SSE连接~', new Date().getTime());
          if (e.status !== 200) {
            try {
              const errorData = await e.json();
              let commonData = {
                ...this.sseParams,
                query: prompt,
              };
              let fillData = {
                ...commonData,
                response: errorData.msg,
              };
              sessionCom.replaceLastData(lastIndex, fillData);
            } catch (e) {
              const text = await e.text();
              this.$message.error(text || i18n.t('sse.error'));
            }

            this.stopEventSource();
            this.setStoreSessionStatus(-1);
            return;
          }
        },
        onmessage: e => {
          if (e && e.data) {
            let data = JSON.parse(e.data);
            console.log('===>', new Date().getTime(), data);
            hasStreamMessage = true;
            this.sseResponse = data;
            if (
              ['agentChat', 'webChat'].includes(this.type) &&
              data.conversationId
            ) {
              this.setActiveAgentStreamParams({
                assistantId:
                  this.type === 'webChat' ? undefined : data.assistantId,
                conversationId: data.conversationId,
              });
            }
            const currentQuery = prompt || data.prompt || data.query || '';
            //待替换的数据，需要前端组装
            let commonData = {
              ...data,
              ...this.sseParams,
              query: currentQuery,
              fileList: this.fileList,
              response: '',
              filepath: data.file_url || '',
              requestFileUrls: this.queryFilePath
                ? [this.queryFilePath]
                : data.requestFileUrls,
              searchList: data.search_list || [],
              gen_file_url_list: data.gen_file_url_list || [],
              thinkText: i18n.t('agent.thinking'),
              toolText: '使用工具中...',
              isOpen: true,
              showScrollBtn: null,
              citations: [],
              subConversions: [], // 初始化子会话列表
              messageSequence: [], // 初始化消息序列，用于平铺渲染
              _lastOrder: -1, // 内部追踪最后一次的 order
            };

            if (data.code === 0) {
              // 处理子会话消息 (eventType !== 0)
              if (
                data.eventType !== AGENT_MESSAGE_CONFIG.MAIN_AGENT.EVENT_TYPE &&
                data.eventData
              ) {
                const {
                  id,
                  name,
                  status,
                  timeCost,
                  profile,
                  order: innerOrder,
                  parentId,
                  errMessage,
                } = data.eventData;
                let subConversion = this._subConversionsMap.get(id);
                let subProcessor = this._subConversionProcessors.get(id);

                if (!subConversion) {
                  subConversion = {
                    id,
                    name,
                    status, // 1开始、2输出中、3结束、4处理失败
                    parentId: parentId || '', // 核心：关联父级ID
                    timeCost,
                    profile, //头像
                    innerOrder: innerOrder, // 内部排序序号
                    response: '',
                    stableChunks: [],
                    activeResponse: '',
                    errMessage: errMessage || '',
                    isOpen:
                      data.eventType ===
                      AGENT_MESSAGE_CONFIG.AGENT_THINK.EVENT_TYPE, // agentThink 默认展开，其他默认收起
                    searchList: data.search_list || [], // 初始化 searchList
                    citationsTagList: [], // 已引用的出处索引
                    conversationType: this.convertConversionType(
                      data.eventType,
                    ),
                    messageSequence: [], // 支持子会话内部穿插序列
                    userToggled: false, // 标记用户是否手动操作过
                  };
                  this._subConversionsMap.set(id, subConversion);

                  // 初始化流处理器
                  subProcessor = new StreamProcessor({
                    lastIndex,
                    md,
                    parseSub: (text, index, searchList) =>
                      parseSubConversation(text, index, searchList, id),
                    convertLatexSyntax,
                    preProcess: processToolResultBlocks,
                    searchList: subConversion.searchList,
                  });
                  this._subConversionProcessors.set(id, subProcessor);
                } else {
                  // 更新状态 (状态单向锁：如果已经是 3 或 4，则不更新为 1 或 2)
                  if (
                    !(subConversion.status === 3 || subConversion.status === 4)
                  ) {
                    subConversion.status = status;
                  }
                  if (
                    (status === 3 || status === 4) &&
                    data.eventType ===
                      AGENT_MESSAGE_CONFIG.AGENT_THINK.EVENT_TYPE &&
                    !subConversion.userToggled // 仅在用户未手动操作过时自动折叠
                  ) {
                    subConversion.isOpen = false;
                  }
                  if (timeCost) subConversion.timeCost = timeCost;
                  if (innerOrder !== undefined)
                    subConversion.innerOrder = innerOrder; // 更新内部排序
                  if (errMessage) subConversion.errMessage = errMessage; // 更新错误信息
                  // 如果后续包中有 search_list，则更新
                  if (data.search_list && data.search_list.length) {
                    subConversion.searchList = data.search_list;
                    subProcessor.updateSearchList(data.search_list);
                  }
                }

                // 累加回复内容并处理流 (针对子会话容器内部进行穿插序列化分段)
                if (data.response) {
                  const innerOrderKey = `${id}_${data.order}`;
                  let chunkProcessor =
                    this._subMainProcessorsMap.get(innerOrderKey);
                  let processedResponse = data.response.replace(/\\n/g, '\n');

                  // 1. 在子会话容器内寻找当前 order 对应的文本片段 (type: 'main')
                  let currentMainChunk = subConversion.messageSequence.find(
                    item => item.type === 'main' && item.order === data.order,
                  );

                  if (!currentMainChunk) {
                    currentMainChunk = {
                      type: 'main',
                      order: data.order,
                      stableChunks: [],
                      activeResponse: '',
                      response: '',
                    };
                    subConversion.messageSequence.push(currentMainChunk);
                    // 按 order 排序，确保输出顺序
                    subConversion.messageSequence.sort(
                      (a, b) => (a.order || 0) - (b.order || 0),
                    );

                    // 为该片段创建专属打字机处理器
                    chunkProcessor = new StreamProcessor({
                      lastIndex,
                      md,
                      parseSub: (text, index, searchList) =>
                        parseSubConversation(text, index, searchList, id),
                      convertLatexSyntax,
                      preProcess: processToolResultBlocks, // 预处理 <<<...>>> 工具结果块
                      searchList: subConversion.searchList,
                    });
                    this._subMainProcessorsMap.set(
                      innerOrderKey,
                      chunkProcessor,
                    );
                  }

                  // 写入并更新渲染块
                  currentMainChunk.response += processedResponse;
                  chunkProcessor.append(processedResponse);
                  const renderResult = chunkProcessor.getRenderResult();

                  // 应对同一 order 下的多包连续推流
                  this.$set(
                    currentMainChunk,
                    'stableChunks',
                    renderResult.stableChunks,
                  );
                  this.$set(
                    currentMainChunk,
                    'activeResponse',
                    renderResult.activeResponse,
                  );

                  // 把打字机状态同步提升到外层 subConversion
                  // 当作为文本分段被父级“吸收”时，外层模板将直接读取 subConversion 的这两个属性进行打字
                  this.$set(
                    subConversion,
                    'stableChunks',
                    renderResult.stableChunks,
                  );
                  this.$set(
                    subConversion,
                    'activeResponse',
                    renderResult.activeResponse,
                  );

                  // 物理累加（兼容旧逻辑，但不作为新版嵌套渲染的主数据源）
                  subConversion.response = subConversion.messageSequence
                    .filter(i => i.type === 'main')
                    .map(i => i.response)
                    .join('');
                  subConversion.citationsTagList = renderResult.citations || [];
                }

                // 处理子会话递归嵌套：将此节点及其 order 注册进父级序列
                if (parentId) {
                  const parentSub = this._subConversionsMap.get(parentId);
                  if (parentSub) {
                    const hasInParent = parentSub.messageSequence.some(
                      item =>
                        (item.type === 'sub' || item.type === 'main') &&
                        item.id === id,
                    );
                    if (!hasInParent) {
                      const isTextChunk =
                        data.eventType ===
                        AGENT_MESSAGE_CONFIG.SUB_TEXT.EVENT_TYPE;

                      // 若为正文片段，直接强转类型，并将其整体obj放入父序列
                      if (isTextChunk) {
                        subConversion.type = 'main';
                        subConversion.order = data.order;
                        parentSub.messageSequence.push(subConversion);
                      } else {
                        // 常规情况：只是放一个引用卡片标志过去
                        parentSub.messageSequence.push({
                          type: 'sub',
                          id: id,
                          order: data.order,
                        });
                      }

                      parentSub.messageSequence.sort(
                        (a, b) => (a.order || 0) - (b.order || 0),
                      );
                    }
                  }
                }

                // 更新消息序列
                let sequence =
                  sessionCom.getSessionData()['history'][lastIndex]
                    ?.messageSequence || [];
                // 仅将顶层子会话加入主消息的平铺序列区，孙级由递归负责渲染
                if (
                  data.order !== undefined &&
                  data.order !== null &&
                  !parentId
                ) {
                  let currentSubItem = sequence.find(
                    item => item.type === 'sub' && item.id === id,
                  );
                  if (!currentSubItem) {
                    currentSubItem = {
                      type: 'sub',
                      id: id,
                      order: data.order,
                    };
                    sequence.push(currentSubItem);
                  }
                }

                // 构造 fillData
                // 获取最新的子会话列表
                const subConversionsList = Array.from(
                  this._subConversionsMap.values(),
                );

                let fillData = {
                  ...commonData,
                  finish:
                    this._currentMainFinish !== undefined
                      ? this._currentMainFinish
                      : 0,
                  subConversions: subConversionsList,
                  messageSequence: sequence,
                };
                sessionCom.replaceLastData(lastIndex, fillData);
                // 如果子智能体结束或失败，可能需要滚动到底部
                if (status === 3 || status === 4) {
                  this.$nextTick(() => sessionCom.scrollBottom());
                }
              } else {
                // 主智能体消息 (eventType === 0 或 undefined)
                // 更新当前主智能体 finish 状态
                this._currentMainFinish = data.finish;

                // 根据 order 获取或创建对应的 processor
                const currentOrder = data.order !== undefined ? data.order : 0;
                let mainProcessor = this._mainProcessors.get(currentOrder);

                if (!mainProcessor) {
                  mainProcessor = new StreamProcessor({
                    lastIndex,
                    md,
                    parseSub,
                    convertLatexSyntax,
                  });
                  this._mainProcessors.set(currentOrder, mainProcessor);
                }

                //finish 0：进行中  1：关闭   2:敏感词关闭
                let _sentence = data.response;
                this._print.print(
                  {
                    response: _sentence,
                    finish: data.finish,
                  },
                  commonData,
                  (worldObj, search_list) => {
                    this.setStoreSessionStatus(0);
                    mainProcessor.updateSearchList(search_list);
                    mainProcessor.append(worldObj.world);

                    const renderResult = mainProcessor.getRenderResult();

                    // 更新消息序列
                    let sequence =
                      sessionCom.getSessionData()['history'][lastIndex]
                        ?.messageSequence || [];

                    if (data.order !== undefined && data.order !== null) {
                      let currentMainItem = sequence.find(
                        item =>
                          item.type === 'main' && item.order === data.order,
                      );

                      if (!currentMainItem) {
                        currentMainItem = {
                          type: 'main',
                          order: data.order,
                          renderedContent: '',
                          stableChunks: [],
                          activeResponse: '',
                        };
                        sequence.push(currentMainItem);
                      }

                      currentMainItem.renderedContent = renderResult.response;
                      currentMainItem.stableChunks = renderResult.stableChunks;
                      currentMainItem.activeResponse =
                        renderResult.activeResponse;
                    }

                    // 获取最新的子会话列表
                    const subConversionsList = Array.from(
                      this._subConversionsMap.values(),
                    );

                    let fillData = {
                      ...commonData,
                      ...renderResult,
                      responseFiles:
                        (this.sseResponse && this.sseResponse.responseFiles) ||
                        commonData.responseFiles ||
                        [],
                      detailId:
                        (this.sseResponse && this.sseResponse.detailId) ||
                        commonData.detailId ||
                        '',
                      finish: worldObj.finish,
                      searchList:
                        search_list && search_list.length
                          ? search_list.map(n => ({
                              ...n,
                              snippet: md.render(n.snippet),
                            }))
                          : [],
                      subConversions: subConversionsList,
                      messageSequence: sequence,
                    };
                    sessionCom.replaceLastData(lastIndex, fillData);
                    if (worldObj.finish !== 0) {
                      if (worldObj.finish === 4) {
                        let fillData = {
                          ...commonData,
                          response: i18n.t('sse.sensitiveTips'),
                          subConversions: subConversionsList,
                          messageSequence: sequence,
                        };
                        sessionCom.replaceLastData(lastIndex, fillData);
                        this.$nextTick(() => {
                          sessionCom.scrollBottom();
                        });
                      }
                      this.setStoreSessionStatus(-1);
                    }

                    if (worldObj.isEnd && worldObj.finish === 1) {
                      this.setStoreSessionStatus(-1);
                      this._currentMainFinish = undefined;
                    }
                  },
                );
              }
            } else if (data.code !== 0) {
              this.setStoreSessionStatus(-1);
              const historyList = sessionCom.getSessionData()['history'] || [];
              const lastData = historyList[lastIndex] || {};
              // 获取最新的子会话列表，防止被覆盖
              const subConversionsList = this._subConversionsMap
                ? Array.from(this._subConversionsMap.values())
                : [];
              let fillData = {
                ...lastData,
                errResponse:
                  data.response || data.message || this.$t('rag.answerFailed'),
                errorDetail: data.message || '',
                subConversions: subConversionsList,
                error: true,
              };
              sessionCom.replaceLastData(lastIndex, fillData);
              this._currentMainFinish = undefined;
              this._print && this._print.stop();
            }
          }
        },
        onclose: () => {
          console.log('===> sendEventSource onClose');
          // 1. 如果打字机仍在运行，等待其自然结束（onPrintEnd 回调）后再收尾
          //    若打字机已经空闲（或未启动），直接同步收尾
          const doFinalize = () => {
            this.setStoreSessionStatus(-1);
            const history = sessionCom.getSessionData()['history'] || [];
            const lastItem = history[lastIndex];
            if (
              options.removeEmptyOnClose &&
              !hasStreamMessage &&
              history.length - 1 === lastIndex
            ) {
              sessionCom.removeLastHistory();
              this.echo = !(sessionCom.getSessionData()['history'] || [])
                .length;
              this.stopBtShow = false;
              this.sseOnCloseCallBack && this.sseOnCloseCallBack();
              this.clearActiveAgentStreamParams();
              return;
            }
            if (lastItem && lastItem.responseLoading) {
              sessionCom.replaceLastData(lastIndex, {
                ...lastItem,
                responseLoading: false,
              });
            }
            this.sseOnCloseCallBack && this.sseOnCloseCallBack();
            this.clearActiveAgentStreamParams();
          };
          // 打字机还在跑（printStatus===1）或队列未排空时，挂载 onPrintEnd 延迟执行
          if (
            this._print &&
            (this._print.printStatus === 1 ||
              this._print.sIndex < this._print.sentenceArr.length)
          ) {
            const originalOnPrintEnd = this._print.onPrintEnd;
            this._print.onPrintEnd = () => {
              originalOnPrintEnd && originalOnPrintEnd();
              doFinalize();
            };
          } else {
            doFinalize();
          }
        },
      });
    },
    // 更新子会话的用户操作状态
    connectEventSource({
      assistantId,
      conversationId,
      query = '',
      lastIndex,
      fileList = [],
      headers = null,
    } = {}) {
      if (!assistantId || !conversationId) return;
      const sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) {
        console.warn('[sseMethod] session-com ref missing');
        return;
      }

      this.stopBtShow = true;
      this.isStoped = false;
      this.echo = false;
      this.fileList = fileList || [];
      this.queryFilePath = '';
      this.setSseParams({
        assistantId,
        conversationId,
      });

      const history = sessionCom.getSessionData()['history'] || [];
      const targetIndex =
        typeof lastIndex === 'number' ? lastIndex : history.length;
      const shouldPushHistory = typeof lastIndex !== 'number';

      const connectApi =
        this.type === 'webChat'
          ? `${OPENURL_API}/agent/${assistantId}/stream/connect`
          : this.chatType === 'test'
            ? AGENT_DRAFT_CONNECT_API_URL
            : AGENT_CONNECT_API_URL;

      this.sendEventSource(query, '', targetIndex, {
        streamApi: connectApi,
        streamData: {
          assistantId,
          conversationId,
        },
        clearInput: false,
        headers,
        skipPushHistory: !shouldPushHistory,
        removeEmptyOnClose: shouldPushHistory,
        pushHistoryData: shouldPushHistory
          ? {
              query,
              pending: true,
              responseLoading: true,
              requestFileUrls: [],
              fileList: fileList || [],
              pendingResponse: '',
              subConversions: [],
              messageSequence: [],
            }
          : null,
      });
    },
    setSubConversionUserToggle(id, isOpen) {
      if (this._subConversionsMap) {
        let subConversion = this._subConversionsMap.get(id);
        if (subConversion) {
          subConversion.isOpen = isOpen;
          subConversion.userToggled = true;
        }
      }
    },
    doExprienceSend(params) {
      this.stopBtShow = true;
      this.isStoped = false;
      this.fileInfoList =
        params.fileInfo?.map(item => ({
          fileName: item.oldFileName || item.fileName,
          fileSize: item.fileSize,
          fileUrl: item.fileUrl,
        })) || [];
      let _history = this.$refs['session-com'].getList();
      this.sendExprienceEventStream(
        params.inputVal,
        '',
        _history.length,
        params.fileList,
        { fileInfo: this.fileInfoList },
      );
    },
    sendExprienceEventStream(
      prompt,
      msgStr,
      lastIndex,
      fileList,
      options = {},
    ) {
      this.sseResponse = {};
      this.setStoreSessionStatus(0);
      const sessionCom = this.$refs['session-com'];
      if (!sessionCom) return;

      let targetIndex = typeof lastIndex === 'number' ? lastIndex : 0;
      let hasPushedHistory = false;
      const {
        streamApi = this.exprience_sseApi,
        streamData = null,
        skipPushHistory = false,
        lazyPushHistory = false,
        pushHistoryData = null,
        initialResponse = '',
      } = options;
      const params = pushHistoryData || {
        query: prompt,
        pending: true,
        responseLoading: true,
        requestFileUrls: [],
        fileList: fileList || [],
        fileInfo: options.fileInfo || [],
        pendingResponse: '',
      };
      if (!skipPushHistory && !lazyPushHistory) {
        sessionCom.pushHistory(params);
        hasPushedHistory = true;
      }
      const ensureHistoryForStream = () => {
        if (hasPushedHistory || skipPushHistory) return;
        targetIndex = sessionCom.getList().length;
        sessionCom.pushHistory(params);
        hasPushedHistory = true;
      };
      let endStr = initialResponse;
      // 初始化推理流处理器
      const reasoningProcessor = this._initReasoningStream({
        lastIndex: targetIndex,
        md,
        parseSub,
        convertLatexSyntax,
      });

      this._print = new Print({
        onPrintEnd: () => {
          // this.setStoreSessionStatus(-1)
        },
      });

      this.eventSource = this.fetchEventSource(
        streamApi,
        streamData || {
          ...this.apiParams,
          content: prompt,
          fileInfo: this.fileInfoList,
        },
        {
          headers: this.getAgentStreamHeaders(),
          onopen: async e => {
            //console.log("已建立SSE连接~",new Date().getTime());
            if (e.status !== 200) {
              try {
                const errorData = await e.json();
                let commonData = {
                  ...this.sseParams,
                  query: prompt,
                  fileList: fileList || [],
                  fileInfo: this.fileInfoList || [],
                };
                let fillData = {
                  ...commonData,
                  response: errorData.msg,
                };
                ensureHistoryForStream();
                sessionCom.replaceLastData(targetIndex, fillData);
              } catch (e) {
                const text = await e.text();
                this.$message.error(text || i18n.t('sse.error'));
              }

              this.stopEventSource();
              this.setStoreSessionStatus(-1);
              return;
            }
          },
          onmessage: e => {
            if (e && e.data) {
              let data;
              try {
                data = JSON.parse(e.data);
                // console.log('===>', new Date().getTime(), data);
              } catch (error) {
                return; // 如果解析失败，直接返回，不处理这条消息
              }

              ensureHistoryForStream();

              const choices = data.choices && data.choices[0];
              const delta = (choices && choices.delta) || {};
              const reasoning = delta.reasoning_content || '';
              const output = delta.content || '';
              // 对齐原逻辑的兜底：如果没有 choices 或符合 stop 条件，标识为结束
              const isFinish =
                !choices ||
                choices.finish_reason === 'stop' ||
                delta.content === 'stop';

              this.setStoreSessionStatus(0);
              this.sseResponse = data;
              //待替换的数据，需要前端组装
              let commonData = {
                ...this.sseResponse,
                ...this.sseParams,
                query: prompt,
                fileList: fileList || [],
                fileInfo: this.fileInfoList || [],
                fileName: '',
                fileSize: '',
                response: '',
                filepath: '',
                requestFileUrls: '',
                searchList:
                  this.sseResponse.data && this.sseResponse.data.searchList
                    ? this.sseResponse.data.searchList
                    : [],
                gen_file_url_list: [],
                thinkText: i18n.t('sse.thinkingText'),
                isOpen: true,
                citations: [],
                qa_type: 0, // 为了组件复用，前端加了标识
              };
              if ([7, -1].includes(data.code)) {
                this.setStoreSessionStatus(-1);
                let fillData = {
                  ...commonData,
                  response: data.message,
                  error: true,
                };
                sessionCom.replaceLastData(targetIndex, fillData);
              } else {
                // 定义推理内容渲染逻辑
                const doRenderReasoning = worldObj => {
                  this.setStoreSessionStatus(0);
                  reasoningProcessor.append(worldObj.world);
                  const reasoningRenderResult =
                    reasoningProcessor.getRenderResult();
                  let fillData = {
                    ...commonData,
                    activeReasoning: reasoningRenderResult.activeResponse || '',
                    stableReasoningChunks:
                      reasoningRenderResult.stableChunks || [],
                    finish: 0,
                  };
                  sessionCom.replaceLastData(targetIndex, fillData);
                };

                // 定义正文渲染逻辑（保持原有非分片拼接特性）
                const doRenderMain = (worldObj, search_list) => {
                  this.setStoreSessionStatus(0);
                  const reasoningRenderResult =
                    reasoningProcessor.getRenderResult();
                  endStr += worldObj.world;
                  endStr = convertLatexSyntax(endStr);
                  endStr = parseSub(endStr, targetIndex);
                  let fillData = {
                    ...commonData,
                    activeReasoning: reasoningRenderResult.activeResponse || '',
                    stableReasoningChunks:
                      reasoningRenderResult.stableChunks || [],
                    response: md.render(endStr),
                    oriResponse: endStr,
                    finish: worldObj.finish ? 1 : 0,
                    searchList:
                      search_list && search_list.length
                        ? search_list.map(n => ({
                            ...n,
                            snippet: n.snippet ? md.render(n.snippet) : '',
                          }))
                        : [],
                  };
                  sessionCom.replaceLastData(targetIndex, fillData);
                  if (worldObj.isEnd && worldObj.finish) {
                    this.setStoreSessionStatus(-1);
                  }
                };

                // 分发处理：如果是推理内容，或者需要缓冲的正文
                this._dispatchReasoningOrOutput({
                  reasoning,
                  output,
                  finish: isFinish ? 1 : 0,
                  commonData,
                  doRenderReasoning,
                  doRenderMain,
                });
              }
            }
          },
        },
      );
    },
    // 重连模型体验 SSE
    connectModelExperienceEventSource() {
      if (!this.sessionId) return;
      const sessionCom = this.$refs['session-com'];
      if (!sessionCom) return;

      this.stopEventSource();
      this._print && this._print.stop();
      this._reasoningPrint && this._reasoningPrint.stop();

      const sessionData = sessionCom.getSessionData() || {};
      const history = sessionData.history || [];
      const hasHistory = history.length > 0;
      const lastIndex = hasHistory ? history.length - 1 : 0;
      const lastItem = hasHistory ? history[lastIndex] || {} : {};
      const query = lastItem.query || '';
      const fileList = lastItem.fileList || [];

      this.stopBtShow = true;
      this.isStoped = false;
      this.sendExprienceEventStream(query, '', lastIndex, fileList, {
        streamApi: EXPRIENCE_CONNECT_API_URL,
        streamData: {
          sessionId: this.sessionId,
        },
        skipPushHistory: hasHistory,
        lazyPushHistory: !hasHistory,
        initialResponse: lastItem.oriResponse || '',
      });
    },
    // 多线程SSE简化版本
    sendEventStreamIsolation(url, params, callbacks = {}, timeout = 0) {
      let fullContent = '';
      let isCompleted = false;
      const { onProgress, onComplete } = callbacks;

      const _print = new Print({});
      const ctrlAbort = new AbortController();

      const handleComplete = content => {
        if (isCompleted) return;
        isCompleted = true;
        ctrlAbort.abort();
        if (onComplete) onComplete(content);
      };

      this.fetchEventSource(`${USER_API}` + url, params, {
        onopen: async response => {
          if (response.status !== 200) {
            try {
              const errorData = await response.json();
              console.log('Network error', errorData);
              this.$message.error(errorData.msg || i18n.t('sse.error'));
            } catch (e) {
              console.error('Failed to parse error response', e);
              this.$message.error(i18n.t('sse.error'));
            }
            handleComplete(fullContent);
          }
        },
        onmessage: e => {
          if (e && e.data) {
            try {
              const data = JSON.parse(e.data);
              _print.print(
                {
                  response: data.response,
                  finish: data.finish,
                },
                {},
                worldObj => {
                  fullContent += worldObj.world;
                  if (onProgress) onProgress(fullContent, worldObj);
                  if (Boolean(worldObj.finish)) {
                    console.log('===> eventSource onComplete');
                    handleComplete(fullContent);
                  }
                },
              );
            } catch (e) {
              console.warn('message json parse fail: ', e);
            }
          }
        },
        onclose: () => {
          console.log('===> eventSource onClose');
          handleComplete(fullContent);
        },
        onerror: e => {
          console.log(i18n.t('sse.connectError'));
          if (e.readyState === EventSource.CLOSED) {
            console.log('connection is closed');
          } else {
            console.warn('Error occured', e);
          }
          handleComplete(fullContent);
        },
        signal: ctrlAbort.signal,
      });

      if (timeout > 0) {
        setTimeout(() => {
          if (!ctrlAbort.signal.aborted) {
            ctrlAbort.abort();
            this.$message.warning(i18n.t('sse.timeoutError'));
            handleComplete(fullContent);
          }
        }, timeout);
      }
    },
    preStop() {
      // 立即置 -1：隐藏停止按钮，避免重复点击触发多次 abort
      this.setStoreSessionStatus(-1);
      this.cancelCurrentRagStream();
      this.cancelCurrentAgentStream();
      this.cancelCurrentModelExperienceStream();
      // 立刻同步断流 + 停掉两个打字机，不等 sseOnCloseCallBack 的异步链路。
      // 这样"点一次停止，字立刻停"：在 reasoning 阶段也不会被残余帧拉回。
      this.ctrlAbort && this.ctrlAbort.abort();
      // 置 null：避免 stopEventSource / handleComplete 里再次 abort 同一个
      // AbortController（虽然幂等，但能让"已经停过了"这件事显式化）。
      this.ctrlAbort = null;
      this._print && this._print.stop();
      this._reasoningPrint && this._reasoningPrint.stop();
      // 强制关闭最后一条消息里所有还在 running 的 ragSteps
      // （思考卡片计时器 / 齿轮动画 / 左侧 running 徽标同时停下）
      // 注意：step 对象必须"原地" mutate，不能 map 出新对象 —— 因为
      // sendRagEventSource 的闭包 ragSteps 与 lastItem.ragSteps 共享 step
      // 引用，如果晚到的 SSE 帧再次 replaceLastData({ ragSteps: [...ragSteps] })
      // 用的还是闭包里的旧对象；原地改才能让两边都看到 done。
      // 下面再 new 一个外层数组只是为了触发 Vue 响应式，和 step 原地修改并不矛盾。
      try {
        const sessionCom = this.sessionComRef || this.$refs['session-com'];
        if (sessionCom) {
          const history = sessionCom.getSessionData().history;
          const lastIndex = history.length - 1;
          const lastItem = history[lastIndex];
          if (lastItem && Array.isArray(lastItem.ragSteps)) {
            const now = Date.now();
            lastItem.ragSteps.forEach(s => {
              if (s && s.status === 'running') {
                s.status = 'done';
                s.endAt = now;
                s.duration = `${((now - (s.startAt || now)) / 1000).toFixed(3)}s`;
              }
            });
            // 新建数组引用以触发 Vue 响应式（对象内部已被原地改）
            // 不能简单改 finish=2：replaceLastData 对 finish!==0 且 response 空时
            // 会兜底写成"无响应数据"，正文尚未开始时会把思考卡片下方糊上那行字。
            // 底部三点加载动画由 `sessionStatus==0` 控制，此处只需保证
            // responseLoading:false + ragSteps:done 即可。
            sessionCom.replaceLastData(lastIndex, {
              ...lastItem,
              responseLoading: false,
              ragSteps: [...lastItem.ragSteps],
            });
          }
        }
      } catch (_e) {
        // 静默：停止按钮的兜底逻辑不应阻断 abort 主流程
      }
      //获取已经拿到的全部回答,一次性回显出来
      this.sseOnCloseCallBack(true);
      this.clearActiveAgentStreamParams();
    },
    sseOnCloseCallBack(isStoped) {
      this.stopEventSource();
      //图文问答不使用打字机
      /* if(this.sseResponse.qa_type === 6){
                return
            }*/
      //主动停止
      if (isStoped) {
        // 手动停止时，将所有进行中的子会话状态置为失败/停止
        if (this._subConversionsMap) {
          let hasUpdate = false;
          for (let sub of this._subConversionsMap.values()) {
            if (sub.status === 1 || sub.status === 2) {
              sub.status = 4;
              hasUpdate = true;
            }
          }
          if (hasUpdate) {
            let sessionCom = this.sessionComRef || this.$refs['session-com'];
            if (sessionCom) {
              let history = sessionCom.getSessionData().history;
              let lastIndex = history.length - 1;
              if (lastIndex >= 0) {
                const subConversionsList = Array.from(
                  this._subConversionsMap.values(),
                );
                let lastItem = history[lastIndex];
                sessionCom.replaceLastData(lastIndex, {
                  ...lastItem,
                  subConversions: subConversionsList,
                });
              }
            }
          }
        }
        this.stopAndEcho();
      } else {
        //收到onclose,且使用的是文生代码
        if (this.sseResponse.qa_type === 4) {
          this.stopAndEcho();
        } else {
          //接口405等
          let history_list = [];
          let lastIndex = history_list.length - 1;
          let lastRQ = history_list[lastIndex];
          let endStr = this._print.getAllworld();
          endStr = convertLatexSyntax(endStr);
          // 替换标签
          endStr = parseSub(endStr);
          // 如果返回有结果，则在结束时不展示“本次回答已终止”
          this.runResponse = md.render(endStr);
          this.runDisabled = false;
          this.setStoreSessionStatus(-1);
        }
      }
    },
    stopAndEcho() {
      //暂存已经收到的所有response
      let endResponse = this._print.getAllworld();

      this._print && this._print.stop();
      // RAG 流还有一个独立的思考打字机，必须一并停掉，否则它会继续
      // 触发 doRender(field='reasoning') → setStoreSessionStatus(0) →
      // 停止按钮被反复拉回可见。
      this._reasoningPrint && this._reasoningPrint.stop();

      setTimeout(() => {
        this.setStoreSessionStatus(-1);

        let history_list = [];
        let lastIndex = history_list.length - 1;
        let lastRQ = history_list[lastIndex];
        if (endResponse) {
          endResponse = convertLatexSyntax(endResponse);
          // 替换标签
          endResponse = parseSub(endResponse);
          this.runResponse = md.render(endResponse);
          this.runDisabled = false;
        } else {
          if (
            Object.keys(this.sseResponse).length !== 0 &&
            this.sseResponse.code !== 7
          ) {
            this.runResponse = '本次回答已被终止';
            this.setStoreSessionStatus(-1);
          } else {
            this.stopEventSource();
            this.setStoreSessionStatus(-1);
            this.$refs['session-com'].stopPending();
          }
        }
      }, 15);
    },
    stopEventSource() {
      this.ctrlAbort && this.ctrlAbort.abort();
      this.eventSource = null;
    },
    refreshLastSession() {
      let endResponse = this._print.getAllworld();
      let history_list = [];
      let lastIndex = history_list.length - 1;
      let lastRQ = history_list[lastIndex];
      // this.$refs['session-com'].replaceLastData(lastIndex, {
      //     ...lastRQ,
      //     response: endResponse
      // })
    },
    setPrompt(data) {
      const editable =
        this.$refs.editable || (this.getEditableRef && this.getEditableRef());
      if (editable) {
        editable.setPrompt(data);
      }
      // this.$refs['editable'].setPrompt(data)
    },
    clearInput() {
      const editable =
        this.$refs.editable || (this.getEditableRef && this.getEditableRef());
      if (editable) {
        editable.clearInput();
        editable.clearFile();
      }
      this.inputVal = '';
      this.fileId = '';
    },
    clearPageHistory() {
      this.$refs['session-com'] && this.$refs['session-com'].clearData();
      // this.$refs.editable && this.clearInput()
      this.clearInput();
    },
    clearHistory() {
      this.stopBtShow = false;
      this.clearPageHistory();
    },
    refresh() {
      let sessionCom = this.sessionComRef || this.$refs['session-com'];
      if (!sessionCom) return;
      let history_list = sessionCom.getList();
      let _history = history_list[history_list.length - 1];
      let inputVal = _history.query;
      let fileInfo = _history.fileInfo ? _history.fileInfo : [];
      let fileList = _history.fileList ? _history.fileList : [];
      this.preSend(inputVal, fileList, fileInfo);
    },
  },
};
