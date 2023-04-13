# multi-factor-auth-go-service

## Prerequisites

In order to run the service properly, the configuration options have to be set accordingly.

For testing purposes, most of the default configuration options should be fine.
In particular, `external.toml` config file should be modified to set API url and
mongoDB connection string, if `mongoDB` persiter type is selected.

## Local testing environment with mongo

The `Makefile` commands can be used to manage the testing setup more easily.
There are commands for running the service locally, there are commands for running the
service with docker, there are also commands for running mongoDB setup with
`docker-compose` deployments. (Check all available `Makefile` commands)

> For production systems make sure to follow the proper infrastructure setup and
> proper security considerations.

### Single mongoDB instance

The default setup consists of a single mongoDB instance, running with docker on host
networking. The `tcs` service can be started locally, without docker.

To start the mongoDB instance run the following:
```bash
# It will start and set up the mongoDB instance.
# It will take several seconds to complete the setup
make compose-new
```

Start the `tcs` service:
```bash
# It will start the service locally
make run
```

### MongoDB cluster

Firstly, the MongoDB connection string has to be updated accordingly. For the current testing
setup the following `URI` can be set in `external.toml` config file: `mongodb://mongodb0:27017,mongodb1:27018,mongodb2:27019/?replicaSet=mongoReplSet`

To start a setup with mongoDB replica set cluster, use the following Makefile command:
```bash
# It will take several seconds to complete the setup
make compose-new db_setup=mongodb-cluster
```

Start the `tcs` service:
```bash
# It will set the docker image and it will launch the container
make docker-new db_setup=mongodb-cluster
```

Make sure to run all related commands with `db_setup=mongodb-cluster`.
