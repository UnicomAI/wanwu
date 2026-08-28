## Description: <br>
Automate web browser interactions using natural language via CLI commands. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[peytoncasper](https://clawhub.ai/user/peytoncasper) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and operators use this skill to let an agent navigate websites, interact with page elements, extract structured page data, capture screenshots, and manage browser sessions through CLI commands. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill gives the agent broad browser control. <br>
Mitigation: Install and use it only when broad browser automation is acceptable, and review browser state or screenshots before continuing sensitive workflows. <br>
Risk: Remote Browserbase mode can be selected automatically when Browserbase API keys are configured. <br>
Mitigation: Confirm whether remote mode is active before using sensitive sites or entering credentials. <br>
Risk: Logged-in browser sessions can persist in the local .chrome-profile directory. <br>
Mitigation: Clear or isolate .chrome-profile between tasks, especially after authenticated browsing. <br>
Risk: Screenshots can capture sensitive page content. <br>
Mitigation: Review screenshots before storing, sharing, or sending them to another system. <br>
Risk: Downloaded files are saved automatically. <br>
Mitigation: Treat files in ./agent/downloads as untrusted until they are verified. <br>


## Reference(s): <br>
- [Browser Automation CLI Reference](artifact/REFERENCE.md) <br>
- [Browser Automation Examples](artifact/EXAMPLES.md) <br>
- [ClawHub Skill Page](https://clawhub.ai/peytoncasper/browser-pc) <br>
- [Publisher Profile](https://clawhub.ai/user/peytoncasper) <br>


## Skill Output: <br>
**Output Type(s):** [text, shell commands, JSON, files, configuration, guidance] <br>
**Output Format:** [CLI commands with JSON responses, PNG screenshots, downloaded files, and setup guidance.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Screenshots are saved under ./agent/browser_screenshots; downloads are saved under ./agent/downloads; local browser sessions may persist in .chrome-profile.] <br>

## Skill Version(s): <br>
1.0.0 (source: ClawHub release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
