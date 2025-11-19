<template>
    <div>
        <el-dialog
        title="Recall parameter configuration"
        :visible.sync="dialogVisible"
        width="50%"
        :before-close="handleClose">
        <span>
            <el-form :model="ruleForm" ref="ruleForm" label-width="100px" class="demo-ruleForm">
                <el-form-item :label="item.label" :prop="item.props" v-for="(item,index) in konwledgeSet" :key="index">
                    <el-row>
                        <el-col :span="1">
                            <el-tooltip class="item" effect="light" :content="item.desc" placement="bottom">
                                <span class="el-icon-question question"></span>
                            </el-tooltip>
                        </el-col>
                        <el-col :span="2">
                            <el-switch v-model="ruleForm[item.btnProps]"></el-switch>
                        </el-col>
                        <el-col :span="20">
                            <el-slider v-model="ruleForm[item.props]" show-input :min="item.min" :max="item.max" :step="item.step"></el-slider>
                        </el-col>
                    </el-row>
                </el-form-item>
            </el-form>
        </span>
        <span slot="footer" class="dialog-footer">
            <el-button @click="dialogVisible = false">Cancel</el-button>
            <el-button type="primary" @click="submit">Confirm</el-button>
        </span>
        </el-dialog>
    </div>
</template>
<script>
export default {
    props:{
        knowledgeConfig:{
            type: Object,
            default:null
        }
    },
    data(){
        return{
            dialogVisible:false,
            ruleForm:{
                maxHistory:0,
                threshold:0.4,
                topK:5,
                maxHistoryEnable:true,
                thresholdEnable:true,
                topKEnable:true
            },
            konwledgeSet: [
                {
                    label:'Max Context Length',
                    desc: 'Maximum number of conversation turns to keep in context.',
                    props: 'maxHistory',
                    btnProps:'maxHistoryEnable',
                    min: 0,
                    max: 100,
                    step: 1,
                },
                {
                    label:' Filter threshold value',
                    desc: 'Minimum relevance score for retrieved results; items below this threshold are filtered out.',
                    props: "threshold",
                    btnProps:"thresholdEnable",
                    precision:1,
                    min: 0,
                    max: 1,
                    step: 0.1,
                },
                {
                    label:' Number of knowledge items',
                    desc: 'Maximum number of knowledge segments to retrieve; if more are found, only this many will be returned.',
                    props: "topK",
                    btnProps:"topKEnable",
                    min:1,
                    max: 20,
                    step: 1,
                }
            ]
        }
    },
    methods:{
        showDialog(){
            this.dialogVisible = true;
            if(this.knowledgeConfig !== null){
                const data = JSON.parse(JSON.stringify(this.knowledgeConfig));
                this.ruleForm = data;
            }
        },
        handleClose(){
            this.dialogVisible = false;
        },
        submit(){
            this.dialogVisible = false;
            this.$emit('setKnowledgeSet',this.ruleForm)
        }
    }
}
</script>
<style lang="scss" scoped>
/deep/{
    .el-input-number--small{
        line-height: 28px!important;
    }
}
.question{
    cursor: pointer;
    color:#ccc;
}
</style>