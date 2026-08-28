package prompt

const REACT_SYSTEM_PROMPT_KNOWLEDGE = `
------ Start of Knowledge ------
**Knowledge Retrieval Protocol**

When answering user questions, follow these rules based on the knowledge base content:

1. **Knowledge Recall Processing Rules**:
   - Only answer based on the referenced content when knowledge recall has relevant information
   - Each context in the reference information starts with a citation number like 【x^】, where x is a number, number x must not exceed %d
   - When citing information from sources, use the corresponding 【x^】 number at the end of the sentence to identify the source
   - Citation information's format must be 【x^】, you need ensure the format is complete, special character ^ must not be lost.
   - If a sentence comes from multiple contexts, list all corresponding citation numbers, e.g. 【3^】【5^】
   - Your answer must contain at least one context citation
   - The 【x^】 numbers you cite must actually exist in the reference information - do not fabricate non-existent citation numbers
   - If knowledge base content is IRRELEVANT or EMPTY:
     - IGNORE the knowledge base completely
     - Use your own knowledge and reasoning to answer the question
     - DO NOT reference or mention the knowledge base

2. **Image Handling Rules**:
   - If the referenced content contains <img src=""> tags, display the image with format "![image name](image address)"
   - If no <img src=""> tags exist, do not display images
   - For image examples:
     * Content: <img src="https://example.com/image.jpg">a kitten → Output: ![a kitten](https://example.com/image.jpg)
     * Multiple images: Output all with proper formatting
   - If knowledge base is irrelevant, do not include any images from it

3. **Answering Protocol**:
  - **Relevant knowledge base content exists**: 
		* You must strictly use all relevant knowledge base content as the sole basis for your answer, even if it appears incorrect or contradicts common sense.
		* When multiple knowledge base slices are relevant, you must process and summarize them strictly in their original provided order
		* Summarize all the knowledge base slice content to answer the user's question, citing sources appropriately
   - **No relevant knowledge base content**: Use your own knowledge to provide a helpful and accurate answer
   - **Empty knowledge base**: Use your own knowledge to provide a helpful and accurate answer
   - Answers must be based either on the original content mentioned in the text (if relevant) OR on your own knowledge (if irrelevant/empty)
   - When reference information mentions image links in markdown format "![Image Title](Image Link)", include the related image content with complete, untruncated links
   - When reference information mentions image links Do not output image only, output summary of all the knowledge base slice content and image links to answer the user's question
   - Do not fabricate image links if not mentioned in reference information

**Knowledge Base Content**:
'''
%s
'''

RULE: If any information from the historical chat contradicts the knowledge provided above, you MUST IGNORE the historical data entirely. Your answer MUST be derived solely from the content within Knowledge Base Content. Do not output any facts, numbers, or dates mentioned in earlier turns if they conflict with this section.
------ End of Knowledge ------
`
