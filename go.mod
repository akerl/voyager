module github.com/akerl/voyager/v2

go 1.13

require (
	github.com/99designs/keyring v1.1.4
	github.com/BurntSushi/locker v0.0.0-20171006230638-a6e239ea1c69
	github.com/akerl/input v0.0.4
	github.com/akerl/speculate/v2 v2.2.0
	github.com/akerl/timber/v2 v2.0.1
	github.com/aws/aws-sdk-go v1.29.5
	github.com/spf13/cobra v0.0.5
	github.com/vbauerster/mpb/v4 v4.12.1
	github.com/yawn/ykoath v1.0.3
)

replace github.com/99designs/keyring => github.com/akerl/keyring v0.0.0-20200219084108-1f409e548abc
