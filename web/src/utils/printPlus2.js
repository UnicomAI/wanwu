/*根据finish为1OR2 when ，判断YesNo打印结束*/
import workerTimer from './worker'
import {parseSub, isSub} from "@/utils/util.js"

const Print = function (opt) {
    this.sentenceArr = []//存储待打印 of 句子 of Array
    this.sIndexMap={} 
    this.timer = opt.timer || 10; //打印速度
    this.t = null;
    this.sIndex = 0 //Record已打印句子 of Index（避免重复打印）
    this.printStatus = 0
    this.fullWord = ''
    this.searchList = []
    this.onPrintEnd = (opt.onPrintEnd && (typeof opt.onPrintEnd === 'function')) ? opt.onPrintEnd : () => {
    };
    this.looper = null
}
Print.prototype = {
    print(sentence,privateData, printingCB, endCB) {
        if(privateData.searchList  && privateData.searchList.length){
            this.searchList = privateData.searchList
        }
        this.sentenceArr.push(sentence)
        this.loop(printingCB, endCB, "truely")
    },
    stop() {
        this.sentenceArr = []
        this.sIndexMap = {}
        this.sIndex = 0
        this.looper && this.looper.stop()
    },
    loop(printingCB, endCB) {

        //If正在打印OR者打印结束
        if (this.printStatus === 1 || this.sIndex >= this.sentenceArr.length) {
            return;
        }

        let curSentence = this.sentenceArr[this.sIndex]
        this.printStatus = 1
        if(!curSentence){
            console.log(this.sIndex, this.sentenceArr)
            return;
        }
        this.looper = new Looper(this.sIndex, curSentence, this.timer, (world) => {
            this.printStatus = 1
            let isEnd = this.sIndex === this.sentenceArr.length -1
            printingCB({world,finish:curSentence.finish, isEnd},this.searchList)
        }, (data) => {
            this.printStatus = 0
            this.sIndex++;
            if (this.sentenceArr[this.sIndex]) {
                this.loop(printingCB, endCB)
            } else {
                this.onPrintEnd()
            }
        },this.sIndexMap)
    },
    getAllworld(){
        let str = ''
        this.sentenceArr.forEach((n,i)=>{
            str += n.response
        })
        return str
    }
}

const Looper = function (sIndex, sentence, timer, printCB, endCB,sIndexMap) {
    this.sIndex = sIndex
    this.sIndexMap=sIndexMap
    this.sentence = sentence ? sentence.response : "" //当前要打印 of 句子
    this.timer = timer
    this.t = null
    this.index = 0 //当前打印到 of 字符位置
    this.printCB = printCB //每打印一个字符 of Callback
    this.endCB = endCB //句子打印结束 of Callback
    this.isCodeBlock = false // Add：标记YesNo为代码块
    this.codeBlockContent = '' // Add：存储代码块Content
    this.animationFrame = null
    this.lastTimestamp = performance.now(); // Add：每次LooperInit when Reset
    // 在Init when 检测YesNo为代码块
    this.detectCodeBlock()
    this.start()
}

Looper.prototype = {
    detectCodeBlock() {
        // CheckYesNoContains MCP ToolName，IfYesThen不按代码块Process
        const mcpToolPattern = /<tool>mcp-tool name：/;
        if (mcpToolPattern.test(this.sentence)) {
            this.isCodeBlock = false;
            return;
        }
        
        // 更宽松 of 代码块Match正Then
        const codeBlockRegex = /\n\n```(?:\w+)?[\s\S]*?```\n\n/s;
        const match = this.sentence.match(codeBlockRegex);
        if (match) {
            this.isCodeBlock = true;
            this.codeBlockContent = match[0]; // 整个代码块Content
            this.sentence = match[0]; // 代码块内部Content（去掉```）
            this.index = this.sentence.length; // Add：代码块Directly打印完毕
        }
    },
    start() {
        if(this.sentence === ''){
            this.printCB('')
            this.stop()
            this.index++;
            return
        }

        this.lastTimestamp = performance.now(); // Add：每次start都Reset

        if (this.isCodeBlock) {
            this.printCB(this.sentence);
            this.stop();
            return;
        }

        // ProcessIndex引文Tag
        if(isSub(this.sentence)){
            this.printCB(parseSub(this.sentence))
            this.stop()
            this.index++;
            return
        }

        // this.printFn();

        
        // const batchSize = 10; // 推荐每次Output30个字符
        // const interval = 15; // 减少Output间隔 when 间
        // this.index = 0;
        // this.t = workerTimer.setInterval(() => {
        //     if (this.index === this.sentence.length) {
        //         this.stop()
        //         return
        //     }
        //     const endIdx = Math.min(this.index + batchSize, this.sentence.length);
        //     const chunk = this.sentence.slice(this.index, endIdx);
        //     this.printCB(chunk);
        //     this.index = endIdx;
        // }, interval,this)
        // 普通TextUse优化后 of 逐字打印
        this.printNormalText();
    },
    printNormalText(){
        if (this.animationFrame) {
            cancelAnimationFrame(this.animationFrame);
        }

        this.index = 0;
        const baseSpeed = 40; // 基础速度
        const maxSpeed = 120; // Max速度

        const printNextChunk = (timestamp) => {
            if (this.index >= this.sentence.length) {
                this.stop();
                return;
            }

            // 动态计算应打印 of 字符数
            const elapsed = timestamp - this.lastTimestamp;
            const progress = this.index / this.sentence.length;
            const currentSpeed = baseSpeed + (maxSpeed - baseSpeed) * Math.min(progress / 0.3, 1);
            const targetChars = Math.ceil(elapsed * currentSpeed / 1000);

            // 计算本次要打印 of 字符
            const endIdx = Math.min(this.index + targetChars, this.sentence.length);
            const currentChunk = this.sentence.slice(this.index, endIdx);
            
            this.index = endIdx;

            // 传递当前这次要打印 of Text片段
            this.printCB(currentChunk);
            this.lastTimestamp = timestamp;

            // 继续下一帧OR结束
            if (this.index < this.sentence.length) {
                this.animationFrame = requestAnimationFrame(printNextChunk);
            } else {
                this.stop();
            }
        };

        this.animationFrame = requestAnimationFrame(printNextChunk);
    },
    printFn(){
        let sentenceArr = this.sentence.split('')
        this.printCB(sentenceArr[this.index])
        this.index++;
        if(this.index !== sentenceArr.length){
            this.t = workerTimer.setTimeout(()=>{
                this.printFn()
            },this.timer,this)
        }else{
            this.stop()
        }
    },
    stop() {
        if (this.animationFrame) {
            cancelAnimationFrame(this.animationFrame);
            this.animationFrame = null;
        }

        if(this.sIndexMap[`${this.sIndex}`]) {
            return;
        }
        this.sIndexMap[`${this.sIndex}`] = true;
        this.endCB({msg: 'end', index: this.sIndex});
        this.t && workerTimer.clearInterval(this.t);
        this.t = null;
    }
}

export default Print;