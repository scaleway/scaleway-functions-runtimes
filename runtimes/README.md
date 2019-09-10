# Scaleway Functions's Runtimes

This directory contains the code related to `language-specific runtimes`, used on Scaleway Functions (Serverless - FAAS platform).

These runtimes are **Developped and Maintained by Scaleway Serverless Team**.

## Use these runtimes

### With Docker (Dockerfiles)

#### Pre-requisites

- Install Docker
- Install make (**note** that you may also build images with shell commands instead of make).

#### Build Core runtime Base image:

First, you will have to build the base `core-runtime` docker image, extended by each of our runtimes.
```bash
make build_container tag_release
```

This should be a docker image `rg.fr-par.scw.cloud/scwserverlessruntimes/core-runtime:$TAG` (Tag is specified as a variable in the [Makefile](../Makefile)).

#### Build the runtimes

For each of the runtime:

```bash
# Inside of a runtimes/${runtime} directory
docker build -t your-image:your-tag .
```

**If you want to add your function handler's code in the runtime**, you will have to update the related `Dockerfile` to copy it at the right place:
```Dockerfile
# ... instructions related to runtime building

COPY ./myfunction .

# Set Environment variables
ENV SCW_HANDLER_PATH=/path/to/my/handler
ENV SCW_HANDLER_NAME=exported_function
```
