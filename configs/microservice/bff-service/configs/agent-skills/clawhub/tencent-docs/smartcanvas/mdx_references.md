==============================================================================
AI INGEST SPECIFICATION
Version: 2.1.0
==============================================================================
本文件定义用于生成与解析 MDX 文档的强制性规范。
该规范主要面向 AI 生成内容使用，同时兼顾人工可读性。
任何未在本规范中明确允许的语法、组件、属性与取值，均视为禁止。

-------------------------------------------------------------------------------
AI 解析约定
-------------------------------------------------------------------------------

本规范使用固定的分隔符来表示文档结构层级。

AI 在解析本规范时，必须将以下分隔符视为结构标记：

    "======"  表示章节(chapter)
    "------"  表示小节(section)

这些分隔符用于定义本规则文档结构层级， 并非装饰性格式。
严禁将该分隔符应用到 MDX 文档来定义结构(禁止)。

===============================================================================
第 0 章：总体原则
===============================================================================

-------------------------------------------------------------------------------
Markdown语法与 MDX 组件使用规则
-------------------------------------------------------------------------------
Markdown 优先
    Markdown 是主要内容表达形式。
    仅当 Markdown 无法表达所需结构或语义(例如：表格、分栏、复杂引用、需要属性的块等)时，才允许使用 MDX。

行内样式一律使用 Mark(强制)
    所有行内样式必须使用 <Mark>。
    禁止使用 Markdown 的 **bold** / *italic* / ~~strike~~ / __underline__ 等行内样式。

数学公式
    优先使用 Markdown 的 $math$ / $$math$$ 等数学公式。

未知组件规则(强制)
    AI 只能使用本规范中定义的组件。
    如果需要表达的结构没有对应组件，必须退化为 Markdown 表达。
    禁止生成任何未在本规范中声明的组件。

-------------------------------------------------------------------------------
缩进、换行规则
-------------------------------------------------------------------------------
缩进单位
    一级缩进固定为 4 个空格；
    禁止使用 Tab。

块缩进规则(强制)
    块级组件必须“三段式多行写法”(强制)
    块级组件禁止写成一行(即使内容很短)。

    错误(禁止):
        <Heading level="1">标题</Heading>
    正确(强制):
        <Heading level="1">
            标题
        </Heading>

块内内容与子块的缩进规则(强制)
    块级组件的“直接内容行”(纯文本 / Mark / Link)必须缩进到开标签下一层：
        内容行缩进 = 开标签缩进 + 4 空格
    子块(嵌套块级组件)同样必须缩进一层：
        子块开标签缩进 = 父块开标签缩进 + 4 空格
    同一层级的兄弟块必须保持一致的缩进深度。
    禁止出现“块级嵌套但无缩进”的写法。

行内内容的换行限制(强制)
    Mark / Link 必须与周围文本处于同一行文本流中。
    禁止为了排版在句子中间插入换行，造成“软换行”。
    单段内容(例如 Callout 内的一段说明)必须保持连续文本流，除非明确需要软换行。

-------------------------------------------------------------------------------
表达式能力限制(强制)
-------------------------------------------------------------------------------
以下全部禁止：
    任何 {...} 表达式属性(例如 level={3})
    任何 MDX expression
    任何 ESM(import / export)
    任何未定义组件

-------------------------------------------------------------------------------
属性语法规则(强制)
-------------------------------------------------------------------------------
禁止使用表达式(再次强调)
    禁止使用 {}, 包括属性值、子表达式等

属性值必须使用双引号(强制)
    错误(禁止):
        <Heading level=1>

    错误(禁止):
        <Heading level='1'>

    错误(禁止):
        <Heading level={1}>

    正确(强制):
        <Heading level="1">

布尔属性不写值(强制):
    以下属性为布尔属性，出现即为 true, 不得写 ="true":
        Mark: bold / italic / underline / strike
        Todo: checked(如有)
    严禁使用 false 显式设置。
    如果后续章节中明确了属性为布尔类型，遵循该规则
        正确:
            <Mark bold>文本</Mark>
            <Todo checked>
                已完成
            </Todo>
        不推荐(禁止生成):
            <Mark bold="true">文本</Mark>
            <Todo checked="true">
                已完成
            </Todo>
        错误(禁止):
            <Mark bold="false">文本</Mark>
            <Todo checked="false">
                已完成
            </Todo>

-------------------------------------------------------------------------------
颜色 Token 规则(强制)
-------------------------------------------------------------------------------
所有颜色相关属性均为 token 白名单。
禁止使用任何 CSS 颜色值(例如 #fff、rgb(...)、red 等)。
颜色表名单会在末尾附录中定义。

===============================================================================
第 1 章：页面级属性(Frontmatter)
===============================================================================
文档的页面级属性使用 frontmatter 来定义

位置与格式(强制)
    文档顶部必须包含 YAML frontmatter。
    frontmatter 必须是文档的第一段内容，前面不得出现任何字符(包括空行)。
    frontmatter 必须以 --- 开始，并以 --- 结束。
    frontmatter 中必须包含 title 字段，且为非空字符串。

允许字段(强制白名单)
    仅允许以下字段(其余字段禁止出现)：
        title
        cover
        icon
        fontFamily
        fontSize
        spacing

取值规则(强制)
    title (required)
        表示文档的标题。
        允许字符串。
    cover (recommended)
        建议都添加，除非文档内容非常不适合添加。
        表示文档的头图，横幅展示。
        允许图片链接。
    icon (optional)
        必须为单个 emoji 字符,禁止多个 emoji。
        禁止文本或图片 URL。

    fontFamily (optional)
        仅允许：
            simsun
            kaiti
            default
        含义：
            simsun  → 宋体
            kaiti   → 楷体
            default → 默认字体(黑体体系)

    fontSize (optional)
        仅允许：
            small
            default
            large

    spacing (optional)
        仅允许：
            compact
            default
            loose

默认行为(重要强制)
    fontFamily / fontSize / spacing 的默认值均为 default。
    如无明确需求：必须省略这三个字段；禁止为了“完整性”自动写入 default。

    示例(正确)：
        ---
        title: React 学习路线
        icon: ⚛️
        cover: upload_image返回的image_id
        ---

    示例(错误：不应显式写默认值)：
        ---
        title: React 学习路线
        cover: upload_image返回的image_id
        fontFamily: default
        fontSize: default
        spacing: default
        ---

==============================================================================
第 2 章: Block Components (块级组件)
==============================================================================

------------------------------------------------------------------------------
Paragraph
------------------------------------------------------------------------------
用途
    段落

属性
    textAlign (文本对齐方式)
    blockColor (段落背景颜色)

取值规则
    textAlign (optional)
        left
        center
        right
    blockColor (optional)
        BLOCK_COLORS

子元素
    Text
    Mark
    Link

示例
    <Paragraph textAlign="right">
        <Mark bold>加粗</Mark>普通文本
    </Paragraph>

限制规则
    如不需要段落级属性：必须不包 Paragraph(直接输出纯文本/Mark/Link)。
    仅当需要段落级属性(例如 textAlign、blockColor)时才使用 Paragraph。
    默认为文本左对齐，不需要显示设置 textAlign 为 left。

------------------------------------------------------------------------------
Heading
------------------------------------------------------------------------------
用途
    标题

属性
    textAlign (文本对齐方式)
    blockColor (段落背景颜色)
    level (标题层级)

取值规则
    textAlign (optional)
        left
        center
        right
    blockColor (optional)
        BLOCK_COLORS
    level (required)
        数字字面量字符串 1-6

子元素
    Text
    Mark
    Link

示例
    <Heading level="1" blockColor="red">
        标题1
    </Heading>

限制规则
    当标题只需要 level 属性且无其他属性时，应优先使用 Markdown 标题语法：
        # 标题1
        ## 标题2
    当需要额外属性（例如 textAlign、blockColor）时，必须使用 Heading 组件。
    标题支持标题 1-6。

    frontmatter.title 定义页面或文档的唯一标题。
    正文中允许使用一级标题 (#)， 但不得将与 frontmatter.title 内容相同的一级标题放在正文开头。
    如果正文第一段为与 frontmatter.title 相同的一级标题 (#)， 则视为重复标题，生成时应避免。
    正文中的一级标题仅用于章节划分，不表示页面标题。

------------------------------------------------------------------------------
BlockQuote
------------------------------------------------------------------------------
用途
    引用块

属性
    textAlign (文本对齐方式)
    blockColor (段落背景颜色)

取值规则
    textAlign (optional)
        left
        center
        right
    blockColor (optional)
        BLOCK_COLORS

子元素
    所有块级元素

示例
    <BlockQuote>
        引用内容

        <BlockQuote>
            子引用内容
        </BlockQuote>
    </BlockQuote>

限制规则
    当引用块中只有一个段落时，且无属性设置时，采用 Markdown 的表达方式。
    当引用块中有嵌套或者多个段落时，采用 Mdx 表达。

------------------------------------------------------------------------------
Callout
------------------------------------------------------------------------------
用途
    高亮块

属性
    blockColor (高亮背景颜色)
    borderColor (高亮边框颜色)
    icon (高亮块左上角 icon)

取值规则
    blockColor (required)
        BLOCK_COLORS
    borderColor (required)
        BORDER_COLORS
    icon (optional)
        单个 Emoji 字符

子元素
    所有块级元素

示例
    <Callout icon="⚠️" blockColor="yellow" borderColor="light_orange">
        警告

        警告内容...
    </Callout>

限制规则
    blockColor 和 borderColor 建议使用同一色系, 除非需要反差场景。
    单段 Callout 文本必须保持连续文本流(禁止句中人为换行)。

------------------------------------------------------------------------------
ColumnList
------------------------------------------------------------------------------
用途
    分栏容器

属性
    无

取值规则
    无

子元素
    Column

示例
    <ColumnList>
        <Column>
            分栏左
        </Column>
        <Column>
            分栏右
        </Column>
    </ColumnList>

限制规则
    仅表达容器，无需设置任何属性。
    子元素中必须为 Column，且必须至少存在一个。

------------------------------------------------------------------------------
Column
------------------------------------------------------------------------------
用途
    分栏实体 item

属性
    width (宽度)

取值规则
    width (recommended)
        带 % 的百分比字符串

子元素
    所有块级元素

示例
    <ColumnList>
        <Column width="20%">
            分栏左
        </Column>
        <Column width="60%">
            分栏中
        </Column>
        <Column width="20%">
            分栏右
        </Column>
    </ColumnList>

限制规则
    不能独立定义，只允许出现在 ColumnList 下。
    width 代表宽度百分比，不设置的会均分剩下的宽度，但建议都根据内容进行合理设置。

------------------------------------------------------------------------------
Divider
------------------------------------------------------------------------------
用途
    分割线

属性
    blockColor (分割线颜色)

取值规则
    blockColor (optional)
       DIVIDER_COLORS

子元素
    无

示例
    <Divider blockColor="sky_blue" />

限制规则
    使用自闭合标签。
    如果不需要设置颜色，使用 Markdown 的 --- 表达方式。

------------------------------------------------------------------------------
Image
------------------------------------------------------------------------------
用途
    图片

属性
    src (图片地址)
    alt (图片说明)
    align (对齐方式)
    width (宽度)
    height (高度)

取值规则
    src (required)
        图片地址字符串
    alt (recommended)
        文字字符串
    align (optional)
        left
        center
        right
    width (optional)
        数字字面量字符串，单位 px
    height (optional)
        数字字面量字符串，单位 px

子元素
    无

示例
    <Image src="https://example.com/image.png" alt="示例图片" align="right" />

限制规则
    图片均采用 Mdx 表达方式，不使用 Markdown 表达。
    align 默认是 center，可不显式设置。
    width 和 height 需要根据原始比例设置，如果获取不到原始比例，只需要设置宽度。不设置的话最大宽度为文档容器宽度。

------------------------------------------------------------------------------
Todo
------------------------------------------------------------------------------
用途
    待办列表

属性
    blockColor (段落背景颜色)
    checked (是否完成)

取值规则
    blockColor (optional)
        BLOCK_COLORS
    checked (optional)
        布尔类型，不用显式设置值，存在属性即代表 true

子元素
    Text
    Mark
    Link

示例
    <Todo>
        任务1
        <Todo checked>
            任务1-1
        </Todo>
        <Todo>
            任务1-2
        </Todo>
    </Todo>
    <Todo checked>
        任务2
    </Todo>


限制规则
    每一个 Todo 代表一个待办项，连续的组成一个视觉列表。
    允许有子待办，但必须放在待办正文的后面。
    Todo 第一个 child 必须为行内文本流或Mark，Link。
    Todo 后面的 child 可以为任意块元素，但建议子任务嵌套或需要混合使用无序列表有序列表的场景。

------------------------------------------------------------------------------
BulletedList
------------------------------------------------------------------------------
用途
    无序列表

属性
    blockColor (段落背景颜色)

取值规则
    blockColor (optional)
        BLOCK_COLORS

子元素
    Text
    Mark
    Link
    所有块级元素

示例
    <BulletedList>
        无序列表
        <BulletedList>
            无序子列表
        </BulletedList>
    </BulletedList>
    <BulletedList>
        无序列表
    </BulletedList>

    错误(禁止):
        <BulletedList>
            无序列表1

            无序列表2

            无序列表3
        </BulletedList>


限制规则
    每一个 BulletedList  代表一个列表项，连续的组成一个视觉列表。
    第一个 child 必须为行内文本流或Mark，Link, 表示列表项文本内容。
    后面的 child 可以为任意块元素，但建议子列表嵌套或需要混合使用待办列表、无序列表、有序列表的场景。

------------------------------------------------------------------------------
NumberedList
------------------------------------------------------------------------------
用途
    有序列表

属性
    blockColor (段落背景颜色)

取值规则
    blockColor (optional)
        BLOCK_COLORS

子元素
    Text
    Mark
    Link
    所有块级元素

示例
    <NumberedList>
        有序列表1
        <NumberedList>
            有序列表1.1
        </NumberedList>
        <NumberedList>
            有序列表1.2
        </NumberedList>
    </NumberedList>
    <NumberedList>
        有序列表2
    </NumberedList>

    错误(禁止):
        <NumberedList>
            有序列表1

            有序列表2

            有序列表3
        </NumberedList>


限制规则
    每一个 NumberedList 代表一个列表项，连续的组成一个视觉列表。
    第一个 child 必须为行内文本流或Mark，Link, 表示列表项文本内容。
    后面的 child 可以为任意块元素，但建议子列表嵌套或需要混合使用待办列表、无序列表、有序列表的场景。
------------------------------------------------------------------------------
Table
------------------------------------------------------------------------------
用途
    表格

属性
    无

取值规则
    无

子元素
    TableRow

示例
    <Table>
        <TableRow>
            <TableCell>
                cell A1
            </TableCell>
            <TableCell>
                cell A2
            </TableCell>
        </TableRow>
    </Table>


限制规则
    表格使用 Mdx 表达，禁止使用 Markdown 语法表达。

------------------------------------------------------------------------------
TableRow
------------------------------------------------------------------------------
用途
    表格行

属性
    无

取值规则
    无

子元素
    TableCell

示例
    <Table>
        <TableRow>
            <TableCell>
                cell A1
            </TableCell>
            <TableCell>
                cell A2
            </TableCell>
        </TableRow>
    </Table>


限制规则
    禁止单独使用，仅可作为 Table 的子元素来表达行容器。

------------------------------------------------------------------------------
TableCell
------------------------------------------------------------------------------
用途
    表格单元格

属性
    无

取值规则
    无

子元素
    除 Table 外的块元素

示例
    <Table>
        <TableRow>
            <TableCell>
                cell A1
            </TableCell>
            <TableCell>
                cell A2
            </TableCell>
        </TableRow>
    </Table>


限制规则
    禁止单独使用，仅可作为 TableRow 的子元素来表达单元格容器。

------------------------------------------------------------------------------
MathBlock
------------------------------------------------------------------------------
用途
    数学公式

属性
    width (宽度)

取值规则
    width (optional)
        数字，单位为像素

子元素
    只允许唯一一个 mardown math

示例
    <MathBlock>
      $$
      i\hbar\frac{\partial}{\partial t}\Psi(\vec{r},t) = \left[-\frac{\hbar^2}{2m}
      abla^2 + V(\vec{r},t)\right]\Psi(\vec{r},t)
      $$
    </MathBlock>

限制规则
    如不指定 width 属性，优先使用 Markdown math 表达，不用包 MathBlock。

==============================================================================
第 3 章: Inline Components (行内组件)
==============================================================================

------------------------------------------------------------------------------
Mark
------------------------------------------------------------------------------
用途
    带样式文本

属性
    bold(加粗)
    italic(斜体)
    underline(下划线)
    strike(中划线)
    color(文本颜色)
    backgroundColor(文本背景色)

取值规则
    bold (optional)
        布尔属性不写值
    italic (optional)
        布尔属性不写值
    underline (optional)
        布尔属性不写值
    strike (optional)
        布尔属性不写值
    color(optional)
        TEXT_COLORS
    backgroundColor (optional)
        BLOCK_COLORS

子元素
    文本

示例
    <Mark bold>重点内容</Mark><Mark color="yellow">警告</Mark>

限制规则
    Mark 必须单行书写(开始标签、内容、结束标签在同一行)。
    Mark 不得被拆行，不得在 Mark 前后额外插入换行造成软换行。

------------------------------------------------------------------------------
Link
------------------------------------------------------------------------------
用途
    超链接

属性
    href (链接地址)

取值规则
    href (required)
        链接文本

子元素
    文本

示例
    <Link href="...">文本</Link>

限制规则
    Link 必须单行书写(开始标签、内容、结束标签在同一行)。
    Link 不得被拆行，不得在 Link 前后额外插入换行造成软换行。

==============================================================================
APPENDIX
==============================================================================

------------------------------------------------------------------------------
BLOCK_COLORS
------------------------------------------------------------------------------
用于：
    blockColor
    Mark.backgroundColor

允许值：
    default
    grey
    light_grey
    dark
    light_blue
    blue
    light_sky_blue
    sky_blue
    light_green
    green
    light_yellow
    yellow
    light_orange
    orange
    light_red
    red
    light_rose_red
    rose_red
    light_purple
    purple


------------------------------------------------------------------------------
BORDER_COLORS
------------------------------------------------------------------------------
用于：
    Callout.borderColor

允许值：
    default
    grey
    blue
    sky_blue
    green
    yellow
    orange
    red
    rose_red
    purple


------------------------------------------------------------------------------
DIVIDER_COLORS
------------------------------------------------------------------------------
用于：
    Divider.blockColor

允许值：
    default
    black
    light_grey
    grey
    light_blue
    blue
    light_sky_blue
    sky_blue
    light_green
    green
    light_yellow
    yellow
    light_orange
    orange
    light_red
    red
    light_rose_red
    rose_red
    light_purple
    purple


------------------------------------------------------------------------------
TEXT_COLORS
------------------------------------------------------------------------------
用于：
    Mark.color

允许值：
    default
    grey
    blue
    sky_blue
    green
    yellow
    orange
    red
    rose_red
    purple

==============================================================================
END
==============================================================================
