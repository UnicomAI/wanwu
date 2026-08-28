# OpenClaw对接飞书

1.  浏览器打开飞书开放平台，并登录。地址：[https://open.feishu.cn/?lang=zh-CN](https://open.feishu.cn/?lang=zh-CN)
    
    ![image](assets/e5a4c59e-d5e8-4fa1-834e-a6335f1cebcd.png)
    
2.  进入开发者后台，点击创建企业自建应用，输入应用名称/描述，选择一个图标。
    
    ![image](assets/a26a058c-59e2-444a-8123-b4855f237c9b.png)
    
3.  点击左侧“添加应用能力”，选择“机器人”。
    
    ![image](assets/d566b2c2-b527-49da-b9d5-20472a951db7.png)
    
4.  点击左侧“权限管理”-“开通权限”，搜索以下权限并添加：
    

*   im:resource
    
*   contact:user.base
    
*   im:chat:readonly
    
*   im:message
    
*   contact:contact.base:readonly
    
*   im:message.p2p\_msg:readonly
    

![image](assets/b3f2172f-c19d-49d5-b31d-5b774e6f7b55.png)

1.  点击左侧“版本管理与发布”-“创建版本”，填写版本号，更新说明，然后保存，确认发布。
    
    ![image](assets/cf7201b7-a6d5-4ce2-9fff-0a31eefaa4d0.png)
    
2.  回到openclaw终端，依次执行以下命令。在飞书开放平台，凭证与基础信息可以看到appID和appSecret。
    
    ```plain
    openclaw plugins install @m1heng-clawd/feishu
    
    openclaw config set channels.feishu.appId "***"
    
    openclaw config set channels.feishu.appSecret "***"
    
    openclaw config set channels.feishu.enabled true
    
    openclaw config set channels.feishu.connectionMode websocket
    ```
    
3.  重启openclaw gateway
    
    ```plain
    openclaw gateway
    ```
    
    ![image](assets/c7d71dee-547e-4c5b-b48b-da7c5c681cc4.png)
    
4.  回到飞书开放平台，点击左侧“事件与回调”-“事件配置”，编辑“订阅方式”，选择“长连接”，保存。
    
    ![image](assets/e509718c-4d56-430f-b449-53ae22b4cf08.png)
    
    点击“添加事件”，输入“接收”过滤，勾选“接收消息”，确认添加。
    
    ![image](assets/aaf614f4-1984-4098-bba7-f284dce8fc37.png)
    
    点击顶部“创建版本”，再提交一个新版本。
    
    ![image](assets/d8d2fc7c-3985-4d44-a4c2-1e518cc0a276.png)
    
5.  打开飞书手机app，可以看到“开发者小助手”里面，点击“打开应用”，对它说句话 。
    

![image](assets/ignore-error,1.jpeg)