[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/agm650/sensu-handler-betteruptime)
![goreleaser](https://github.com/agm650/sensu-handler-betteruptime/workflows/goreleaser/badge.svg)
[![Go Test](https://github.com/agm650/sensu-handler-betteruptime/workflows/Go%20Test/badge.svg)](https://github.com/agm650/sensu-handler-betteruptime/actions?query=workflow%3A%22Go+Test%22)
[![goreleaser](https://github.com/agm650/sensu-handler-betteruptime/workflows/goreleaser/badge.svg)](https://github.com/agm650/sensu-handler-betteruptime/actions?query=workflow%3Agoreleaser)

# BetterUptime

## Table of Contents

- [BetterUptime](#betteruptime)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Files](#files)
  - [Usage examples](#usage-examples)
  - [Configuration](#configuration)
    - [Asset registration](#asset-registration)
    - [Handler definition](#handler-definition)
    - [Annotations](#annotations)
      - [Examples](#examples)
  - [Installation from source](#installation-from-source)
  - [Additional notes](#additional-notes)
  - [Contributing](#contributing)

## Overview

sensu-handler-betteruptime is a [Sensu Handler][6] built using the [Sensu Plugin SDK][2].
This is a way to create incident using API on better uptime.

## Files

## Usage examples

## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```bash
sensuctl asset add agm650/sensu-handler-betteruptime
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/agm650/sensu-handler-betteruptime].

### Handler definition

```yml
---
type: Handler
api_version: core/v2
metadata:
  name: sensu-handler-betteruptime
  namespace: default
spec:
  command: sensu-handler-betteruptime --token better_uptime_token
  type: pipe
  runtime_assets:
  - agm650/sensu-handler-betteruptime
```

### Annotations

All arguments for this handler are tunable on a per entity or check basis based on annotations.  The
annotations keyspace for this handler is `betteruptime/config/name`.

#### Examples

To change the example argument for a particular check, for that checks's metadata add the following:

```yml
type: CheckConfig
api_version: core/v2
metadata:
  annotations:
    betteruptime/config/name: "Name of the Event"
    betteruptime/config/summary: "Summary of the Event"
    betteruptime/config/description: "Description of the incident"
    betteruptime/config/email: "requester.email@domain.com"
[...]
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-handler-betteruptime repository:

```bash
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/handler-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/handler-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/handlers/
[7]: https://github.com/sensu-community/handler-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
