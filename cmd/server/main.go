package main

import (
	"context"
	"github.com/JkLondon/mcp-stocks-info-server/pkg/db"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	repositories2 "github.com/JkLondon/mcp-stocks-info-server/internal/core/ports/repositories"

	"github.com/JkLondon/mcp-stocks-info-server/internal/adapters/mcp"
	"github.com/JkLondon/mcp-stocks-info-server/internal/adapters/repositories"
	"github.com/JkLondon/mcp-stocks-info-server/internal/adapters/repositories/apis"
	"github.com/JkLondon/mcp-stocks-info-server/internal/adapters/services"
	"github.com/JkLondon/mcp-stocks-info-server/internal/config"
	"github.com/JkLondon/mcp-stocks-info-server/pkg/cache"
)

func main() {
	// Определяем путь к конфигурационному файлу
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// Если не удалось загрузить конфигурацию, используем значения по умолчанию
		log.Printf("Не удалось загрузить конфигурацию: %v. Используем значения по умолчанию.", err)
		cfg = &config.Config{}
		cfg.Cache.DefaultTTL = 5 * time.Minute
		cfg.Server.Port = 8080
		cfg.MOEX.BaseURL = "https://iss.moex.com/iss"
		cfg.NewsAPI.BaseURL = "https://newsapi.org/v2"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем кэш
	var cacheClient cache.Cache
	if cfg.Cache.RedisURI != "" {
		// Если указан URI Redis, используем Redis для кэширования
		cacheClient, err = cache.NewRedisCache(cfg.Cache.RedisURI, cfg.Cache.RedisDB)
		if err != nil {
			log.Fatalf("Ошибка инициализации Redis: %v", err)
		}
		log.Printf("Инициализирован Redis-кэш: %s", cfg.Cache.RedisURI)
	} else {
		// В противном случае используем in-memory кэш
		cacheClient = cache.NewInMemoryCache(cfg.Cache.DefaultTTL)
		log.Printf("Инициализирован in-memory кэш с TTL %v", cfg.Cache.DefaultTTL)
	}

	// Создаем подключение к MongoDB
	var mongoDB *db.MongoDB
	if cfg.Database.URI != "" {
		mongoDB, err = db.NewMongoDB(
			cfg.Database.URI,
			cfg.Database.Database,
			cfg.Database.Collection,
			cfg.Database.Timeout,
		)
		if err != nil {
			log.Fatalf("Ошибка подключения к MongoDB: %v", err)
		}
		defer mongoDB.Close(ctx)
		log.Printf("Подключение к MongoDB: %s/%s", cfg.Database.URI, cfg.Database.Database)
	} else {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: URI базы данных не указан, будет использоваться только кэш")
		// Можно создать заглушку для базы данных
	}

	// Создаем API-клиенты
	moexAPI := apis.NewMOEXAPIClient(cfg, cacheClient)
	newsAPI := apis.NewNewsAPIClient(cfg, cacheClient)

	// Создаем репозитории
	var stockRepo repositories2.StockRepository
	var newsRepo repositories2.NewsRepository

	if mongoDB != nil {
		// Если есть подключение к MongoDB, используем его
		stockRepo = repositories.NewStockRepository(
			mongoDB.GetDatabase(),
			cacheClient,
			moexAPI,
			cfg.Cache.StocksTTL,
			true,
		)

		newsRepo = repositories.NewNewsRepository(
			mongoDB.GetDatabase(),
			cacheClient,
			newsAPI,
			cfg.Cache.NewsTTL,
			true,
		)
	} else {
		// Иначе создаем заглушки для репозиториев
		// Здесь должна быть реализация mock-репозиториев
		log.Fatalf("В текущей версии требуется MongoDB для работы сервера")
	}

	// Создаем сервисы
	stockService := services.NewStockService(stockRepo)
	newsService := services.NewNewsService(newsRepo)

	// Создаем MCP сервер
	mcpServer := mcp.NewMCPServer(cfg, stockService, newsService)

	// Обработка сигналов для корректного завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем MCP сервер
	go func() {
		log.Println("Запуск MCP сервера...")
		if err := mcpServer.Start(); err != nil {
			log.Fatalf("Ошибка запуска MCP сервера: %v", err)
		}
	}()

	// Ожидаем сигнала для завершения
	<-sigChan
	log.Println("Получен сигнал завершения. Останавливаем сервер...")
	cancel() // Отменяем контекст для корректного завершения всех операций
	log.Println("Сервер остановлен")
}
