# multi-factor-auth-go-service

## Prerequisites

In order to run the service properly, the configuration options have to be set accordingly.

For testing purposes, most of the default configuration options should be fine.
In particular, `external.toml` config file should be modified to set API url and
mongoDB connection string, if `mongoDB` persiter type is selected.

## Infrastructure

MFA (multi-factor-auth) service uses MongoDB for persistent storage and Redis for caching.
The configuration file for setting up the interaction with external services
(like MongoDB and Redis) can be found 
[here](https://github.com/multiversx/mx-multi-factor-auth-go-service/blob/main/cmd/multi-factor-auth/config/external.toml).

### Redis caching

Redis is used for rate limiting. If maximum number of failed attempts for two factor
authentication codes is reached (based on the specified value in config), the user will
be blocked for several minutes (as specified in config). This is done by setting a redis
key with TTL for the blocked user.

The key is composed of `user address` and `IP address`. In order for this key to be set
correctly, we need to make sure that the user IP is as close to the real one as possible.

There is a separate section in the main configuration file for http headers handling.
Please check `HTTPHeaders` section in [external.toml](https://github.com/multiversx/mx-multi-factor-auth-go-service/blob/main/cmd/multi-factor-auth/config/external.toml) for more details.

If the server is exposed directly to the internet, you can set `ForwardedByClientIP` 
to `false` and `TrustedPlatform` as empty. This way the client IP will be set directly
from the network address that sent the last request (which will be the client in
this case, or an untrusted proxy).

If the server is exposed through one or multiple trusted proxies, make sure to set
`TrustedPlatform`, `RemoteIPHeaders` and `TrustedProxies` fields accordingly. If
they are not set explicitly, the default `gin` values will be set. Please check
also `gin` documentation for these fields.

> **Important**
> For this to work properly, make sure to align `Gin` configuration section 
> with your infrastructure setup.

## Local testing environment

The `Makefile` commands can be used to manage the testing setup more easily.
There are commands for running the service locally, there are commands for running the
service with docker containers, there are also commands for running mongoDB and redis setup with
`docker-compose` deployments. (Check all available `Makefile` commands)

> Tested with:
> * docker-compose version 1.25.0
> * Docker version 20.10.12

> **Warning**
> For production systems make sure to follow the proper infrastructure setup and
> proper security considerations.

The local testing environment can be setup in two ways:
- mongoDB single, redis single
- mongoDB cluster, redis single

### Single mongoDB instance and redis

The default setup consists of a single mongoDB and a single redis instnace, running with
docker on host networking mode. The `mfa` service can be started locally, without docker.

To start the mongoDB instance run the following:
```bash
# It will start and set up the mongoDB instance.
# It will take several seconds to complete the setup
make compose-new
```

Start the `mfa` service:
```bash
# It will start the service locally
make run

# It will start the service locally with docker
make docker-run
```

### MongoDB cluster and redis

Firstly, the MongoDB connection string has to be updated accordingly. For the current testing
setup the following `URI` can be set in `external.toml` config file: `mongodb://mongodb0:27017,mongodb1:27018,mongodb2:27019/?replicaSet=mongoReplSet`

To start a setup with mongoDB replica set cluster, use the following Makefile command:
```bash
# It will take several seconds to complete the setup
make compose-new db_setup=cluster
```

Start the `mfa` service:
```bash
# It will set the docker image and it will launch the container
make docker-new db_setup=cluster
```

> **Note**
> Make sure to run all related commands with `db_setup=cluster`.
