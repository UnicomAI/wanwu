## Description: <br>
Tencent Meeting helps agents schedule, update, cancel, inspect, and troubleshoot Tencent Meeting sessions, participants, recordings, transcripts, smart minutes, and recording permission requests. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[wemeeting](https://clawhub.ai/user/wemeeting) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Employees and external collaborators use this skill to manage Tencent Meeting workflows, including creating or changing meetings, reviewing attendees, retrieving recordings or transcripts, and applying for recording access when permission is missing. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The Tencent Meeting token enables meeting management and access to recordings or transcripts through Tencent's MCP service. <br>
Mitigation: Use a token appropriate for the intended account and install the skill only where that level of meeting and recording access is acceptable. <br>
Risk: Meeting changes, cancellations, feedback reports, and recording permission requests can affect real meetings or expose sensitive meeting context. <br>
Mitigation: Review the skill's confirmation prompts before these actions, and require explicit user approval for sensitive operations. <br>
Risk: Feedback or troubleshooting flows may involve privacy-sensitive meeting details. <br>
Mitigation: Redact personal identifiers, meeting topics, links, participant details, and other private content before reporting or sharing feedback. <br>
Risk: Tool calls may include client OS, agent, and model metadata. <br>
Mitigation: Deploy only in environments where sending this client metadata with Tencent Meeting MCP calls is acceptable. <br>
Risk: Local update preferences can change when or whether users see version update prompts. <br>
Mitigation: Follow the documented update preference flow and respect confirmation prompts before optional updates unless the user has preauthorized automatic updates. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/wemeeting/tencent-meeting-skill) <br>
- [Tencent Meeting homepage](https://meeting.tencent.com/) <br>
- [Tencent Meeting MCP token endpoint](https://mcp.meeting.tencent.com/mcp/wemeet-open/v1) <br>
- [API references](references/api_references.md) <br>
- [Error dictionary](references/error_dictionary.md) <br>
- [Feedback rules](references/feedback_rules.md) <br>
- [Privacy policy](references/privacy_policy.md) <br>
- [Version management](references/version_management.md) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, shell commands, configuration, guidance] <br>
**Output Format:** [Markdown and structured tool-call guidance with shell command examples] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Uses Tencent Meeting token configuration and may include trace identifiers returned by Tencent Meeting MCP calls.] <br>

## Skill Version(s): <br>
1.0.11 (source: release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
