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

