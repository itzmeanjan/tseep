#!/bin/bash
echo "Preparing all !"

for i in {1..6}
do
docker run --name client_${i} -d --env-file client.env client
done

echo "Prepared all !"
exit 0
