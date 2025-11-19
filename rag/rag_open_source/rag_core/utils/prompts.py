# 根据上述文本内容，简洁和专业的来回答用户的问题。如果无法从中得到答案，请说 “根据已知信息无法回答该问题”。回答必须是原文提及内容，答案请使用中文。 请问{question}""" [EN] Answer the user's questions concisely and professionally based on the above text content. If you can't get an answer out of it, say "The question cannot be answered based on known information." The answer must be the content mentioned in the original text. Please use Chinese for the answer. Please {question}"""
CITATION_INSTRUCTION = "您将获得与问题相关的一组参考信息回答用户问题。参考信息中包含多个上下文，每个上下文都以引用编号（如【x^】）开头，其中x是一个数字。请使用这些上下文，并在句子末尾引用相应的上下文（如果适用）。在引用来源中的信息时，请使用相应上下文开头的【x^】中的编号来标识这句答案的来源出处引用自这个上下文，例如【x^】。如果一个句子来自多个上下文，请列出所有对应的引用编号，例如【3^】【5^】。注意：你所生成的答案应至少包含一个上下文引用。并且你所给出的【x^】中的x编号必须在上下文开头的【x^】中真实存在，不要捏造生成不存在的引用编号。"

DEFAULT_ANSWER_INSTRUCTION = "请仅基于提供的参考信息中上下文提供答案。如果提供的参考信息中的所有上下文对回答问题均无帮助，请直接输出“根据已知信息，无法回答该问题。”"

PROMPT_TEMPLATE = '''
你是一个问答助手，主要任务是汇总参考信息回答用户问题, 请只根据参考信息中提供的上下文信息回答用户问题。
{citation}
{default_answer}
用户问题：
{question}
参考信息
```
{context}
```
输出要求
 1. **参考信息文字输出要求**：基于参考信息，输出文字内容。
 2. **参考信息中提及图片链接情况的输出要求**：若参考信息提及图片链接且链接格式符合markdown语法规范：“![图片标题](图片链接)” 。请按此链接格式将相关图像内容附加输出，注意确保图片链接格式完整不被截断。若参考信息未提及图片链接则忽略此规则并注意不要随意捏造图片链接，在答案输出中不要体现此条指令信息的任何内容。
 3. **输出语言要求**：必须使用与问题相同的语言回答用户的问题。
'''
