# End

## Node Overview

Responsible for presenting workflow processing results to users or downstream systems in the most appropriate way.

## Configuration Guidelines

The End node is the last node of the workflow, and it determines how the final output of the workflow is consumed.

It provides two completely different return patterns to adapt to diverse application scenarios.

##### Mode One:

Return Variable

In this mode, the End node will output the processed and structured data from the workflow in **JSON format**.

1. Select **"Return Variable"** mode in the End node.

2. In the **"Output Variables"** area, add the variables you want to ultimately output.

3. These variables usually need to **reference upstream node** output results (for example, reference Code node calculation results, LLM node answers, etc.).

4. Supports multiple basic types, including string (String), numbers (Integer, Number), boolean values (Boolean), time (Time), objects (Object), arrays (Array), and files (File).

![image-20250820180822543](image-20250820180822543.png)

##### Mode Two:

Return Text

In this mode, the End node will directly output a preset **text content** as the final reply to the user by the agent.

1. Select **"Return Text"** mode in the End node.

2. **Configure Output Variables**:

   - The "Output Variables" here are mainly used for **data pass-through**. When you need to use workflow results for **binding cards**, even if you selected the "Return Text" mode, you still need to define the variables to be passed to the card here.

3. **Write Reply Content**:

   - This is the final reply to the user and **cannot be empty**.
   
   - **Dynamic Reference**: In the reply content, you can use `{{Variable Name}}` syntax to reference End node output variables.

   - **Example**: Assuming the End node output variable is `weather_report`. You can write in the reply content:

   `According to the query results, today's weather is as follows: {{weather_report}}. Wish you a pleasant life!`

   - During workflow execution, `{{weather_report}}` will be replaced with the actual weather report content.

![image-20250820181140457](image-20250820181140457.png)

## Best Practices and Scenario Selection

| Scenario | Recommended Mode | Reason |
|:-----------------------------------|:---------------------------|:-----------------------------------------------------------|
| **Building pure text chatbots** | **Return Text** | You need to precisely control every reply of the chatbot, ensuring accurate tone, style, and information points. |
| **Results need to be rendered into custom cards** | **Return Variable** or **Return Text** | **Return Variable**: Card components need JSON data for dynamic rendering.<br>**Return Text**: If card content is fixed and only needs to pass small amounts of data, this mode can also be used. |
| **Performing complex data analysis or calculation** | **Return Variable** | The core output of the workflow is data (such as analysis reports, calculation results), and JSON format is most convenient for display, archiving, or being called by other systems. |
| **Need LLM to "humanize" the results** | **Return Variable** | Let LLM do what it does best: transform cold data (JSON) into warm, natural language replies. |
