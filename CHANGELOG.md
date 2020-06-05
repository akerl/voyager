# 2.10.4 / 2020-06-05

* [ENHANCEMENT] Update deps
* [BUGFIX] Fix handling for empty wmenu selection

# 2.10.3 / 2020-05-08

* [ENHANCEMENT] Update deps

# 2.10.2 / 2020-05-07

* [ENHANCEMENT] Update deps

# 2.10.1 / 2020-05-07

* [ENHANCEMENT] Update deps

# 2.10.0 / 2020-05-06

* [ENHANCEMENT] Add ResolveOptions for Grapher, to enable overriding region setting

# 2.9.0 / 2020-04-15

* [BUGFIX] Add quirks fix for fuzzyfinder terminal bug on Macs
* [ENHANCEMENT] Update deps

# 2.8.0 / 2020-03-23

* [FEATURE] Use STS regional endpoints for API requests

# 2.7.1 / 2020-03-05

* [ENHANCEMENT] Prompt for profile before confirmation when rotating

# 2.7.0 / 2020-03-05

* [FEATURE] Implement profile rotation

# 2.6.1 / 2020-03-02

* [ENHANCEMENT] Improve text for new rotated AWS key

# 2.6.0 / 2020-03-02

* [FEATURE] Add profile management commands

# 2.5.0 / 2020-02-20

* [BUGFIX] Correct handling of quotes in xargs commands
* [ENHANCEMENT] Update deps

# 2.4.1 / 2020-02-20

* [BUGFIX] Fix go.sum file

# 2.4.0 / 2020-02-20

* [ENHANCEMENT] Show prompt message for fuzzy prompt

# 2.3.1 / 2020-02-19

* [ENHANCEMENT] Update deps

# 2.3.0 / 2020-02-19

* [ENHANCEMENT] Update deps

# 2.2.0 / 2020-02-19

* [FEATURE] Add user-agent support
* [BUGFIX] Update to patched keychain library which resolves Catalina deprecation warning
* [ENHANCEMENT] Update deps

# 2.1.6 / 2020-02-18

* [ENHANCEMENT] Update deps
* [ENHANCEMENT] Switch to GitHub Actions for CI

# 2.1.5 / 2020-01-08

* [ENHANCEMENT] Update deps

# 2.1.4 / 2019-08-20

* [BUGFIX] Update deps

# 2.1.3 / 2019-08-20

* [BUGFIX] Additional version setting fix

# 2.1.2 / 2019-08-20

* [BUGFIX] Fix version setting

# 2.1.1 / 2019-08-19

* [FEATURE] Show confirmation prompt before running xargs

# 2.1.0 / 2019-08-19

* [FEATURE] Add xargs subcommand
* [FEATURE] Mutex for thread-safe credential lookups

# 2.0.9 / 2019-08-15

* [FEATURE] Update speculate to handle platform-specific env vars

# 2.0.8 / 2019-08-13

* [BUGFIX] Trim all whitespace from profiles/prompt, including on Windows

# 2.0.7 / 2019-08-12

* [BUGFIX] Bring in speculate upgrade to fix MultiMfaPrompt chaining bug

# 2.0.6 / 2019-08-12

* [ENHANCEMENT] Support speculate MultiMfaPrompt

# 2.0.5 / 2019-08-12

* [BUGFIX] Update speculate dep to fix empty parameter bug

# 2.0.4 / 2019-08-12

* [BUGFIX] Update timber dependency to fix logging bug

# 2.0.3 / 2019-08-12

* [BUGFIX] Update input dependency to fix prompt formatting bug

# 2.0.2 / 2019-08-11

* [ENHANCEMENT] Update go.mod to remove cruft

# 2.0.1 / 2019-08-11

* [BUGFIX] Fix SessionName typo

# 2.0.0 / 2019-08-11

* [BREAKING] Updated to speculate v2
* [BREAKING] Uses https://github.com/akerl/input rather than built-in prompts
* [BREAKING] Travel/Itinerary/Hop rebuild as Grapher/Path/Hop. Grapher.Resolve() resolves the path to the target role/account, Path.Traverse() executes the necessary role assumptions. This enables the addition of hop caching
* [FEATURE] Caching is supported by Path objects, allowing credentials to be saved and reused for intermediate hops
* [FEATURE] Fail with meaningful error if schema format has changed
* [ENHANCEMENT] Improved debug/info log messages

