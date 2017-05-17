voyager
=========

[![Build Status](https://img.shields.io/circleci/project/akerl/voyager/master.svg)](https://circleci.com/gh/akerl/voyager)
[![GitHub release](https://img.shields.io/github/release/akerl/voyager.svg)](https://github.com/akerl/voyager/releases)
[![MIT Licensed](https://img.shields.io/badge/license-MIT-green.svg)](https://tldrlegal.com/license/mit-license)

Helper for assuming roles on AWS accounts

## Usage

```
* account: 1234567890
  roles:
    readonly:
      mfa: false
    admin:
      mfa: true
  region: us-east-1
  source: 0987654321/admin
  tags:
    group: cool-servers
    owners: my-team
```

## Installation

## License

voyager is released under the MIT License. See the bundled LICENSE file for details.
