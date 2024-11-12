#!/bin/bash

set -o allexport
source .env
set -o allexport

IFS=',' read -r -a ports <<< "$REDIS_CLUSTER_NODES"
Network="127.0.0.1"

# Ensure the number of ports provided can support 3 masters and 3 slaves
if [ "${#ports[@]}" -lt 6 ]; then
    echo "Error: You need at least 6 ports specified in REDIS_CLUSTER_NODES."
    exit 1
fi

# Setup Redis configuration and directories for each node
for i in "${ports[@]}"; do 
    mkdir -p $i
    cd $i

    cat <<EOL > redis.conf
bind $Network
port $i
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
appendonly yes
EOL

    if ! redis-server ./redis.conf --daemonize yes; then
        echo "Failed to start Redis on port $i"
        exit 1
    fi

    echo "Running Redis at $i"
    ufw allow $i
    cd ..
done

# Avoid Split Brain situation
# https://redis.io/learn/operate/redis-at-scale/scalability/exercise-1#step-2
redis-cli --cluster create $(printf "$Network:%s " "${ports[@]}") --cluster-replicas 1

redis-cli -p "${ports[0]}" cluster info

echo "Cluster created with 3 master and 3 slave nodes. To enable read operations from slaves, use the 'readonly' command in your Redis client connections when connecting to replicas."
