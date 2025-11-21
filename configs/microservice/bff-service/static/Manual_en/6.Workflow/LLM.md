# LLM

## Node Overview

Core features: In workflows, inject a "brain", giving it capabilities for understanding, reasoning, generation, and decision-making.

## Configuration Guidelines

Configuring the LLM node is essentially **selecting an appropriate expert and giving them clear, complete instructions**.

##### 1. Select Model

- **How to operate**: In the node configuration area "Model" dropdown menu, select a large language model.

  For detailed model import guide, see [Model Import Guide - Detailed](../Model Import Guide - Detailed.md)

- Supports users to select all imported platform LLMs and perform parameter configuration.

- **Recommendation**:

  - **Select as needed**: There is no "best" model, only the "most suitable" model.
  
    For simple text polishing, basic models are sufficient; for complex logical reasoning or code generation, more advanced models need to be selected.

##### 2. Configure Prompt

- **System Prompt**

  - **Function**: Defines the model's **core persona, role, and basic principles**. It sets a macro framework for the model, affecting all its subsequent thinking and answers.

  - **How to write**:

    - **Clear Role**: Directly tell the model "You are an XX".
      
      Example: "You are a professional tech blogger."

    - **Define Tasks**: Clearly instruct its core responsibilities.
      
      Example: "Your task is to interpret complex technology concepts in easy-to-understand language for general readers."

    - **Set Style**: Define the response style.
