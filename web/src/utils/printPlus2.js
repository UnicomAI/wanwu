/*根据finishis1OR2 when ，CheckYesNoprintend*/
import workerTimer from './worker'
import {parseSub, isSub} from "@/utils/util.js"

const Print = function (opt) {
    this.sentenceArr = []//存储待print of 句子 of Array
    this.sIndexMap={} 
    this.timer = opt.timer || 10; //print速度
    this.t = null;
    this.sIndex = 0 //Recordalreadyprint句子 of Index（avoidduplicateprint）
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

        //Ifin progressprintOR者printend
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
    this.sentence = sentence ? sentence.response : "" //currenttoprint of 句子
    this.timer = timer
    this.t = null
    this.index = 0 //currentprintto of 字符位置
    this.printCB = printCB //eachprint一字符 of Callback
    this.endCB = endCB //句子printend of Callback
    this.isCodeBlock = false // Add：markYesNois代码块
    this.codeBlockContent = '' // Add：存储代码块Content
    this.animationFrame = null
    this.lastTimestamp = performance.now(); // Add：each次LooperInit when Reset
    // 在Init when 检测YesNois代码块
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
            this.codeBlockContent = match[0]; // 整代码块Content
            this.sentence = match[0]; // 代码块内部Content（去掉```）
            this.index = this.sentence.length; // Add：代码块Directlyprint完毕
        }
    },
    start() {
        if(this.sentence === ''){
            this.printCB('')
            this.stop()
            this.index++;
            return
        }

        this.lastTimestamp = performance.now(); // Add：each次start都Reset

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

        
        // const batchSize = 10; // 推荐each次Output30字符
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
        // 普通TextUse优化after of 逐字print
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

            // 动态计算应print of 字符数
            const elapsed = timestamp - this.lastTimestamp;
            const progress = this.index / this.sentence.length;
            const currentSpeed = baseSpeed + (maxSpeed - baseSpeed) * Math.min(progress / 0.3, 1);
            const targetChars = Math.ceil(elapsed * currentSpeed / 1000);

            // 计算本次toprint of 字符
            const endIdx = Math.min(this.index + targetChars, this.sentence.length);
            const currentChunk = this.sentence.slice(this.index, endIdx);
            
            this.index = endIdx;

            // 传递current这次toprint of TextSegment
            this.printCB(currentChunk);
            this.lastTimestamp = timestamp;

            // Continue下一帧ORend
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