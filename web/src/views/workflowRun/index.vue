<template>
  <div class="workflowRun" style="overflow: hidden;">
    <!--调试-->
    <DebugModal
      ref="debug"
      @doDebug="doDebug"
    />
    <div class="workflow-result" v-if="!isInit">
      <div class="node-status" v-if="nodeData.node_status ">
        <div class="header">
          <i
            v-if="nodeData.node_status === 'loading'"
            class="el-icon-warning loading status-icon"
          ></i>
          <i
            v-if="nodeData.node_status === 'success'"
            class="el-icon-success success status-icon"
          ></i>
          <i
            v-if="nodeData.node_status === 'failed'"
            class="el-icon-error failed status-icon"
          ></i>
          <i
            v-if="nodeData.node_status === 'init'"
            class="el-icon-warning init status-icon"
          ></i>
          {{ statusObj[nodeData.node_status] }}
        </div>

        <div>
          <div
            v-if="nodeData.node_status === 'loading'"
            style="line-height: 140px; text-align: center"
          >
            <i
              class="el-icon-loading"
              style="color: cornflowerblue; margin: auto; font-size: 26px"
            ></i>
          </div>
          <div class="node-status-content" v-else>
            <!--ErrorTip-->
            <div
              class="params node-message"
              v-if="nodeData.node_status === 'failed'"
            >
              ErrorInformation: <pre v-html="nodeData.res_inputs"></pre>
            </div>
            <!--正常输出-->
            <div v-else>
              <div class="params">
                <p>
                  <i class="el-icon-caret-bottom"></i> 输出：
                  <i
                    class="el-icon-document-copy copy-icon"
                    @click="preCopy(nodeData.res_inputs)"
                  ></i>
                </p>
                <div>
                  <pre v-html="nodeData.res_inputs"></pre>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <el-empty class="noData" v-if="isInit" description="click左侧 '运row' Button，start运rowWorkflow"></el-empty>
  </div>
</template>

<script>
import DebugModal from "./run/DebugDrawer";
import { useWorkFlow } from "@/api/workflow";
export default {
  components: {
    DebugModal,
  },
  data() {
    return {
      workflowId: '',
      statusObj: {
        success: "Success",
        failed: "Failed",
        init: "waiting",
        loading: "Running",
        running_skip: "not运row",
      },
      isInit: true,
      nodeData: {
        node_status: 'init',
      }
    }
  },
  computed: {},
  created() {
    this.workflowId = this.$route.query.id
  },
  mounted() {},
  methods: {
    preCopy(val) {
      this.$copy(JSON.stringify(val));
      this.$message.success("Content copied to clipboard");
    },
    async doDebug(data) {
      this.nodeData.node_status = 'loading'
      this.isInit = false

      const res = await useWorkFlow({
        data,
        workflowID: this.workflowId,
      });

      this.$refs["debug"].runDisabled = false
      if (res.code === 0) {
        this.nodeData.res_inputs = res.data
        this.nodeData.node_status = 'success'
      } else {
        this.nodeData.res_inputs = res.data
        this.nodeData.node_status = 'failed'
      }
    },
  }
};
</script>

<style lang="scss">
.workflow-result {
  margin-left: 450px;
  margin-top: 20px;
  margin-bottom: 20px;
  width: calc(100% - 460px);
  border-radius: 10px;
  background-color: #fff;
  min-height: 200px;
  padding: 20px;
  box-shadow: 0 1px 4px 0 rgba(0,0,0,0.15);
}
.node-status {
  //position: absolute;
  width: 100%;
  border-radius: 10px;
  border: 1px solid #ddd;
  background-color: #333;
  color: #fff;
  bottom: -255px;
  left: 0;
  right: 0;
  min-height: 260px;
  .node-status-content {
    max-height: calc(100vh - 110px);
    overflow: auto;
  }
  .header {
    position: relative;
    line-height: 28px;
    background: #666;
    padding: 0 10px;
    border-radius: 6px 6px 0 0;

  }
  .status-icon {
    font-size: 18px;
  }
  .success {
    color: #5a9600;
  }
  .failed {
    color: red;
  }
  .init {
    color: orange;
  }
  .params {
    padding: 10px;
    .copy-icon {
      cursor: pointer;
      &:hover {
        color: cornflowerblue;
      }
    }
  }
  .loading {
    color: cornflowerblue;
  }
}
.node-header {
  display: flex;
  justify-content: space-between;
  .node-header-icon {
    width: 30px;
    height: 20px;
    object-fit: contain;
    border-radius: 10px;
    vertical-align: middle;
    margin: 0 6px 4px 0;
  }
  .controls {
    cursor: pointer;
  }
}
.node-params {
  position: relative;
  padding: 10px;
  background-color: #f9f9fb;
  color: #151b26;
  border-radius: 10px;
  margin-top: 10px;
  h2{
    font-weight:400;
  }
  /*.iten-item{
    line-height:30px;
    background:#ffffff;
    margin-top:10px;
    padding-left:15px;
    border-radius:3px;
    height: 30px;
  }*/
  .params-type {
    .params-type-span {
      margin-left: 10px;
    }
  }
  .params-content {
    margin-top: 4px;
    .params-content-item {
      padding: 4px 0;
      display: flex;
      span {
        font-size: 12px;
        display: inline-block;
        vertical-align: middle;
      }
      span:nth-child(1) {
        color: #876300;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 106px;
      }
      span:nth-child(2) {
        margin-left: 6px;
        padding: 0 5px;
        white-space: nowrap;
        border-radius: 4px;
        background-color: #e8e9eb;
        color: #5c5f66;
        max-width: 75%;
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }
    .params-content-value {
      display: block;
      box-sizing: border-box;
      width: fit-content;
      max-width: 100%;
      padding: 0 4px;
      /*border: 1px solid #e8e9eb;*/
      border-radius: 4px;
      background-color: #fff;
      color: #333;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }
  .value-box {
    border: 1px solid red;
    min-height: 100px;
  }
}
.workflowRun .noData {
  width: calc(100% - 460px);
  margin-left: 450px;
  text-align: center;
  margin-top: calc(50vh - 180px);
  /deep/ .el-empty__description p {
    color: #B3B1BC;
  }
}
</style>
