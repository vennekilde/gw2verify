# gw2verify

A Guild Wars 2 verification backend that allows a user to link multiple Guild Wars 2 accounts.

Integrating the backend to an application is done using the provided OpenAPI v3 spec found at [/api/openapi.yaml](https://github.com/vennekilde/gw2verify/api/openapi.yaml)

Intended to be used with multiple platforms and has so far been used on Far Shiverpeaks to integrate Website, Teamspeak & Discord to the same backend.

## Building

### Docker Image

`make package`

The code will be compiled during the docker build process

### Target: Host Machine

`make build`

### Target: Linux

`make build_linux`

### Target: Windows

`make build_windows`
