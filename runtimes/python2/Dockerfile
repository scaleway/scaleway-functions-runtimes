FROM rg.fr-par.scw.cloud/scwfunctionsruntimes/python-dep2:v1.1.0

RUN mkdir -p /home/app

# Wrapper/boot-strapper
WORKDIR /home/app

COPY requirements.txt   .
RUN pip install -r requirements.txt
COPY index.py           .
COPY bootup.sh          .
RUN chmod +x bootup.sh

# Allow user to add his dependencies to `/package` at root of function archive
ENV PYTHONPATH=/home/app/function:/home/app/function/package
ENV PORT=8080
ENV SCW_UPSTREAM_PORT=8081
ENV SCW_HANDLER_PATH=/home/app/function/handler
ENV SCW_HANDLER_NAME=handle
ENV SCW_RUNTIME_BINARY=/bin/sh
ENV SCW_RUNTIME_BRIDGE=/home/app/bootup.sh
