# qdrant-cloud-buf-plugins

Collection of [Buf plugins](https://buf.build/docs/cli/buf-plugins/overview/) used by Qdrant Cloud APIs.

## Development

This project leverages Make to automate common development tasks. To view all available commands, run:

``` sh
make help
```

### Setup

To work with this project locally, you need to have [Go](https://go.dev/doc/install) installed.
Additionally, there are other required dependencies that you can install running:

``` sh
make bootstrap
```

### Running tests

To run the tests, execute:

``` sh
make test
```

### Formatting & linting code

To format and lint the code of the project, execute:

``` sh
make fmt
make lint
```
