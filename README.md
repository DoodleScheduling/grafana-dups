# Unpack manifests from crossplane resources
[![release](https://github.com/doodlescheduling/grafana-dups/actions/workflows/release.yaml/badge.svg)](https://github.com/doodlescheduling/grafana-dups/actions/workflows/release.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/doodlescheduling/grafana-dups)](https://goreportcard.com/report/github.com/doodlescheduling/grafana-dups)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/DoodleScheduling/grafana-dups/badge)](https://api.securityscorecards.dev/projects/github.com/DoodleScheduling/grafana-dups)
[![Coverage Status](https://coveralls.io/repos/github/DoodleScheduling/grafana-dups/badge.svg?branch=master)](https://coveralls.io/github/DoodleScheduling/grafana-dups?branch=master)

This small utility which looks for duplicate grafana dashboards provisioned via kubernetes ConfigMaps.
Its checks for:
* Any duplicate dashboard uid's
* Any duplicate dashboard titles
* Any duplicate title/folder combination if --folder-annotation is set

## Installation

### Brew
```
brew tap doodlescheduling/grafana-dups
brew install grafana-dups
```

### Docker
```
docker pull ghcr.io/doodlescheduling/grafana-dups:v0
```

## Arguments

| Flag           | Short        | Env            | Default      | Description   |
| ------------- | ------------- | ------------- | ------------- | ------------- |
| `--file`  | `-f`  | `FILE` | `/dev/stdin` | Path to input |
| `--allow-failure`  | ``  | `ALLOW_FAILURE` | `false` | Do not exit > 0 if an error occurred |
| `--folder-annotation`  | `-a`  | `FOLDER_ANNOTATION` | `` | Name of the folder annotation key |
| `--label-selector`  | `-l`  | `LABEL_SELECTOR` | `` | Filter resources by labels |