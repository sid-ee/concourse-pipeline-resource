# Concourse Pipeline Resource

Interact with concourse pipelines from concourse.

## Installing

This resource is only compatible with Concourse versions 0.74.0 and higher
as no BOSH release is provided. Use this resource by adding the following to
the `resource_types` section of a pipeline config:

```yaml
resource_types:
- name: concourse-pipeline
  type: docker-image
  source:
    repository: robdimsdale/concourse-pipeline-resource
    tag: v0.6.3
```

The value for `tag` should be the latest tag in this repository.

The docker image is `robdimsdale/concourse-pipeline-resource`; the images are
available on
[dockerhub](https://hub.docker.com/r/robdimsdale/concourse-pipeline-resource).

The rootfs of the docker image is available with each release on the
[releases page](https://github.com/robdimsdale/concourse-pipeline-resource/releases).

The docker image is semantically versioned; these versions correspond to the git tags
in this repository.

## Source Configuration

* `target`: *Required.*  URL of your concourse instance e.g. `https://my-concourse.com`.

* `username`: *Required.*  Basic auth username for logging in to Concourse.
  Basic Auth must be enabled on the Concourse installation.

* `password`: *Required.*  Basic auth password for logging in to Concourse.
  Basic Auth must be enabled on the Concourse installation.

### Example Pipeline Configuration

#### Check

``` yaml
---
resources:
- name: my-pipelines
  type: concourse-pipeline
  source:
    target: https://my-concourse.com
    username: some-user
    password: some-password
```

## Behavior

### `check`: Check for changes to the pipelines.

Return a checksum of the concatenated contents of all pipelines.

## Developing

### Prerequisites

A valid install of golang >= 1.5 is required.

### Dependencies

Dependencies are vendored in the `vendor` directory, according to the
[golang 1.5 vendor experiment](https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&cad=rja&uact=8&ved=0ahUKEwi7puWg7ZrLAhUN1WMKHeT4A7oQFggdMAA&url=https%3A%2F%2Fgolang.org%2Fs%2Fgo15vendor&usg=AFQjCNEPCAjj1lnni5apHdA7rW0crWs7Zw).

If using golang 1.6, no action is required.

If using golang 1.5 run the following command:

```
export GO15VENDOREXPERIMENT=1
```

### Running the tests

Install the ginkgo executable with:

```
go get -u github.com/onsi/ginkgo/ginkgo
```

The tests require a concourse API server to test against, and a valid
basic auth username/password for that concourse deployment.

The tests also require that you build the fly CLI as a binary.
This CLI must be compatible with the chosen concourse deployment.
The source for the fly CLI can be found [here](https://github.com/concourse/fly).
`FLY_LOCATION` should be set to the location of the compiled binary.

Run the tests with the following command:

```
FLY_LOCATION=/path/to/fly/cli \
TARGET=https://my-concourse.com \
USERNAME=my-basic-auth-user \
PASSWORD=my-basic-auth-password \
./bin/test
```

### Project management

The CI for this project can be found at https://concourse.robdimsdale.com/pipelines/concourse-pipeline-resource
and the scripts can be found in the
[robdimsdale-ci repository](https://github.com/robdimsdale/robdimsdale-ci).

The roadmap is captured in [Pivotal Tracker](https://www.pivotaltracker.com/projects/1549921).
