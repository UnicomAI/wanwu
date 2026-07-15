# OpenClaw通过wsl导出

## 一、前置准备

1.  已安装wsl2环境
    
2.  已安装ubuntu20.04（其他版本也可以）
    

## 二、Ubuntu默认安装位置（了解即可无需操作）

查看wsl 子系统Ubuntu20.04默认安装位置

#### 通过文件资源管理器查找

1.  打开资源管理器
    
2.  地址栏输入：
    
    ```plain
    %LOCALAPPDATA%\Packages
    ```
    
3.  查找以 `CanonicalGroupLimited.` 开头的文件夹，例如：
    

```plain
CanonicalGroupLimited.Ubuntu20...
```

示例位置如下：

![image](assets/e34c7f2c-f311-46a1-bf99-9886bf41d388.png)

## 三、wsl导出ubuntu系统镜像

1.  先查看所有WSL 
    
    ```shell
    wsl -l --all -v
    ```
    
    或者
    
    ```shell
    wsl --list --verbose
    ```
    
2.  停止正在运行的wsl
    
    ```shell
    wsl --shutdown
    ```
    
3.  导出到自定义路径
    
    ```shell
    wsl --export <发行版Linux> <tar包路径>
    ```
    
    **注意：发行版Linux**对应的是**NAME**字段的值
    
    示例：wsl --export Ubuntu-20.04 c:\ubuntu20.04.tar
    
    ![image.png](assets/12de86af-c823-4b42-99ab-f90029702c3e.png)
    
    ubuntu20.04.tar就是最终导出的tar镜像文件
    
4.  （可选）导出完成之后，如果需要**删除原有环境**，那执行以下操作，将原有的Linux注销。
    

```shell
wsl --unregister <发行版Linux>
```

示例：wsl --unregister Ubuntu-20.04