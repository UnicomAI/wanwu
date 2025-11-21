<template>
    <div>
        <el-dialog
        title="ModelConfiguration"
        :visible.sync="dialogVisible"
        width="50%"
        :before-close="handleClose">
        <span>
            <el-form :model="ruleForm" ref="ruleForm" label-width="100px" class="demo-ruleForm">
                <el-form-item :label="item.label" :prop="item.props" v-for="(item,index) in modelSet" :key="index">
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
                            <el-slider v-model="ruleForm[item.props]" show-input  :min="item.min" :max="item.max" :step="item.step"></el-slider>
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
        modelConfig:{
            type: Object,
            default:null
        }
    },
    data(){
        return{
            dialogVisible:false,
            ruleForm:{
                temperature:0.14,
                topP:0.85,
                frequencyPenalty:1.1,
                temperatureEnable:false,
                topPEnable:false,
                frequencyPenaltyEnable:false
            },
            modelSet: [
                {
                    label:'Temperature',
                    desc: 'Increasing temperature will make the model\'s responses more creative',
                    props: 'temperature',
                    btnProps:'temperatureEnable',
                    min: 0,
                    max: 1,
                    step: 0.01,
                },
                {
                    label:'多样性',
                    desc: 'Probability threshold for nucleus sampling during generation. Higher values increase randomness; lower values increase determinism',
                    props: "topP",
                    btnProps:"topPEnable",
                    min: 0,
                    max: 10,
                    step: 0.01,
                },
                {
                    label:'duplicate惩罚',
                    desc: 'Used to control the repetition rate of words already used by the model. Increasing this reduces repetition of the same words in output.',
                    props: "frequencyPenalty",
                    btnProps:"frequencyPenaltyEnable",
                    min: 1,
                    max: 10,
                    step: 0.1,
                }
            ]
        }
    },
    methods:{
        showDialog(){
            this.dialogVisible = true;
            if(this.modelConfig !== null){
                const data = JSON.parse(JSON.stringify(this.modelConfig))
                this.ruleForm = data;
            }
        },
        handleClose(){
            this.dialogVisible = false;
        },
        submit(){
            this.dialogVisible = false;
            this.$emit('setModelSet',this.ruleForm)
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