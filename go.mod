module github.com/akerl/voyager/v2

go 1.13

// Needed until https://github.com/99designs/keyring/pull/59 is merged
replace github.com/99designs/keyring => github.com/akerl/keyring v0.0.0-20200219084108-1f409e548abc

// Needed until https://github.com/ktr0731/go-fuzzyfinder/pull/13 is merged
replace github.com/ktr0731/go-fuzzyfinder => github.com/akerl/go-fuzzyfinder v0.1.2-0.20200220111247-2e90b475f471

require (
	github.com/99designs/keyring v1.1.4
	github.com/BurntSushi/locker v0.0.0-20171006230638-a6e239ea1c69
	github.com/akerl/input v0.0.6
	github.com/akerl/speculate/v2 v2.3.2
	github.com/akerl/timber/v2 v2.0.1
	github.com/aws/aws-sdk-go v1.29.15
	github.com/keybase/go-keychain v0.0.0-20200218013740-86d4642e4ce2 // indirect
	github.com/ktr0731/go-fuzzyfinder v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.0.3 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mdp/qrterminal/v3 v3.0.0
	github.com/pquerna/otp v1.2.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vbauerster/mpb/v4 v4.12.1
	github.com/yawn/ykoath v1.0.4
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073 // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/text v0.3.2 // indirect
)
