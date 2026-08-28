## Description: <br>
Automate web browser interactions using natural language via CLI commands. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[peytoncasper](https://clawhub.ai/user/peytoncasper) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and agents use this skill to navigate websites, click or type into page elements, extract structured web data, take screenshots, and automate browser workflows from CLI commands. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill combines broad browser automation with an unverified global CLI install. <br>
Mitigation: Inspect and trust the actual CLI package exposed by npm install and npm link before installation or use. <br>
Risk: Persistent browser profile data, screenshots, downloads, and remote-browser mode can expose sensitive browsing activity or credentials. <br>
Mitigation: Use a dedicated browser profile and test accounts, explicitly choose local or remote mode, avoid sensitive sites unless necessary, and clear stored profile data when finished. <br>
Risk: Automated actions can submit forms, change accounts, download files, or scrape websites without sufficient review. <br>
Mitigation: Manually confirm any form submission, account change, download, or scraping workflow before allowing the command to proceed. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/peytoncasper/browser-automation) <br>
- [Browser Automation Examples](EXAMPLES.md) <br>
- [Browser Automation CLI Reference](REFERENCE.md) <br>


## Skill Output: <br>
**Output Type(s):** [shell commands, JSON, screenshots, guidance] <br>
**Output Format:** [Markdown guidance with inline shell commands and JSON command output] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Browser commands may create screenshot and download files; extracted page data can be returned as structured JSON.] <br>

## Skill Version(s): <br>
1.0.1 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
