# TODO

- Versioning
  - We should allow users to lock their version of particular rulesets. We should be able to do this
  - with a symbol before the version, or a simple attribute that says "locked".
  - Keep this simple no need to reimplement full versioning system.
- Rule Remediation
  - Allow remediation to have more than one line
- Clean up and add more documentation. A video or text tutorial on how to write rules would be best UX as it
  stands its kinda hard to understand.
- Add ability to recursively grab files when specifying lint paths.
- Language server (gives this the ability to embed this into an IDE free of charge).
- Add nocolor option
- Add concurrency to linting, we should be able to run many rules at the same time for a single file. Allow the user to set this.
- Think about allowing a pager view of the humanized output
- Take input from stdin?
  - What was the use case here?
- Allow users to add comments to hcl files such that they can suppress some rules/rulesets.
- Can we check terminal size before hand and avoid running the spinner for insufficently small terminals?
  (This causes the spinner to render poorly)
- Formatter's printerror should take an error and expand it into a string, so that we can pass around errors not strings.
- Add timeouts for potentially long actions like downloads
- When we update the ruleset we should check that all rules are actually installed properly and if not just go ahead and recompile the ones that aren't listed
- Create a delete command for troubleshooting
