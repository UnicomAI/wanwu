## Description: <br>
This skill helps agents create, edit, analyze, visualize, and recalculate spreadsheet files, including XLSX, XLSM, CSV, and TSV files. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[wu-uk](https://clawhub.ai/user/wu-uk) <br>

### License/Terms of Use: <br>
Proprietary <br>


## Use Case: <br>
Employees, external users, developers, and analysts use this skill to build spreadsheet models, preserve or update workbook formatting and formulas, analyze tabular data, and recalculate formulas through LibreOffice. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill can modify spreadsheet workbooks during editing and formula recalculation. <br>
Mitigation: Use the skill on trusted workbooks, keep backups before edits or recalculation, and review resulting files before relying on them. <br>
Risk: The recalculation workflow can add a persistent LibreOffice macro to the user's LibreOffice profile. <br>
Mitigation: Run the skill in a sandbox or disposable LibreOffice profile when possible, and only install it if persistent local LibreOffice configuration changes are acceptable. <br>


## Reference(s): <br>
- [ClawHub release page](https://clawhub.ai/wu-uk/financial-modeling-qa-xlsx) <br>


## Skill Output: <br>
**Output Type(s):** [Text, Markdown, Code, Shell commands, Configuration, Guidance, Files] <br>
**Output Format:** [Markdown guidance with Python code snippets, shell commands, JSON recalculation results, and generated or modified spreadsheet files] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May modify workbook files and may configure a persistent LibreOffice macro in the user's LibreOffice profile.] <br>

## Skill Version(s): <br>
0.1.0 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
