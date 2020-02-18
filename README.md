voyager
=========

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/akerl/voyager/Build)](https://github.com/akerl/voyager/actions))
[![GitHub release](https://img.shields.io/github/release/akerl/voyager.svg)](https://github.com/akerl/voyager/releases)
[![License](https://img.shields.io/github/license/akerl/voyager)](https://github.com/akerl/voyager/blob/master/LICENSE)

Helper for assuming roles on AWS accounts

## Usage

```
{
  "account": "1234567890",
  "region": "us-east-1",
  "source": "auth",
  "roles": {
    "admin": {
      "mfa": true
    },
    "readonly": {
      "mfa": false
    }
  },
  "tags": {
    "owner": "sadteam",
    "env": "prod",
    "group": "databases"
  }
},
```

## Installation

## License

voyager is released under the MIT License. See the bundled LICENSE file for details.
