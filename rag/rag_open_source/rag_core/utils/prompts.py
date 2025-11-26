# Answer the user's questions concisely and professionally based on the above text content. If you can't get an answer out of it, say "The question cannot be answered based on known information." The answer must be the content mentioned in the original text. Please use English for the answer. Please {question}"""
CITATION_INSTRUCTION = "You will be given a set of reference information related to the question. The reference information contains multiple contexts, each starting with a citation number (e.g., [x^]), where x is a number. Please use these contexts and cite the corresponding context at the end of the sentence (if applicable). When citing information from a source, use the number from the [x^] at the beginning of the corresponding context to identify the source of the answer, e.g., [x^]. If a sentence comes from multiple contexts, list all corresponding citation numbers, e.g., [3^][5^]. Note: Your generated answer should contain at least one context citation. The x number in [x^] you provide must actually exist in the [x^] at the beginning of the context; do not fabricate non-existent citation numbers."

DEFAULT_ANSWER_INSTRUCTION = "Please provide answers based only on the provided reference information context. If all contexts in the provided reference information are not helpful for answering the question, please directly output: Based on the available information, I cannot answer your question."

PROMPT_TEMPLATE = '''
You are a Q&A assistant. Your main task is to summarize reference information to answer user questions. Please only answer user questions based on the context provided in the reference information.
{citation}
{default_answer}
User Question:
{question}
Reference Information
```
{context}
```
Output Requirements:
 1. **Text Output Requirement**: Based on the reference information, output text content.
 2. **Image Link Handling**: Do NOT include any image links in your response unless the user specifically asks about images, diagrams, or visual content. Ignore any markdown image links (![...](url)) in the reference information. Do not output any image URLs or markdown image syntax.
 3. **Output Language Requirement**: You MUST always respond in English, regardless of the language of the question.
'''
