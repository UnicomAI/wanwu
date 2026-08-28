# 如何在万悟中接入OpenClaw机器人

## 一、导入镜像

下载Ubuntu.24.04_amd64.tar.gz镜像

文件下载地址：https://pan.baidu.com/e/1cupIcEP2RBwi_hOr4xQnFQ?pwd=ae86

```shell
# 解压 Ubuntu.24.04_amd64.tar.gz
tar -zxvf Ubuntu.24.04_amd64.tar.gz
# wsl --import <名称>  <安装位置>  <tar文件名>
wsl --import Ubuntu-24.04 .\Ubuntu-24.04 .\Ubuntu.24.04_amd64.tar --version 2
# 启动wsl
wsl -d Ubuntu-24.04
```

用户名：user/root 密码：123456

## 二、启动openclaw 插件

```shell
# 进入openclaw插件目录
cd ~/clawchat
# 启动openclaw插件
./deploy.sh
```

出现如下日志即为启动成功

![image.png](assets/ca6b461e-a91a-4858-aac7-c013d1918499.png)

## 三、万悟平台使用openclaw

在万悟平台选择模型后即可与openclaw进行对话

![image.png](assets/bcff1e85-cf34-4b47-b077-6220b9747223.png)