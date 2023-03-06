#!/bin/bash

set -x

mongosh --host mongodb0 --eval <<EOF
var config = {
    "_id": "mongoReplSet",
    "version": 1,
    "members": [
        {
            "_id": 0,
            "host": "mongodb0"
        },
        {
            "_id": 1,
            "host": "mongodb1"
        },
        {
            "_id": 2,
            "host": "mongodb2"
        }
    ]
};
rs.initiate(config, { force: true });
EOF
