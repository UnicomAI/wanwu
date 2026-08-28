## Description: <br>
ima skills helps agents manage IMA notes and knowledge bases through IMA OpenAPI workflows, including note search and writes, file or URL imports, and knowledge-base search. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[iampennyli](https://clawhub.ai/user/iampennyli) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
External users and developers who use IMA use this skill to let an agent search, read, create, append, upload, and organize notes and knowledge-base content after they provide IMA OpenAPI credentials. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill can access and modify IMA notes and knowledge bases using user-provided credentials. <br>
Mitigation: Install only if the publisher is trusted, protect the API key, and review note creation, note appends, file uploads, URL imports, and knowledge-base writes before approving them. <br>
Risk: The security summary says the skill can use IMA credentials more broadly than the docs disclose. <br>
Mitigation: Keep IMA_BASE_URL unset or restricted to the official IMA service, and allow requests only to ima.qq.com and scoped COS endpoints returned by the IMA API. <br>
Risk: Incorrect write or upload handling can change user content, upload unintended files, or produce garbled note text. <br>
Mitigation: Confirm target notes and knowledge bases before writes, validate UTF-8 before notes writes, preserve binary file content during uploads, and reject unsupported file types. <br>


## Reference(s): <br>
- [ClawHub release page](https://clawhub.ai/iampennyli/ima-skills) <br>
- [IMA homepage](https://ima.qq.com) <br>
- [IMA OpenAPI credential setup](https://ima.qq.com/agent-interface) <br>
- [Knowledge Base API reference](knowledge-base/references/api.md) <br>
- [Notes API reference](notes/references/api.md) <br>
- [Tencent COS signature documentation](https://cloud.tencent.com/document/product/436/7778) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, code, shell commands, configuration, guidance] <br>
**Output Format:** [Markdown guidance with shell commands, JSON request examples, and API response handling instructions.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Requires IMA_OPENAPI_CLIENTID and IMA_OPENAPI_APIKEY or matching local credential files; operations may read, create, append, upload, and import user content through IMA APIs.] <br>

## Skill Version(s): <br>
1.1.7 (source: server release metadata and artifact meta.json) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
