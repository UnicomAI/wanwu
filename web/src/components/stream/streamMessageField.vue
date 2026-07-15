<!--问答消息框-->
<template>
  <div class="session rl">
    <div v-if="supportClear" class="session-setting">
      <el-link
        class="right-setting"
        @click="gropdownClick"
        type="primary"
        :underline="false"
        :class="{ 'is-disabled': sessionStatus === 0 }"
        style="color: var(--color); top: 0"
      >
        <span class="el-icon-delete"></span>
        {{ $t('app.clearChat') }}
      </el-link>
    </div>
    <div
      class="history-box showScroll"
      :id="scrollContainerId"
      v-loading="loading"
      ref="timeScroll"
      @click="handleGlobalClick"
      @mouseover="handleCitationHover"
      @mouseout="handleCitationLeave"
      :style="{ 'max-height': historyBoxHeight }"
    >
      <div v-for="(n, i) in session_data.history" :key="`${i}sdhs`">
        <!--问题-->
        <div v-if="n.query" class="session-question">
          <div :class="['session-item', 'rl']">
            <img class="logo" :src="userAvatarSrc" />
            <div class="answer-content">
              <div class="answer-content-query">
                <div class="echo-doc-box" v-if="hasFiles(n)">
                  <el-button
                    v-show="canScroll(i, n.showScrollBtn)"
                    icon="el-icon-arrow-left "
                    @click="prev($event, i)"
                    circle
                    class="scroll-btn left"
                    size="mini"
                    type="primary"
                    style="z-index: 10"
                  ></el-button>
                  <div class="imgList" :ref="`imgList-${i}`">
                    <div
                      v-for="(file, j) in n.fileList"
                      :key="`${j}sdsl`"
                      class="docInfo-img-container"
                    >
                      <el-image
                        v-if="hasImgs(n, file)"
                        :src="file.fileUrl"
                        class="docIcon imgIcon"
                        :preview-src-list="[file.fileUrl]"
                        fit="cover"
                      />
                      <div v-else class="docInfo-container">
                        <FileIcon
                          :type="getFileIconType(file)"
                          size="30px"
                          class="docIcon"
                        />
                        <div class="docInfo">
                          <p class="docInfo_name">
                            {{ $t('knowledgeManage.fileName') }}:{{ file.name }}
                          </p>
                          <p class="docInfo_size">
                            {{ $t('knowledgeManage.fileSize') }}:{{
                              getFileSizeDisplay(file.size)
                            }}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                  <el-button
                    v-show="canScroll(i, n.showScrollBtn)"
                    icon="el-icon-arrow-right"
                    @click="next($event, i)"
                    circle
                    class="scroll-btn right"
                    size="mini"
                    type="primary"
                  ></el-button>
                </div>
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
                  >
                    <i class="el-icon-s-order"></i>
                    &nbsp;
                    {{ $t('agent.copyToInput') }}
                  </p>
                  <span
                    slot="reference"
                    class="answer-text"
                    style="display: inline-block; margin-top: 5px"
                  >
                    {{ n.query }}
                  </span>
                </el-popover>
              </div>
            </div>
          </div>
        </div>
        <!--loading-->
        <div v-if="n.responseLoading && !n.error" class="session-answer">
          <div class="session-answer-wrapper">
            <img class="logo" :src="modelIconUrl || avatarSrc(defaultUrl)" />
            <div class="answer-content"><i class="el-icon-loading"></i></div>
          </div>
        </div>
        <!--pending-->
        <div v-if="n.pendingResponse" class="session-answer">
          <div class="session-answer-wrapper">
            <img class="logo" :src="modelIconUrl || avatarSrc(defaultUrl)" />
            <div class="answer-content" style="padding: 10px; color: #e6a23c">
              {{ n.pendingResponse }}
            </div>
          </div>
        </div>

        <!--回答 文字+图片-->
        <div
          v-if="
            n.error ||
            n.response ||
            n.msg_type ||
            (n.subConversions && n.subConversions.length) ||
            n.activeReasoning ||
            (n.stableReasoningChunks && n.stableReasoningChunks.length)
          "
          class="session-answer"
          :id="'message-container' + i"
        >
          <div class="session-answer-wrapper">
            <img class="logo" :src="modelIconUrl || avatarSrc(defaultUrl)" />
            <div class="session-wrap" style="width: calc(100% - 30px)">
              <!-- 思考块显示 (msg_type 逻辑) -->
              <div
                class="deepseek"
                v-if="
                  n.msg_type &&
                  ['qa_start', 'qa_finish', 'knowledge_start'].includes(
                    n.msg_type,
                  )
                "
              >
                <img
                  :src="require('@/assets/imgs/think-icon.png')"
                  class="think_icon"
                />
                {{ getTitle(n.msg_type) }}
              </div>
              <div
                v-else-if="
                  !(n.messageSequence && n.messageSequence.length) &&
                  !(n.ragSteps && n.ragSteps.length) &&
                  (showDSBtn(n.response || '') ||
                    n.activeReasoning ||
                    (n.stableReasoningChunks && n.stableReasoningChunks.length))
                "
              >
                <div class="deepseek" @click="toggle($event, i)">
                  <img
                    :src="require('@/assets/imgs/think-icon.png')"
                    class="think_icon"
                  />
                  {{
                    n.activeReasoning ||
                    (n.stableReasoningChunks && n.stableReasoningChunks.length)
                      ? n.finish === 0 &&
                        !n.response &&
                        !n.activeResponse &&
                        (!n.stableChunks || n.stableChunks.length === 0)
                        ? n.thinkText || $t('agent.thinking')
                        : $t('agent.thinked')
                      : n.thinkText
                  }}
                  <i
                    v-bind:class="{
                      'el-icon-arrow-down': !n.isOpen,
                      'el-icon-arrow-up': n.isOpen,
                    }"
                  ></i>
                </div>
              </div>

              <!-- 消息序列渲染 -->
              <div
                v-if="n.messageSequence && n.messageSequence.length"
                class="message-sequence-wrapper"
              >
                <template v-for="(item, idx) in n.messageSequence">
                  <!-- 子会话渲染块 -->
                  <div
                    v-if="item.type === 'sub'"
                    :key="'sub-' + item.id + idx"
                    class="sub-conversion-box order-sub"
                  >
                    <sub-conversion
                      :conversion="findSubData(n, item.id)"
                      :parents-index="i"
                      :all-sub-conversions="n.subConversions"
                      @toggle-conversion="toggleSubConversion"
                      @collapse-click="collapseClick"
                    ></sub-conversion>
                  </div>

                  <!-- 主会话渲染块 -->
                  <div
                    v-else-if="item.type === 'main'"
                    :key="'main-' + idx"
                    class="order-main-chunk"
                  >
                    <!-- 片段内的思考按钮 -->
                    <div
                      v-if="
                        showDSBtn(
                          item.renderedContent || item.activeResponse || '',
                        )
                      "
                    >
                      <div class="deepseek" @click="toggle($event, i)">
                        <img
                          :src="require('@/assets/imgs/think-icon.png')"
                          class="think_icon"
                        />
                        {{ n.thinkText }}
                        <i
                          v-bind:class="{
                            'el-icon-arrow-down': !n.isOpen,
                            'el-icon-arrow-up': n.isOpen,
                          }"
                        ></i>
                      </div>
                    </div>
                    <template
                      v-if="
                        (item.stableChunks && item.stableChunks.length) ||
                        item.activeResponse
                      "
                    >
                      <div class="answer-content">
                        <div
                          v-for="(chunk, cIdx) in item.stableChunks"
                          :key="'stable-' + cIdx"
                          class="chunk_stable"
                          @click="$refs.imagePreview.handleImageClick($event)"
                          v-bind:class="{ 'ds-res': showDSBtn(chunk) }"
                          v-html="
                            showDSBtn(chunk) ? replaceHTML(chunk, n) : chunk
                          "
                        ></div>
                        <div
                          v-if="item.activeResponse"
                          class="chunk_active"
                          @click="$refs.imagePreview.handleImageClick($event)"
                          v-bind:class="{
                            'ds-res': showDSBtn(item.activeResponse),
                          }"
                          v-html="
                            showDSBtn(item.activeResponse)
                              ? replaceHTML(item.activeResponse, n)
                              : item.activeResponse
                          "
                        ></div>
                      </div>
                    </template>
                    <!-- 历史内容 -->
                    <div
                      v-else-if="item.renderedContent"
                      class="answer-content order-main-renderedContent"
                      @click="$refs.imagePreview.handleImageClick($event)"
                      v-bind:class="{
                        'ds-res': showDSBtn(item.renderedContent),
                        hideDs: !n.isOpen,
                      }"
                      v-html="
                        showDSBtn(item.renderedContent)
                          ? replaceHTML(item.renderedContent, n)
                          : item.renderedContent
                      "
                    ></div>

                    <!-- 局部错误卡片 -->
                    <ErrorMsgCard
                      v-if="item.errMsg || item.errResponse"
                      :title="item.errResponse || item.response"
                      :desc="item.errMsg"
                      :class="{ 'agent-error-card': chatType === 'agent' }"
                    />
                  </div>
                </template>
              </div>

              <!-- 如果没有 messageSequence(rag) -->
              <template v-else>
                <!-- 子会话渲染区域 -->
                <div
                  v-if="n.subConversions && n.subConversions.length"
                  class="sub-conversion-box"
                >
                  <sub-conversion-list
                    parent-id=""
                    :all-sub-conversions="n.subConversions"
                    :parents-index="i"
                    @toggle-conversion="toggleSubConversion"
                    @collapse-click="collapseClick"
                  ></sub-conversion-list>
                </div>

                <!-- RAG 过程卡片：知识库检索 + 深度思考（可折叠） -->
                <template
                  v-if="chatType === 'rag' && n.ragSteps && n.ragSteps.length"
                >
                  <rag-step-card
                    v-for="(step, sIdx) in n.ragSteps"
                    :key="'rag-step-' + sIdx"
                    :type="step.type"
                    :status="step.status"
                    :start-at="step.startAt"
                    :duration="step.duration"
                    :should-collapse="
                      step.type === 'thinking' && hasFinalAnswerStarted(n)
                    "
                  >
                    <template v-if="step.type === 'qa_search'">
                      <div
                        v-if="n.qaSearchList && n.qaSearchList.length"
                        class="rag-source-list"
                        @click="$refs.imagePreview.handleImageClick($event)"
                      >
                        <div
                          v-for="(m, j) in n.qaSearchList"
                          :key="`rag-qa-${j}`"
                          class="rag-source-item"
                          :data-citation-index="j + 1"
                        >
                          <div class="rag-source-header">
                            <span class="rag-source-index">{{ j + 1 }}</span>
                            <span
                              class="rag-source-title"
                              :title="m.title || m.link || ''"
                            >
                              {{ m.title || m.link || '—' }}
                            </span>
                            <a
                              v-if="m.link"
                              :href="m.link"
                              target="_blank"
                              rel="noopener noreferrer"
                              class="rag-source-download"
                              :title="m.link"
                            >
                              <i class="el-icon-download"></i>
                            </a>
                          </div>
                          <div
                            v-if="typeof m.score === 'number'"
                            class="rag-source-meta"
                          >
                            <span class="rag-source-pill rag-source-score">
                              Score: {{ formatScore(m.score) }}
                            </span>
                          </div>
                          <div
                            v-if="m.snippet"
                            :ref="'ragQASnippet_' + i + '_' + j"
                            class="rag-source-snippet"
                            :class="{
                              'is-collapsed':
                                !ragSnippetExpanded['qa-' + i + '-' + j],
                            }"
                            v-html="m.snippet"
                          ></div>
                          <div
                            v-if="
                              m.snippet &&
                              ragSnippetOverflow['qa-' + i + '-' + j]
                            "
                            class="rag-source-expand-btn"
                            @click="toggleRagSnippet('qa-' + i + '-' + j)"
                          >
                            {{
                              ragSnippetExpanded['qa-' + i + '-' + j]
                                ? $t('common.button.fold')
                                : $t('common.button.viewAll')
                            }}
                          </div>
                        </div>
                      </div>
                      <div
                        v-else-if="step.status === 'running'"
                        class="rag-source-loading"
                      >
                        <i class="el-icon-loading"></i>
                        <span>{{ $t('rag.step.running') }}</span>
                      </div>
                      <div v-else class="rag-source-empty">
                        {{ $t('rag.step.noHit') }}
                      </div>
                    </template>
                    <template v-else-if="step.type === 'knowledge_search'">
                      <div
                        v-if="n.searchList && n.searchList.length"
                        class="rag-source-list"
                        @click="$refs.imagePreview.handleImageClick($event)"
                      >
                        <div
                          v-for="(m, j) in n.searchList"
                          :key="`rag-src-${j}`"
                          class="rag-source-item"
                          :data-citation-index="j + 1"
                        >
                          <div class="rag-source-header">
                            <span class="rag-source-index">{{ j + 1 }}</span>
                            <span
                              class="rag-source-title"
                              :title="m.title || m.link || ''"
                            >
                              {{ m.title || m.link || '—' }}
                            </span>
                            <a
                              v-if="m.link"
                              :href="m.link"
                              target="_blank"
                              rel="noopener noreferrer"
                              class="rag-source-download"
                              :title="m.link"
                            >
                              <i class="el-icon-download"></i>
                            </a>
                          </div>
                          <div
                            v-if="m.user_kb_name || typeof m.score === 'number'"
                            class="rag-source-meta"
                          >
                            <span
                              v-if="m.user_kb_name"
                              class="rag-source-pill rag-source-kb"
                              :title="m.user_kb_name"
                            >
                              {{ m.user_kb_name }}
                            </span>
                            <span
                              v-if="typeof m.score === 'number'"
                              class="rag-source-pill rag-source-score"
                            >
                              Score: {{ formatScore(m.score) }}
                            </span>
                          </div>
                          <div
                            v-if="m.snippet"
                            :ref="'ragSnippet_' + i + '_' + j"
                            class="rag-source-snippet"
                            :class="{
                              'is-collapsed': !ragSnippetExpanded[i + '-' + j],
                            }"
                            v-html="m.snippet"
                          ></div>
                          <div
                            v-if="m.snippet && ragSnippetOverflow[i + '-' + j]"
                            class="rag-source-expand-btn"
                            @click="toggleRagSnippet(i + '-' + j)"
                          >
                            {{
                              ragSnippetExpanded[i + '-' + j]
                                ? $t('common.button.fold')
                                : $t('common.button.viewAll')
                            }}
                          </div>
                        </div>
                      </div>
                      <!-- running 阶段：显示 loading，避免尚未完成就错判"未命中" -->
                      <div
                        v-else-if="step.status === 'running'"
                        class="rag-source-loading"
                      >
                        <i class="el-icon-loading"></i>
                        <span>{{ $t('rag.step.running') }}</span>
                      </div>
                      <!-- done 且 searchList 为空：判定未命中 -->
                      <div v-else class="rag-source-empty">
                        {{ $t('rag.step.noHit') }}
                      </div>
                    </template>
                    <template v-else-if="step.type === 'thinking'">
                      <div
                        v-for="(chunk, idx) in n.stableReasoningChunks"
                        :key="'rc-' + idx"
                        class="chunk_stable"
                        @click="$refs.imagePreview.handleImageClick($event)"
                        v-html="unwrapCodeImages(chunk)"
                      ></div>
                      <div
                        v-if="n.activeReasoning"
                        class="chunk_active"
                        @click="$refs.imagePreview.handleImageClick($event)"
                        v-html="unwrapCodeImages(n.activeReasoning)"
                      ></div>
                    </template>
                  </rag-step-card>
                </template>

                <!-- 主会话-->
                <!-- 透传的独立思考过程区域（非 RAG 或无 ragSteps 时才使用此内联展示） -->
                <template
                  v-if="
                    !(chatType === 'rag' && n.ragSteps && n.ragSteps.length) &&
                    ((n.stableReasoningChunks &&
                      n.stableReasoningChunks.length) ||
                      n.activeReasoning)
                  "
                >
                  <div
                    class="answer-content no-order-chunk-answer reasoning-area ds-res"
                    v-show="n.isOpen"
                  >
                    <section class="reasoning-area-content">
                      <div
                        v-for="(chunk, idx) in n.stableReasoningChunks"
                        :key="'r-' + idx"
                        class="chunk_stable"
                        @click="$refs.imagePreview.handleImageClick($event)"
                        v-html="chunk"
                      ></div>
                      <div
                        v-if="n.activeReasoning"
                        class="chunk_active"
                        @click="$refs.imagePreview.handleImageClick($event)"
                        v-html="n.activeReasoning"
                      ></div>
                    </section>
                  </div>
                </template>

                <template
                  v-if="
                    (n.stableChunks && n.stableChunks.length) ||
                    n.activeResponse
                  "
                >
                  <div
                    class="answer-content no-order-chunk-answer"
                    :class="{ 'rag-answer': chatType === 'rag' }"
                    @click="$refs.imagePreview.handleImageClick($event)"
                  >
                    <div
                      v-for="(chunk, idx) in n.stableChunks"
                      :key="idx"
                      class="chunk_stable"
                      v-bind:class="{ 'ds-res': showDSBtn(chunk) }"
                      v-html="showDSBtn(chunk) ? replaceHTML(chunk, n) : chunk"
                    ></div>
                    <div
                      v-if="n.activeResponse"
                      class="chunk_active"
                      v-bind:class="{ 'ds-res': showDSBtn(n.activeResponse) }"
                      v-html="
                        showDSBtn(n.activeResponse)
                          ? replaceHTML(n.activeResponse, n)
                          : n.activeResponse
                      "
                    ></div>
                  </div>
                </template>
                <div
                  v-else-if="n.response"
                  class="answer-content history-answer"
                  v-bind:class="{
                    'ds-res': showDSBtn(n.response),
                    hideDs: !n.isOpen,
                    'rag-answer': chatType === 'rag',
                  }"
                  @click="$refs.imagePreview.handleImageClick($event)"
                >
                  <div
                    v-html="
                      showDSBtn(n.response)
                        ? replaceHTML(n.response, n)
                        : n.response
                    "
                  ></div>
                </div>
              </template>

              <!-- 整个回答下方的兜底主错误卡片 -->
              <ErrorMsgCard
                v-if="n.error && n.errResponse"
                :title="n.errResponse || n.response"
                :desc="n.errorDetail"
                :class="{ 'agent-error-card': chatType === 'agent' }"
              />
            </div>
          </div>
          <!--文件-->
          <div
            v-if="n.gen_file_url_list && n.gen_file_url_list.length"
            class="file-path response-file"
          >
            <el-image
              v-for="(g, k) in n.gen_file_url_list"
              :key="k"
              :src="g"
              :preview-src-list="[g]"
            ></el-image>
          </div>
          <!-- 出处 -->
          <div
            v-if="
              n.searchList &&
              n.searchList.length &&
              n.finish === 1 &&
              (chatType !== 'agent' || !hasNewAgentKnowledge(n)) &&
              !(chatType === 'rag' && n.ragSteps && n.ragSteps.length)
            "
            class="search-list"
          >
            <h2
              class="recommended-question-title"
              v-if="n.msg_type && ['qa_finish'].includes(n.msg_type)"
            >
              {{ $t('app.recommendedQuestion') }}
            </h2>
            <div
              v-for="(m, j) in n.searchList"
              :key="`${j}sdsl`"
              class="search-list-item"
            >
              <div
                v-if="m.content_type && m.content_type === 'qa'"
                class="qa_content"
                @click="handleRecommendedQuestion(m)"
              >
                <span>{{ j + 1 }}. {{ m.question }}</span>
              </div>
              <template v-else>
                <div
                  class="serach-list-item"
                  v-if="showSearchList(j, n.citations)"
                >
                  <span @click="collapseClick(n, m, j)">
                    <i
                      :class="[
                        '',
                        m.collapse
                          ? 'el-icon-caret-bottom'
                          : 'el-icon-caret-right',
                      ]"
                    ></i>
                    {{ $t('agent.source') }}：
                  </span>
                  <a
                    v-if="m.link"
                    :href="m.link"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="link"
                  >
                    {{ m.link }}
                  </a>
                  <span
                    v-if="m.title"
                    @click.stop="handleSourceTitleClick(n, m, j, i)"
                  >
                    <sub
                      class="subTag"
                      :data-parents-index="i"
                      :data-collapse="m.collapse ? 'true' : 'false'"
                    >
                      {{ j + 1 }}
                    </sub>
                    {{ m.title }}
                  </span>
                </div>
                <el-collapse-transition>
                  <div
                    v-show="!!m.collapse"
                    class="snippet"
                    @click="$refs.imagePreview.handleImageClick($event)"
                  >
                    <p v-html="m.snippet"></p>
                  </div>
                </el-collapse-transition>
              </template>
            </div>
          </div>
          <!-- 主体内容后的slot -->
          <div
            v-if="
              n.finish === 1 &&
              (i !== session_data.history.length - 1 || sessionStatus !== 0)
            "
            class="answer-operation"
          >
            <slot
              name="afterContent"
              :responseFiles="n.responseFiles"
              :item="n"
              :index="i"
            />
          </div>
          <!--loading-->
          <div
            v-if="
              n.finish === 0 &&
              sessionStatus == 0 &&
              i === session_data.history.length - 1
            "
            class="text-loading"
          >
            <div></div>
            <div></div>
            <div></div>
          </div>
          <!--停止生成 重新生成 点赞   session code 是0时不可操作-->
          <div class="answer-operation gap-10px">
            <div class="opera-left">
              <span
                v-if="
                  i === session_data.history.length - 1 && sessionStatus !== 0
                "
                class="restart"
                @click="refresh"
              >
                <img :src="require('@/assets/imgs/refresh-icon.png')" />
              </span>
              <span
                class="preStop"
                @click="preStop"
                v-if="
                  supportStop &&
                  i === session_data.history.length - 1 &&
                  sessionStatus === 0
                "
              >
                <img :src="require('@/assets/imgs/stop-icon.png')" />
              </span>
            </div>
            <div
              class="opera-right"
              style="flex: 0"
              @click="
                () => {
                  copy(n.oriResponse) && copycb();
                }
              "
            >
              <img :src="require('@/assets/imgs/copy-icon.png')" />
            </div>
            <svg-icon
              v-if="
                chatType === 'agent' && (n.finish === 1 || sessionStatus !== 0)
              "
              icon-class="trash"
              class="del-icon"
              @click="handleDelConversation(n)"
            />
            <!--提示话术-->
            <div class="answer-operation-tip">
              {{ $t('agent.answerOperationTip') }}
            </div>
          </div>
          <!-- 推荐问题 -仅最后一条回答显示 -->
          <div
            v-if="
              sessionStatus === -1 &&
              ((recommendConfig.list && recommendConfig.list.length) ||
                recommendConfig.loading) &&
              i === session_data.history.length - 1
            "
            class="session-section-wrapper recommend-question"
          >
            <div
              v-for="(item, index) in recommendConfig.list"
              :key="index"
              :class="[
                'recommend-question-item',
                { 'is-tips': item.type === 'tips' },
              ]"
              @click="
                item.type !== 'tips' &&
                $emit('handleRecommendClick', item.content)
              "
            >
              {{ item.content }}
            </div>
            <div
              v-if="recommendConfig.loading"
              class="text-loading recommend-question-loading"
            >
              <div></div>
              <div></div>
              <div></div>
            </div>
          </div>
        </div>

        <!-- 回答 仅图片-->
        <div
          v-if="
            !n.response && n.gen_file_url_list && n.gen_file_url_list.length
          "
          class="session-answer"
        >
          <div class="session-answer-wrapper">
            <img class="logo" :src="modelIconUrl || avatarSrc(defaultUrl)" />
            <div class="answer-content">
              <div
                v-if="n.gen_file_url_list && n.gen_file_url_list.length"
                class="file-path response-file no-response"
              >
                <el-image
                  v-for="(g, k) in n.gen_file_url_list"
                  :key="k"
                  :src="g"
                  :preview-src-list="[g]"
                ></el-image>
              </div>
            </div>
          </div>
          <!--仅图片时只有 重新生成-->
          <div class="answer-operation">
            <div class="opera-left">
              <span
                v-if="i === session_data.history.length - 1"
                class="restart"
              >
                <i class="el-icon-refresh" @click="refresh">
                  &nbsp;
                  {{ $t('agent.refresh') }}
                </i>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!-- RAG 引用 hover 气泡：单例，定位到被 hover 的 .citation 附近 -->
    <transition name="rag-tip-fade">
      <div
        v-if="ragCitationTip.visible"
        class="rag-citation-popover"
        :class="'placement-' + ragCitationTip.placement"
        :style="{
          left: ragCitationTip.x + 'px',
          top: ragCitationTip.y + 'px',
          '--citation-arrow-offset': ragCitationTip.arrowOffset + 'px',
        }"
      >
        <div class="rag-citation-popover-head">
          <span class="rag-citation-popover-num">
            {{ ragCitationTip.number }}
          </span>
          <span
            class="rag-citation-popover-title"
            :title="ragCitationTip.title"
          >
            {{ ragCitationTip.title || $t('rag.citation.source') }}
          </span>
        </div>
        <div v-if="ragCitationTip.snippet" class="rag-citation-popover-snippet">
          {{ ragCitationTip.snippet }}
        </div>
        <div class="rag-citation-popover-hint">
          {{ $t('rag.citation.viewSource') }}
        </div>
      </div>
    </transition>
    <!-- 图片预览组件 -->
    <ImagePreview ref="imagePreview" />
  </div>
</template>

<script>
import smoothscroll from 'smoothscroll-polyfill';
import { md } from '@/mixins/markdown-it';
import 'highlight.js/styles/atom-one-dark.css';
import commonMixin from '@/mixins/common';
import { mapGetters } from 'vuex';
import { avatarSrc, formatScore, getFileIconType } from '@/utils/util';
import SubConversion from './subConversion/index.vue';
import SubConversionList from './subConversion/SubConversionList.vue';
import RagStepCard from './ragStepCard.vue';
import ErrorMsgCard from './errorMsgCard.vue';
import ImagePreview from '@/components/ImagePreview.vue';
import FileIcon from '@/components/FileIcon.vue';
import { AGENT_MESSAGE_CONFIG } from '@/components/stream/constants';

export default {
  mixins: [commonMixin],
  props: {
    defaultUrl: {
      type: String,
      default: '',
    },
    chatType: {
      type: String,
      default: '',
    },
    recommendConfig: {
      type: Object,
      default: () => ({
        reqController: null,
        list: [],
        loading: false,
      }),
    },
    modelIconUrl: {},
    supportStop: {},
    modelSessionStatus: {},
    supportClear: {
      type: Boolean,
      default: true,
    },
  },
  components: {
    SubConversion,
    SubConversionList,
    RagStepCard,
    ErrorMsgCard,
    ImagePreview,
    FileIcon,
  },
  data() {
    return {
      md: md,
      autoScroll: true,
      scrollTimeout: null,
      loading: false,
      session_data: {
        tool: '',
        searchList: [],
        history: [],
        response: '',
      },
      c: null,
      ctx: null,
      canvasShow: false,
      cv: null,
      currImg: {
        url: '',
        width: 0, // 原始宽高
        height: 0,
        w: 0, // 压缩后的宽高
        h: 358,
        roteX: 0, // 压缩后的比例
        roteY: 0,
      },
      imgConfig: ['jpeg', 'PNG', 'png', 'JPG', 'jpg', 'bmp', 'webp'],
      audioConfig: ['mp3', 'wav'],
      fileScrollStateMap: {},
      resizeTimer: null,
      scrollContainerId: `timeScroll-${this._uid}`,
      // 复制提示计时器map
      copyTimerMap: new Map(),
      historyBoxHeight: '', // 动态历史会话容器高度
      // 引用角标 hover 气泡（RAG / Agent 复用）
      ragCitationTip: {
        visible: false,
        x: 0,
        y: 0,
        placement: 'top', // 'top' | 'bottom'
        title: '',
        snippet: '',
        number: '',
        arrowOffset: 0,
      },
      // RAG 引用片段展开状态：key = `${messageIdx}-${sourceIdx}`，对齐模板里的 ragSnippet_ ref 命名
      // ragSnippetOverflow 一次性检测后缓存，不再重算——snippet 到达后不再变化，且可避免展开态反作用于检测逻辑
      ragSnippetExpanded: {},
      ragSnippetOverflow: {},
      _ragCitationTipHideTimer: null,
    };
  },
  computed: {
    ...mapGetters('user', ['userAvatar']),
    // ...mapState('app', ['sessionStatus']),
    sessionStatus() {
      return ['number', 'string'].includes(typeof this.modelSessionStatus)
        ? this.modelSessionStatus
        : this.$store.state.app.sessionStatus;
    },
    userAvatarSrc() {
      return this.userAvatar
        ? avatarSrc(this.userAvatar)
        : require('@/assets/imgs/robot-icon.png');
    },
    isStreaming() {
      const history = this.session_data.history;
      if (history.length === 0) return false;
      const lastItem = history[history.length - 1];
      return lastItem.finish === 0 && this.sessionStatus === 0;
    },
  },
  watch: {
    'session_data.history': {
      handler() {
        this.$nextTick(() => {
          this.updateAllFileScrollStates();
          console.log(this.session_data.history);
        });
      },
      deep: true,
    },
  },
  mounted() {
    this.setupScrollListener();
    smoothscroll.polyfill();
    document.addEventListener('click', this.handleCitationClick);
    addEventListener('resize', this.handleWindowResize);
    // 兜底：外层 wheel / resize 时立即隐藏引用气泡（避免 fixed 定位错位）
    addEventListener('wheel', this.onWindowWheelHideTip, {
      passive: true,
    });
    this.updateAllFileScrollStates();
  },
  updated() {
    // RAG 引用卡片的 snippet 渲染后检测是否溢出 6 行——checkRagSnippetOverflow 对已检测 key 早退，
    // 流式过程每条新帧触发 updated 的开销可忽略
    this.$nextTick(() => this.checkRagSnippetOverflow());
  },
  beforeDestroy() {
    if (this.handleCitationClick) {
      document.removeEventListener('click', this.handleCitationClick);
    }
    const container = document.getElementById(this.scrollContainerId);
    if (container) {
      container.removeEventListener('scroll', this.handleScroll);
    }
    clearTimeout(this.scrollTimeout);

    removeEventListener('resize', this.handleWindowResize);
    removeEventListener('wheel', this.onWindowWheelHideTip);
    if (this.resizeTimer) {
      clearTimeout(this.resizeTimer);
    }
    // 移除图片错误事件监听器
    if (this.imageErrorHandler) {
      document.body.removeEventListener('error', this.imageErrorHandler, true);
    }
    // 清除复制提示计时器
    this.copyTimerMap.forEach(timerId => {
      clearTimeout(timerId);
    });
    this.copyTimerMap.clear();
    // 清除引用角标 hover tooltip 的延迟隐藏计时器
    if (this._ragCitationTipHideTimer) {
      clearTimeout(this._ragCitationTipHideTimer);
      this._ragCitationTipHideTimer = null;
    }
  },
  methods: {
    getFileIconType,
    avatarSrc,
    formatScore,
    getTitle(type) {
      if (type === 'qa_start') {
        return this.$t('app.qaSearching');
      } else if (type === 'knowledge_start') {
        return this.$t('app.knowledgeSearch');
      } else if (type === 'qa_finish') {
        return this.$t('knowledgeManage.qaDatabase.title');
      } else {
        return this.$t('menu.knowledge');
      }
    },
    handleRecommendedQuestion(m) {
      this.$emit('handleRecommendedQuestion', m.question);
    },
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
          this.$set(
            this.session_data.history[index],
            'showScrollBtn',
            canScroll,
          );
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
    canScroll(i, showScrollBtn) {
      if (showScrollBtn !== null && showScrollBtn !== undefined) {
        return showScrollBtn;
      }
      return this.fileScrollStateMap[i] || false;
    },
    prev(e, i) {
      e.stopPropagation();
      const refKey = `imgList-${i}`;
      const containerArray = this.$refs[refKey];
      if (containerArray && containerArray.length > 0) {
        const container = containerArray[0];
        container.scrollBy({
          left: -200,
          behavior: 'smooth',
        });
      }
    },
    next(e, i) {
      e.stopPropagation();
      const refKey = `imgList-${i}`;
      const containerArray = this.$refs[refKey];
      if (containerArray && containerArray.length > 0) {
        const container = containerArray[0];
        container.scrollBy({
          left: 200,
          behavior: 'smooth',
        });
      }
    },
    hasFiles(n) {
      return n.fileList && n.fileList.length > 0;
    },
    hasImgs(n, file) {
      if (!n.fileList || n.fileList.length === 0 || !file || !file.name) {
        return false;
      }
      let type = file.name.split('.').pop().toLowerCase();
      return this.imgConfig.map(t => t.toLowerCase()).includes(type);
    },
    handleCitationClick(e) {
      const target = e.target;

      // 处理引用气泡内部图标点击 (兼容代码，实际不触发)
      if (target.classList.contains('citation-tips-content-icon')) {
        const index = target.dataset.index;
        const citation = Number(target.dataset.citation);
        const historyItem = this.session_data.history[index];
        if (historyItem && historyItem.searchList) {
          const searchItem = historyItem.searchList[citation - 1];
          if (searchItem) {
            const j = historyItem.searchList.indexOf(searchItem);
            this.collapseClick(historyItem, searchItem, j);
          }
        }
        e.stopPropagation();
        return;
      }

      // 处理引用标签点击
      const citationTarget = target.closest('.citation');

      if (!citationTarget) return;

      if (citationTarget.dataset.pid) {
        // 子会话引用点击处理(思考子会话等)
        const pid = citationTarget.dataset.pid;
        const parentsIndex = Number(citationTarget.dataset.parentsIndex);
        const citationIndex = Number(citationTarget.textContent);
        const historyItem = this.session_data.history[parentsIndex];

        if (historyItem && historyItem.subConversions) {
          const citationContext = this.resolveCitationSourceConversion(
            historyItem,
            pid,
            citationIndex,
          );

          if (
            citationContext &&
            citationContext.dataOwner &&
            citationContext.dataOwner.searchList &&
            citationContext.dataOwner.searchList[citationIndex - 1]
          ) {
            const { dataOwner, displayOwner } = citationContext;
            const searchItem = dataOwner.searchList[citationIndex - 1];

            if (
              displayOwner &&
              displayOwner.conversationType ===
                AGENT_MESSAGE_CONFIG.AGENT_KNOWLEDGE.CONVERSATION_TYPE
            ) {
              this.$set(displayOwner, 'isOpen', true);
              this.scrollToKnowledgeCitation(displayOwner.id, citationIndex);
            } else {
              this.collapseClick(dataOwner, searchItem, citationIndex - 1);
              this.scrollToSubConversionCitation(
                (displayOwner && displayOwner.id) || dataOwner.id,
                citationIndex,
              );
            }

            e.stopPropagation();
            return;
          }
        }
      }

      // 获取引用所属的历史消息项
      const parentsIndex = Number(citationTarget.dataset.parentsIndex);
      const historyItem = this.session_data.history[parentsIndex];

      // Agent 主会话引用（无 data-pid，来自顶层回答）跳转到对应的知识库子会话
      if (
        this.chatType === 'agent' &&
        !citationTarget.dataset.pid &&
        this.hasNewAgentKnowledge(historyItem)
      ) {
        const index = Number(citationTarget.textContent);
        if (historyItem && historyItem.subConversions) {
          const knowledgeSub = historyItem.subConversions.find(
            a =>
              a.conversationType ===
              AGENT_MESSAGE_CONFIG.AGENT_KNOWLEDGE.CONVERSATION_TYPE,
          );

          if (knowledgeSub) {
            this.$set(knowledgeSub, 'isOpen', true);
            this.$nextTick(() => {
              const container = document.querySelector(
                `.sub-conversion-item[data-sub-id="${knowledgeSub.id}"]`,
              );
              if (container) {
                const targetSearchItem = container.querySelector(
                  `.knowledge-item[data-index="${index - 1}"]`,
                );
                if (targetSearchItem) {
                  targetSearchItem.scrollIntoView({
                    behavior: 'smooth',
                    block: 'center',
                  });
                }
              }
            });
            e.stopPropagation();
            return;
          }
        }
        return;
      }

      // RAG 引用点击：展开知识库检索卡片 + 滚动到对应 source + 高亮闪烁
      if (
        this.chatType === 'rag' &&
        historyItem &&
        historyItem.ragSteps &&
        historyItem.ragSteps.length &&
        historyItem.searchList &&
        historyItem.searchList.length
      ) {
        const index = Number(citationTarget.textContent);
        if (index && historyItem.searchList[index - 1]) {
          this.scrollToRagCitation(parentsIndex, index);
          e.stopPropagation();
          return;
        }
      }

      // 通用引用点击处理
      this.$handleCitationClick(e, {
        sessionStatus: this.sessionStatus,
        sessionData: this.session_data,
        citationSelector: '.citation',
        scrollElementId: this.scrollContainerId,
        onToggleCollapse: (item, collapse) => {
          this.$set(item, 'collapse', collapse);
        },
      });
    },
    // 子会话展开收起
    toggleSubConversion(conversion) {
      const newState = !conversion.isOpen;
      this.$set(conversion, 'isOpen', newState);
      this.$emit('sub-conversion-toggle', {
        id: conversion.id,
        isOpen: newState,
      });
    },
    showSearchList(j, citations) {
      return (citations || []).includes(j + 1);
    },
    setCitations(index) {
      let citation = `#message-container${index} .citation`;
      const allCitations = document.querySelectorAll(citation);
      const citationsSet = new Set();

      allCitations.forEach(element => {
        const text = element.textContent.trim();
        if (text) {
          citationsSet.add(Number(text));
        }
      });

      return Array.from(citationsSet);
    },
    setupScrollListener() {
      const container = document.getElementById(this.scrollContainerId);
      container.addEventListener('scroll', this.handleScroll);
    },
    handleScroll(e) {
      // 滚动时立即隐藏引用气泡：popover 用 fixed 定位 + viewport 坐标，
      // 滚动时 citation 已经移位，气泡停在原地会严重错位。
      this.hideRagCitationTip();
      const container = document.getElementById(this.scrollContainerId);
      const { scrollTop, clientHeight, scrollHeight } = container;
      const nearBottom = scrollHeight - (scrollTop + clientHeight) < 5;
      if (!nearBottom) {
        this.autoScroll = false;
      }
      clearTimeout(this.scrollTimeout);
      this.scrollTimeout = setTimeout(() => {
        if (nearBottom) {
          this.autoScroll = true;
          this.scrollBottom();
        }
      }, 500);
    },
    replaceHTML(data, n) {
      const thinkStart = /<think>/i;
      const thinkEnd = /<\/think>/i;
      const toolStart = /<tool>/i;
      const toolEnd = /<\/tool>/i;

      // 处理 think 标签
      if (thinkEnd.test(data)) {
        n.thinkText = this.$t('agent.thinked');
        if (!thinkStart.test(data)) {
          data = '<think>\n' + data;
        }
      }

      // 新增处理 tool 标签
      if (toolEnd.test(data)) {
        n.thinkText = this.$t('agent.thinked');
        if (!toolStart.test(data)) {
          data = '<tool>\n' + data;
        }
      }

      // 统一替换为 section 标签
      return data
        .replaceAll('think>', 'section>')
        .replaceAll('tool>', 'section>');
    },
    showDSBtn(data) {
      const pattern = /<(think|tool)(\s[^>]*)?>|<\/(think|tool)>/;
      const matches = data.match(pattern);
      return !!matches;
    },
    toggle(event, index) {
      const name = event.target.className;
      if (
        name === 'deepseek' ||
        name === 'el-icon-arrow-up' ||
        name === 'el-icon-arrow-down'
      ) {
        this.session_data.history[index].isOpen =
          !this.session_data.history[index].isOpen;
        this.$set(
          this.session_data.history,
          index,
          this.session_data.history[index],
        );
      }
    },
    queryCopy(text) {
      this.$emit('queryCopy', text);
    },
    getSessionData() {
      return this.session_data;
    },
    copy(text) {
      const processedText = text.replaceAll('<br/>', '\n');
      return this.$copy(processedText);
    },
    copycb() {
      this.$message.success(this.$t('agent.copyTips'));
    },
    /**
     * 处理出处列表的展开/折叠点击事件
     * @param {Object} sourceContainer - 包含出处列表的容器对象（如主历史项或子会话对象）
     * @param {Object} searchItem - 当前点击的出处条目对象
     * @param {number} index - 当前条目在 searchList 中的索引
     */
    collapseClick(sourceContainer, searchItem, index) {
      this.$set(sourceContainer.searchList, index, {
        ...searchItem,
        collapse: !searchItem.collapse,
      });
    },
    doLoading() {
      this.loading = true;
    },
    scrollBottom() {
      this.loading = false;
      if (!this.autoScroll) return;
      this.$nextTick(() => {
        document.getElementById(this.scrollContainerId).scrollTop =
          document.getElementById(this.scrollContainerId).scrollHeight;
      });
    },

    codeScrollBottom() {
      this.$nextTick(() => {
        this.loading = false;
        document.getElementsByTagName('code').scrollTop =
          document.getElementsByTagName('code').scrollHeight;
      });
    },
    pushHistory(data) {
      this.session_data.history.push(data);
      this.scrollBottom();
    },
    replaceLastData(index, data) {
      this.$set(this.session_data.history, index, data);
      this.scrollBottom();
      this.codeScrollBottom();
      if (data.finish === 1) {
        this.$nextTick(() => {
          const setCitations = this.setCitations(index);
          this.$set(
            this.session_data.history[index],
            'citations',
            setCitations,
          );
        });
      }
    },
    getFileSizeDisplay(fileSize) {
      if (!fileSize || typeof fileSize !== 'number' || Number.isNaN(fileSize)) {
        return '...';
      }
      return fileSize > 1024
        ? `${(fileSize / (1024 * 1024)).toFixed(2)} MB`
        : `${fileSize} bytes`;
    },
    replaceData(data) {
      this.session_data = data;
      this.scrollBottom();
    },
    replaceHistory(data) {
      this.session_data.history = data;
      this.$nextTick(() => {
        this.session_data.history.forEach((n, index) => {
          const setCitations = this.setCitations(index);
          this.$set(
            this.session_data.history[index],
            'citations',
            setCitations,
          );
        });
        this.scrollBottom();
      });
    },
    removeLastHistory() {
      this.session_data.history.pop();
    },
    replaceHistoryWithImg(data) {
      this.session_data.history = data;
      this.$nextTick(() => {
        this.preTagging(data[0].annotation);
      });
    },
    clearData() {
      this.session_data = {
        tool: '',
        searchList: [],
        history: [],
        response: '',
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
      if (this.sessionStatus === 0) return;
      this.$emit('clearHistory');
    },
    getList() {
      return structuredClone(
        this.session_data.history.filter(item => {
          delete item.operation;
          return item;
        }),
      );
    },
    getAllList() {
      return structuredClone(this.session_data.history);
    },
    stopLoading() {
      this.session_data.history = this.session_data.history.filter(item => {
        return !item.pending;
      });
    },
    stopPending() {
      this.session_data.history = this.session_data.history.filter(item => {
        if (item.pending) {
          item.responseLoading = false;
          item.pendingResponse = this.$t('app.stopStream');
        }
        return item;
      });
    },
    refresh() {
      if (this.sessionStatus === 0) {
        return;
      }
      this.$emit('refresh');
    },
    preStop() {
      if (this.sessionStatus === 0) {
        this.$emit('preStop');
      }
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
    initCanvasUtil() {
      this.canvasShow = true;
      this.$nextTick(() => {
        this.cv &&
          this.cv.destroy() &&
          this.cv.clearPre() &&
          this.cv.clearLabels() &&
          (this.cv = null);
        this.cv = new CanvasUtil(this);
      });
    },
    preTagging(response) {
      this.currImg = {
        url: '',
        width: 0,
        height: 0,
        w: 0,
        h: 358,
        roteX: 0,
        roteY: 0,
        dx: 0,
        dy: 0,
      };
      let image = new Image();
      image.src = response.annotationImg;
      image.onload = () => {
        this.currImg.width = image.width;
        this.currImg.height = image.height;
        this.c = document.getElementById('mycanvas');
        this.ctx = this.c.getContext('2d');
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
      let contain = document.getElementById('mycantain');
      if (currImg.width > contain.offsetWidth) {
        this.currImg.roteX = currImg.width / contain.offsetWidth;
        currImg.w = contain.offsetWidth;
        currImg.h = (currImg.height * contain.offsetWidth) / currImg.width;
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
        currImg.roteY = currImg.height / currImg.h;
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
    // 初始化 history 列表（模型体验）
    initHistoryList(list) {
      this.$set(this.session_data, 'history', list);
      this.$nextTick(() => {
        this.updateAllFileScrollStates();
      });
      // 组件重建后内容渲染可能需要更长时间，设置滚动到底部
      clearTimeout(this.scrollTimeout);
      this.scrollTimeout = setTimeout(() => {
        this.autoScroll = true;
        this.scrollBottom();
      }, 100);
    },
    handleGlobalClick(e) {
      // 复制
      if (e.target.classList.contains('copy-btn')) {
        const btn = e.target;
        if (this.copyTimerMap.has(btn)) {
          clearTimeout(this.copyTimerMap.get(btn));
        }
        let innerText = btn.dataset.clipboardText
          ? decodeURIComponent(btn.dataset.clipboardText)
          : btn.parentNode.nextElementSibling.innerText;
        this.copy(innerText);
        this.$message.success(this.$t('agent.copyTips'));
        btn.innerText = this.$t('agent.copySuccess');
        const timerId = setTimeout(() => {
          btn.innerText = this.$t('agent.copy');
          this.copyTimerMap.delete(btn);
        }, 1500);
        this.copyTimerMap.set(btn, timerId);
      }
    },
    // 获取子会话数据
    findSubData(n, id) {
      if (!n.subConversions) return null;
      return n.subConversions.find(sub => sub.id === id);
    },
    // 检测是否包含新式 Agent 知识库子会话
    hasNewAgentKnowledge(n) {
      if (!n.subConversions || !n.subConversions.length) return false;
      return n.subConversions.some(
        sub =>
          sub.conversationType ===
          AGENT_MESSAGE_CONFIG.AGENT_KNOWLEDGE.CONVERSATION_TYPE,
      );
    },
    /**
     * 解析子会话引用的“数据宿主”和“展示宿主”。
     * 当前规则限定在 subText 所属的父级 subAgent 范围内：
     * 1. 若当前子会话自己持有 searchList，则直接使用自己；
     * 2. 否则只回退到直属父级 subAgent，且不再继续向更高祖先查找；
     * 3. 若父级 subAgent 下存在可命中的 agentKnowledge，则优先将其作为展示宿主。
     *
     * @param {Object} historyItem - 当前历史消息项，包含完整的 subConversions 列表
     * @param {string} subId - 被点击引用所属的子会话 id（通常是 subText 或普通子会话）
     * @param {number} citationIndex - 引用序号，和 searchList 下标一一对应（需 -1 取值）
     * @returns {{dataOwner: Object, displayOwner: Object} | null}
     * dataOwner 负责提供 searchList 数据；displayOwner 负责实际展开和滚动定位。
     */
    resolveCitationSourceConversion(historyItem, subId, citationIndex) {
      if (
        !historyItem ||
        !Array.isArray(historyItem.subConversions) ||
        !subId ||
        !citationIndex
      ) {
        return null;
      }

      const currentSub = historyItem.subConversions.find(
        sub => sub.id === subId,
      );
      if (!currentSub) {
        return null;
      }

      if (
        Array.isArray(currentSub.searchList) &&
        currentSub.searchList[citationIndex - 1]
      ) {
        return {
          dataOwner: currentSub,
          displayOwner: currentSub,
        };
      }

      if (!currentSub.parentId) {
        return null;
      }

      const parentSubAgent = historyItem.subConversions.find(
        sub => sub.id === currentSub.parentId,
      );
      if (
        !parentSubAgent ||
        !Array.isArray(parentSubAgent.searchList) ||
        !parentSubAgent.searchList[citationIndex - 1]
      ) {
        return null;
      }

      const knowledgeSub = historyItem.subConversions.find(
        sub =>
          sub.parentId === parentSubAgent.id &&
          sub.conversationType ===
            AGENT_MESSAGE_CONFIG.AGENT_KNOWLEDGE.CONVERSATION_TYPE &&
          Array.isArray(sub.searchList) &&
          sub.searchList[citationIndex - 1],
      );

      return {
        dataOwner: parentSubAgent,
        displayOwner: knowledgeSub || parentSubAgent,
      };
    },
    /**
     * 滚动到普通子会话出处列表中的指定条目。
     * 适用于 search-list 结构的子会话，不处理 knowledge-item 结构；
     * knowledge 子会话使用 scrollToKnowledgeCitation。
     *
     * @param {string} subId - 目标子会话 id，用于定位 .sub-conversion-item 容器
     * @param {number} citationIndex - 引用序号，对应 .search-list-item[data-citation-index]
     */
    scrollToSubConversionCitation(subId, citationIndex) {
      if (!subId || !citationIndex) return;

      this.$nextTick(() => {
        const container = document.querySelector(
          `.sub-conversion-item[data-sub-id="${subId}"]`,
        );
        if (!container) return;

        const targetSearchItem = container.querySelector(
          `.search-list-item[data-citation-index="${citationIndex}"]`,
        );
        if (!targetSearchItem) return;

        targetSearchItem.scrollIntoView({
          behavior: 'smooth',
          block: 'center',
        });
      });
    },
    scrollToKnowledgeCitation(subId, citationIndex) {
      if (!subId || !citationIndex) return;

      this.$nextTick(() => {
        const container = document.querySelector(
          `.sub-conversion-item[data-sub-id="${subId}"]`,
        );
        if (!container) return;

        const targetKnowledgeItem = container.querySelector(
          `.knowledge-item[data-index="${citationIndex - 1}"]`,
        );
        if (!targetKnowledgeItem) return;

        targetKnowledgeItem.scrollIntoView({
          behavior: 'smooth',
          block: 'center',
        });
      });
    },
    /**
     * 引用角标悬浮入口。
     * 通过历史容器的事件委托命中 .citation，再按 chatType 从 Vue 状态中解析对应来源，
     * 避免把标题和片段内容重复塞入 data-* 属性。
     *
     * @param {MouseEvent} e - 容器 mouseover 事件
     */
    handleCitationHover(e) {
      if (!['agent', 'rag'].includes(this.chatType)) return;
      const target =
        e.target && e.target.closest ? e.target.closest('.citation') : null;
      if (!target) return;

      const tipData = this.resolveCitationTipData(target);
      if (!tipData) {
        this.hideRagCitationTip();
        return;
      }

      this.showCitationTip(target, tipData);
    },
    /**
     * 引用角标移出入口。
     * 鼠标在多个引用角标之间移动时不立即隐藏，交给下一个 hover 覆盖；
     * 真正离开引用角标后隐藏当前气泡。
     *
     * @param {MouseEvent} e - 容器 mouseout 事件
     */
    handleCitationLeave(e) {
      if (!['agent', 'rag'].includes(this.chatType)) return;
      const target =
        e.target && e.target.closest ? e.target.closest('.citation') : null;
      if (!target) return;

      const related = e.relatedTarget;
      if (related && related.closest && related.closest('.citation')) {
        return;
      }
      this.hideRagCitationTip();
    },
    /**
     * 按当前会话类型分发引用 tooltip 数据解析逻辑。
     *
     * @param {HTMLElement} citationTarget - 被悬浮的 .citation 元素
     * @returns {{number: string, title: string, snippet: string} | null}
     */
    resolveCitationTipData(citationTarget) {
      if (this.chatType === 'rag') {
        return this.resolveRagCitationTipData(citationTarget);
      }
      if (this.chatType === 'agent') {
        return this.resolveAgentCitationTipData(citationTarget);
      }
      return null;
    },
    /**
     * 根据 RAG 引用角标解析 tooltip 展示数据。
     * RAG 角标只保留消息索引和引用序号，标题/片段从当前 historyItem.searchList 中按序号读取。
     *
     * @param {HTMLElement} citationTarget - 被悬浮的 .citation 元素
     * @returns {{number: string, title: string, snippet: string} | null}
     */
    resolveRagCitationTipData(citationTarget) {
      const parentsIndex = Number(citationTarget.dataset.parentsIndex);
      const citationIndex = Number(citationTarget.textContent.trim());
      if (
        !Number.isInteger(parentsIndex) ||
        !citationIndex ||
        citationIndex < 1
      ) {
        return null;
      }

      const historyItem = this.session_data.history[parentsIndex];
      const source =
        historyItem &&
        historyItem.searchList &&
        historyItem.searchList[citationIndex - 1];
      if (!source) return null;

      return {
        number: citationTarget.textContent.trim(),
        title: this.cleanCitationTipText(
          source.title ||
            source.file_name ||
            source.fileName ||
            source.user_kb_name ||
            source.link ||
            '',
          80,
        ),
        snippet: this.cleanCitationTipText(
          source.snippet || source.content || '',
          160,
        ),
      };
    },
    /**
     * 根据智能体引用角标解析 tooltip 展示数据。
     *
     * 查找优先级：
     * 1. 子会话引用：通过 data-pid + 引用序号复用 resolveCitationSourceConversion；
     * 2. 主回答引用：优先查 agentKnowledge 子会话的 searchList；
     * 3. 兜底查当前 historyItem.searchList。
     *
     * @param {HTMLElement} citationTarget - 被悬浮的 .citation 元素
     * @returns {{number: string, title: string, snippet: string} | null}
     */
    resolveAgentCitationTipData(citationTarget) {
      const parentsIndex = Number(citationTarget.dataset.parentsIndex);
      const citationIndex = Number(citationTarget.textContent.trim());
      if (
        !Number.isInteger(parentsIndex) ||
        !citationIndex ||
        citationIndex < 1
      ) {
        return null;
      }

      const historyItem = this.session_data.history[parentsIndex];
      if (!historyItem) return null;

      let source = null;
      const pid = citationTarget.dataset.pid;

      if (pid) {
        const citationContext = this.resolveCitationSourceConversion(
          historyItem,
          pid,
          citationIndex,
        );
        source =
          citationContext &&
          citationContext.dataOwner &&
          citationContext.dataOwner.searchList &&
          citationContext.dataOwner.searchList[citationIndex - 1];
      } else if (this.hasNewAgentKnowledge(historyItem)) {
        const knowledgeSub = historyItem.subConversions.find(
          sub =>
            sub.conversationType ===
            AGENT_MESSAGE_CONFIG.AGENT_KNOWLEDGE.CONVERSATION_TYPE,
        );
        source =
          knowledgeSub &&
          knowledgeSub.searchList &&
          knowledgeSub.searchList[citationIndex - 1];
      }

      if (!source && historyItem.searchList) {
        source = historyItem.searchList[citationIndex - 1];
      }
      if (!source) return null;

      return {
        number: citationTarget.textContent.trim(),
        title: this.cleanCitationTipText(
          source.title ||
            source.file_name ||
            source.fileName ||
            source.user_kb_name ||
            source.link ||
            '',
          80,
        ),
        snippet: this.cleanCitationTipText(
          source.snippet ||
            source.content ||
            [source.question, source.answer].filter(Boolean).join('\n\n') ||
            '',
          160,
        ),
      };
    },
    /**
     * 清洗并截断引用 tooltip 文本。
     * 用于把来源中的 Markdown/HTML 片段转为适合气泡展示的纯文本摘要。
     *
     * @param {string} value - 原始标题或片段内容
     * @param {number} maxLength - 最大展示字符数
     * @returns {string}
     */
    cleanCitationTipText(value, maxLength = 160) {
      const raw = String(value || '')
        .replaceAll(/!\[[^\]]*\]\([^)]+\)/g, '')
        .replaceAll(/\[([^\]]+)\]\([^)]+\)/g, '$1')
        .replaceAll(/<\/?[a-zA-Z][^>]*>/g, '')
        .replaceAll(/\s+/g, ' ')
        .trim();
      if (!raw) return '';

      const textarea = document.createElement('textarea');
      textarea.innerHTML = raw;
      const decoded = textarea.value.replaceAll(/\s+/g, ' ').trim();
      return decoded.length > maxLength
        ? decoded.slice(0, maxLength) + '…'
        : decoded;
    },
    /**
     * 在指定引用角标附近展示引用 tooltip。
     * 复用现有 ragCitationTip 状态和气泡样式，只负责定位与赋值。
     *
     * @param {HTMLElement} target - 被悬浮的引用角标
     * @param {{title?: string, snippet?: string, number?: string}} tipData
     */
    showCitationTip(target, { title = '', snippet = '', number = '' } = {}) {
      if (this._ragCitationTipHideTimer) {
        clearTimeout(this._ragCitationTipHideTimer);
        this._ragCitationTipHideTimer = null;
      }

      const rect = target.getBoundingClientRect();
      const TIP_HEIGHT = 140;
      const MARGIN = 8;
      const TIP_WIDTH = Math.min(320, innerWidth - MARGIN * 2);
      const TIP_HALF_WIDTH = TIP_WIDTH / 2;
      const placement = rect.top < TIP_HEIGHT + MARGIN ? 'bottom' : 'top';
      const rawX = rect.left + rect.width / 2;
      let x = rawX;
      x = Math.min(
        Math.max(x, TIP_HALF_WIDTH + MARGIN),
        innerWidth - TIP_HALF_WIDTH - MARGIN,
      );
      const y = placement === 'top' ? rect.top - MARGIN : rect.bottom + MARGIN;
      const arrowOffset = rawX - x;

      this.ragCitationTip = {
        visible: true,
        x,
        y,
        placement,
        title,
        snippet,
        number,
        arrowOffset,
      };
    },
    /**
     * 立即隐藏引用气泡并清掉未触发的定时器。
     * 在 mouseleave / 容器滚动 / 窗口 resize / 整页 wheel 等时机都被调用。
     */
    onWindowWheelHideTip() {
      // 独立方法而非 inline 箭头：保证 add/remove 用同一引用
      this.hideRagCitationTip();
    },
    hideRagCitationTip() {
      if (this._ragCitationTipHideTimer) {
        clearTimeout(this._ragCitationTipHideTimer);
        this._ragCitationTipHideTimer = null;
      }
      if (this.ragCitationTip.visible) {
        this.ragCitationTip.visible = false;
      }
    },
    /**
     * RAG 引用片段展开/收起：key = `${messageIdx}-${sourceIdx}`。
     */
    toggleRagSnippet(key) {
      this.$set(this.ragSnippetExpanded, key, !this.ragSnippetExpanded[key]);
    },
    /**
     * RAG 引用片段溢出检测：DOM 稳定后对比 scrollHeight/clientHeight，标记是否需要"展开全文"按钮。
     * 缓存已检测过的 key，避免展开态 toggle class 后回流导致检测结果翻转，也省掉重复 DOM 读。
     * 调用时机：updated() 钩子里 $nextTick 后触发——snippet 到达那次渲染会命中，之后的文本流式更新早退。
     */
    checkRagSnippetOverflow() {
      const refs = this.$refs || {};
      Object.keys(refs).forEach(name => {
        if (!name.startsWith('ragSnippet_')) return;
        const key = name.slice('ragSnippet_'.length).replace('_', '-');
        if (key in this.ragSnippetOverflow) return;
        const entry = refs[name];
        const el = Array.isArray(entry) ? entry[0] : entry;
        if (!el) return;
        if (!el.clientHeight) return; // 元素隐藏中，等下次 updated 再检测，不缓存
        const isOverflow = el.scrollHeight > el.clientHeight + 1;
        this.$set(this.ragSnippetOverflow, key, isOverflow);
      });
    },

    /**
     * 判断该消息是否已经开始流式输出"正式回答"。
     * 用于驱动 thinking 卡片在回答开始的同一帧收起。
     */
    hasFinalAnswerStarted(n) {
      if (!n) return false;
      return !!(
        (n.stableChunks && n.stableChunks.length) ||
        n.activeResponse ||
        n.response
      );
    },
    /**
     * 思考过程里 LLM 常把 ![image](url) 放在反引号里 → markdown-it 解析为 <code>...</code>
     * 识别这种 <code>![alt](url)</code> 还原成 <img>，否则图片永远出不来。
     * 仅用于 RAG thinking 卡片（不影响其他地方的代码高亮）。
     */
    unwrapCodeImages(html) {
      if (!html || typeof html !== 'string') return html || '';
      // URL 白名单：只允许 http(s)、data:image/、或根路径，避免 javascript:/vbscript: 等 XSS 向量
      const SAFE_URL_RE = /^(https?:\/\/|data:image\/|\/)/i;
      // alt 文本转义：防止 alt="..."><script>... 之类的属性注入
      const escapeAttr = s =>
        String(s)
          .replaceAll('&', '&amp;')
          .replaceAll('<', '&lt;')
          .replaceAll('>', '&gt;')
          .replaceAll('"', '&quot;');
      const replacer = (match, inner) => {
        const m = inner.match(/^\s*!\[([^\]]*)\]\(([^)\s]+)\)\s*$/);
        if (!m) return match;
        const url = m[2].replaceAll('&amp;', '&');
        if (!SAFE_URL_RE.test(url)) return match;
        const alt = escapeAttr(m[1]);
        // url 也转义引号以阻断属性破坏（& < > 通常在 url 里是合法字符，仅转 "）
        const safeUrl = url.replaceAll('"', '&quot;');
        return `<img src="${safeUrl}" alt="${alt}" />`;
      };
      // 1) 块级缩进代码：<pre><code>![image](url)</code></pre>
      let out = html.replaceAll(
        /<pre(?:\s[^>]*)?>\s*<code(?:\s[^>]*)?>([^<]+)<\/code>\s*<\/pre>/g,
        replacer,
      );
      // 2) 行内反引号代码：<code>![image](url)</code>
      out = out.replaceAll(/<code(?:\s[^>]*)?>([^<]+)<\/code>/g, replacer);
      // 3) 占位 <img>：src 不是合法 URL（http/https/data/ 根路径） → 还原为字面文本
      //    用于 LLM 在思考中写 ![title](url) 这种"语法示例"的情况，避免显示破图标
      out = out.replaceAll(
        /<img\b[^>]*?\bsrc="([^"]*)"[^>]*>/gi,
        (match, src) => {
          if (/^(https?:\/\/|data:|\/)/i.test(src)) return match;
          const altMatch = match.match(/\balt="([^"]*)"/i);
          const alt = altMatch ? altMatch[1] : '';
          // 解码再编码，避免 &amp; 之类丢失
          const decode = s =>
            s
              .replaceAll('&lt;', '<')
              .replaceAll('&gt;', '>')
              .replaceAll('&quot;', '"')
              .replaceAll('&#39;', "'")
              .replaceAll('&amp;', '&');
          const escape = s =>
            s
              .replaceAll('&', '&amp;')
              .replaceAll('<', '&lt;')
              .replaceAll('>', '&gt;');
          return escape(`![${decode(alt)}](${decode(src)})`);
        },
      );
      return out;
    },
    /**
     * 点击 RAG 回答中的引用角标 → 展开 knowledge_search 卡片 → 滚到对应 source → 高亮闪烁
     * @param {number} parentsIndex - 消息下标（对应 #message-container{i}）
     * @param {number} citationIndex - 引用序号（从 1 开始）
     */
    scrollToRagCitation(parentsIndex, citationIndex) {
      if (citationIndex == null) return;
      const container = document.querySelector(
        `#message-container${parentsIndex}`,
      );
      if (!container) return;

      const card = container.querySelector(
        '[data-rag-step-type="knowledge_search"]',
      );
      if (!card) return;

      // 若处于收起态，先点击 header 展开
      const body = card.querySelector('.rag-step-body');
      const bodyVisible = body && body.offsetHeight > 0;
      if (!bodyVisible) {
        const header = card.querySelector('.rag-step-header');
        if (header) header.click();
      }

      // 过渡结束再滚动 & 高亮（collapse-transition 约 300ms）
      setTimeout(
        () => {
          const target = card.querySelector(
            `.rag-source-item[data-citation-index="${citationIndex}"]`,
          );
          if (!target) return;
          target.scrollIntoView({ behavior: 'smooth', block: 'center' });
          target.classList.remove('rag-source-flash');
          // 强制重绘以重放动画
          // eslint-disable-next-line no-unused-expressions
          target.offsetWidth;
          target.classList.add('rag-source-flash');
          setTimeout(() => target.classList.remove('rag-source-flash'), 1600);
        },
        bodyVisible ? 0 : 320,
      );
    },
    // 动态设置滚动容器高度
    setHistoryBoxHeight(inputHeight) {
      if (inputHeight) {
        const baseInputHeight = 56;
        const offset = Math.max(0, inputHeight - baseInputHeight);
        this.historyBoxHeight = `calc(100% - ${46 + offset}px)`;
      } else {
        this.historyBoxHeight = '';
      }
      this.scrollBottom();
    },
    // 引用结果标题点击
    handleSourceTitleClick(n, m, j, i) {
      if (n.subConversions && n.subConversions.length > 0) {
        // 打开子会话
        n.subConversions.forEach(sub => {
          if (
            sub.conversationType ===
            AGENT_MESSAGE_CONFIG.AGENT_KNOWLEDGE.CONVERSATION_TYPE
          ) {
            this.$set(sub, 'isOpen', true);
          }
        });

        // 滚动到指定位置
        this.$nextTick(() => {
          const container = document.getElementById('message-container' + i);
          if (container) {
            const target = container.querySelector(
              `.knowledge-item[data-index="${j}"]`,
            );
            if (target) {
              target.scrollIntoView({
                behavior: 'smooth',
                block: 'center',
              });
            }
          }
        });
      }
    },
    // 删除单条会话
    async handleDelConversation(n) {
      if (n.detailId || n.id) {
        this.$emit('delConversationQA', n.detailId || n.id);
      }
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
    font-style: normal;
  }
}

::v-deep {
  pre {
    white-space: pre-wrap !important;
    min-height: 50px;
    word-wrap: break-word;
    padding: 0;
    background: none;
    &.hljs {
      resize: vertical;
    }
    .hljs {
      max-height: 300px !important;
      white-space: pre-wrap !important;
      min-height: 50px;
      word-wrap: break-word;
      resize: vertical;
      color: #abb2bf;
      background: #282c34;
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
      max-width: 100%;
      max-height: 400px;
      height: auto;
      object-fit: contain;
      cursor: zoom-in;
    }
    section li,
    li {
      list-style-position: inside !important; /* 将标记符号放在内容框内 */
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
  display: flex;
  flex-direction: column;
  .session-item {
    min-height: 80px;
    display: flex;
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
          background: linear-gradient(135deg, #7c8cff 0%, #6171e6 100%);
          color: #fff;
          padding: 7px 12px 7px 14px;
          border-radius: 12px 4px 12px 12px;
          margin: 0 !important;
          line-height: 1.5;
          font-weight: 400;
          box-shadow:
            0 2px 6px rgba(97, 113, 230, 0.25),
            inset 0 1px 0 rgba(255, 255, 255, 0.18);
          letter-spacing: 0.2px;
          transition:
            box-shadow 0.2s ease,
            transform 0.2s ease;
          &:hover {
            box-shadow:
              0 4px 12px rgba(97, 113, 230, 0.32),
              inset 0 1px 0 rgba(255, 255, 255, 0.2);
          }
        }
        .session-setting-id {
          color: rgba(98, 98, 98, 0.5);
          font-size: 12px;
          margin-top: -8px;
        }
        .echo-doc-box {
          margin-bottom: 10px;
          width: 100%;
          max-width: 100%;
          display: flex;
          gap: 8px;
          justify-content: space-between;
          align-items: center;
          position: relative;
          .scroll-btn {
            position: absolute;
            top: 50%;
            transform: translateY(-15px);
            &.left {
              left: 5px;
            }
            &.right {
              right: 5px;
            }
          }
          .imgList {
            width: 100%;
            gap: 10px;
            overflow-x: hidden;
            scroll-behavior: smooth;
            display: flex;
            flex-wrap: nowrap;
          }
          .docInfo-container {
            display: flex;
            align-items: center;
            background: #fff;
            border: 1px solid rgb(235, 236, 238);
            padding: 5px 10px 5px 5px;
            border-radius: 5px;
          }
          .docInfo-img-container {
            flex-shrink: 0; /* 防止图片被压缩 */
            // 单张图片
            &:first-child:last-child {
              width: 100%;
              ::v-deep .el-image {
                width: auto !important;
                height: auto !important;
                max-width: 100%;
                display: block;
                float: right;
                border-radius: 6px;

                .el-image__inner {
                  width: 100% !important;
                  height: 100% !important;
                }
              }
            }
            // 多张图片
            &:not(:first-child:last-child) {
              width: auto;
              ::v-deep .el-image {
                width: 70px !important;
                height: 70px !important;
                display: block;
                border-radius: 6px;

                .el-image__inner {
                  width: 100% !important;
                  height: 100% !important;
                  object-position: left top;
                }
              }
            }
            p {
              text-align: center;
              color: $color;
              font-size: 12px;
            }
          }
          .docIcon {
            flex: 0 0 auto;
            width: 30px;
            height: 30px;
          }
          .docInfo {
            margin-left: 5px;
            .docInfo_name {
              color: #333;
            }
            .docInfo_size {
              color: #bbbbbb;
              text-align: left !important;
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
      .qa_content {
        display: flex;
        gap: 10px;
        margin-top: 5px;
      }
      .recommended-question-title {
        border-bottom: 1px solid #e5e5e5;
        padding: 5px 0;
      }
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
          ::v-deep img {
            cursor: zoom-in;
          }
        }
      }
    }
    /*操作*/
    .answer-operation {
      display: flex;
      align-items: center;
      padding: 5px 20px 15px 63px;
      color: #777;
      .opera-left {
        .restart,
        .preStop {
          cursor: pointer;
          img {
            width: 20px;
            height: 20px;
            padding: 2px;
          }
        }
      }
      .opera-right {
        cursor: pointer;
        display: inline-flex;
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
      .answer-operation-tip {
        padding-bottom: 4px;
        font-size: 12px;
        color: #999;
      }
    }
  }

  /*图片*/
  .file-path {
    .el-image {
      height: 200px !important;
      background-color: #f9f9f9;
      ::v-deep.el-image__inner,
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

  // 与 ragStepCard 视觉风格对齐：12px 圆角 + 渐变底 + 细边 + 微阴影；
  // 用红色系替换原紫色系以保留错误语义。

  .history-box {
    height: calc(100% - 46px);
    flex: 1;
    overflow-y: auto !important;
    padding: 20px 4px;
  }
  /*删除历史...*/
  .session-setting {
    position: relative;
    height: 36px;
    right: 50px;
    .right-setting {
      position: absolute;
      right: 10px;
      top: -5px;
      color: #ff2324;
      font-size: 16px;
      cursor: pointer;
      ::v-deep {
        .el-dropdown-menu {
          width: 100px;
        }
        .el-dropdown-menu__item {
          padding: 0 15px !important;
        }
      }
      &.is-disabled {
        opacity: 0.5;
        cursor: not-allowed;
        pointer-events: none;
      }
    }
  }

  .think_icon {
    width: 12px !important;
    height: 12px !important;
    margin-right: 3px;
  }
  .ds-res {
    ::v-deep section {
      color: #8b8b8b;
      position: relative;
      font-size: 12px;
      * {
        font-size: 12px;
      }
    }
    ::v-deep section::before {
      content: '';
      position: absolute;
      height: 100%;
      width: 1px;
      background: #ddd;
      left: -8px;
    }
    ::v-deep .hideDs {
      display: none;
    }
  }

  .deepseek {
    font-size: 13px;
    color: #8b8b8b;
    font-weight: bold;
    margin: 0 0 10px 6px;
    cursor: pointer;
    display: inline-block;
  }

  .sub-conversion-box {
    border-radius: 8px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
}
/* 仅通过样式调整位置：
   问题在右侧（内容在右、头像在最右），答案在左侧（默认） */
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
    gap: 10px; /* 头像和内容之间10px距离 */
    padding: 20px 20px 0 20px;
    min-height: 80px;
    background: none; /* 确保外层容器无背景色 */

    .logo {
      width: 30px;
      height: 30px;
      border-radius: 6px;
      object-fit: cover;
      flex-shrink: 0; /* 防止头像被压缩 */
      background: none; /* 头像无背景色 */
    }

    .answer-content {
      flex: 1;
      background-color: #eceefe; /* 只有内容区域有背景色 */
      border-radius: 0 10px 10px 10px;
      padding: 20px;
      line-height: 1.6;

      ::v-deep table {
        border-collapse: collapse;
        margin: 12px 0;
        width: 100%;
        font-size: 13px;

        th,
        td {
          border: 1px solid #e5e7eb;
          padding: 7px 12px;
          text-align: left;
        }

        tr {
          background: rgba(99, 102, 241, 0.06);
          &:nth-child(2n) {
            background: rgba(99, 102, 241, 0.12);
          }
        }

        th {
          background: rgba(99, 102, 241, 0.06);
          color: #1f2937;
          font-weight: 600;
        }
      }
    }
  }
}

/* 图片加载失败时的样式 */
img.failed {
  position: relative;
  border: 2px dashed #ff6b6b;
  background-color: #fff5f5;
  opacity: 0.5;
}

img.failed::after {
  content: '图片加载失败';
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
  width: 54px;
  height: 18px;
  margin: 6px 0 0 55px;
}

.text-loading.la-dark {
  color: #e8e8e8;
}

.text-loading > div {
  display: inline-block;
  float: none;
  background-color: currentColor;
  border: 0 solid currentColor;
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

.session-section-wrapper {
  padding: 5px 20px 15px 63px;
}

.recommend-question {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
  &-item {
    display: inline-flex;
    font-size: 12px;
    padding: 4px 6px;
    background: #f2f2f2;
    cursor: pointer;
    border-radius: 8px;
    align-items: center;
    line-height: 14px;
    &:hover {
      background: #eceefe;
    }
    &.is-tips {
      cursor: default;
      opacity: 0.7;
      &:hover {
        background: #f2f2f2;
      }
    }
  }
  &-loading {
    margin: 0;
  }
}

.message-sequence-wrapper {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.reasoning-area {
  margin-bottom: 8px;
}

/* RAG 知识库检索卡片内的 source 列表（紫色系，与 thinking 卡片一致） */
.rag-source-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

// 思考过程块：支持图片预览
.chunk_stable,
.chunk_active {
  ::v-deep img {
    cursor: zoom-in;
  }
}
.rag-source-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  background: #fff;
  border: 1px solid rgba(99, 102, 241, 0.12);
  border-radius: 8px;
}
.rag-source-header {
  display: flex;
  align-items: center;
  gap: 8px;

  .rag-source-index {
    flex-shrink: 0;
    width: 18px;
    height: 18px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid rgba(99, 102, 241, 0.5);
    color: #4f46e5;
    border-radius: 50%;
    font-size: 12px;
    font-weight: 500;
    line-height: 1;
  }
  .rag-source-title {
    flex: 1;
    min-width: 0;
    font-size: 14px;
    font-weight: 500;
    color: #303133;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .rag-source-download {
    flex-shrink: 0;
    color: #6b7280;
    font-size: 14px;
    text-decoration: none;

    &:hover {
      color: #4f46e5;
    }
  }
}
.rag-source-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;

  .rag-source-pill {
    display: inline-block;
    padding: 2px 10px;
    border-radius: 20px;
    font-size: 12px;
    line-height: 1.5;
    white-space: nowrap;
    background: rgba(99, 102, 241, 0.08);
    color: #4f46e5;
  }
  .rag-source-kb {
    max-width: calc(100% * 2 / 3);
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
.rag-source-snippet {
  font-size: 13px;
  line-height: 1.6;
  color: #606266;
  word-break: break-word;

  // 折叠态：限 6 行 + 末尾省略号；展开态去掉所有限制，让 v-html 内容完整显示
  &.is-collapsed {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 6;
    line-clamp: 6;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  ::v-deep img {
    max-width: 100%;
    max-height: 180px;
    height: auto;
    object-fit: contain;
    border-radius: 4px;
    cursor: zoom-in;
  }
}
.rag-source-expand-btn {
  margin-top: 4px;
  font-size: 12px;
  color: #4f46e5;
  cursor: pointer;
  display: inline-block;
  user-select: none;

  &:hover {
    opacity: 0.8;
  }
}
.rag-source-empty {
  color: #9ca3af;
  font-size: 13px;
  font-style: italic;
  padding: 6px 0;
}
.rag-source-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  color: #6b7280;
  font-size: 13px;

  i.el-icon-loading {
    color: #4f46e5;
    font-size: 14px;
  }
}

/* ===== RAG 回答区视觉升级（仅 rag 作用域） ===== */
.answer-content.rag-answer {
  background: #ffffff !important;
  border: 1px solid rgba(99, 102, 241, 0.14);
  border-radius: 4px 14px 14px 14px !important;
  padding: 18px 22px !important;
  box-shadow:
    0 6px 24px -12px rgba(79, 70, 229, 0.18),
    0 2px 6px -2px rgba(0, 0, 0, 0.04);
  color: #1f2937;
  font-size: 15px;
  line-height: 1.75;

  ::v-deep {
    p {
      margin: 9px 0;
      word-break: break-word;
    }
    p:first-child {
      margin-top: 0;
    }
    p:last-child {
      margin-bottom: 0;
    }
    // 段落首元素为 img（后可跟引用角标 sup）时收缩到图片宽度，消除白色边框感
    p:has(> img:first-child) {
      display: table;
    }

    strong {
      color: #111827;
      font-weight: 600;
    }

    code {
      background: rgba(99, 102, 241, 0.08);
      color: #4f46e5;
      padding: 1px 6px;
      border-radius: 4px;
      font-size: 13px;
      font-family:
        'SFMono-Regular', 'JetBrains Mono', Menlo, Consolas, monospace;
    }
    pre {
      background: #0f172a;
      color: #e2e8f0;
      padding: 14px 16px;
      border-radius: 10px;
      overflow-x: auto;
      font-size: 13px;
      line-height: 1.6;
      margin: 12px 0;
    }
    pre code {
      background: transparent;
      color: inherit;
      padding: 0;
    }

    ul,
    ol {
      padding-left: 22px;
      margin: 10px 0;
    }
    li {
      margin: 4px 0;
    }

    blockquote {
      margin: 12px 0;
      padding: 8px 14px;
      border-left: 3px solid rgba(99, 102, 241, 0.5);
      background: rgba(99, 102, 241, 0.04);
      color: #4b5563;
      border-radius: 0 6px 6px 0;
    }

    a {
      color: #4f46e5;
      text-decoration: none;
      &:hover {
        text-decoration: underline;
      }
    }

    img {
      max-width: 70%;
      max-height: 280px;
      width: auto;
      height: auto;
      object-fit: contain;
      display: block;
      margin: 14px 0;
      border-radius: 8px;
      box-shadow:
        0 4px 16px -6px rgba(0, 0, 0, 0.18),
        0 1px 3px rgba(0, 0, 0, 0.06);
      cursor: zoom-in;
      transition:
        transform 0.25s cubic-bezier(0.22, 1, 0.36, 1),
        box-shadow 0.25s ease;
      &:hover {
        transform: translateY(-1px) scale(1.005);
        box-shadow:
          0 10px 28px -10px rgba(0, 0, 0, 0.24),
          0 2px 6px rgba(0, 0, 0, 0.08);
      }
    }

    // 引用角标：14×14 小圆点上标，font 9px，与正文不抢眼
    .citation {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 14px;
      height: 14px;
      min-width: 14px;
      padding: 0;
      margin: 0 1px;
      border-radius: 50%;
      background: rgba(99, 102, 241, 0.08);
      border: 1px solid rgba(99, 102, 241, 0.28);
      color: #4f46e5;
      font-size: 9px;
      font-weight: 600;
      line-height: 1;
      text-align: center;
      cursor: pointer;
      vertical-align: super;
      position: relative;
      top: 1px;
      transform: none !important;
      transition:
        background 0.18s ease,
        color 0.18s ease,
        border-color 0.18s ease,
        box-shadow 0.18s ease;
      text-decoration: none;

      &:hover {
        background: #4f46e5;
        color: #fff;
        border-color: #4f46e5;
        box-shadow: 0 2px 6px rgba(79, 70, 229, 0.32);
      }
    }
  }
}

/* RAG 引用 hover 气泡 */
.rag-citation-popover {
  position: fixed;
  z-index: 2050;
  transform: translate(-50%, -100%);
  width: 320px;
  max-width: calc(100vw - 16px);
  box-sizing: border-box;
  padding: 10px 12px;
  background: #ffffff;
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 10px;
  box-shadow:
    0 12px 32px -12px rgba(79, 70, 229, 0.28),
    0 2px 8px rgba(0, 0, 0, 0.06);
  pointer-events: none;
  font-size: 13px;
  line-height: 1.5;
  color: #1f2937;

  &.placement-bottom {
    transform: translate(-50%, 0);
  }

  &::after {
    content: '';
    position: absolute;
    left: calc(50% + var(--citation-arrow-offset, 0px));
    width: 10px;
    height: 10px;
    background: #ffffff;
    border: 1px solid rgba(99, 102, 241, 0.2);
    transform: translateX(-50%) rotate(45deg);
  }
  &.placement-top::after {
    bottom: -6px;
    border-top: none;
    border-left: none;
  }
  &.placement-bottom::after {
    top: -6px;
    border-bottom: none;
    border-right: none;
  }

  .rag-citation-popover-head {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 6px;

    .rag-citation-popover-num {
      flex-shrink: 0;
      min-width: 18px;
      height: 18px;
      padding: 0 5px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      background: #4f46e5;
      color: #fff;
      font-size: 11px;
      font-weight: 600;
      border-radius: 9px;
      line-height: 1;
    }
    .rag-citation-popover-title {
      flex: 1;
      min-width: 0;
      font-size: 13px;
      font-weight: 600;
      color: #111827;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
  .rag-citation-popover-snippet {
    color: #4b5563;
    font-size: 12.5px;
    line-height: 1.6;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 4;
    line-clamp: 4;
    overflow: hidden;
  }
  .rag-citation-popover-hint {
    margin-top: 6px;
    color: #6366f1;
    font-size: 11.5px;
    font-weight: 500;
  }
}
.rag-tip-fade-enter-active,
.rag-tip-fade-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.rag-tip-fade-enter,
.rag-tip-fade-leave-to {
  opacity: 0;
}
.rag-tip-fade-enter.placement-top,
.rag-tip-fade-leave-to.placement-top {
  transform: translate(-50%, calc(-100% + 4px));
}
.rag-tip-fade-enter.placement-bottom,
.rag-tip-fade-leave-to.placement-bottom {
  transform: translate(-50%, -4px);
}

/* 引用跳转后的闪烁高亮 */
.rag-source-item.rag-source-flash {
  animation: rag-source-flash-anim 1.5s ease-out;
}
@keyframes rag-source-flash-anim {
  0% {
    background: #ffffff;
    box-shadow: 0 0 0 0 rgba(99, 102, 241, 0);
  }
  20% {
    background: rgba(99, 102, 241, 0.14);
    box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.25);
  }
  100% {
    background: #ffffff;
    box-shadow: 0 0 0 0 rgba(99, 102, 241, 0);
  }
}

.del-icon {
  cursor: pointer;
  color: rgb(155, 155, 155);
  font-size: 16px;
}

.gap-10px {
  gap: 10px;
}

::v-deep .agent-error-card {
  .session-error-toggle {
    .el-icon-arrow-down {
      transform: rotate(90deg);
      &.is-expanded {
        transform: rotate(0deg);
      }
    }
  }
}
</style>
