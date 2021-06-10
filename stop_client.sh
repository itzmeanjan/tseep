#!/bin/bash
echo "Stopping all !"

for i in {1..6}
do
docker stop client_${i}
done

echo "Stopped all !"
exit 0
