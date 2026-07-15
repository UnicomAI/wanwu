# OpenClaw 接入 Skill 教程

OpenClaw 接入 Skill 主要有两种方式：**手动导入指定 Skill**、**创建自定义 Skill**，以下是详细操作步骤。

## 一、手动导入 Skill

### 前提准备

提前准备好需要导入的 Skill 文件夹，可从以下推荐网址获取 Skill 资源：

*   [Anthropics Skills 仓库](https://github.com/anthropics/skills)
    
*   [MCPMarket Skills 平台](https://mcpmarket.cn/skills/)
    

### 操作步骤

1.  打开 OpenClaw 界面 UI，进入 `config` → `Skills` 页面；
    
2.  选择 `load` 标签，点击 `Add` 按钮；
    
3.  在下方空白处填写本地 Skill 包的路径，建议开启 `watch skills` 选项，最后点击 `save` 保存；
    

> 示例：导入本地 pptx 技能包（配图） ![image](assets/b3aff098-194b-41a3-886d-34e6528a2085.png)

### 备选方式（Raw 模式/配置文件修改）

也可选择 raw 模式，或直接修改本地 `openclaw.json` 配置文件，添加如下配置：

```json
"skills": {
  "allowBundled": [
    ""
  ],
  "load": {
    "extraDirs": [
      "本地skill路径" // 替换为实际的Skill文件夹路径
    ],
    "watch": true
  }
}

```

### 验证结果

保存后，即可在 OpenClaw 的 `skills` 菜单中找到刚导入的 Skill。

> ![image](assets/1c1e4f52-0e41-44eb-9fc4-518de8dc15b7.png)

## 二、创建自定义 Skill（推荐）

该方式适合构建个性化专属 Skill，只需向 OpenClaw 描述需求，即可自动生成对应的 Skill 包，操作简单高效。

### 操作示例：创建“根据图片生成古诗”的 Skill

1.  **提出需求**：向 OpenClaw 明确说明“创建一个可根据图片文件生成古诗的 Skill”；
    
    > ![image](assets/11af1c7f-bd35-4c43-a24e-5c11d327d8c5.png)
    
2.  **等待生成**：Skill 包及对应程序的生成需要一定时间，请耐心等待；
    
    > ![image](assets/28f0fd86-c413-4758-afca-0b3dbb8cb4d3.png)
    
3.  **验证生成结果**：生成完成后，界面会显示新 Skill 的文件路径，可检查文件是否已成功生成；
    
    > ![image](assets/268e6c42-a6ea-4161-82df-780e86e81375.png)
    
4.  **查看 Skills 模块**：在 OpenClaw 的 `skills` 模块中可确认新 Skill 已添加；
    
    > ![image](assets/1bc9308e-241d-4bcf-a29f-50fefc91d62f.png)
    
5. **验证 Skill 效果**：使用该 Skill 测试“根据图片生成古诗”的功能是否正常；

> ![image](assets/65640362-9961-47a0-96cf-84702eadd03c.png)![image](https://alidocs.oss-cn-zhangjiakou.aliyuncs.com/res/8K4nyeZpmJdNjnLb/img/50d268c6-52bf-44ed-a633-0247c00766b8.png)