# Start

## Node Overview

Responsible for defining all conditions required to start the workflow, and is the source of data inflow.

## Configuration Guidelines

- **1. Add Parameters**

  - **Manual Addition**: Set **Variable Name** and **Variable Type**.
  
  - **Batch Import (Efficient)**: If you already have a clear parameter structure, you can click the **Import JSON** icon. In the pop-up panel, paste your JSON data structure, and the system will automatically parse and create all parameters for you, greatly improving configuration efficiency.

- **2. Set Data Type**

  - Supports multiple basic types, including string (String), numbers (Integer, Number), boolean values (Boolean), time (Time), objects (Object), arrays (Array), and files (File).
  
  - Powerful `Object` type supports up to **3 levels of nesting**, which can meet the definition requirements of complex data structures (such as address information, product details).

- **3. Write Parameter Description**

  - A high-quality description allows the model to more accurately understand the parameter's purpose and expected format.
  
  - **Examples**:
    - **Poor Description**: `city`
    - **Good Description**: `The target city for which the user wants to query weather, such as: Beijing, Shanghai, New York.`

- **4. Set Required or Optional**

  - **Required**: After checking, if the user input fails to provide this parameter information, the workflow will not be triggered. This is suitable for core business logic that cannot be missing parameters (such as "city" when querying weather).
  
  - **Optional**: If the parameter is not provided, the workflow will still start, and the parameter value will be empty. This is suitable for enhancement or supplementary information.

![image-20250820175404325](image-20250820175404325.png)
