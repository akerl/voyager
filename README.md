voyager
=========

[![Build Status](https://img.shields.io/travis/com/akerl/voyager.svg)](https://travis-ci.com/akerl/voyager)
[![GitHub release](https://img.shields.io/github/release/akerl/voyager.svg)](https://github.com/akerl/voyager/releases)
[![MIT Licensed](https://img.shields.io/badge/license-MIT-green.svg)](https://tldrlegal.com/license/mit-license)

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
