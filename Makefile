.PHONY: build run docker-build docker-run docker-compose-up docker-compose-down

# Переменные
BINARY_NAME=mcp-stocks-server
DOCKER_IMAGE=mcp-stocks-server
DOCKER_COMPOSE=docker-compose

# Сборка приложения
build:
	go build -o $(BINARY_NAME) ./cmd/server

# Запуск приложения
run: build
	./$(BINARY_NAME) config.yaml

# Сборка Docker образа
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Запуск Docker контейнера
docker-run: docker-build
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# Запуск всех сервисов через Docker Compose
docker-compose-up:
	$(DOCKER_COMPOSE) up -d

# Остановка всех сервисов Docker Compose
docker-compose-down:
	$(DOCKER_COMPOSE) down

# Проверка статуса сервисов
docker-compose-ps:
	$(DOCKER_COMPOSE) ps

# Просмотр логов всех сервисов
docker-compose-logs:
	$(DOCKER_COMPOSE) logs

# Просмотр логов приложения
docker-compose-logs-app:
	$(DOCKER_COMPOSE) logs app

# Очистка
clean:
	rm -f $(BINARY_NAME)
	docker rmi $(DOCKER_IMAGE) || true 