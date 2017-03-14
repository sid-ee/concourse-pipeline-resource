# Concourse Pipeline Resource

Get and set concourse pipelines from concourse.

## Installing

This resource is only compatible with Concourse versions 0.74.0 and higher
as no BOSH release is provided. Use this resource by adding the following to
the `resource_types` section of a pipeline config:

```yaml
---
resource_types:
- name: concourse-pipeline
  type: docker-image
  source:
    repository: robdimsdale/concourse-pipeline-resource
    tag: latest-final
```

See [concourse docs](http://concourse.ci/configuring-resource-types.html) for more details
on adding `resource_types` to a pipeline config.

**Using `tag: latest-final` will pull the latest final release, which can be
found on the
[releases page](https://github.com/robdimsdale/concourse-pipeline-resource/releases)**

**To avoid automatically upgrading, use a fixed tag instead e.g. `tag: v0.6.3`**

The docker image is `robdimsdale/concourse-pipeline-resource`; the images are
available on
[dockerhub](https://hub.docker.com/r/robdimsdale/concourse-pipeline-resource).

The rootfs of the docker image is available with each release on the
[releases page](https://github.com/robdimsdale/concourse-pipeline-resource/releases).

The docker image is semantically versioned; these versions correspond to the git tags
in this repository.

## Source configuration

Check returns the versions of all pipelines. Configure as follows:

```yaml
---
resources:
- name: my-pipelines
  type: concourse-pipeline
  source:
    target: https://my-concourse.com
    insecure: "false"
    teams:
    - name: team-1
      username: some-user
      password: some-password
    - name: team-2
      username: other-user
      password: other-password
```

* `target`: *Optional.* URL of your concourse instance e.g. `https://my-concourse.com`.
  If not specified, the resource defaults to the `ATC_EXTERNAL_URL` environment variable,
  meaning it will always target the same concourse that created the container.

* `insecure`: *Optional.* Connect to Concourse insecurely - i.e. skip SSL validation.
  Must be a [boolean-parseable string](https://golang.org/pkg/strconv/#ParseBool).
  Defaults to "false" if not provided.

* `teams`: *Required.* At least one team must be provided, with the following parameters:

  * `name`: *Required.* Name of team.
    Equivalent of `-n team-name` in `fly login` command.

  * `username`: *Required.*  Basic auth username for logging in to the team.
    Basic Auth must be enabled for the team.

  * `password`: *Required.*  Basic auth password for logging in to the team.
    Basic Auth must be enabled for the team.

## `in`: Get the configuration of the pipelines

Get the config for each pipeline; write it to the local working directory (e.g.
`/tmp/build/get`) with the filename derived from the pipeline name and team name.

For example, if there are two pipelines `foo` and `bar` belonging to `team-1`
and `team-2` respectively, the config for the first will be written to
`team-1-foo.yml` and the second to `team-2-bar.yml`.

```yaml
---
resources:
- name: my-pipelines
  type: concourse-pipeline
  source: ...

jobs:
- name: download-my-pipelines
  plan:
  - get: my-pipelines
```

## `out`: Set the configuration of the pipelines

Set the configuration for each pipeline provided in the `params` section.

Configuration can be either static or dynamic.
Static configuration has the configuration fixed in the pipeline config file,
whereas dynamic configuration reads the pipeline configuration from the provided file.

One of either static or dynamic configuration must be provided; using both is not allowed.

### static

```yaml
---
resources:
- name: my-pipelines
  type: concourse-pipeline
  source:
    teams:
    - name: team-1

jobs:
- name: set-my-pipelines
  plan:
  - put: my-pipelines
    params:
      pipelines:
      - name: my-pipeline
        team: team-1
        config_file: path/to/config/file
        vars_files:
        - path/to/optional/vars/file/1
        - path/to/optional/vars/file/2
```

* `pipelines`: *Required.* Array of pipelines to configure.
Must be non-nil and non-empty. The structure of the `pipeline` object is as follows:

 - `name`: *Required.* Name of pipeline to be configured.
 Equivalent of `-p my-pipeline-name` in `fly set-pipeline` command.

 - `team`: *Required.* Name of the team to which the pipeline belongs.
 Equivalent of `-n my-team` in `fly login` command.
 Must match one of the `teams` provided in `source`.

 - `config_file`: *Required.* Location of config file.
 Equivalent of `-c some-config-file.yml` in `fly set-pipeline` command.

 - `vars_files`: *Optional.* Array of strings corresponding to files
 containing variables to be interpolated via `{{ }}` in `config_file`.
 Equivalent of `-l some-vars-file.yml` in `fly set-pipeline` command.

### dynamic

Resource configuration as above for Check, with the following job configuration:

```yaml
---
jobs:
- name: set-my-pipelines
  plan:
  - put: my-pipelines
    params:
      pipelines_file: path/to/pipelines/file
```

* `pipelines_file`: *Required.* Path to dynamic configuration file.
  The contents of this file should have the same structure as the
  static configuration above, but in a file.

## Developing

### Prerequisites

A valid install of golang >= 1.6 is required.

### Dependencies

Dependencies are vendored in the `vendor` directory, according to the
[golang 1.5 vendor experiment](https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&cad=rja&uact=8&ved=0ahUKEwi7puWg7ZrLAhUN1WMKHeT4A7oQFggdMAA&url=https%3A%2F%2Fgolang.org%2Fs%2Fgo15vendor&usg=AFQjCNEPCAjj1lnni5apHdA7rW0crWs7Zw).

#### Updating dependencies

Install [gvt](https://github.com/FiloSottile/gvt) and make sure it is available
in your $PATH, e.g.:

```
go get -u github.com/FiloSottile/gvt
```

To add a new dependency:
```
gvt fetch
```

To update an existing dependency to a specific version:

```
gvt delete <import_path>
gvt fetch -revision <revision_number> <import_path>
```

### Running the tests

Install the ginkgo executable with:

```
go get -u github.com/onsi/ginkgo/ginkgo
```

The tests require a concourse API server to test against, and a valid
basic auth username/password for that concourse deployment.

Run the tests with the following command:

```
TARGET=https://my-concourse.com \
USERNAME=my-basic-auth-user \
PASSWORD=my-basic-auth-password \
./bin/test
```

### Contributing

Please make all pull requests to the `develop` branch, and
[ensure the tests pass locally](https://github.com/robdimsdale/concourse-pipeline-resource#running-the-tests).

### Project management

The CI for this project can be found at https://concourse.robdimsdale.com/pipelines/concourse-pipeline-resource
and the scripts can be found in the
[robdimsdale-ci repository](https://github.com/robdimsdale/robdimsdale-ci).

The roadmap is captured in [Pivotal Tracker](https://www.pivotaltracker.com/projects/1549921).
