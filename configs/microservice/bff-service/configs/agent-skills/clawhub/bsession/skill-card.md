## Description: <br>
Browser automation - setup the bsession environment, fetch info from a website one time, create scripted automations, or debug existing sessions from any repository. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[gaxxx](https://clawhub.ai/user/gaxxx) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and automation engineers use this skill to set up and operate Docker-backed browser automation sessions for one-time website extraction, recurring scripted checks, and debugging existing browser sessions. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill installs and operates a persistent Docker browser environment with broad local control. <br>
Mitigation: Review the exact release before setup, keep Docker access limited to trusted users, and inspect recurring automations before enabling them. <br>
Risk: Browser profiles, configuration files, and automation scripts may persist sensitive session data or credentials. <br>
Mitigation: Prefer a VNC password, avoid storing secrets in configuration files unless they are masked, and remove stale profiles or workspace data when no longer needed. <br>
Risk: The installer may write skill files for multiple agent environments. <br>
Mitigation: Review the target skill directories after installation and keep only the integrations needed for the deployment. <br>


## Reference(s): <br>


## Skill Output: <br>
**Output Type(s):** [Text, Markdown, Code, Shell commands, Configuration, Files] <br>
**Output Format:** [Markdown with inline shell commands and code snippets] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May create or modify browser automation scripts, session configuration, and local workspace files.] <br>

## Skill Version(s): <br>
0.1.0 (source: server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
