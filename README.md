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

The key is composed of user bech32 address and ip address. In order for this key to be correct,
we need to make sure that the user ip is as close to the real one as possible.

There is a separate section in the main configuration file for http headers handling.
Please check `HTTPHeaders` section in [config.toml](https://github.com/multiversx/mx-multi-factor-auth-go-service/blob/main/cmd/multi-factor-auth/config/external.toml) for more details.

If the server is exposed directly to the internet, you can set custom headers list as
empty and the number of trusted proxies to 0. This way the client IP will be set directly
from the network address that send the last request (which will be the client in this case).

If the server is exposed through one or multiple trusted proxies, make sure to set custom
headers list (for example with `cf-connecting-ip` if using cloudflare), and number of
intermediate proxies field (for x-forwarded-for header, if it will end up being used;
as decribed below, x-forwarded-for header will be used if custom headers will
not contain the necessary information).

There are several steps considered when trying to fetch the client real IP:
* firstly, we try to fetch client ip based on custom http headers from trusted proxies;
custom http headers can be set in http headers config section
* if not found in custom http headers, search in `x-forwarded-for` header based on the
variables in http headers config section
* if not found in `x-forwarded-for` header or if `x-forwarded-for` header is not trusted
(this can be done by setting number of
trusted proxies in config to 0 and header will not be checked at all),
it will fetch the client ip from gin's RemoteAddr, which
contains the network address that send the last request; this field is relevant and safe when
there are no intermediate proxies between client and server

> **Important**
> Make sure to allign `mfa` configuration for `HTTPHeaders` with infrastructure.

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
