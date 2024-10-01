start:
	docker compose --env-file .env.example up -d --build
stop:
	docker compose --env-file .env.example down
