#!/bin/bash

# Correctly define the array without a trailing comma
ports=(3002 3003 3004 3005 3006 3007 3008 3009 3010 3011 3012)

# Change to the correct directory
cd cmd/server || exit 1

# Loop through the ports
for port in "${ports[@]}"; do
  echo "Starting instance on port $port..."
  # Start the server on each port in the background
  go run main.go --port="$port" &
done

# Wait for all background processes to finish
wait
