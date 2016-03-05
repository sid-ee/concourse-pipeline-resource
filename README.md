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

### Project management

The CI for this project can be found at https://concourse.robdimsdale.com/pipelines/concourse-pipeline-resource
and the scripts can be found in the
[robdimsdale-ci repository](https://github.com/robdimsdale/robdimsdale-ci).

The roadmap is captured in [Pivotal Tracker](https://www.pivotaltracker.com/projects/1549921).
