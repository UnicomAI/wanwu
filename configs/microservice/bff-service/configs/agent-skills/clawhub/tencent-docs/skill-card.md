## Description: <br>
腾讯文档 TENCENT DOCS helps agents create, edit, search, import, export, and manage Tencent Docs documents, sheets, slides, smart canvases, diagrams, spaces, and related file workflows. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[liyang58](https://clawhub.ai/user/liyang58) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and end users use this skill through an agent to operate Tencent Docs cloud documents: create and edit content, manage files and spaces, read or search documents, upload local files, and clip web pages. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill requires Tencent Docs OAuth access and can use sensitive credentials. <br>
Mitigation: Install only with a Tencent Docs account and token scope you are comfortable using for agent-driven document operations. <br>
Risk: The skill can upload local files, export documents, delete documents or space nodes, and change sharing permissions. <br>
Mitigation: Require explicit user confirmation before uploads, exports, deletions, recursive space operations, or public read/edit link changes. <br>
Risk: Unsupported requests may be reported to Tencent without a separate notice. <br>
Mitigation: Avoid entering sensitive prompts or document details unless you accept this reporting behavior. <br>
Risk: Web clipping and image workflows can download and upload web content into Tencent Docs. <br>
Mitigation: Confirm source URLs, content rights, and sensitivity before web clipping, image upload, or web image search workflows. <br>


## Reference(s): <br>
- [ClawHub Release Page](https://clawhub.ai/liyang58/tencent-docs) <br>
- [Tencent Docs Homepage](https://docs.qq.com/home) <br>
- [Tencent Docs OAuth Setup](https://docs.qq.com/scenario/open-claw.html?nlc=1) <br>
- [Tencent Docs MCP Endpoint](https://docs.qq.com/openapi/mcp) <br>
- [Authentication Guide](artifact/references/auth.md) <br>
- [Common Workflows](artifact/references/workflows.md) <br>
- [File Management Reference](artifact/references/manage_references.md) <br>
- [Smart Canvas Guide](artifact/smartcanvas/entry.md) <br>
- [Sheet Guide](artifact/sheet/entry.md) <br>
- [Slide Reference](artifact/references/slide_references.md) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, code, shell commands, configuration, guidance] <br>
**Output Format:** [Markdown and text guidance with JSON tool arguments, shell commands, and Tencent Docs document URLs or IDs] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Requires Tencent Docs OAuth credentials and may invoke local helper scripts for setup, slide generation, and file import workflows.] <br>

## Skill Version(s): <br>
1.0.31 (source: frontmatter and server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
