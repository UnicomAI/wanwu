## Description: <br>
Creates, reads, edits, validates, converts, and annotates Word .docx documents, including formatted reports, tables of contents, images, comments, and tracked changes. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[zda2019](https://clawhub.ai/user/zda2019) <br>

### License/Terms of Use: <br>
Proprietary <br>


## Use Case: <br>
Developers, document authors, and agents use this skill to create polished Word documents or inspect and modify existing .docx files while preserving document structure, formatting, comments, and tracked changes. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: A helper compiles and injects native code from predictable temporary paths. <br>
Mitigation: Review the helper before installation and avoid or harden the LibreOffice conversion and accept-changes flows until the predictable temporary-path behavior is fixed. <br>
Risk: Shell-based Office automation can modify important documents or produce changes that are not obvious on first inspection. <br>
Mitigation: Work on copies of important documents and review generated or edited files before sharing. <br>
Risk: Tracked-change and comment author metadata may misrepresent provenance if the default author is left unchanged. <br>
Mitigation: Override the tracked-change or comment author when document provenance matters. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/zda2019/docx-2) <br>
- [Artifact license terms](artifact/LICENSE.txt) <br>


## Skill Output: <br>
**Output Type(s):** [Text, Markdown, Code, Shell commands, Files, Guidance] <br>
**Output Format:** [Markdown guidance with code and shell-command examples; generated or edited .docx files when applied.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May use local Office tooling and document-validation helpers; review generated documents before sharing.] <br>

## Skill Version(s): <br>
0.1.0 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
