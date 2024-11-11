#!/bin/bash

set -o allexport
source .env
set -o allexport

IFS=',' read -r -a ports <<< "$REDIS_CLUSTER_NODES"

for i in "${ports[@]}"; do 
    echo "Stopping Redis on port $i..."
    redis-cli -p "$i" shutdown

    if [ $? -eq 0 ]; then
        echo "Redis on port $i stopped successfully."
    else
        echo "Failed to stop Redis on port $i or it was not running."
    fi

    echo "Deleting directory $i..."
    rm -rf "$i"

    echo "Removing firewall rule for port $i..."
    ufw delete allow "$i"
done

echo "All Redis instances and associated resources have been deleted."
