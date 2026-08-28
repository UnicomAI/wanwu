## Description: <br>
Guides agents in using Microsoft Playwright CLI for browser automation, including page navigation, element interaction, screenshots, PDF generation, operation recording, and tests. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[Michael-C-Matias](https://clawhub.ai/user/Michael-C-Matias) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and QA engineers use this skill to plan and execute Playwright CLI workflows for browser navigation, screenshots, PDF export, code generation, test runs, and browser management. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Authenticated browser automation can expose or misuse account state when run on sites the user is not authorized to test. <br>
Mitigation: Use the skill only on sites the operator owns or is explicitly authorized to test, and review automated actions before running them against authenticated sessions. <br>
Risk: Generated scripts, screenshots, PDFs, reports, or auth.json storage-state files may contain private data. <br>
Mitigation: Protect and delete saved login state when it is no longer needed, and do not commit generated artifacts that may contain secrets, credentials, or private page content. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/Michael-C-Matias/playwright-cli-openclaw) <br>


## Skill Output: <br>
**Output Type(s):** [shell commands, code, configuration, guidance, files] <br>
**Output Format:** [Markdown with inline bash and code blocks] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May reference generated screenshots, PDFs, reports, scripts, and browser storage-state files.] <br>

## Skill Version(s): <br>
1.0.0 (source: server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
