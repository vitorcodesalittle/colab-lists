#!/usr/bin/bash
set -o errexit
[ -z "$(docker ps -q)" ] && docker stop $(docker ps -q)
docker run --rm -it \
	-v "/etc/letsencrypt:/etc/letsencrypt" \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-p 80:80 \
	certbot/certbot certonly --standalone --agree-tos --email vitormaia1890@gmail.com -n -d lists.vilmasoftware.com.br
cp /etc/letsencrypt/archive/lists.vilmasoftware.com.br ./data/live/ -R
[ -z "$(docker ps -q)" ] && docker stop $(docker ps -q)
bash start-app.sh
