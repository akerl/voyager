module github.com/akerl/voyager/v2

go 1.13

// Needed until https://github.com/99designs/keyring/pull/59 is merged
replace github.com/99designs/keyring => github.com/akerl/keyring v0.0.0-20200219084108-1f409e548abc

// Needed until https://github.com/ktr0731/go-fuzzyfinder/pull/13 is merged
replace github.com/ktr0731/go-fuzzyfinder => github.com/akerl/go-fuzzyfinder v0.1.2-0.20200220111247-2e90b475f471

require (
	github.com/99designs/keyring v0.0.0-00010101000000-000000000000
	github.com/BurntSushi/locker v0.0.0-20171006230638-a6e239ea1c69
	github.com/akerl/input v0.0.7
	github.com/akerl/speculate/v2 v2.5.1
	github.com/akerl/timber/v2 v2.0.1
	github.com/aws/aws-sdk-go v1.30.7
	github.com/mdp/qrterminal/v3 v3.0.0
	github.com/micmonay/keybd_event v1.1.0 // indirect
	github.com/pquerna/otp v1.2.0
	github.com/spf13/cobra v1.0.0
	github.com/vbauerster/mpb/v4 v4.12.2
	github.com/yawn/ykoath v1.0.4
)
