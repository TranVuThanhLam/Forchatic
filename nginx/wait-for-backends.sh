#!/bin/sh

BACKENDS="backend:8080 backend2:8080"  # Thêm nhiều hơn nếu cần: backend3:8080 backend4:8080

for backend in $BACKENDS; do
  host=$(echo $backend | cut -d: -f1)
  port=$(echo $backend | cut -d: -f2)
  echo "Waiting for $host:$port..."
  while ! nc -z "$host" "$port"; do
    echo "Still waiting for $host:$port..."
    sleep 2
  done
  echo "$host:$port is up!"
done

echo "All backends are up! Starting Nginx..."
exec nginx -g 'daemon off;'
