## Description: <br>
Headless browser automation CLI optimized for AI agents with accessibility tree snapshots and ref-based element selection. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[MaTriXy](https://clawhub.ai/user/MaTriXy) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and AI agents use this skill to automate browser workflows, inspect accessibility snapshots, select elements deterministically by reference, and extract page data through the agent-browser CLI. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Saved browser state and cookies can represent sensitive account access. <br>
Mitigation: Keep saved authentication files out of source control and logs, and use isolated sessions or test accounts for automation. <br>
Risk: The skill depends on the external agent-browser package source. <br>
Mitigation: Install only when the package source is trusted and review commands before running them on logged-in websites. <br>


## Reference(s): <br>
- [Agent Browser upstream repository](https://github.com/vercel-labs/agent-browser) <br>
- [ClawHub skill page](https://clawhub.ai/MaTriXy/agent-browser-clawdbot) <br>
- [Publisher profile](https://clawhub.ai/user/MaTriXy) <br>


## Skill Output: <br>
**Output Type(s):** [Text, Markdown, Shell commands, Configuration, Guidance] <br>
**Output Format:** [Markdown with inline shell commands and JSON examples] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May guide browser state persistence, session isolation, navigation, interaction, data extraction, screenshots, PDFs, network control, cookies, and storage workflows.] <br>

## Skill Version(s): <br>
0.1.0 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
