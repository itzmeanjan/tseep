#!/bin/bash
echo "Starting all !"

for i in {1..6}
do
docker start client_${i}
done

echo "Started all !"
exit 0
