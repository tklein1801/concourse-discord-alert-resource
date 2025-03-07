[![Go Report Card](https://goreportcard.com/badge/github.com/tklein1801/concourse-discord-alert-resource)](https://goreportcard.com/report/github.com/tklein1801/concourse-discord-alert-resource)

# concourse-discord-alert-resource

> [!NOTE]  
> This is a fork of [arbourd/concourse-slack-alert-resource](https://github.com/arbourd/concourse-slack-alert-resource) which got modified in order to send messages as notifications to Discord using webhooks

A structured and opinionated Discord notification resource for [Concourse](https://concourse-ci.org/).

<img src="./img/default.png" width="100%" style="border-radius: 5px">

The message is built by using Concourse's [resource metadata](https://concourse-ci.org/implementing-resource-types.html#resource-metadata) to show the pipeline, job, build number and a URL.

## Installing

Use this resource by adding the following to the resource_types section of a pipeline config:

```yaml
resource_types:
  - name: discord-alert
    type: registry-image
    source:
      repository: ghcr.io/tklein1801/concourse-discord-alert-resource
```

See the [Concourse docs](https://concourse-ci.org/resource-types.html) for more details on adding `resource_types` to a pipeline config.

## Source Configuration

- `url`: _Required._ Slack webhook URL.
- `concourse_url`: _Optional._ The external URL that points to Concourse. Defaults to the env variable `ATC_EXTERNAL_URL`.
- `username`: _Optional._ Concourse local user (or basic auth) username. Required for non-public pipelines if using alert type `fixed` or `broke`
- `password`: _Optional._ Concourse local user (or basic auth) password. Required for non-public pipelines if using alert type `fixed` or `broke`
- `disable`: _Optional._ Disables the resource (does not send notifications). Defaults to `false`.

## Behavior

### `check`: No operation.

### `in`: No operation.

### `out`: Send a message to Discord.

Sends a structured message to Slack based on the alert type.

#### Parameters

- `alert_type`: _Optional._ The type of alert to send to Slack. See [Alert Types](#alert-types). Defaults to `default`.
- `message`: _Optional._ The status message at the top of the alert. Defaults to name of alert type.
- `message_file`: _Optional._ File containing text which overrides `message`. If the file cannot be read, `message` will be used instead.
- `text`: _Optional._ Additional text below the message of the alert. Defaults to an empty string.
- `text_file`: _Optional._ File containing text which overrides `text`. If the file cannot be read, `text` will be used instead.
- `color`: _Optional._ The color of the notification bar as a hexadecimal. Defaults to the icon color of the alert type.
- `disable`: _Optional._ Disables the alert. Defaults to `false`.

#### Alert Types

- `default`

  <img src="./img/default.png" width="100%" style="border-radius: 5px">

- `success`

  <!-- <img src="./img/success.png" width="50%"> -->

- `failed`

  <!-- <img src="./img/failed.png" width="50%"> -->

- `started`

  <!-- <img src="./img/started.png" width="50%"> -->

- `aborted`

  <!-- <img src="./img/aborted.png" width="50%"> -->

- `errored`

  <!-- <img src="./img/errored.png" width="50%"> -->

- `fixed`

  Fixed is a special alert type that only alerts if the previous build did not succeed. Fixed requires `username` and `password` to be set for the resource if the pipeline is not public.

  <!-- <img src="./img/fixed.png" width="50%"> -->

- `broke`

  Broke is a special alert type that only alerts if the previous build succeed. Broke requires `username` and `password` to be set for the resource if the pipeline is not public.

  <!-- <img src="./img/broke.png" width="50%"> -->

## Examples

### Out

Using the default alert type with custom message and color:

```yaml
resources:
  - name: notify
    type: discord-alert
    source:
      url: https://discord.com/api/webhooks/********/****

jobs:
  # ...
  plan:
    - put: notify
      params:
        message: Completed
        color: '#eeeeee'
```

Using built-in alert types with appropriate build hooks:

```yaml
resources:
  - name: notify
    type: discord-alert
    source:
      url: https://discord.com/api/webhooks/********/****

jobs:
  # ...
  plan:
    - put: notify
      params:
        alert_type: started
    - put: some-other-task
      on_success:
        put: notify
        params:
          alert_type: success
      on_failure:
        put: notify
        params:
          alert_type: failed
      on_abort:
        put: notify
        params:
          alert_type: aborted
      on_error:
        put: notify
        params:
          alert_type: errored
```

Using the `fixed` alert type:

```yaml
resources:
  - name: notify
    type: discord-alert
    source:
      url: https://discord.com/api/webhooks/********/****
      # `alert_type: fixed` requires Concourse credentials if pipeline is private
      username: concourse
      password: concourse

jobs:
  # ...
  plan:
    - put: some-other-task
      on_success:
        put: notify
        params:
          # will only alert if build was successful and fixed
          alert_type: fixed
```
