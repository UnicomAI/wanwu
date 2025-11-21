# Workflow

## Node Overview

- **Core Features**: It allows you to embed another workflow (subprocess) within a workflow (main process), implementing the powerful feature of "workflow calling workflow".

## Configuration Guidelines

#### 1. Input and Output: Data Transfer

- **Structured**: The input and output structure of the Workflow node is completely determined by the **sub-workflow** it calls, and you cannot modify it in the parent workflow.

- **Configure Input**:

  You need to define **required input parameters** for the sub-workflow and specify data sources in the parent Workflow node.
  
  Data sources support two methods:

  - **Fixed Value**: Directly input a static value, such as `Hello World` or `2024`.
  
  - **Reference Variable**: Reference upstream other node output results to implement dynamic data transfer.

#### 2. Batch Processing Mode: From "Single Processing" to "Batch Production"

By default, the Workflow node only executes once.

However, after enabling **Batch Processing Mode**, it can operate like a production line, processing each item in the input list you provide according to the sub-workflow logic.
