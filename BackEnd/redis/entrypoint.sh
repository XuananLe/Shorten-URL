#!/bin/bash

set -o allexport
source .env
set -o allexport

IFS=',' read -r -a ports <<< "$REDIS_CLUSTER_NODES"
Network="127.0.0.1"

# Loop through the ports and start Redis instances
for i in "${ports[@]}"; do 
    mkdir -p $i

    cd $i

    cat <<EOL > redis.conf
port $i
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
appendonly yes
EOL
    redis-server ./redis.conf --daemonize yes
    echo "Running Redis at $i"
    ufw allow $i
    cd ..
done

redis-cli --cluster create $Network:${ports[0]} \
  $Network:${ports[1]} \
  $Network:${ports[2]} \
  $Network:${ports[3]} \
  $Network:${ports[4]} \
  $Network:${ports[5]} \
  --cluster-replicas 1
