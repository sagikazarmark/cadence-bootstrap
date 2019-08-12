# Cadence Bootstrap

[![CircleCI](https://circleci.com/gh/sagikazarmark/cadence-bootstrap.svg?style=svg)](https://circleci.com/gh/sagikazarmark/cadence-bootstrap)
[![Go Report Card](https://goreportcard.com/badge/github.com/sagikazarmark/cadence-bootstrap?style=flat-square)](https://goreportcard.com/report/github.com/sagikazarmark/cadence-bootstrap)
[![GolangCI](https://golangci.com/badges/github.com/sagikazarmark/cadence-bootstrap.svg)](https://golangci.com/r/github.com/sagikazarmark/cadence-bootstrap)
[![GoDoc](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/sagikazarmark/cadence-bootstrap)

Cadence Bootstrap helps setting up a Cadence instance.
It's mostly helpful in development environments (eg. to setup domains in a Docker Compose setup).


## Usage

```bash
docker run --rm -it -e CADENCE_HOST=docker.for.mac.host.internal -e CADENCE_DOMAIN=cadence-samples -e CADENCE_RETENTION=3 sagikazarmark/cadence-bootstrap cadence-bootstrap
```


## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
