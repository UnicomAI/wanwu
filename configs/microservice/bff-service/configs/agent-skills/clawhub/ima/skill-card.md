## Description: <br>
IMA personal notes API skill for searching notes, browsing notebooks, reading note content, creating notes, and appending content to existing notes. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[laineyboy](https://clawhub.ai/user/laineyboy) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
External users and agents use this skill to manage a user's IMA notes through the IMA OpenAPI, including search, notebook browsing, reading, creation, and appending workflows. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill can read and change private IMA notes. <br>
Mitigation: Install only when the agent should access those notes, keep API keys private, and prefer revocable credentials. <br>
Risk: Broad trigger guidance may cause unintended note reads or writes. <br>
Mitigation: Ask the agent to confirm the target note or notebook and exact content before creating, appending, or revealing note content. <br>
Risk: Using the skill in shared chats could expose sensitive note content. <br>
Mitigation: Avoid shared-chat use for sensitive notes and limit displays to titles or summaries unless the user explicitly authorizes full content. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/laineyboy/ima) <br>
- [IMA homepage](https://ima.qq.com) <br>
- [IMA agent interface](https://ima.qq.com/agent-interface) <br>
- [IMA notes API reference](references/api.md) <br>


## Skill Output: <br>
**Output Type(s):** [guidance, shell commands, configuration, API calls, markdown, text] <br>
**Output Format:** [Markdown guidance with shell command examples and JSON API request bodies] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Uses IMA_OPENAPI_CLIENTID and IMA_OPENAPI_APIKEY credentials to call the IMA notes API.] <br>

## Skill Version(s): <br>
1.0.0 (source: server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
