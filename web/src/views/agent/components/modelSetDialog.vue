<template>
    <div>
        <el-dialog
        title="模型配置"
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
            <el-button @click="dialogVisible = false">取 消</el-button>
            <el-button type="primary" @click="submit">确 定</el-button>
        </span>
        </el-dialog>
    </div>
</template>
<script>
export default {
    props:{
        modelform:{
            type:Object,
            default:null
        }
    },
    data(){
        return{
            dialogVisible:false,
            ruleForm:{
                temperature:0.7,
                topP:1,
                frequencyPenalty:0,
                presencePenalty:0,
                maxTokens:512,
                temperatureEnable:true,
                topPEnable:true,
                presencePenaltyEnable:true,
                maxTokensEnable:true,
                frequencyPenaltyEnable:true
            },
            modelSet: [
                {
                    label:'温度',
                    desc: '增加温度将使模型的回答更具创造性',
                    props: 'temperature',
                    btnProps:'temperatureEnable',
                    min: 0,
                    max: 2,
                    step: 0.1,
                },
                {
                    label:'TopP',
                    desc: '生成过程中核采样方法概率阈值。取值越大，生成的随机性越高；取值越小，生成的确定性越高',
                    props: "topP",
                    btnProps:'topPEnable',
                    min: 0,
                    max: 1,
                    step: 0.01,
                },
                {
                    label:'频率惩罚',
                    desc: '用于控制模型已使用字词的重复率。提高此项可以降低模型在输出中重复相同字词的重复度。',
                    props: "frequencyPenalty",
                    btnProps:'frequencyPenaltyEnable',
                    min: -2,
                    max: 2,
                    step: 1,
                },
                {
                    label:'存在惩罚',
                    desc: '用于控制模型生成时的重复度，提高此项可以降低模型生成的重复度',
                    props: "presencePenalty",
                    btnProps:'presencePenaltyEnable',
                    min: -2,
                    max: 2,
                    step: 1,
                },
                {
                    label:'最大标记',
                    desc: '模型回答的tokens的最大长度',
                    props: "maxTokens",
                    btnProps:'maxTokensEnable',
                    min: 1,
                    max: 4096,
                    step: 1,
                }
            ]
        }
    },
    methods:{
        showDialog(){
            this.dialogVisible = true;
            if(this.modelform !== null){
                this.ruleForm = this.modelform
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