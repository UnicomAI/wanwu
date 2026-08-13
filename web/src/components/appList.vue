<template>
  <div :class="['app-card-container', 'scroll-card-container']">
    <div class="app-card scroll-card-pr">
      <div class="smart rl smart-create" v-if="isShowTool">
        <div class="app-card-create" @click="showCreate">
          <div class="create-img-wrap">
            <img
              v-if="imgObj[type]"
              class="create-img"
              :src="imgObj[type]"
              alt=""
            />
          </div>
          <span>{{ `${$t('common.button.add')}${apptype[type] || ''}` }}</span>
        </div>
      </div>
      <template v-if="listData && listData.length">
        <div
          class="smart rl"
          v-for="(n, i) in listData"
          :key="`${i}sm`"
          :style="`cursor: ${isCanClick(n) ? 'pointer' : 'default'} !important;`"
          @click.stop="isCanClick(n) && toEdit(n)"
        >
          <!-- 类型标签 -->
          <span :class="['ribbon', conversionRibbon(n.appType, n.category)]">
            <span>{{ conversionRibbonText(n.appType, n.category) }}</span>
          </span>

          <div class="smart-card-header">
            <el-image
              v-if="n.avatar && n.avatar.path"
              class="logo"
              lazy
              :src="avatarSrc(n.avatar.path)"
              :key="`${i}-${n.appId}-avatar`"
            ></el-image>
            <div class="info-right">
              <p class="name-wrap" :title="n.name">
                <span class="name">{{ n.name }}</span>
                <i
                  v-if="isShowPublished && n.publishType"
                  class="el-icon-success published-icon"
                />
              </p>
              <!-- 未发布的开发状态显示版本号 -->
              <div v-if="isDev" class="info-tag">
                {{ n.publishType ? n.version : $t('common.state.draft') }}
              </div>
            </div>
          </div>
          <el-tooltip
            v-if="n.desc"
            popper-class="instr-tooltip tooltip-cover-arrow"
            effect="dark"
            :content="n.desc"
            placement="bottom-start"
          >
            <p class="desc">{{ n.desc }}</p>
          </el-tooltip>

          <div class="smart-card-footer">
            <span :class="['smartDate']">{{ n.createdAt }}</span>
            <div class="smart-card-footer-right">
              <div
                :class="['tags', { 'is-showTool-tags': isExploreShowTool(n) }]"
              >
                <div v-if="!isShowTool" class="favorite-wrap">
                  <el-tooltip
                    class="item"
                    effect="dark"
                    :content="n.user.userName"
                    placement="top-start"
                  >
                    <span class="user-name">
                      {{
                        n.user
                          ? n.user.userName.length > 6
                            ? n.user.userName.substring(0, 6) + '...'
                            : n.user.userName
                          : ''
                      }}
                    </span>
                  </el-tooltip>
                  <img
                    v-if="!n.isFavorite"
                    class="favorite"
                    src="@/assets/imgs/like.png"
                    alt=""
                    @click="handelMark($event, n, i)"
                  />
                  <img
                    v-else
                    class="favorite"
                    src="@/assets/imgs/like_active.png"
                    alt=""
                    @click="handelMark($event, n, i)"
                  />
                </div>
              </div>

              <div v-if="isShowPublished && n.publishType" class="publishType">
                <span
                  v-if="n.publishType === 'private'"
                  class="publishType-tag"
                >
                  <span class="el-icon-lock"></span>
                  {{ $t('appSpace.private') }}
                </span>
                <span v-else class="publishType-tag">
                  <span class="el-icon-unlock"></span>
                  {{ $t('appSpace.public') }}
                </span>
              </div>

              <div class="editor" v-if="isShowTool">
                <el-dropdown @command="handleClick($event, n)" placement="top">
                  <span class="el-dropdown-link">
                    <i class="el-icon-more icon edit-icon" @click.stop />
                  </span>
                  <el-dropdown-menu slot="dropdown">
                    <el-dropdown-item command="edit" v-if="isCanClick(n)">
                      {{ $t('common.button.edit') }}
                    </el-dropdown-item>
                    <el-dropdown-item command="delete">
                      {{ $t('common.button.delete') }}
                    </el-dropdown-item>
                    <el-dropdown-item command="copy">
                      {{ $t('common.button.copy') }}
                    </el-dropdown-item>
                    <!--不在卡片进行发布-->
                    <!--<el-dropdown-item
                      command="publish"
                      v-if="n.appType === workflow && !n.publishType"
                    >
                      {{$t('common.button.publish')}}
                    </el-dropdown-item>-->
                    <el-dropdown-item command="publishSet" v-if="n.publishType">
                      {{ $t('appSpace.publishSet') }}
                    </el-dropdown-item>
                    <el-dropdown-item
                      command="export"
                      v-if="[workflow, chat].includes(n.appType)"
                    >
                      {{ $t('common.button.export') }}
                    </el-dropdown-item>
                    <el-dropdown-item
                      command="transform"
                      v-if="[workflow, chat].includes(n.appType)"
                    >
                      {{
                        $t('common.button.transform') +
                        (n.appType === workflow
                          ? $t('appSpace.chat')
                          : $t('appSpace.workflow'))
                      }}
                    </el-dropdown-item>
                    <el-dropdown-item
                      command="convertToSkill"
                      v-if="
                        [agent, rag, workflow].includes(n.appType) &&
                        n.publishType
                      "
                    >
                      {{ $t('common.button.convertToSkill') }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </el-dropdown>
              </div>

              <div class="editor editor-explore" v-if="isExploreShowTool(n)">
                <el-dropdown @command="handleClick($event, n)" placement="top">
                  <span class="el-dropdown-link">
                    <i class="el-icon-more icon edit-icon" @click.stop />
                  </span>
                  <el-dropdown-menu slot="dropdown">
                    <el-dropdown-item command="copy">
                      {{ $t('common.button.copy') }}
                    </el-dropdown-item>
                    <el-dropdown-item command="export">
                      {{ $t('common.button.export') }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </el-dropdown>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
    <el-empty
      class="noData"
      v-if="!(listData && listData.length)"
      :description="$t('common.noData')"
    ></el-empty>
    <el-dialog
      :title="$t('list.tips')"
      :visible.sync="dialogVisible"
      width="400px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
      class="createTotalDialog"
    >
      <div style="margin-top: -20px">
        <div
          v-for="item in publishList"
          :key="item.key"
          style="margin-bottom: 5px"
        >
          <el-radio :label="item.key" v-model="publishType">
            {{ item.value }}
          </el-radio>
        </div>
        <div style="text-align: right; margin-top: 15px; margin-bottom: -10px">
          <el-button size="mini" type="primary" @click="doPublish">
            {{ $t('common.button.confirm') }}
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import {
  deleteApp,
  appCancelPublish,
  appPublish,
  copyTextQues,
  copyAgentApp,
} from '@/api/appspace';
import {
  copyWorkFlow,
  exportWorkflow,
  transformWorkflow,
} from '@/api/workflow';
import { setFavorite } from '@/api/explore';
import {
  AppType,
  AGENT,
  RAG,
  CHAT,
  WORKFLOW,
  DIGITAL_EMPLOYEE,
} from '@/utils/commonSet';
import { avatarSrc, resDownloadFile } from '@/utils/util';

export default {
  props: {
    type: String,
    showCreate: Function,
    appData: {
      type: Array,
      required: true,
    },
    isShowTool: false,
    isShowPublished: false,
    appFrom: {
      type: String,
      default: '',
    },
    isDev: false,
  },
  watch: {
    appData: {
      handler: function (val) {
        this.listData = val;
      },
      immediate: true,
      deep: true,
    },
  },
  data() {
    return {
      apptype: AppType,
      basePath: this.$basePath,
      workflow: WORKFLOW,
      chat: CHAT,
      agent: AGENT,
      rag: RAG,
      digitalEmployee: DIGITAL_EMPLOYEE,
      listData: [],
      row: {},
      publishType: 'private',
      dialogVisible: false,
      publishList: [
        { key: 'private', value: this.$t('workflow.publishText') },
        { key: 'organization', value: this.$t('workflow.publicOrgText') },
        { key: 'public', value: this.$t('workflow.publicTotalText') },
      ],
      imgObj: {
        [WORKFLOW]: require(`@/assets/imgs/card_create_icon_workflow.svg`),
        [CHAT]: require(`@/assets/imgs/card_create_icon_chatflow.svg`),
        [AGENT]: require(`@/assets/imgs/card_create_icon_agent.svg`),
        [RAG]: require(`@/assets/imgs/card_create_icon_rag.svg`),
      },
    };
  },
  methods: {
    avatarSrc,
    handleClose() {
      this.dialogVisible = false;
    },
    isExploreShowTool(n) {
      return this.appFrom === 'explore' && [WORKFLOW, CHAT].includes(n.appType);
    },
    isCanClick(n) {
      return true;
      /*this.isShowTool
        ? ([WORKFLOW, CHAT].includes(n.appType) && !n.publishType) ||
            ![WORKFLOW, CHAT].includes(n.appType)
        : true;*/
    },
    // 公用删除方法
    async handleDelete() {
      const params = {
        appId: this.row.appId,
        appType: this.row.appType,
      };
      const res = await deleteApp(params);
      if (res.code === 0) {
        this.$message.success(this.$t('list.delSuccess'));
        this.$emit('reloadData');
      }
    },
    workflowEdit(row) {
      const querys = {
        id: row.appId,
      };
      this.$router.push({ path: '/workflow', query: querys });
    },
    workflowDelete(row) {
      this.row = row;
      this.$alert(this.$t('list.deleteTips'), this.$t('list.tips'), {
        confirmButtonText: this.$t('list.confirm'),
        callback: action => {
          if (action === 'confirm') {
            this.handleDelete();
          }
        },
      });
    },
    async workflowCopy(row) {
      const params = { workflow_id: row.appId };
      const res = await copyWorkFlow(
        params,
        row.appType,
        this.appFrom !== 'explore',
      );

      if (res.code === 0) {
        this.$router.push({
          path: '/workflow',
          query: { id: res.data.workflow_id },
        });
      }
    },
    workflowPublish(row) {
      this.row = row;
      this.dialogVisible = true;
      this.publishType = 'private';
    },
    async doPublish() {
      const params = {
        appId: this.row.appId,
        appType: this.row.appType,
        publishType: this.publishType,
      };
      const res = await appPublish(params);
      if (res.code === 0) {
        this.$message.success(this.$t('list.publicSuccess'));
        this.handleClose();
        this.$emit('reloadData');
      }
    },
    async cancelPublish(row) {
      let confirmed = true;
      const params = {
        appId: row.appId,
        appType: row.appType,
      };

      //工作流取消发布，需弹窗提示
      if (row.appType === WORKFLOW) {
        confirmed = await this.showDeleteConfirm(this.$t('list.cancelHint'));
      }

      if (confirmed) {
        const res = await appCancelPublish(params);
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          this.$emit('reloadData');
        }
      }
    },
    workflowExport(row) {
      exportWorkflow(
        { workflow_id: row.appId },
        row.appType,
        this.appFrom !== 'explore',
      ).then(response => {
        resDownloadFile(response, `${row.name}.json`);
      });
    },
    jumpToWorkflowPublicSet(row) {
      this.$router.push({
        path: `/workflow/publishSet`,
        query: { appId: row.appId, appType: row.appType, name: row.name },
      });
    },
    workflowTransform(row) {
      transformWorkflow({ workflow_id: row.appId }, row.appType).then(() => {
        this.$emit('reloadData');
      });
    },
    workflowOperation(method, row) {
      switch (method) {
        case 'edit':
          this.workflowEdit(row);
          break;
        case 'delete':
          this.workflowDelete(row);
          break;
        case 'copy':
          this.workflowCopy(row);
          break;
        case 'publish':
          this.workflowPublish(row);
          break;
        case 'cancelPublish':
          this.cancelPublish(row);
          break;
        case 'publishSet':
          this.jumpToWorkflowPublicSet(row);
          break;
        case 'export':
          this.workflowExport(row);
          break;
        case 'transform':
          this.workflowTransform(row);
          break;
        case 'convertToSkill':
          this.convertToSkill(row);
          break;
      }
    },
    chatDelete(row) {
      this.row = row;
      this.$alert(this.$t('list.deleteChatTips'), this.$t('list.tips'), {
        confirmButtonText: this.$t('list.confirm'),
        callback: action => {
          if (action === 'confirm') {
            this.handleDelete();
          }
        },
      });
    },
    chatOperation(method, row) {
      switch (method) {
        case 'edit':
          this.workflowEdit(row);
          break;
        case 'delete':
          this.chatDelete(row);
          break;
        case 'copy':
          this.workflowCopy(row);
          break;
        case 'cancelPublish':
          this.cancelPublish(row);
          break;
        case 'publishSet':
          this.jumpToWorkflowPublicSet(row);
          break;
        case 'export':
          this.workflowExport(row);
          break;
        case 'transform':
          this.workflowTransform(row);
          break;
        case 'convertToSkill':
          this.convertToSkill(row);
          break;
      }
    },
    async showDeleteConfirm(tips) {
      try {
        await this.$alert(tips, this.$t('list.tips'), {
          confirmButtonText: this.$t('list.confirm'),
        });
        return true;
      } catch (err) {
        return false;
      }
    },
    intelligentEdit(row) {
      this.$router.push({ path: `/agent/test?id=${row.appId}` });
    },
    intelligentDelete(row) {
      this.row = row;
      this.handleDelete();
    },
    intelligentCopy(row) {
      copyAgentApp({ assistantId: row.appId })
        .then(res => {
          if (res.code === 0) {
            const id = res.data.assistantId;
            this.$message.success(this.$t('list.copySuccess'));
            this.$router.push({ path: `/agent/test?id=${id}` });
          }
        })
        .catch(() => {});
    },
    intelligentOperation(method, row) {
      switch (method) {
        case 'edit':
          // 智能体编辑
          this.intelligentEdit(row);
          break;
        case 'delete':
          // 智能体删除
          this.intelligentDelete(row);
          break;
        case 'copy':
          // 智能体复制
          this.intelligentCopy(row);
          break;
        case 'cancelPublish':
          this.cancelPublish(row);
          break;
        case 'publishSet':
          //发布设置
          this.$router.push({
            path: `/agent/publishSet`,
            query: { appId: row.appId, appType: row.appType, name: row.name },
          });
          break;
        case 'convertToSkill':
          this.convertToSkill(row);
          break;
      }
    },
    txtQuesEdit(row) {
      this.$router.push({ path: `/rag/test?id=${row.appId}` });
    },
    txtQuesDelete(row) {
      this.row = row;
      this.handleDelete();
    },
    txtQuesCopy(row) {
      copyTextQues({ ragId: row.appId })
        .then(res => {
          if (res.code === 0) {
            const id = res.data.ragId;
            this.$message.success(this.$t('list.copySuccess'));
            this.$router.push({ path: `/rag/test?id=${id}` });
          }
        })
        .catch(() => {});
    },
    txtQuesOperation(method, row) {
      switch (method) {
        case 'edit':
          // 文本问答编辑
          this.txtQuesEdit(row);
          break;
        case 'delete':
          // 文本问答删除
          this.txtQuesDelete(row);
          break;
        case 'copy':
          // 文本问答复制
          this.txtQuesCopy(row);
          break;
        case 'cancelPublish':
          this.cancelPublish(row);
          break;
        case 'publishSet':
          this.$router.push({
            path: `/rag/publishSet`,
            query: { appId: row.appId, appType: row.appType, name: row.name },
          });
          break;
        case 'convertToSkill':
          this.convertToSkill(row);
          break;
      }
    },
    jumpToWorkflowRun(row) {
      this.$router.push({
        path: '/explore/workflow',
        query: { id: row.appId },
      });
    },
    commonToChat(row) {
      const type = row.appType;
      switch (type) {
        case AGENT:
          this.$router.push({
            path: '/explore/agent',
            query: { id: row.appId },
          });
          break;
        case RAG:
          this.$router.push({ path: '/explore/rag', query: { id: row.appId } });
          break;
        case WORKFLOW:
          this.jumpToWorkflowRun(row);
          break;
        case CHAT:
          this.jumpToWorkflowRun(row);
          break;
        case DIGITAL_EMPLOYEE:
          this.$router.push({
            path: '/explore/digitalEmployee',
            query: { employeeId: row.appId },
          });
          break;
      }
    },
    commonMethods(method, row) {
      const type = row.appType;

      switch (type) {
        case AGENT:
          this.intelligentOperation(method, row);
          break;
        case RAG:
          this.txtQuesOperation(method, row);
          break;
        case WORKFLOW:
          this.workflowOperation(method, row);
          break;
        case CHAT:
          this.chatOperation(method, row);
          break;
      }
    },
    // 转化为 Skill
    convertToSkill(row) {
      this.$emit('convertToSkill', row);
    },
    handleClick(command, row) {
      this.commonMethods(command, row);
    },
    toEdit(row) {
      if (this.appFrom === 'explore') {
        this.commonToChat(row);
      } else {
        this.commonMethods('edit', row);
      }
    },
    handelMark(e, n, i) {
      e.stopPropagation();
      this.$confirm(
        n.isFavorite
          ? this.$t('explore.unFavorite')
          : this.$t('explore.favorite'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      )
        .then(() => {
          setFavorite({
            appId: n.appId,
            appType: n.appType,
            isFavorite: !n.isFavorite,
          }).then(res => {
            if (res.code === 0) {
              this.$message.success(
                n.isFavorite
                  ? this.$t('explore.delSuccess')
                  : this.$t('explore.setSuccess'),
              );
              const list = [...this.listData];
              list[i].isFavorite = !n.isFavorite;
              this.listData = [...list];
              // this.getHistoryList();
            }
          });
        })
        .catch(() => {});
    },
    // 转换飘带样式
    conversionRibbon(_appType, category = -1) {
      switch (_appType) {
        case AGENT:
          return category === 2 ? 'gold' : 'blue';
        case RAG:
          return 'cyan';
        case WORKFLOW:
          return 'blue';
        case CHAT:
          return 'purple';
        case DIGITAL_EMPLOYEE:
          return 'green';
      }
    },
    // 转换飘带文字内容（为区分多智能体）
    conversionRibbonText(_appType, category = -1) {
      if (!this.apptype[_appType]) {
        return '';
      }
      return _appType === AGENT
        ? category === 2
          ? this.$t('appSpace.multiAgent')
          : this.$t('appSpace.singleAgent')
        : AppType[_appType];
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/commonCard.scss';
.noData {
  padding: 30px 0;
}
</style>
