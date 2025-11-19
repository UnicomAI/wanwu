<template>
    <div>
        <div>
            <el-button type="primary" icon="el-icon-plus" size="mini" @click="showDialog(null)">Create</el-button>
            <el-table
                :data="tableData"
                style="width: 100%;margin-top:15px;"
                :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
              >
                <el-table-column
                  prop="name"
                  label="AppName"
                >
                </el-table-column>
                <el-table-column
                  prop="description"
                  label="Description"
                >
                </el-table-column>
                <el-table-column
                  prop="suffix"
                  label="AccessUrl"
                >
                <template slot-scope="scope">
                    <span>{{scope.row.suffix}}</span>
                    <span class="el-icon-copy-document copy" @click="handleCopy(scope.row) && copycb()"></span>
                </template>
                </el-table-column>
                <el-table-column
                  prop="expiredAt"
                  label=" 期 when 间"
                  width="180"
                >
                </el-table-column>
                <el-table-column
                  prop="status"
                  label="Status"
                  width="180"
                >
                <template slot-scope="scope">
                    <el-switch
                        v-model="scope.row.status"
                        @change="statusChange($event,scope.row)"
                        active-color="var(--color)"
                        >
                    </el-switch>
                </template>
                </el-table-column>
                <el-table-column
                  :label="$t('knowledgeManage.operate')"
                  width="260"
                >
                  <template slot-scope="scope">
                    <el-button
                      size="mini"
                      round
                      @click="showDialog(scope.row)"
                    >Edit</el-button>
                    <el-button
                      size="mini"
                      round
                      @click="handleDel(scope.row)"
                    >{{$t('common.button.delete')}}</el-button>
                  </template>
                </el-table-column>
              </el-table>
        </div>
        <el-dialog
          :title="title"
          :visible.sync="dialogVisible"
          width="30%"
          :before-close="handleClose">
          <el-form ref="form" :model="form" class="formUrl" label-width="100px">
            <el-form-item label="AppName" 
              prop="name"
              :rules="[{ required: true, message: 'Please enterAppName', trigger: 'blur' }]"
            >
              <el-input v-model="form.name" placeholder="Please enterAppName"></el-input>
            </el-form-item>
            <el-form-item label="AppDescription" 
              prop="description"
            >
              <el-input v-model="form.description" placeholder="Please enterAppDescription" type="textarea" :rows="2"></el-input>
            </el-form-item>
            <el-form-item label=" 期 when 间" prop="expiredAt">
               <el-date-picker
                v-model="form.expiredAt"
                type="datetime"
                value-format="yyyy-MM-dd HH:mm:ss"
                placeholder="url生效 when 间">
              </el-date-picker>
            </el-form-item>
            <!-- <el-form-item label="Knowledge Base出处Details">
              <el-switch
                v-model="value"
                active-color="var(--color)">
              </el-switch>
            </el-form-item> -->
            <!-- <el-form-item label="WorkflowDetails">
              <el-switch
                v-model="value"
                active-color="var(--color)">
              </el-switch>
            </el-form-item> -->
            <div class="online-item">
              <el-form-item  prop="copyright">
                <template #label>
                  <span>版权</span>
                  <el-tooltip class="item" effect="dark" content="YesNo在界面中Show版权Information" placement="top-start">
                    <span class="el-icon-question tips"></span>
                  </el-tooltip>
                </template>
                <el-input v-model="form.copyright" placeholder="Please enter版权Information"></el-input>
              </el-form-item>
              <el-form-item prop="copyrightEnable">
                <el-switch
                  :disabled="!form.copyright"
                  v-model="form.copyrightEnable"
                  active-color="var(--color)">
                </el-switch>
              </el-form-item>
            </div>
            <div class="online-item">
              <el-form-item prop="privacyPolicy">
                <template #label>
                  <span>隐私协议</span>
                  <el-tooltip class="item" effect="dark" content="YesNo在界面中Show隐私协议Information" placement="top-start">
                    <span class="el-icon-question tips"></span>
                  </el-tooltip>
                </template>
                <el-input v-model="form.privacyPolicy" placeholder="Please enter隐私政策Link" @blur="urlBlur"></el-input>
              </el-form-item>
              <el-form-item prop="privacyPolicyEnable">
                <el-switch
                  :disabled="!form.privacyPolicy"
                  v-model="form.privacyPolicyEnable"
                  active-color="var(--color)">
                </el-switch>
              </el-form-item>
            </div>
            <div class="online-item">
              <el-form-item prop="disclaimer">
                <template #label>
                  <span>免责声明</span>
                  <el-tooltip class="item" effect="dark" content="YesNo在界面中Show免责声明" placement="top-start">
                    <span class="el-icon-question tips"></span>
                  </el-tooltip>
                </template>
                <el-input v-model="form.disclaimer" type="textarea" :rows="2" placeholder="Please enter免责声明"></el-input>
              </el-form-item>
              <el-form-item prop="disclaimerEnable">
                <el-switch
                  :disabled="!form.disclaimer"
                  v-model="form.disclaimerEnable"
                  active-color="var(--color)">
                </el-switch>
              </el-form-item>
            </div>
          </el-form>
          <span slot="footer" class="dialog-footer">
            <el-button @click="handleClose">取 消</el-button>
            <el-button type="primary" @click="submit('form')">确 定</el-button>
          </span>
        </el-dialog>
    </div>
</template>
<script>
import { getOpenurl,delOpenurl,editOpenurl,createOpenurl,switchOpenurl } from "@/api/agent";
export default {
    props:['appId','appType'],
    data(){
        return{
          form:{
            appId:'',
            appType:'',
            copyright:'',
            copyrightEnable:false,
            disclaimer:'',
            disclaimerEnable:false,
            expiredAt:'',
            name:'',
            description:'',
            privacyPolicy:'',
            privacyPolicyEnable:false
          },
          title:'CreateURL',
          dialogVisible:false,
          tableData:[],
          urlId:''
        }
    },
    created(){
      this.form.appId = this.appId;
      this.form.appType = this.appType;
      this.getList();
    },
    methods:{
        urlBlur(){
          const text = this.form.privacyPolicy;
          if(!this.isValidUrl(text)){
            this.$message.warning('Link效验不合格')
            this.form.privacyPolicy = '';
          }
        },
        isValidUrl(string) {
          const pattern = /^https?:\/\/(?:[-\w.])+(?:\:[0-9]+)?(?:\/(?:[\w/_.])*(?:\?(?:[\w&=%.])*)?(?:\#(?:[\w.])*)?)?$/;
          return pattern.test(string.trim());
        },
        handleCopy(row){
          let text = row.suffix;
          var textareaEl = document.createElement("textarea");
          textareaEl.setAttribute("readonly", "readonly");
          textareaEl.value = text;
          document.body.appendChild(textareaEl);
          textareaEl.select();
          var res = document.execCommand("copy");
          document.body.removeChild(textareaEl);
          return res;
        },
        copycb() {
          this.$message.success("Content已Copy到粘贴板");
        },
       getList(){
          getOpenurl({appId:this.appId,appType:this.appType}).then(res =>{
            if(res.code === 0){
              this.tableData = res.data || []
            }
          }).catch(() =>{

          })
       },
       statusChange(status,row){
          switchOpenurl({status,urlId:row.urlId}).then(res =>{
            if(res.code === 0){
              this.$message.success('OperationSuccess');
              this.getList();
            }
          }).catch(() =>{
            
          })
       },
      showDialog(row=null){
        this.dialogVisible = true
        if(row === null){
          this.title = 'CreateURL';
          this.urlId = '';
          this.$nextTick(() =>{
            if (this.$refs.form) {
              this.$refs.form.resetFields();
              this.$refs.form.clearValidate();
              this.clear();
            }
          })
        }else{
          this.title = 'EditURL';
          this.urlId = row.urlId;
          Object.keys(row).forEach(key  =>{
            if(this.form.hasOwnProperty(key)){
              this.form[key] = row[key]
            }
          })
        }
      },
      clear(){
        this.form.copyright = '';
        this.form.copyrightEnable = false;
        this.form.disclaimer = '';
        this.form.disclaimerEnable = false;
        this.form.expiredAt = '';
        this.form.name = '';
        this.form.privacyPolicy = '';
        this.form.privacyPolicyEnable = false;
      },
      submit(formName){
        this.$refs[formName].validate((valid) => {
          if (valid) {
            if(this.urlId === ''){
              this.createUrl()
            }else{
              this.eidtUrl()
            }
          } else {
            return false;
          }
        });
      },
      createUrl(){
        createOpenurl(this.form).then(res =>{
          if(res.code === 0){
            this.$message.success('OperationSuccess');
            this.dialogVisible = false;
            this.getList();
          }
        }).catch(() =>{

        })
      },
      eidtUrl(){
        const data = {
          ...this.form,
          urlId:this.urlId
        }
        editOpenurl(data).then(res =>{
          if(res.code === 0){
            this.$message.success('OperationSuccess');
            this.dialogVisible = false;
            this.getList();
          }
        }).catch(() =>{

        })
      },
      handleDel(row){
          this.$confirm(
          "确定要Delete当前AccessURL吗？",
          this.$t("knowledgeManage.tip"),
          {
            confirmButtonText: "确定",
            cancelButtonText: "Cancel",
            type: "warning",
          }
        )
          .then(() => {
            delOpenurl({ urlId: row.urlId }).then((res) => {
              if (res.code === 0) {
                this.$message.success("DeleteSuccess");
                this.getList();
              }
            });
          })
          .catch((error) => {
            this.getList();
          });
        },
      handleClose(){
        this.dialogVisible = false;
      }
    }
}
</script>
<style lang="scss" scoped>
.copy{
  cursor: pointer;
  margin-left:5px;
  color: $color;
}
.formUrl{
  .el-date-editor{
    width:100%;
  }
  .online-item{
    display: flex;
    justify-content: space-between;
    .tips{
      margin-left: 2px;
      color: #aaadcc;
      cursor: pointer;
    }
  }
  .online-item > :nth-child(1){
    width:80%;
  }
  .online-item > :nth-child(2){
    width:20%;
    display:flex;
    justify-content:flex-end;
    /deep/.el-form-item__content{
      margin-left:0!important;
    }
  }
}

</style>