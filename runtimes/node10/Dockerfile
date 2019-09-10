FROM rg.fr-par.scw.cloud/scwserverlessruntimes/core:v1.2.2 AS runtime
LABEL intermediate=true

# Build step, build TypeScript files into JavaScript
FROM rg.fr-par.scw.cloud/scwserverlessruntimeslib/node:10.12.0-alpine as build

WORKDIR /home/app

COPY package.json .

RUN npm install

COPY tsconfig.json .
COPY src ./src

RUN npm run build

# Run Step, prepare the environment for the function runtime running in production
FROM rg.fr-par.scw.cloud/scwserverlessruntimeslib/node:10.12.0-alpine

RUN apk add libc6-compat

WORKDIR /root/

RUN mkdir -p /home/app

# Wrapper/boot-strapper
WORKDIR /home/app
COPY package.json ./

# This ordering means the npm installation is cached for the outer function handler.
RUN npm i --production

ENV PORT=8080
ENV SCW_HANDLER_PATH=/home/app/function/handler
ENV SCW_HANDLER_NAME=handle
ENV SCW_RUNTIME_BINARY=/usr/local/bin/node
ENV SCW_RUNTIME_BRIDGE=/home/app/index.js
ENV SCW_UPSTREAM_PORT=8081

# Copy outer function handler
COPY --from=build /home/app/dist .
COPY --from=runtime /home/app /home/app