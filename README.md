# Docker Janitor

Simple docker cleaner for images, volumes, networks and containers. Written as a maintenance utility that relays inside as a container or a standalone binary.

You can specify time interval to cleanup, time delay in which container is safe after creation and give a filter list to exclude specific image names or tags from being deleted.

# Usage

## Binary, package

* You can download the latest release and run binary as is:
  ```bash
  docker-janitor -interval 1h -freshness 30m -clear-images
  ```
* TODO Add package instructions after goreleaser integrations
## Docker
Or use docker image. To run it ad-hoc (once and exit):
```bash
docker run \
  --name=docker-janitor
  -e ONCE=true \
  -e CLEAR_IMAGES=true \
  -e EXCLUDE_TAGS=latest,stable,5.22 \
  -e FRESHNESS=30m \
  -v /var/run/docker.sock:/var/run/docker.sock -d houstonheat/docker-janitor:latest
```
To run the latest version in a standalone host you can use docker-compose example file:
```bash
docker-compose -f deploy/janitor-compose.yml up -d
```

## Command line flags

Name                           | Environment Variable Name | Description
-------------------------------|---------------------------|-----------------
`-once`                        | ONCE                      | Execute cleaning just once
`-debug`                       | DEBUG                     | Set log level to debug
`-dry-run`                     | DRY_RUN                   | Do not change anything
`-clear-containers`            | CLEAR_CONTAINERS          | Clear unused containers, same as `docker container prune`
`-clear-networks`              | CLEAR_NETWORKS            | Clear unused networks, same as `docker network prune`
`-clear-volumes`               | CLEAR_VOLUMES             | Clear unused volumes, same as `docker volume prune`
`-clear-images`                | CLEAR_IMAGES              | Clear unused images, same as `docker image prune -a` without any filters
`-exclude-fullnames`           | EXCLUDE_FULLNAMES         | Comma separated list of images fullnames `(repo[:port]/path:tag)` to exclude from cleaning (e.g. `-exclude-names registry.domain:9000/path/name:v1.0.0`). This option only makes sense when `-clear-images` flag is set
`-exclude-names`               | EXCLUDE_NAMES             | Comma separated list of images names `(repo[:port]/path)` to exclude from cleaning (e.g. `-exclude-names registry.domain/path/name,path/name/,ubuntu,myimage`). This option only makes sense when `-clear-images` flag is set
`-exclude-tags`                | EXCLUDE_TAGS              | Comma separated list of images tags to exclude from cleaning (e.g. `-exclude-tags latest,stable,5.22`). This option only makes sense when `-clear-images` flag is set
`-freshness`                   | FRESHNESS                 | Freshness will keep images that were created in the given time period (default 1h)
`-interval`                    | INTERVAL                  | Interval to check on unused elements (default 12h). This option only makes sense when the -once flag is not set.

# Development

## Project
* TODO add development instructions
## Testing

```bash
# The most common case: run all tests
go test -v ./...
```

# Contributing

Easiest way to contribute is to provide feedback! Create an issue or [ping
houston\_heat\_ on Twitter](https://twitter.com/houston_heat_). Any contributions you make are greatly appreciated:

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'houstonheat/docker-janitor#issue_number | Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

# License
Distributed under the MIT License. See [LICENSE](LICENSE) for more information.

# Contact
[@houston_heat_](https://twitter.com/houston_heat_) - houstonheat@yandex.ru

Project Link: https://github.com/houstonheat/docker-janitor

# TODO

- Add unit tests for:
  - boolean flags
  - docker package
