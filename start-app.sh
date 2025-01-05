docker pull ghcr.io/vitorcodesalittle/colab-lists:latest;
docker run -d -p 443:8080 -v ./data/:/app/data --restart always --entrypoint '/app/main' ghcr.io/vitorcodesalittle/colab-lists:latest \
	-tls \
	-certificate /app/data/live/lists.vilmasoftware.com.br/fullchain1.pem \
	-private-key /app/data/live/lists.vilmasoftware.com.br/privkey1.pem \
	-smtp-host smtp.gmail.com \
	-smtp-port 587 \
	-smtp-noreply 'vilmasoftware@gmail.com' \
	-smtp-username 'vilmasoftware@gmail.com' \
	-smtp-password 'dlya tskh oryn orra'

# docker run -p 443:8080 -v ./data/:/app/data -i --entrypoint '/bin/sh' ghcr.io/vitorcodesalittle/colab-lists:latest
