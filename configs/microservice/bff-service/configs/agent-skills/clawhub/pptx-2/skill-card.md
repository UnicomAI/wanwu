## Description: <br>
Create, edit, and analyze .pptx presentation files, including slide content, layouts, comments, speaker notes, and theme details. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[liuyingduo](https://clawhub.ai/user/liuyingduo) <br>

### License/Terms of Use: <br>
Proprietary <br>


## Use Case: <br>
Agents supporting presentation work use this skill to inspect, create, edit, validate, and visually QA PowerPoint decks. It is intended for workflows involving .pptx files, templates, slide XML, speaker notes, comments, thumbnails, and generated presentation outputs. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The helper tools can edit, repack, convert, and validate local Office documents through shell commands. <br>
Mitigation: Use the skill on copies of presentations and review generated presentations, PDFs, images, thumbnails, and unpacked XML before sharing. <br>
Risk: Bundled Office helpers include behavior beyond the advertised PPTX scope, including DOCX/XLSX-capable validation and packing paths. <br>
Mitigation: Avoid using the helpers on DOCX or XLSX files unless that broader Office-document processing is explicitly intended. <br>
Risk: The LibreOffice conversion helper may apply an LD_PRELOAD shim in restricted environments. <br>
Mitigation: Install and run the skill only in environments where this local conversion authority is acceptable. <br>


## Reference(s): <br>
- [Pptx Skill Instructions](artifact/SKILL.md) <br>
- [Editing Presentations](artifact/editing.md) <br>
- [PptxGenJS Tutorial](artifact/pptxgenjs.md) <br>
- [ClawHub Skill Release](https://clawhub.ai/liuyingduo/pptx-2) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, code, shell commands, configuration, guidance, files] <br>
**Output Format:** [Markdown guidance with command snippets, code examples, and Office-document file outputs when helper scripts are run] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May produce or transform .pptx files, unpacked Office XML, thumbnails, PDFs, or slide images during validation and QA workflows.] <br>

## Skill Version(s): <br>
0.1.1 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
