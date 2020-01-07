# cruisemic

A tool to capture and parse a research ship's underway data feed.
Separate data feeds will be written in the format specified at [https://github.com/armbrustlab/tsdataformat](https://github.com/armbrustlab/tsdataformat).

## Install

From a cloned copy of this repo,
run `go install ./...` to install the binary `cruisemic` at `$GOPATH/bin`.

## CLI help

Run `cruisemic -h`

## Install as a launchd service

Copy `service-files/launchd/local.cruisemic.plist` to `~/Library/LaunchAgents`.
Modify appropriate values.

Launch the service with `launchctl load -w ~/Library/LaunchAgents/local.cruisemic.plist`.

Stop the service with `launchctl unload -w ~/Library/LaunchAgents/local.cruisemic.plist`.
