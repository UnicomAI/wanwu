# 文档（SmartCanvas）工具完整参考文档

腾讯文档（SmartCanvas）提供了一套完整的在线文档操作工具，支持创建、编辑智能文档。**内容格式使用 MDX，向下兼容全部 Markdown 语法**——标题、列表、表格、代码块、引用、图片、链接等标准 Markdown 可直接使用，同时支持分栏、高亮块、待办等高级排版组件。

---

## 目录

- [概念说明](#概念说明)
- [创建智能文档 — create_smartcanvas_by_mdx](#创建智能文档--create_smartcanvas_by_mdx)
- [统一编辑工具（推荐）](#统一编辑工具推荐)
  - [smartcanvas.read - 读取页面内容](#smartcanvasread)
  - [smartcanvas.find - 搜索文档内容](#smartcanvasfind)
  - [smartcanvas.edit - 编辑文档内容](#smartcanvasedit)
    - [边界场景处理规范](#边界场景处理规范)
- [典型工作流示例](#典型工作流示例)
  - [工作流一：用户指定了编辑位置（有查询意图）](#工作流一用户指定了编辑位置有查询意图)
  - [工作流二：用户未指定编辑位置（无查询意图）](#工作流二用户未指定编辑位置无查询意图)
  - [工作流三：在「XXX」后插入内容](#工作流三在xxx后插入内容)
  - [工作流四：修改「XXX」为新内容](#工作流四修改xxx为新内容)
  - [工作流五：删除「XXX」](#工作流五删除xxx)
  - [工作流六：直接追加内容到文档末尾](#工作流六直接追加内容到文档末尾)
  - [工作流七：创建分栏布局](#工作流七创建分栏布局)
  - [工作流八：向已有分栏中添加内容](#工作流八向已有分栏中添加内容)
  - [工作流九：修改分栏列数或宽度比例](#工作流九修改分栏列数或宽度比例)

---

## 概念说明

| 概念 | 说明 |
|------|------|
| `file_id` | 文档的唯一标识符，每个文档有唯一的 file_id |
| `page_id` | 页面 ID，Page 是文档的基本容器单元，可通过 `smartcanvas.read` 读取页面内容 |
| `Block ID` | 块 ID，`smartcanvas.read` / `smartcanvas.find` 返回的 MDX 中 `id` 属性值，用于 `smartcanvas.edit` 定位锚点 |

**文档结构**：

```
file_id（文档）
└── Page（页面）
    ├── Heading（标题，level 1-6）
    ├── Paragraph / Text（段落/文本）
    ├── BulletedList / NumberedList（列表）
    ├── Todo（待办事项）
    ├── Table（表格）
    ├── Callout（高亮块）
    ├── ColumnList（分栏布局）
    ├── Image（图片）
    └── ...（更多组件详见 mdx_references.md）
```

> ⚠️ **重要约束**：
> - 所有内容块（Block）必须挂载在 `Page` 下
> - `Page` 可以不指定父节点（挂载到根节点）
> - 完整的组件列表和规范详见 `mdx_references.md`

---

## 创建智能文档 — create_smartcanvas_by_mdx

**【创建文档的首选工具】** 创建排版丰富的在线智能文档。

**【格式说明】** 统一使用 mdx 格式（`content_format="mdx"`，默认值，无需显式传入）。

- **MDX 向下兼容全部 Markdown 语法**：标题、列表、表格、代码块、引用、图片、链接等标准 md 语法可直接写入 `mdx` 字段，无需转换
- **MDX 同时支持高级组件**：分栏布局 `ColumnList`、高亮块 `Callout`、待办列表 `Todo`、表格 `Table`、带样式文本 `Mark` 等丰富排版组件，适用于需要复杂排版和视觉效果的场景
- 生成包含 MDX 高级组件的内容时，须严格遵循 `mdx_references.md` 规范，并对照规范逐条自校验；纯 Markdown 语法无此约束

**【图片约束】** 所有图片禁止直接使用 http/https 外链，必须先调用 `upload_image` 工具上传获取 `image_id`，再填入对应位置：
- **MDX 组件**：封面图 `cover: image_id值`，正文图片 `<Image src='image_id值' alt='描述' />`
- **标准 Markdown 图片**：`![描述](image_id值)`
- 如果图片过大导致上传失败，必须先本地压缩图片再重新上传，严禁回退使用 URL。

**📖 MDX 规范详见：** `mdx_references.md`

### 工作流

【统一使用 mdx 格式（content_format 默认 "mdx"，兼容全部 Markdown 语法）】

步骤 1：【模板匹配 - 必须优先执行】
  根据用户需求，在下方【模板列表】中查找最匹配的模板：

  匹配优先级：精确匹配 > 场景匹配 > 分类匹配 > 通用生成
  - 精确匹配：用户需求与模板标题高度一致 → 直接读取对应引用文件
  - 场景匹配：用户需求与模板场景相符但细节不同 → 读取最接近的模板文件作为结构参考
  - 分类匹配：用户需求属于某分类但无精确模板 → 读取同类模板文件参考结构
  - 通用生成：无匹配模板 → 跳过，直接进入步骤 2 自由生成

  【找到匹配模板】→ 读取 smartcanvas/template/<引用文件名>
  - 以模板的 frontmatter 配置（icon、layout 等）为参考
  - 以模板的章节结构和 MDX 组件类型为骨架
  - ⚠️ 将模板中所有示例数据替换为用户实际信息，禁止照搬示例内容

  【模板列表】

| # | 模板标题 | 示例 Prompt | 引用文件 |
|---|----------|------------|----------|
| 1 | 阶段性工作总结 | 帮我生成一份Q1季度阶段性工作总结，岗位为市场运营，总结内容包括本季度核心目标达成情况、关键项目进展与成果、数据指标对比分析、团队协作亮点、存在的不足与改进方向、下一阶段工作规划，要有具体的数据支撑和案例说明。 | `q1_quarterly_marketing_operations_summary.mdx` |
| 2 | 晋升述职报告 | 帮我生成一份从P6晋升P7的述职报告，岗位为后端开发工程师，内容包括个人基本信息与晋升时间线、核心项目经历及个人贡献、技术能力成长与突破、业务价值创造与量化成果、团队影响力与mentor经验、未来发展规划与目标，突出技术深度和业务影响力。 | `p6_to_p7_promotion_report.mdx` |
| 3 | 实习生实习报告 | 帮我生成一份大三暑期实习报告，实习岗位为数据分析实习生，实习单位为一家互联网科技公司，内容包括实习单位简介、实习岗位职责、主要参与的项目与工作内容、学到的技能与工具、遇到的挑战与解决过程、个人成长与收获、对未来职业发展的思考。 | `summer_internship_report_data_analyst.mdx` |
| 4 | 试用期转正总结 | 帮我生成一份为期三个月的试用期转正工作总结，岗位为UI设计师，内容包括试用期工作概述、主要参与项目及设计成果、工作技能提升情况、团队协作与沟通表现、对公司文化的理解与融入、自我评价与不足反思、转正后的工作目标与计划。 | `ui_designer_probation_summary.mdx` |
| 5 | 项目复盘报告 | 帮我生成一份App 2.0版本改版项目的复盘报告，内容包括项目背景与目标、项目时间线与里程碑、核心成果与数据表现、项目过程中的亮点与创新、遇到的问题与踩坑记录、根因分析与改进措施、经验教训总结、后续迭代建议。 | `app_2_project_retrospective.mdx` |
| 6 | 品牌宣传方案 | 帮我生成一份新消费茶饮品牌的年度品牌宣传方案，品牌定位为年轻时尚健康，目标受众为18-30岁都市年轻人，内容包括品牌现状分析、年度宣传目标、核心传播策略、线上线下整合营销计划、KOL及社交媒体投放策略、重点campaign创意概念、预算分配建议、效果评估指标，全面地展示所有信息，整体篇幅约4000字。 | `new_tea_brand_annual_promotion_plan.mdx` |
| 7 | 产品需求文档 | 帮我生成一份电商平台会员积分系统的产品需求文档（PRD），内容包括需求背景与目标、用户场景分析、功能范围与优先级、核心功能详细描述（积分获取规则、积分消耗方式、会员等级体系、积分商城）、业务流程图说明、数据埋点需求、非功能性需求、版本迭代规划。 | `ecommerce_membership_points_prd.mdx` |
| 8 | 市场营销推广方案 | 帮我生成一份在线教育平台暑期大促的市场营销推广方案，活动周期为一个月，内容包括市场环境分析、目标用户画像、活动主题与核心卖点、推广渠道策略（信息流广告、社交媒体、KOL合作、社群运营）、促销机制设计、内容营销计划、预算分配与ROI预估、执行时间表、风险预案，论据充分，篇幅不少于4000字。 | `online_education_summer_marketing_plan.mdx` |
| 9 | 活动策划方案 | 帮我生成一份公司五周年庆典活动策划方案，参与人数约200人，包含线下晚宴和团建环节，内容包括活动主题与定位、时间地点安排、活动流程与环节设计（签到、开场表演、领导致辞、颁奖典礼、互动游戏、抽奖、晚宴）、场地布置方案、物料清单、人员分工、预算明细、应急预案。 | `company_5th_anniversary_event_plan.mdx` |
| 10 | 运营规划方案 | 帮我生成一份社区团购小程序的年度运营规划方案，内容包括业务现状与数据分析、年度运营目标与KPI拆解、用户增长策略、用户留存与活跃策略、供应链运营优化、团长管理体系、内容运营计划、数据驱动运营体系搭建、季度里程碑与资源需求、风险评估与应对策略。 | `community_group_buying_annual_operation_plan.mdx` |
| 11 | 商业计划书 | 帮我生成一份智能家居IoT创业项目的商业计划书，内容包括执行摘要、公司简介与愿景、市场分析与行业趋势、目标市场与用户画像、产品与服务介绍、核心竞争优势与壁垒、商业模式与盈利方式、营销与推广策略、团队介绍、财务预测与融资需求、风险分析与应对措施、发展规划与里程碑，整体篇幅在4000字左右。 | `smart_home_iot_business_plan.mdx` |
| 12 | 个人自媒体运营方案 | 帮我生成一份个人美食探店类自媒体账号的运营方案，目标平台为小红书和抖音，内容包括账号定位与人设打造、目标受众分析、内容规划与选题方向、拍摄与制作标准、发布频率与最佳发布时间、涨粉策略、互动运营技巧、变现路径规划、月度内容排期表、竞品账号分析与差异化策略。 | `food_review_self_media_operation_plan.mdx` |
| 13 | 副业计划 | 帮我生成一份针对上班族的知识付费副业计划，方向为职场技能培训，内容包括副业定位与目标、个人优势与资源盘点、目标受众与需求分析、产品体系设计（课程、社群、咨询）、平台选择与入驻策略、内容生产计划、推广引流方案、时间管理与精力分配、收入目标与成本预算、阶段性里程碑，整体篇幅约5000字。 | `office_worker_knowledge_side_business_plan.mdx` |
| 14 | 竞品分析报告 | 帮我生成一份短视频平台的竞品分析报告，分析对象为抖音、快手、视频号三个平台，内容包括分析目的与方法论、行业背景与市场规模、竞品基本信息对比、产品定位与核心功能对比、用户画像与用户规模、商业模式与变现能力分析、运营策略差异、技术能力对比、SWOT分析、对自身产品的策略建议，整体篇幅约5000字。 | `short_video_platform_competitive_analysis_2026.mdx` |
| 15 | 行业趋势分析报告 | 帮我生成一份2025年人工智能行业趋势分析报告，内容包括全球AI市场规模与增长趋势、核心技术发展方向（大模型、多模态、AI Agent）、重点应用场景与商业化进展、主要玩家竞争格局、投融资热点与资本动向、政策法规与监管趋势、行业面临的挑战与风险、未来3-5年趋势预测与机会点，信息要足够详细和全面，篇幅在4000-5000字。 | `2025_ai_industry_trend_analysis_report.mdx` |
| 16 | 用户调研报告 | 帮我生成一份在线办公协作工具的用户调研报告，调研方式包括问卷调查和深度访谈，内容包括调研背景与目标、调研方法与样本说明、用户基本画像分析、使用习惯与行为分析、核心需求与痛点挖掘、满意度与NPS分析、竞品使用情况对比、用户典型场景与案例、关键发现与洞察总结、产品优化建议，整体篇幅约4000字。 | `online_office_tool_user_research_report.mdx` |
| 17 | 市场可行性分析 | 帮我生成一份社区生鲜即时配送项目的市场可行性分析报告，内容包括项目概述与目标、市场环境分析（宏观环境PEST分析、行业现状）、目标市场规模测算、竞争格局与进入壁垒、目标用户需求验证、商业模式与盈利能力分析、运营模式与成本结构、风险评估与应对策略、投资回报预测、可行性结论与建议，信息应全面，论据充分，整体篇幅约5000字。 | `community_fresh_delivery_feasibility_report.mdx` |
| 18 | 产品体验评测报告 | 帮我生成一份智能手表产品体验评测报告，评测对象为Apple Watch和华为Watch GT系列，内容包括评测背景与方法、外观设计与做工对比、屏幕显示效果、健康监测功能体验（心率、血氧、睡眠）、运动追踪精准度、智能功能与生态体验、续航能力实测、佩戴舒适度、性价比分析、综合评分与推荐建议，内容详细信息全面，整体篇幅约4000字。 | `smartwatch_comparison_apple_watch_vs_huawei_gt.mdx` |
| 19 | 目标人群画像分析 | 帮我生成一份母婴电商平台目标人群画像分析报告，内容包括分析目的与数据来源、人群基本属性（年龄、地域、收入、学历）、消费行为特征（消费频次、客单价、品类偏好）、媒介触达习惯、决策因素与购买动机、典型用户分群与画像描述、用户生命周期阶段分析、营销触达策略建议，篇幅在4000-5000字。 | `maternity_ecommerce_user_persona_report.mdx` |
| 20 | 选址/选品分析报告 | 帮我生成一份咖啡店选址分析报告，备选地址为三个商圈（CBD写字楼区、大学城周边、社区商业街），内容包括选址标准与评估维度、各备选地址周边环境分析、人流量与客群分析、竞争对手分布情况、租金成本与性价比、交通便利性与可达性、商圈发展潜力评估、综合评分与排名、选址建议与风险提示，篇幅在3000字左右。 | `coffee_shop_location_analysis_report.mdx` |
| 21 | 商业模式分析报告 | 帮我生成一份共享充电宝行业的商业模式分析报告，内容包括行业概述与发展历程、主要玩家与市场份额、商业模式画布分析（价值主张、客户细分、渠道通路、收入来源、成本结构、关键资源、核心活动、重要伙伴）、盈利模式与单位经济模型、核心竞争要素分析、行业挑战与发展瓶颈、未来演变趋势与创新方向，信息要足够详细和全面，篇幅在4000-5000字。 | `shared_powerbank_business_model_report.mdx` |
| 22 | 求职自荐信 | 帮我生成一份应聘互联网公司产品经理岗位的求职自荐信，应聘者为有3年经验的产品经理，内容包括自我介绍与求职意向、与岗位匹配的核心能力、代表性项目经历与成果、对目标公司和岗位的理解、个人职业热情与发展期望、结尾致谢与联系方式，语言真诚有感染力且突出个人亮点。 | `internet_product_manager_cover_letter.mdx` |
| 23 | 推荐信 | 帮我生成一份由大学教授为学生撰写的研究生入学推荐信，被推荐人为计算机科学专业大四学生，内容包括推荐人自我介绍与推荐关系说明、对被推荐人学术能力的评价、研究项目参与情况与表现、个人品质与团队合作能力、与其他学生的横向比较、对其研究生阶段发展的期望、推荐结论，语言正式客观且有说服力。 | `graduate_admission_recommendation_letter.mdx` |
| 24 | 个人职业规划书 | 帮我生成一份应届毕业生的五年职业规划书，专业背景为金融学，目标行业为互联网金融，内容包括自我分析（兴趣、能力、价值观）、行业与职业分析、SWOT个人分析、职业目标设定（短期1年、中期3年、长期5年）、实现路径与行动计划、所需资源与技能提升计划、可能遇到的障碍与应对策略、评估调整机制，篇幅约3000字。 | `finance_graduate_career_plan.mdx` |
| 25 | 面试常见问题准备清单 | 帮我生成一份互联网公司产品经理岗位的面试常见问题准备清单，涵盖自我介绍、行为面试题（STAR法则）、专业能力题（产品设计、数据分析、用户研究）、案例分析题、压力面试题、反问环节建议，每个模块下准备多个问题并附带回答思路和框架，帮助面试者系统化准备。 | `internet_product_manager_interview_checklist.mdx` |
| 26 | 英文自我介绍 | 帮我生成一份适用于外企面试的英文自我介绍模板，时长约2-3分钟，内容包括基本信息与教育背景、工作经验概述、核心技能与专业优势、代表性成就、对目标岗位的热情与匹配度、简短的个人特质展示，提供不同场景版本（正式面试版、社交场合版），语言地道流畅有感染力。 | `english_self_introduction_for_interview.mdx` |
| 27 | 旅行攻略 | 帮我生成一份三天两晚的泉州旅行攻略，从深圳出发，注重自然景观和人文艺术和当地美食，经典容易出片，推荐旅行地点的交通方式以及住宿区域，行程可以安排得相对紧凑，同时罗列一些注意事项。 | `quanzhou_3_day_travel_guide.mdx` |
| 28 | 婚礼策划清单 | 帮我生成一份中式现代风格婚礼的策划清单，预算约15万元，婚礼规模约150人，在酒店举办，内容包括婚礼时间线与筹备进度表（婚前6个月到婚礼当天）、场地布置方案、婚庆团队选择要点、婚纱礼服与造型准备、婚礼流程安排（迎亲、仪式、宴席）、宾客管理与座位安排、婚品采购清单、预算分配明细、注意事项与避坑指南。 | `chinese_modern_wedding_planning_guide.mdx` |
| 29 | 生日派对策划清单 | 帮我生成一份小朋友6岁生日派对的策划清单，主题为太空探险，参与人数约20个小朋友和家长，在家中举办，内容包括派对主题设计与装饰方案、邀请函设计、场地布置清单（气球、横幅、桌布等）、派对流程与互动游戏设计、生日蛋糕与美食菜单、伴手礼准备、拍照打卡区设置、安全注意事项、预算清单、物品采购链接建议。 | `space_theme_6th_birthday_party_plan.mdx` |
| 30 | 家庭年度预算规划 | 帮我生成一份三口之家的年度家庭预算规划，家庭月收入约3万元，坐标二线城市有房贷，内容包括家庭财务现状盘点、年度收入预估、固定支出梳理（房贷、保险、教育）、弹性支出预算（餐饮、交通、娱乐、购物）、储蓄与投资目标、各月预算分配表、应急资金规划、大额支出计划（旅行、家电更换）、节流建议与开源思路、预算执行跟踪方法。 | `2026_family_annual_budget_plan.mdx` |
| 31 | 搬家物品整理清单 | 帮我生成一份从合租房搬到新家的搬家物品整理清单，内容包括搬家前准备工作时间线、物品分类整理方案（客厅、卧室、厨房、卫生间、书房）、需要打包的物品清单、需要丢弃或捐赠的物品筛选标准、搬家公司选择与比价要点、搬家当天流程安排、新家入住前需采购物品清单、水电气网络过户提醒、搬家后整理收纳建议。 | `shared_apartment_moving_checklist.mdx` |
| 32 | 健身训练计划 | 帮我生成一份为期12周的增肌健身训练计划，适合有半年健身基础的男性，每周训练5天，内容包括训练目标与身体数据记录、每周训练部位分配、每日训练动作详细安排（动作名称、组数、次数、休息时间）、热身与拉伸建议、饮食配合建议（蛋白质摄入、碳水循环）、补剂建议、每周进度检查指标、常见错误与纠正提示。 | `12_week_muscle_building_workout_plan.mdx` |
| 33 | 读书笔记 | 帮我生成一份《原则》(Ray Dalio)的读书笔记，内容包括书籍基本信息与推荐理由、作者简介、全书核心主旨与结构概览、各章节要点提炼、核心原则归纳（生活原则、工作原则、管理原则）、精彩语句摘录、个人感悟与思考、与自身工作生活的关联与应用、推荐阅读的相关书籍，整体篇幅约3000字。 | `principles_ray_dalio_book_notes.mdx` |
| 34 | 电影/书籍推荐清单 | 帮我生成一份适合职场人士的成长类书籍和电影推荐清单，包含10本书和10部电影，内容包括推荐主题分类（思维提升、沟通表达、领导力、时间管理、心理健康）、每个推荐作品的基本信息、一句话推荐语、核心看点与收获、适合阅读/观看的场景、难度和时间投入参考、按优先级排序的阅读/观看顺序建议。 | `career_growth_books_and_movies_recommendations.mdx` |
| 35 | 个人年度目标规划 | 帮我生成一份2026年个人年度目标规划，涵盖职业发展、财务管理、健康运动、学习成长、人际关系、生活品质六大维度，内容包括上一年度回顾与反思、各维度年度目标设定、目标拆解为季度和月度里程碑、关键行动计划与习惯养成、所需资源与支持、潜在障碍与应对策略、奖励机制设计、月度复盘检查模板。 | `2026_personal_annual_goal_plan.mdx` |
| 36 | 宠物养护指南 | 帮我生成一份新手养猫全面养护指南，适合第一次养英短蓝猫的铲屎官，内容包括接猫前的准备工作与必备用品清单、猫咪到家后的适应期指南、日常喂养方案（猫粮选择、喂食量、饮水）、疫苗驱虫计划、日常护理（梳毛、剪指甲、清洁耳朵）、常见疾病预防与识别、绝育建议与注意事项、行为习惯解读与训练建议、每月养猫费用预估，尽量详细，篇幅在4000字左右。 | `british_shorthair_cat_care_guide.mdx` |
| 37 | 家庭食谱/每周菜单 | 帮我生成一份家庭一周健康菜单规划，三口之家包含一个6岁儿童，注重营养均衡和荤素搭配，内容包括一周七天的三餐加下午茶安排、每餐的菜品搭配与营养分析、重点菜品的简易做法、每周食材采购清单与预估费用、食材保鲜与储存建议、儿童营养补充要点、周末亲子烹饪活动建议、节约时间的备餐技巧。 | `family_weekly_healthy_meal_plan.mdx` |
| 38 | 节日祝福文案集 | 帮我生成一份全年节日祝福文案集，涵盖春节、元宵节、情人节、妇女节、清明节、劳动节、母亲节、父亲节、端午节、七夕、中秋节、国庆节、重阳节、圣诞节等主要节日，每个节日提供3-5条不同风格的祝福文案（正式商务版、亲友温馨版、朋友圈文艺版、幽默趣味版），同时提供节日相关知识小科普。 | `annual_holiday_greeting_messages.mdx` |

步骤 2：【阅读 MDX 规范】
  阅读 mdx_references.md，了解 MDX 组件规范（组件列表、属性、取值白名单、格式约束）

步骤 3：【生成 MDX 内容】
  按规范生成包含 Frontmatter 和 MDX 组件的内容：
  - 有模板参考时：以模板结构为骨架，填入用户实际内容
  - 无模板参考时：根据文档类型自由设计结构

步骤 4：【图片处理 - 必须执行】

  **封面图（frontmatter cover）**：
  - **默认必须设置**：根据文档主题自行通过网络搜索合适图片并下载 → 调用 `upload_image` 上传 → image_id 填入 `cover: image_id值`
  - 仅当搜索或下载/上传失败时，才去掉 cover 字段

  **正文图片（`<Image>` 元素）**：
  - 若 MDX 中包含 `<Image>`：必须先调用 `upload_image` 上传获取 image_id，填入 `src` 属性（约束详见工具说明开头）

步骤 5：【自校验】
  对照 mdx_references 逐条自校验，确保格式合规（重点检查 cover 和 `<Image src>` 均为 image_id 而非 URL）

步骤 6：【调用工具创建文档】
  调用 create_smartcanvas_by_mdx 创建文档（传入 title + MDX 内容）
  从返回结果中获取 file_id 和 url
```

### 参数说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | ✅ | 文档标题。参数名必须为 `title`，不要使用 `doc_title`、`name`、`file_name` 等其他名称 |
| `mdx` | string | ✅ | 文档正文内容。MDX 向下兼容全部 Markdown 语法，标准 md 内容可直接填入；使用 MDX 高级组件（`Callout` / `ColumnList` / `Todo` / `Table` 等）时须严格遵循 `mdx_references` 规范并逐条自校验。图片约束见工具说明开头 |
| `content_format` | string | | 内容格式。默认 `"mdx"`，建议始终使用默认值——MDX 已向下兼容全部 Markdown 语法，无需切换 |

### 调用示例

```json
{
  "title": "项目需求文档",
  "mdx": "---\ntitle: 项目需求文档\nicon: 📋\n---\n\n# 项目需求\n\n<Callout icon=\"📌\" blockColor=\"light_blue\" borderColor=\"blue\">\n    本项目旨在开发一套智能文档管理系统。\n</Callout>\n\n## 功能需求\n\n<BulletedList>\n    文档创建功能\n</BulletedList>\n<BulletedList>\n    文档编辑功能\n</BulletedList>\n<BulletedList>\n    协作功能\n</BulletedList>"
}
```

### 返回值说明

```json
{
  "file_id": "doc_1234567890",
  "url": "https://docs.qq.com/doc/DV2h5cWJ0R1lQb0lH",
  "error": "",
  "trace_id": "trace_1234567890"
}
```

---

## 统一编辑工具（推荐）

> 💡 **推荐使用统一编辑工具**：`smartcanvas.read` + `smartcanvas.find` + `smartcanvas.edit` 组合，支持 MDX 格式内容、更简洁的 API 设计。

### smartcanvas.read

**功能**：读取智能文档指定页面的完整 MDX 格式内容。一次调用即返回页面全部内容。

**使用场景**：
- 在编辑文档前先阅读全文，了解文档结构和内容
- 获取页面完整内容用于分析、总结或摘要
- `smartcanvas.find` 找不到目标内容时，降级用本工具获取全文查找

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_id` | string | ✅ | 智能文档的唯一标识符 |
| `page_id` | string | | 要读取的页面 ID，为空时自动获取文档的第一个页面 |

**返回字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `content` | string | 页面完整的 MDX 格式文本内容 |
| `error` | string | 错误信息 |
| `trace_id` | string | 调用链追踪 ID |

**调用示例（读取文档第一个页面）**：

```json
{
  "file_id": "your_file_id"
}
```

**调用示例（读取指定页面）**：

```json
{
  "file_id": "your_file_id",
  "page_id": "page_abc123"
}
```

**返回示例**：

```json
{
  "content": "## 项目背景\n\n本项目旨在提升用户体验...\n\n## 总结\n\n以上是文档的全部内容。"
}
```

---

### smartcanvas.find

**功能**：根据文本搜索智能文档中的 Block，返回匹配 Block 的 ID 和 MDX 格式内容。搜索结果中的 Block ID 可作为锚点，用于 `smartcanvas.edit` 的精准编辑操作。

**使用场景**：
- 定位文档中某段内容的位置，获取 Block ID 作为编辑锚点
- 搜索包含特定关键词的内容块

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_id` | string | ✅ | 智能文档的唯一标识符 |
| `query` | string | ✅ | 搜索文本，系统将在文档所有页面中搜索包含该文本的 Block |

**返回字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `blocks` | array | 匹配的 Block 列表 |
| `blocks[].id` | string | Block 的唯一标识符（锚点 ID） |
| `blocks[].content` | string | Block 的 MDX 格式内容 |
| `error` | string | 错误信息 |
| `trace_id` | string | 调用链追踪 ID |

**调用示例**：

```json
{
  "file_id": "your_file_id",
  "query": "项目背景"
}
```

**返回示例**：

```json
{
  "blocks": [
    {
      "id": "block_abc123",
      "content": "## 项目背景\n\n本项目旨在提升用户体验..."
    }
  ]
}
```

---

### smartcanvas.edit

**功能**：编辑智能文档，支持 4 种操作类型：在指定位置前/后插入、删除、修改。

**操作类型说明**：

| Action | 说明 | id 参数 | content 参数 |
|--------|------|---------|----------|
| `INSERT_BEFORE` | 在指定 Block 前插入内容 | 锚点 Block ID（必填） | MDX 格式内容（必填） |
| `INSERT_AFTER` | 在指定 Block 后插入内容 | 锚点 Block ID（为空则追加到文档末尾） | MDX 格式内容（必填） |
| `DELETE` | 删除指定 Block | 要删除的 Block ID（必填，⚠️ 必须先通过 find/read 获取） | 不需要 |
| `UPDATE` | 修改指定 Block 的内容 | 要修改的 Block ID（必填，⚠️ 必须先通过 find/read 获取） | 新的 MDX 格式内容（必填） |

> ⚠️ **强制约束**：`UPDATE` 和 `DELETE` 操作的 `id` 参数**必须**来源于 `smartcanvas.find` 或 `smartcanvas.read` 的返回结果，**禁止**在未获取文档数据的情况下直接传入 id 执行 UPDATE 或 DELETE 操作。

> ⚠️ **readonly 约束**：当 `smartcanvas.find` 或 `smartcanvas.read` 返回的 MDX 内容中，某个块级组件（如 `<Table>`）带有 `readonly` 属性时，表示该组件及其所有子元素为只读状态。**禁止**使用只读组件或其内部子元素的 `id` 作为 `smartcanvas.edit` 的锚点（INSERT_BEFORE / INSERT_AFTER / UPDATE / DELETE 均不可用）。如需在只读组件附近操作，应选择只读组件上方或下方的非只读 Block 作为锚点。

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_id` | string | ✅ | 智能文档的唯一标识符 |
| `action` | enum | ✅ | 操作类型：INSERT_BEFORE / INSERT_AFTER / DELETE / UPDATE |
| `id` | string | 条件 | 锚点 Block ID，见上表说明 |
| `content` | string | 条件 | MDX 格式内容，见上表说明 |

**返回字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `error` | string | 错误信息 |
| `trace_id` | string | 调用链追踪 ID |

**调用示例（在指定 Block 后插入内容）**：

```json
{
  "file_id": "your_file_id",
  "action": "INSERT_AFTER",
  "id": "block_abc123",
  "content": "## 新章节\n\n这是插入的新内容。"
}
```

**调用示例（追加到文档末尾）**：

```json
{
  "file_id": "your_file_id",
  "action": "INSERT_AFTER",
  "content": "追加到文档末尾的内容"
}
```

**调用示例（删除指定 Block）**：

```json
{
  "file_id": "your_file_id",
  "action": "DELETE",
  "id": "block_abc123"
}
```

**调用示例（修改指定 Block）**：

```json
{
  "file_id": "your_file_id",
  "action": "UPDATE",
  "id": "block_abc123",
  "content": "## 修改后的标题\n\n这是更新后的内容。"
}
```

---

### 边界场景处理规范

#### 1. `ColumnList` 分栏删除边界场景

- **删除后只剩 1 个 Column**：不允许 `ColumnList` 中只有一个 `Column`，必须将整个 `ColumnList`（包含剩余 `Column` 内的所有内容）用 `UPDATE` 操作替换为普通块内容（将 `Column` 内的子块直接平铺输出，去掉 `ColumnList` / `Column` 容器）。
- **删除后剩余 2 个或更多 Column**：需要用 `UPDATE` 操作更新整个 `ColumnList`，重新均分或合理分配各 `Column` 的 `width`（例如两列各 `50%`，三列各 `33%`）。
- **操作方式**：上述两种情况均不能只 `DELETE` 单个 `Column`，必须对 `ColumnList` 整体执行 `UPDATE`，传入调整后的完整 MDX 内容。

```json
// 删除一列后只剩一列 → 将 ColumnList 整体替换为普通块内容
{
  "file_id": "your_file_id",
  "action": "UPDATE",
  "id": "columnlist_id",
  "content": "剩余 Column 内子块平铺后的 MDX 内容"
}
```

```json
// 删除一列后仍剩多列 → 更新整个 ColumnList 并重新分配 width
{
  "file_id": "your_file_id",
  "action": "UPDATE",
  "id": "columnlist_id",
  "content": "<ColumnList>\n    <Column width=\"50%\">\n        左列内容\n    </Column>\n    <Column width=\"50%\">\n        右列内容\n    </Column>\n</ColumnList>"
}
```

#### 2. `Callout` 内容清空边界场景

- 当用户要删除 `Callout` 内的全部内容时，`Callout` 本身也应一并删除，不允许保留空的 `Callout` 容器。
- **操作方式**：对 `Callout` 的 `id` 执行 `DELETE` 操作，而非仅删除其内部子块。

```json
{
  "file_id": "your_file_id",
  "action": "DELETE",
  "id": "callout_id"
}
```

#### 3. `BlockQuote` 内容清空边界场景

- 当 `BlockQuote` 内的全部内容被删除时，`BlockQuote` 本身也应一并删除，不允许保留空的 `BlockQuote` 容器。
- **操作方式**：对 `BlockQuote` 的 `id` 执行 `DELETE` 操作，而非仅删除其内部子块。

```json
{
  "file_id": "your_file_id",
  "action": "DELETE",
  "id": "blockquote_id"
}
```

#### 4. 列表（`BulletedList` / `NumberedList` / `Todo`）删除含子项的列表项边界场景

- 当删除某个列表项时，若该列表项下存在子列表项（嵌套的 `BulletedList` / `NumberedList` / `Todo`），子项不能悬空独立存在。
- **操作方式**：对该列表项的 `id` 执行 `DELETE` 操作，系统会连同其所有子项一并删除；若需保留子项内容，应先用 `UPDATE` 将子项内容提升到父级或平铺为独立块，再执行 `DELETE`。

```json
// 直接删除父项（子项一并删除）
{
  "file_id": "your_file_id",
  "action": "DELETE",
  "id": "parent_list_item_id"
}
```

#### 5. `TableRow` / `TableCell` 使用边界场景

- `TableRow` 禁止单独存在，只能作为 `Table` 的直接子元素；`TableCell` 禁止单独存在，只能作为 `TableRow` 的直接子元素。
- **禁止**使用 `TableRow` 或 `TableCell` 的 `id` 作为 `INSERT_BEFORE` / `INSERT_AFTER` 的锚点，向表格内部插入非表格结构的内容。
- **禁止**单独对 `TableRow` 或 `TableCell` 执行 `DELETE` 操作（删除单行/单格），如需修改表格结构，应对整个 `Table` 执行 `UPDATE`，传入调整后的完整表格 MDX 内容。
- **禁止**单独对 `TableCell` 执行 `UPDATE` 操作修改单元格内容，同样应对整个 `Table` 执行 `UPDATE`，传入完整表格 MDX 内容。
- 注意：`Table` 通常带有 `readonly` 属性，此时 `TableRow` / `TableCell` 的 `id` 同样不可用，任何操作均需绕开只读表格，选择其上方或下方的非只读 Block 作为锚点。

```json
// 修改表格内容（如删除某行、修改某单元格）→ 对整个 Table 执行 UPDATE
{
  "file_id": "your_file_id",
  "action": "UPDATE",
  "id": "table_id",
  "content": "<Table>\n    <TableRow>\n        <TableCell>\n            列1内容\n        </TableCell>\n        <TableCell>\n            列2内容\n        </TableCell>\n    </TableRow>\n</Table>"
}
```

---

### 图片编辑说明

编辑场景与创建场景图片约束一致（详见工具说明开头）：必须先调用 `upload_image` 上传获取 `image_id`，再设置到 `src` 属性中，严禁使用外部 URL。

**调用示例（使用 upload_image 上传后插入图片）**：

```json
// 步骤 1：先调用 upload_image 获取 image_id
// 步骤 2：将 image_id 设置到 content 的 Image src 属性中
{
  "file_id": "your_file_id",
  "action": "INSERT_AFTER",
  "id": "block_abc123",
  "content": "<Image src=\"upload_image返回的image_id\" alt=\"示例图片\" />"
}
```

---

## 典型工作流示例

> ⚠️ **定位策略**：有关键词 → 优先 `find`，找不到降级 `read`；无关键词 → 直接 `read`。**UPDATE / DELETE 前必须先通过 `find` 或 `read` 获取真实 Block ID，禁止跳过。**

### 工作流一：用户指定了编辑位置（有查询意图）

```
步骤 1：使用 find 搜索目标 Block
  → smartcanvas.find(file_id, query="用户指定的关键词")
  → 检查搜索结果

步骤 2A：find 找到匹配 Block
  → 将 find 返回的 Block 列表展示给用户确认
  → 用户确认锚点位置后，调用 smartcanvas.edit 传入确认的锚点 ID 执行操作

步骤 2B：find 未找到匹配 Block（降级）
  → 调用 smartcanvas.read(file_id) 读取文档全部内容
  → 在返回的 content 中查找目标内容
  → 根据找到的内容分析并猜测合适的锚点位置
  → 调用 smartcanvas.edit 执行编辑操作
```

### 工作流二：用户未指定编辑位置（无查询意图）

```
步骤 1：读取文档全部内容（⚠️ UPDATE/DELETE 操作此步骤为必须）
  → smartcanvas.read(file_id)
  → 返回的 content 即为页面完整 MDX 内容，了解文档结构

步骤 2：根据文档内容和用户意图猜测锚点位置，执行编辑操作
  → 插入到文档最前面：smartcanvas.edit(action=INSERT_BEFORE, id=首个Block ID, content=MDX内容)
  → 插入到文档最后面：smartcanvas.edit(action=INSERT_AFTER, id为空, content=MDX内容)
  → 插入到特定位置：smartcanvas.edit(action=INSERT_BEFORE/INSERT_AFTER, id=猜测的锚点ID, content=MDX内容)
  → 修改特定内容：smartcanvas.edit(action=UPDATE, id=目标Block ID, content=新MDX内容)【id 必须来自 find/read 结果】
  → 删除特定内容：smartcanvas.edit(action=DELETE, id=目标Block ID)【id 必须来自 find/read 结果】
```

### 工作流三：在「XXX」后插入内容

```
步骤 1：搜索定位目标 Block
  → smartcanvas.find(file_id, query="XXX")

步骤 2A：找到匹配 Block
  → 展示 find 结果给用户确认锚点位置
  → 用户确认后，调用 smartcanvas.edit(action=INSERT_AFTER, id=确认的锚点ID, content=MDX内容)

步骤 2B：未找到匹配 Block（降级）
  → smartcanvas.read(file_id) 获取全文
  → 根据全文内容猜测"XXX"附近的锚点位置
  → smartcanvas.edit(action=INSERT_AFTER, id=猜测的锚点ID, content=MDX内容)
```

### 工作流四：修改「XXX」为新内容

```
步骤 1：搜索定位目标 Block
  → smartcanvas.find(file_id, query="XXX")

步骤 2A：找到匹配 Block
  → 展示 find 结果给用户确认目标 Block
  → 用户确认后，调用 smartcanvas.edit(action=UPDATE, id=确认的Block ID, content=新MDX内容)

步骤 2B：未找到匹配 Block（降级）
  → smartcanvas.read(file_id) 获取全文
  → 根据全文内容定位目标位置
  → smartcanvas.edit(action=UPDATE, id=目标Block ID, content=新MDX内容)
```

### 工作流五：删除「XXX」

```
步骤 1：搜索定位目标 Block
  → smartcanvas.find(file_id, query="XXX")

步骤 2A：找到匹配 Block
  → 展示 find 结果给用户确认要删除的 Block
  → 用户确认后，调用 smartcanvas.edit(action=DELETE, id=确认的Block ID)

步骤 2B：未找到匹配 Block（降级）
  → smartcanvas.read(file_id) 获取全文
  → 根据全文内容定位目标位置
  → smartcanvas.edit(action=DELETE, id=目标Block ID)
```

### 工作流六：直接追加内容到文档末尾

```
步骤 1：直接追加到文档末尾（无需定位）
  → smartcanvas.edit(file_id, action=INSERT_AFTER, id为空, content=MDX内容)
```

### 工作流七：创建分栏布局

> 适用场景：用户希望在文档中新增一个左右分栏区域（如「左边放说明，右边放示例」）。

```
步骤 1：确定插入位置
  → 若用户指定了位置关键词：smartcanvas.find(file_id, query="关键词") 获取锚点 Block ID
  → 若用户未指定位置：smartcanvas.read(file_id) 获取全文，根据文档结构选择合适锚点

步骤 2：构造 ColumnList MDX 内容
  → 两列等宽示例（各 50%）：
     <ColumnList>
         <Column width="50%">
             左列内容（可包含 Heading、Paragraph、BulletedList 等任意块）
         </Column>
         <Column width="50%">
             右列内容
         </Column>
     </ColumnList>
  → 三列等宽示例（各 33%）：
     <ColumnList>
         <Column width="33%">
             第一列内容
         </Column>
         <Column width="33%">
             第二列内容
         </Column>
         <Column width="34%">
             第三列内容
         </Column>
     </ColumnList>
  ⚠️ 注意：ColumnList 至少需要 2 个 Column，width 之和应为 100%

步骤 3：调用 smartcanvas.edit 插入分栏
  → smartcanvas.edit(file_id, action=INSERT_AFTER, id=锚点Block ID, content=ColumnList MDX内容)
  → 若插入到文档末尾：id 为空
```

**调用示例**：

```json
// 在 block_abc123 后插入一个两列分栏
{
  "file_id": "your_file_id",
  "action": "INSERT_AFTER",
  "id": "block_abc123",
  "content": "<ColumnList>\n    <Column width=\"50%\">\n        ## 功能说明\n\n        这里描述功能的详细说明。\n    </Column>\n    <Column width=\"50%\">\n        ## 代码示例\n\n        这里放对应的代码示例。\n    </Column>\n</ColumnList>"
}
```

### 工作流八：向已有分栏中添加内容

> 适用场景：用户希望在某个已存在的分栏（ColumnList）的某一列中追加或修改内容。

```
步骤 1：读取文档内容，获取目标 ColumnList 的完整 MDX 结构
  → smartcanvas.find(file_id, query="分栏内已知的关键词")
    或 smartcanvas.read(file_id) 获取全文
  → 找到目标 ColumnList 的 id 及其完整 MDX 内容

步骤 2：在原有 MDX 基础上修改目标列的内容
  → 保持 ColumnList / Column 结构不变
  → 仅在目标 Column 内追加或修改子块内容
  ⚠️ 注意：不能单独对 Column 内的子块执行 INSERT_BEFORE/INSERT_AFTER，
     必须对整个 ColumnList 执行 UPDATE，传入完整的新 MDX 内容

步骤 3：调用 smartcanvas.edit 更新整个 ColumnList
  → smartcanvas.edit(file_id, action=UPDATE, id=ColumnList的Block ID, content=更新后的完整ColumnList MDX)
```

**调用示例**：

```json
// 在右列末尾追加一条说明（对整个 ColumnList 执行 UPDATE）
{
  "file_id": "your_file_id",
  "action": "UPDATE",
  "id": "columnlist_block_id",
  "content": "<ColumnList>\n    <Column width=\"50%\">\n        ## 功能说明\n\n        这里描述功能的详细说明。\n    </Column>\n    <Column width=\"50%\">\n        ## 代码示例\n\n        这里放对应的代码示例。\n\n        > 注意：示例仅供参考，请根据实际情况调整。\n    </Column>\n</ColumnList>"
}
```

### 工作流九：修改分栏列数或宽度比例

> 适用场景：用户希望将两列分栏改为三列，或调整各列宽度比例（如从 50/50 改为 30/70）。

```
步骤 1：读取文档内容，获取目标 ColumnList 的完整 MDX 结构
  → smartcanvas.find(file_id, query="分栏内已知的关键词")
    或 smartcanvas.read(file_id) 获取全文
  → 找到目标 ColumnList 的 id 及其完整 MDX 内容

步骤 2：构造调整后的完整 ColumnList MDX
  → 增加列：在原有 Column 基础上新增 Column，重新分配 width（各列 width 之和为 100%）
  → 调整宽度：修改各 Column 的 width 属性值
  → 减少列（删除后剩余 ≥ 2 列）：移除目标 Column，重新均分剩余列的 width
  → 减少列（删除后只剩 1 列）：将 ColumnList 整体替换为普通块内容（参见边界场景处理规范第 1 条）
  ⚠️ 注意：width 之和必须为 100%，且 ColumnList 至少保留 2 个 Column

步骤 3：调用 smartcanvas.edit 更新整个 ColumnList
  → smartcanvas.edit(file_id, action=UPDATE, id=ColumnList的Block ID, content=调整后的完整ColumnList MDX)
```

**调用示例（两列改三列）**：

```json
{
  "file_id": "your_file_id",
  "action": "UPDATE",
  "id": "columnlist_block_id",
  "content": "<ColumnList>\n    <Column width=\"33%\">\n        ## 第一列\n\n        第一列内容。\n    </Column>\n    <Column width=\"33%\">\n        ## 第二列\n\n        第二列内容。\n    </Column>\n    <Column width=\"34%\">\n        ## 第三列\n\n        新增的第三列内容。\n    </Column>\n</ColumnList>"
}
```

**调用示例（调整宽度比例为 30/70）**：

```json
{
  "file_id": "your_file_id",
  "action": "UPDATE",
  "id": "columnlist_block_id",
  "content": "<ColumnList>\n    <Column width=\"30%\">\n        ## 侧边说明\n\n        简短的辅助说明内容。\n    </Column>\n    <Column width=\"70%\">\n        ## 主要内容\n\n        详细的主体内容区域。\n    </Column>\n</ColumnList>"
}
```

---

> 📌 **提示**：`file_id` 可通过 `manage.search_file` 搜索获取，或从创建文档的返回结果中获取。所有内容块必须挂载在 `Page` 下，完整组件列表详见 `mdx_references.md`。

---
