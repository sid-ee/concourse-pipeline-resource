# Concourse Pipeline Resource (DEPRECATING)

***This resource is deprecating in favor of the [set_pipeline step](https://concourse-ci.org/jobs.html#schema.step.set-pipeline-step.set_pipeline).***

***If there are limitations to the set_pipeline step that don't let you switch to it right now, please let us know in [concourse/rfcs#31](https://github.com/concourse/rfcs/pull/31)***

Get and set concourse pipelines from concourse.

## Installing

Use this resource by adding the following to
the `resource_types` section of a pipeline config:

```yaml
---
resource_types:
- name: concourse-pipeline
  type: docker-image
  source:
    repository: concourse/concourse-pipeline-resource
```

See [concourse docs](https://concourse-ci.org/resource-types.html) for more details
on adding `resource_types` to a pipeline config.

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

  * `username`: Basic auth username for logging in to the team.
    If this and `password` are blank, team must have no authentication configured.

  * `password`: Basic auth password for logging in to the team.
    If this and `username` are blank, team must have no authentication configured.

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
        vars:
          my_var: "foo"
          my_complex_var: {abc: 123}
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

 - `vars`: *Optional.* Map of keys and values corresponding to variables
 to be interpolated via `(( ))` in `config_file`. Values can arbitrary
 YAML types.
 Equivalent of `-y "foo=bar"` in `fly set-pipeline` command.

 - `unpaused`: *Optional.* Boolean specifying if the pipeline should
 be unpaused after the creation. If it is set to `true`, the command
 `unpause-pipeline` will be executed for the specific pipeline.

 - `exposed`: *Optional.* Boolean specifying if the pipeline should
 be exposed after the creation. If it is set to `true`, the command
 `expose-pipeline` will be executed for the specific pipeline.

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

* golang is *required* - version 1.13.x is tested; earlier versions may also
  work.
* docker is *required* - version 17.06.x is tested; earlier versions may also
  work.

### Dependencies

Dependencies are handled using [go modules](https://github.com/golang/go/wiki/Modules).

#### Updating dependencies

```
go mod download
```

To add or update a specific dependency version, follow the go modules instructions for [Daily Workflow](https://github.com/golang/go/wiki/Modules#daily-workflow)

### Running the tests

#### Using a local environment

The acceptance tests require a running Concourse configured with basic auth to test against.

Run the tests with the following command (optionally also setting `INSECURE=true`):

```
FLY_LOCATION=path/to/fly \
TARGET=https://my-concourse.com \
USERNAME=my-basic-auth-user \
PASSWORD=my-basic-auth-password \
./bin/test
```

#### Using a Dockerfile

**Note**: the `Dockerfile` tests do not run the acceptance tests, but ensure a consistent environment across any `docker` enabled platform. When the docker
image builds, the tests run inside the docker container, and on failure they
will stop the build.

The tests need to be ran from one directory up from the directory of the repo. They will also need the fly
linux tarball (from https://github.com/concourse/concourse/releases) to be present in the `fly/` folder e.g:

```
$cwd/
├── fly/
│   └── fly-5.0.0-linux-amd64.tgz
└── concourse-pipeline-resource/
    ├── .git/
    │    └── ... 
    ├── dockerfiles/
    │    ├── alpine/
    │    │    └── Dockerfile
    │    └── ubuntu/
    │         └── Dockerfile
    └── ...
```

Run the tests with the following commands for both `alpine` and `ubuntu` images:

```sh
docker build -t concourse-pipeline-resource -f concourse-pipeline-resource/dockerfiles/alpine/Dockerfile .
docker build -t concourse-pipeline-resource -f concourse-pipeline-resource/dockerfiles/ubuntu/Dockerfile .
```

### Contributing

Please [ensure the tests pass locally](#running-the-tests).
