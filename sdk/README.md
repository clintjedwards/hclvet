# Creating new rules and rulesets

HCL vet uses a logical grouping of rules called a _ruleset_ to manage what the linter checks for.

An example ruleset can be found [here](https://github.com/clintjedwards/hclvet-ruleset-example).

A ruleset is just a simple directory and can be either local or hosted on most remote
filestores/repositories. This is extremely powerful as it means you can host your own rulesets or
download other's.

## How to create a ruleset

### 1) Use the create command

To create your own custom ruleset, create the directory the ruleset will exist within and run:

`$ hclvet ruleset create <name>`

This will create the required folders and files needed.

### 2) Customize your ruleset

Ruleset settings and information is kept in the `ruleset.hcl` file in the root of the directory.

Within it you will find two attributes: `version` and `name`.

- _Version_ should be changed in the same fashion as most semver applications. Anytime you change, add, or remove
  a rule bump the version to convey that there is a newer version.
- _Name_ is the 20 character maximum, alphanumeric name for your ruleset and should not be changed once set.

## How to create a rule

### 1) Creating a new rule

Rules are written in Golang and kept in the `rules` folder found in the root of a ruleset directory.

Within this folder, each rule is just a miniature golang program and kept in a folder on its own.
The directory name of the rule determines which ID it gets mapped to within HCLvet.

You can run the `hclvet rule create <rule_name>` command to create a new rule from the root of the ruleset directory.

### 2) Customizing your new rule

Created rules are just simple, testable golang programs with only a few pieces that you need to
finish implementing. These parts are highlighted and explained with comments in the
main.go file within the rule directory upon using the generator to create a new rule.

#### **The Check function**

The check function is where the logic for the lint rule is stored. It receives the file to be linted
as an argument and then returns a list of linting errors pertaining to that file.

The implementation of the linting logic should be simple as the sdk offers hcl file parsers that return an easy to walk list of all blocks and attributes within the given file.

#### **The Main function**

The main function simply contains details about the linting rule and registers the rule with the
`NewRule` function located in the SDK.
