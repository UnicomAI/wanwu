## Description: <br>
Reads and displays the contents of local text, configuration, and code files when the user asks to view or cat a file. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[jwu26](https://clawhub.ai/user/jwu26) <br>

### License/Terms of Use: <br>
Apache 2.0 <br>


## Use Case: <br>
Developers and agent users use this skill to inspect local text, configuration, and source-code files that they intentionally want the agent to read. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill can reveal sensitive local file contents if pointed at private keys, tokens, browser or session files, or other confidential paths. <br>
Mitigation: Use it only for files the user intentionally wants the agent to inspect, and avoid sensitive paths unless disclosure is intended. <br>


## Reference(s): <br>
- [ClawHub release page](https://clawhub.ai/jwu26/file-reader) <br>


## Skill Output: <br>
**Output Type(s):** [text, shell commands, guidance] <br>
**Output Format:** [Plain text file contents with command-line usage guidance] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Reads one user-supplied file path at a time.] <br>

## Skill Version(s): <br>
1.0.0 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
