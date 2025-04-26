# MOEX Stocks & News MCP Server

MCP (Model Context Protocol) сервер для доступа к данным о котировках акций на Московской бирже (MOEX) и финансовым новостям. Этот сервер предоставляет интерфейс для взаимодействия с LLM (Large Language Models) для получения актуальной информации о котировках и новостях.

## Возможности

- Получение информации о котировках акций на MOEX
- Получение списка растущих/падающих акций
- Поиск акций по названию или тикеру
- Получение финансовых новостей за текущий день
- Поиск новостей по ключевым словам
- Получение новостей, связанных с конкретными акциями
- Анализ акций на основе котировок и новостей
- Обзор рынка с ключевыми показателями
- Анализ финансовых новостей

## Особенности

- Использование MCP (Model Context Protocol) для интеграции с LLM
- Кэширование данных для быстрого доступа
- Поддержка как Redis, так и in-memory кэша
- Хранение исторических данных в MongoDB
- Чистая архитектура с разделением на слои
- API ключи для доступа к внешним источникам данных
- Контейнеризация с использованием Docker и Docker Compose

## Требования

- Docker и Docker Compose
- API ключи для доступа к внешним сервисам (опционально)

## Установка и запуск с помощью Docker

1. Клонируйте репозиторий:

```bash
git clone https://github.com/jklondon/mcp-stocks-info-server.git
cd mcp-stocks-info-server
```

2. Создайте файл с переменными окружения:

```bash
# Создайте .env файл из примера
cp .env.example .env

# Отредактируйте .env файл, установив свой API ключ NewsAPI
# NEWSAPI_KEY=your_news_api_key_here
```

3. Запустите сервисы с помощью Docker Compose:

```bash
docker-compose up -d
```

4. Проверьте работу сервисов:
```bash
# Посмотреть логи всех сервисов
docker-compose logs

# Посмотреть логи конкретного сервиса
docker-compose logs app
```

## Запуск без Docker

### Требования

- Go 1.21 или выше
- MongoDB (опционально, но рекомендуется)
- Redis (опционально)

### Установка

1. Клонируйте репозиторий:

```bash
git clone https://github.com/jklondon/mcp-stocks-info-server.git
cd mcp-stocks-info-server
```

2. Установите зависимости:

```bash
go mod download
```

3. Создайте конфигурационный файл `config.yaml` (см. пример ниже)

4. Соберите сервер:

```bash
go build -o mcp-stocks-server ./cmd/server
```

### Запуск сервера

```bash
./mcp-stocks-server config.yaml
```

### Пример конфигурационного файла

```yaml
server:
  port: 8080
  host: "localhost"
  timeoutSeconds: 30

database:
  uri: "mongodb://localhost:27017"
  database: "mcp_stocks"
  collection: "stocks"
  timeout: "5s"

cache:
  redisURI: "localhost:6379"
  redisDB: 0
  defaultTTL: "5m"
  stocksTTL: "15m"
  newsTTL: "30m"

moex:
  baseURL: "https://iss.moex.com/iss"
  timeout: "10s"
  useCache: true
  apiKey: "" # Опционально

newsAPI:
  baseURL: "https://newsapi.org/v2"
  timeout: "10s"
  useCache: true
  apiKey: "your_news_api_key_here" # Требуется для доступа к NewsAPI
  sources: ["rbc", "vedomosti", "kommersant"]

apiKeys:
  moexKey: "" # Опционально
  newsAPIKey: "your_news_api_key_here" # Дублирует newsAPI.apiKey

logLevel: "info"
environment: "development"
```

## Интеграция с LLM

Для интеграции с LLM ваш клиент должен поддерживать протокол MCP. Вы можете использовать любой MCP-совместимый клиент для взаимодействия с этим сервером.

### Доступные инструменты (tools)

- `get_stock_info` - получение информации о котировке акции
- `get_top_gainers` - получение списка топ растущих акций
- `get_top_losers` - получение списка топ падающих акций
- `search_stocks` - поиск акций по названию или тикеру
- `get_today_news` - получение финансовых новостей за сегодня
- `search_news` - поиск новостей по ключевому слову
- `get_news_by_ticker` - получение новостей, связанных с указанным тикером

### Доступные шаблоны (prompts)

- `stock_analysis` - анализ котировок акции
- `market_overview` - общий обзор состояния рынка
- `news_analysis` - анализ финансовых новостей за сегодня

## Участие в разработке

Проект является открытым, и любой может внести свой вклад. Если у вас есть предложения или исправления, создайте issue или pull request.

## Лицензия

MIT