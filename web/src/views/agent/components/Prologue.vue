<!--开场白-->
<template>
    <div class="history-box">
        <div class="session-answer" >
            <div :class="['session-item','rl']">
                <img class="logo" :src="editForm.avatar.path ? `/user/api`+ editForm.avatar.path : '@/assets/imgs/bg-logo.png'" />
                <div class="answer-content">
                    <p class="name">{{editForm.name || '无Information'}}</p>
                    <p class="systemPrompt">{{editForm.prologue || ''}}</p>
                    <div class="recommend">
                        <template v-if="recommendQuestion.length > 0">
                            <p class="recommend-p" 
                            v-for="(n,i) in recommendQuestion" 
                            :key="`${i}rml`" 
                            @click="setProloguePrompt(n.value)" 
                            @mouseenter="mouseEnter(n)"
                            @mouseleave="mouseLeave(n)">
                                <img :src="n.active ? require('@/assets/imgs/robot_active.png') :require('@/assets/imgs/robot_unactive.png')" />
                                {{n && n.value }}
                            </p>
                        </template>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
<script>
    export default {
        props:['basicForm','expandForm','isBigModel',"editForm"],
        data(){
            return{
                basePath: this.$basePath,
                recommendQuestion:[],
                debounceTimer:null
            }
        },
        watch:{
            'editForm.recommendQuestion':{
                handler(newVal) {
                    if(newVal){
                        if(this.debounceTimer){
                            clearTimeout(this.debounceTimer)
                        }
                        this.debounceTimer = setTimeout(() =>{
                            this.recommendQuestion = newVal.filter(item => item.value !== '').map(item =>({...item,active:false}))
                        },500)
                    }
                },
                immediate: true,
                deep:true
            }
        },
        methods:{
            setProloguePrompt(val){
                this.$emit('setProloguePrompt',val)
            },
            mouseEnter(n){
                n.active = true;
            },
            mouseLeave(n){
                n.active = false;
            }
        }
    }
</script>
<style lang="scss" scoped>
    .history-box{
        height: calc(100% - 46px);
        overflow-y: auto;
        display:flex;
        align-items:center;
        justify-content:center;
    }
    .echo{
        word-break: break-all;
        height: 100%;
        overflow-y: auto;
        width:auto;
        .session-item{
            width:600px;
            min-height: 80px;
            display: flex;
            justify-content:center;
            flex-wrap: wrap;
            padding: 20px;
            line-height: 28px;
            img{
                width: 100px;
                height: 100px;
                object-fit: cover;
            }
            .logo{
                border-radius: 50%;
                border:1px solid #eee;
            }
            .answer-content{
                width: calc(100% - 30px);
                position: relative;
                margin-left: 14px;
                color: #333;
                margin:0 auto;
                .name{
                    font-size: 18px;
                    font-weight: bold;
                    text-align: center;
                    padding: 12px 0;
                }
                .systemPrompt{
                    line-height: 26px;
                    text-align: center;
                    font-size: 16px;
                    font-weight: bold;
                }
                .recommend{
                    color: #425466 ;
                    text-align: left;
                    cursor: pointer;
                    width:calc(100% - 50px);
                    margin:20px auto;
                    .recommend-p{
                        width:100%;
                        margin: 10px 0;
                        padding:  5px 0 5px 15px ;
                        background:rgba(83,108,143,.08);
                        border-radius:6px;
                        display:flex;
                        align-items:center;   
                        border:1px solid transparent;
                        img{
                            width:16px;
                            height:16px;
                            margin-right: 5px;
                            flex-shrink: 0;
                        }
                    }
                    .recommend-p:hover{
                        color:$color;
                        border:1px solid $color;
                    }
                }
            }
        }
        
        @media (max-width: 620px) {
            .history-box .session-item,.session-item {
                width: 95vw!important;
            }
        }
    }

</style>

