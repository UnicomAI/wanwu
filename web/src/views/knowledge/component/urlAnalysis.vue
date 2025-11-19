<template>
  <div class="urlAnalysis">
    <el-form
      :model="dynamicValidateForm"
      ref="dynamicValidateForm"
      label-width="80px"
      size="mini"
      class="urlAnalysisForm"
    >
      <p class="urlTips">{{$t('knowledgeManage.addLimitTips')}}</p>
      <el-form-item
        v-for="(domain, index) in dynamicValidateForm.domains"
        label="urladdress"
        :key="domain.key"
        :prop="'domains.' + index + '.value'"
        :rules="[
          {
            required: true,
            message: $t('knowledgeManage.notEmpty'),
            trigger: 'blur',
          },
          { validator: validateUrl, trigger: 'blur' },
        ]"
      >
        <el-input
          v-model="domain.value"
          class="url_txt"
          @change="handleChange(domain, index)"
          :disabled="loading.url"
        ></el-input
        >&nbsp;
        <i
          class="el-icon-loading"
          v-if="urlConut > 1 && domain.urlLoading === true"
        ></i>
        <i
          class="el-icon-error"
          @click="removeDomain(domain)"
          v-if="urlConut > 1 && domain.urlLoading !== true"
        ></i>
        <i
          class="el-icon-success"
          v-if="domain.back && domain.urlLoading !== true"
        ></i>
        <i
          class="el-icon-warning"
          v-if="domain.back === false && domain.urlLoading !== true"
        ></i>
      </el-form-item>
      <el-form-item>
        <el-button
          @click="addDomain"
          v-if="urlConut < 10"
          icon="el-icon-plus"
          size="mini"
          :disabled="loading.url"
          >{{$t('knowledgeManage.addOnePiece')}}</el-button
        >
        <el-button
          @click="submitForm('dynamicValidateForm')"
          type="primary"
          size="mini"
          :loading="loading.url"
          >{{$t('knowledgeManage.startAnalysis')}}</el-button
        >
      </el-form-item>
    </el-form>
  </div>
</template>
<script>
import { setUploadURL,setAnalysis } from "@/api/knowledge";
export default {
  props: {
    categoryId: {
      type: String,
      default: "",
    }
  },
  data() {
    return {
      urlConut: 1,
      backResult: [],
      loading: {
        url: false,
      },
      oldList: [{ value: "" }], //Save上一次urlResult
      dynamicValidateForm: {
        domains: [
          {
            value: "",
            back: null,
            urlLoading: false,
          },
        ],
      },
    };
  },
  mounted() {},
  methods: {   
    handleChange(item, index) {
      if ( this.oldList[index].value && item.value !== this.oldList[index].value) {
        this.dynamicValidateForm.domains[index].back = null;
      }
    },
    // startParse
    async submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (valid) {
          this.loading.url = true;
          this.concurRequest(this.dynamicValidateForm.domains, 3)
            .then((res) => {
              this.loading.url = false;
              if (res.length > 0) {
                this.backResult = res;
                this.$emit('handleSetData',this.backResult)
              }
            }).catch((err) => {
              this.loading.url = false;
            });
          this.oldList = JSON.parse(
            JSON.stringify(this.dynamicValidateForm.domains)
          );
        } else {
          console.log("error submit!!");
          return false;
        }
      });
    },
    concurRequest(urls, maxNum) {
      const _this = this;
      return new Promise((resolve) => {
        if (urls.length === 0) {
          resolve([]);
        }
        const results = []; // 存放RequestResult
        let index = 0; // 下一Request of urlAddress of 下标
        let count = 0; // alreadyComplete of RequestCount
        async function request() {
          if (index === urls.length) return;
          const i = index; // SaveNo.，使resultAndurls相对应
          const url = index < urls.length && urls[i].value;
          index++;
          try {
            if (urls[i].back !== true) {
              _this.dynamicValidateForm.domains[i].urlLoading = true;
              const resp = await setAnalysis({
                urlList: [url],
                knowledgeId: _this.categoryId,
              });

              if (Number(resp.code) === 0) {
                _this.dynamicValidateForm.domains[i].back = true;
                _this.dynamicValidateForm.domains[i].fileSize = resp.data.urlList[0].fileSize;
                _this.dynamicValidateForm.domains[i].url = resp.data.urlList[0].url;
                _this.dynamicValidateForm.domains[i].fileName = resp.data.urlList[0].fileName;
                results[i] = resp.data.urlList[0];
                results[i].back = true;
              } else {
                _this.dynamicValidateForm.domains[i].back = false;
                results[i] = urls[i];
              }
              _this.dynamicValidateForm.domains[i].urlLoading = false;
            } else {
              results[i] = urls[i];
            }
          } catch (err) {
            console.log(err)
          } finally {
            count++; // CheckYesNo所Has of Request都alreadyComplete
            if (count === urls.length) {
              resolve(results);
            }
            request();
          }
        }
        const times = Math.min(maxNum, urls.length);
        for (let i = 0; i < times; i++) {
          request();
        }
      });
    },
    resetForm(formName) {
      this.$refs[formName].resetFields();
      this.dynamicValidateForm = {
        domains: [
          {
            value: "",
            back: null,
          },
        ],
      };
      this.oldList = [{ value: "" }];
      this.urlConut = 1;
      this.loading.url = false;
    },    
    removeDomain(item) {
      var index = this.dynamicValidateForm.domains.indexOf(item);
      if (index !== -1) {
        this.dynamicValidateForm.domains.splice(index, 1);
        this.oldList.splice(index, 1);
        this.backResult.splice(index, 1);
      }
      this.urlConut -= 1;
    },
    addDomain() {
      if (this.urlConut < 10) {
        this.dynamicValidateForm.domains.push({
          value: "",
          key: Date.now(),
          back: null,
          urlLoading: false,
        });
        this.oldList.push({ value: "" });
        this.urlConut += 1;
      }
    },
    handleSave() {
      this.$refs["dynamicValidateForm"].validate((valid) => {
        if (valid) {
          if (this.backResult.length > 0 && !this.isDabled) {
            this.$emit("handleLoading", true);
            setUploadURL({
              urls: this.backResult,
              categoryId: this.categoryId,
            })
              .then((res) => {
                if (res.code === 0) {
                  this.$message.success("OperationSuccess");
                } else {
                  this.$message.error(res.msg);
                }
                this.$emit("handleLoading", false, "success");
              })
              .catch((err) => {
                this.$message.error(err);
                this.$emit("handleLoading", false);
              });
          } else {
            this.$message.warning(this.$t('knowledgeManage.analysisTips'));
          }
        } else {
        }
      });
    },
    validateUrl(rule, value, callback) {
      function isValidUrl(url) {
        try {
          new URL(url);
          return true;
        } catch {
          return false;
        }
      }
      if (!isValidUrl(value)) {
        callback(new Error(this.$t('knowledgeManage.urlTest')));
      } else {
        callback();
      }
    },
  },
  computed: {
    result() {
      let obj = {
        change: false,
        list: [],
      };
      for (let i = 0; i < this.dynamicValidateForm.domains.length; i++) {
        if (
          this.oldList[i].value &&
          this.dynamicValidateForm.domains[i].value !== this.oldList[i].value
        ) {
          this.dynamicValidateForm.domains[i].back = null;
        }
      }
      return obj;
    },
    isDabled() {
      if (this.backResult.length > 0) {
        for (let i = 0; i < this.backResult.length; i++) {
          if (this.backResult[i].back === false) {
            return true;
          }
        }
        return false;
      } else {
        return true;
      }
    }
  },
};
</script>
<style lang="scss">
.urlAnalysisForm {
  max-height:300px;
  overflow-y:auto;
  .urlTips{
    padding:10px 0;
    text-align:left;
    font-size:14px;
  }
  .url_txt {
    width: 85%;
  }
  .el-form-item__content {
    display:flex !important;
    justify-content:flex-start!important;
    align-items:center!important;
    &:hover {
      cursor: pointer;
      .el-icon-error {
        display: inline-block;
      }
      .el-icon-success {
        display: none;
      }
      .el-icon-warning {
        display: none;
      }
    }
  }
  .el-icon-warning {
    line-height: 30px;
    display: inline-block;
    color: #e60001;
    font-size: 16px;
  }
  .el-icon-success {
    line-height: 30px;
    display: inline-block;
    color: #67c23a;
    font-size: 16px;
  }
  .el-icon-error {
    line-height: 30px;
    display: none;
    font-size: 16px;
  }
  .el-icon-loading{
    margin-left:5px;
  }
}
</style>