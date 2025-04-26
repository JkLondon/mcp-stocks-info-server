package mcp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JkLondon/mcp-stocks-info-server/internal/config"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/ports/services"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server представляет собой MCP сервер для работы с акциями и новостями
type Server struct {
	server       *server.MCPServer
	stockService services.StockService
	newsService  services.NewsService
	config       *config.Config
}

// NewMCPServer создает новый экземпляр MCP сервера
func NewMCPServer(cfg *config.Config, stockService services.StockService, newsService services.NewsService) *Server {
	// Создаем MCP сервер

	// Логирование запросов
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		fmt.Printf("beforeAny: %s, %v, %v\n", method, id, message)
	})
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		fmt.Printf("onSuccess: %s, %v, %v, %v\n", method, id, message, result)
	})
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		fmt.Printf("onError: %s, %v, %v, %v\n", method, id, message, err)
	})
	hooks.AddBeforeInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest) {
		fmt.Printf("beforeInitialize: %v, %v\n", id, message)
	})
	hooks.AddAfterInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
		fmt.Printf("afterInitialize: %v, %v, %v\n", id, message, result)
	})
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		fmt.Printf("afterCallTool: %v, %v, %v\n", id, message, result)
	})
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		fmt.Printf("beforeCallTool: %v, %v\n", id, message)
	})

	mcpServer := server.NewMCPServer(
		"Stocks & News API",
		"1.0.0",
		// Добавляем hooks
		server.WithHooks(hooks),
	)

	return &Server{
		server:       mcpServer,
		stockService: stockService,
		newsService:  newsService,
		config:       cfg,
	}
}

// Start запускает MCP сервер
func (s *Server) Start() error {
	// Регистрируем инструменты (tools)
	s.registerTools()

	// Регистрируем шаблоны (prompts)
	s.registerPrompts()

	// Запускаем сервер
	return server.ServeStdio(s.server)
}

// registerTools регистрирует инструменты (tools) в MCP сервере
func (s *Server) registerTools() {
	// Регистрируем инструменты для работы с акциями
	s.registerStockTools()

	// Регистрируем инструменты для работы с новостями
	s.registerNewsTools()
}

// registerStockTools регистрирует инструменты для работы с акциями
func (s *Server) registerStockTools() {
	// Инструмент для получения информации об акции
	getStockTool := mcp.NewTool("get_stock_info",
		mcp.WithDescription("Получить информацию о котировке акции на MOEX"),
		mcp.WithString("ticker",
			mcp.Required(),
			mcp.Description("Тикер акции (например, SBER, GAZP, LKOH)"),
		),
	)

	s.server.AddTool(getStockTool, s.handleGetStockInfo)

	// Инструмент для получения топ растущих акций
	getTopGainersTool := mcp.NewTool("get_top_gainers",
		mcp.WithDescription("Получить список топ растущих акций на MOEX"),
		mcp.WithNumber("limit",
			mcp.Description("Количество акций в списке (по умолчанию 10)"),
		),
	)

	s.server.AddTool(getTopGainersTool, s.handleGetTopGainers)

	// Инструмент для получения топ падающих акций
	getTopLosersTool := mcp.NewTool("get_top_losers",
		mcp.WithDescription("Получить список топ падающих акций на MOEX"),
		mcp.WithNumber("limit",
			mcp.Description("Количество акций в списке (по умолчанию 10)"),
		),
	)

	s.server.AddTool(getTopLosersTool, s.handleGetTopLosers)

	// Инструмент для поиска акций
	searchStocksTool := mcp.NewTool("search_stocks",
		mcp.WithDescription("Поиск акций по названию или тикеру"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Поисковый запрос (часть названия или тикера)"),
		),
	)

	s.server.AddTool(searchStocksTool, s.handleSearchStocks)
}

// registerNewsTools регистрирует инструменты для работы с новостями
func (s *Server) registerNewsTools() {
	// Инструмент для получения новостей за сегодня
	getTodayNewsTool := mcp.NewTool("get_today_news",
		mcp.WithDescription("Получить финансовые новости за сегодня"),
		mcp.WithNumber("limit",
			mcp.Description("Количество новостей (по умолчанию все)"),
		),
	)

	s.server.AddTool(getTodayNewsTool, s.handleGetTodayNews)

	// Инструмент для поиска новостей по ключевому слову
	searchNewsTool := mcp.NewTool("search_news",
		mcp.WithDescription("Поиск новостей по ключевому слову"),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("Ключевое слово для поиска"),
		),
	)

	s.server.AddTool(searchNewsTool, s.handleSearchNews)

	// Инструмент для получения новостей по тикеру
	getNewsByTickerTool := mcp.NewTool("get_news_by_ticker",
		mcp.WithDescription("Получить новости, связанные с указанным тикером"),
		mcp.WithString("ticker",
			mcp.Required(),
			mcp.Description("Тикер акции (например, SBER, GAZP, LKOH)"),
		),
	)

	s.server.AddTool(getNewsByTickerTool, s.handleGetNewsByTicker)
}

// registerPrompts регистрирует шаблоны в MCP сервере
func (s *Server) registerPrompts() {
	// Шаблон для анализа акции
	stockAnalysisPrompt := mcp.NewPrompt("stock_analysis",
		mcp.WithPromptDescription("Анализ котировок акции"),
		mcp.WithArgument("ticker",
			mcp.ArgumentDescription("Тикер акции для анализа"),
			mcp.RequiredArgument(),
		),
	)

	s.server.AddPrompt(stockAnalysisPrompt, s.handleStockAnalysisPrompt)

	// Шаблон для обзора рынка
	marketOverviewPrompt := mcp.NewPrompt("market_overview",
		mcp.WithPromptDescription("Общий обзор состояния рынка"),
	)

	s.server.AddPrompt(marketOverviewPrompt, s.handleMarketOverviewPrompt)

	// Шаблон для анализа новостей
	newsAnalysisPrompt := mcp.NewPrompt("news_analysis",
		mcp.WithPromptDescription("Анализ финансовых новостей за сегодня"),
	)

	s.server.AddPrompt(newsAnalysisPrompt, s.handleNewsAnalysisPrompt)
}

// Обработчики инструментов для акций

// handleGetStockInfo обрабатывает запрос на получение информации об акции
func (s *Server) handleGetStockInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ticker, ok := request.Params.Arguments["ticker"].(string)
	if !ok {
		return mcp.NewToolResultError("параметр ticker должен быть строкой"), nil
	}

	stock, err := s.stockService.GetStockInfo(ctx, ticker)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("не удалось получить информацию об акции: %v", err)), nil
	}

	if stock == nil {
		return mcp.NewToolResultError(fmt.Sprintf("акция с тикером %s не найдена", ticker)), nil
	}

	// Формируем результат
	result := fmt.Sprintf(`Информация об акции %s (%s):
Цена: %.2f ₽
Изменение: %.2f (%.2f%%)
Объем торгов: %d
Дата обновления: %s`,
		stock.Ticker, stock.Name,
		stock.Price,
		stock.Change, stock.ChangePerc,
		stock.Volume,
		stock.UpdatedAt.Format("2006-01-02 15:04:05"),
	)

	return mcp.NewToolResultText(result), nil
}

// handleGetTopGainers обрабатывает запрос на получение топ растущих акций
func (s *Server) handleGetTopGainers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 10 // Значение по умолчанию
	if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
		limit = int(limitVal)
	}

	stocks, err := s.stockService.GetMOEXTopGainers(ctx, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("не удалось получить список растущих акций: %v", err)), nil
	}

	if len(stocks) == 0 {
		return mcp.NewToolResultText("Не найдено растущих акций"), nil
	}

	// Формируем результат
	result := fmt.Sprintf("Топ %d растущих акций на MOEX:\n\n", len(stocks))
	for i, stock := range stocks {
		result += fmt.Sprintf("%d. %s (%s): %.2f ₽ (%.2f%%)\n",
			i+1, stock.Ticker, stock.Name, stock.Price, stock.ChangePerc)
	}

	return mcp.NewToolResultText(result), nil
}

// handleGetTopLosers обрабатывает запрос на получение топ падающих акций
func (s *Server) handleGetTopLosers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 10 // Значение по умолчанию
	if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
		limit = int(limitVal)
	}

	stocks, err := s.stockService.GetMOEXTopLosers(ctx, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("не удалось получить список падающих акций: %v", err)), nil
	}

	if len(stocks) == 0 {
		return mcp.NewToolResultText("Не найдено падающих акций"), nil
	}

	// Формируем результат
	result := fmt.Sprintf("Топ %d падающих акций на MOEX:\n\n", len(stocks))
	for i, stock := range stocks {
		result += fmt.Sprintf("%d. %s (%s): %.2f ₽ (%.2f%%)\n",
			i+1, stock.Ticker, stock.Name, stock.Price, stock.ChangePerc)
	}

	return mcp.NewToolResultText(result), nil
}

// handleSearchStocks обрабатывает запрос на поиск акций
func (s *Server) handleSearchStocks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return mcp.NewToolResultError("параметр query должен быть строкой"), nil
	}

	stocks, err := s.stockService.SearchStocks(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("не удалось выполнить поиск акций: %v", err)), nil
	}

	if len(stocks) == 0 {
		return mcp.NewToolResultText("По запросу не найдено акций"), nil
	}

	// Формируем результат
	result := fmt.Sprintf("Результаты поиска по запросу '%s':\n\n", query)
	for i, stock := range stocks {
		result += fmt.Sprintf("%d. %s (%s): %.2f ₽ (%.2f%%)\n",
			i+1, stock.Ticker, stock.Name, stock.Price, stock.ChangePerc)
	}

	return mcp.NewToolResultText(result), nil
}

// Обработчики инструментов для новостей

// handleGetTodayNews обрабатывает запрос на получение новостей за сегодня
func (s *Server) handleGetTodayNews(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 0 // 0 означает все новости
	if limitVal, ok := request.Params.Arguments["limit"].(float64); ok {
		limit = int(limitVal)
	}

	news, err := s.newsService.GetTodayNews(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("не удалось получить новости: %v", err)), nil
	}

	if len(news) == 0 {
		return mcp.NewToolResultText("На сегодня нет финансовых новостей"), nil
	}

	// Применяем лимит, если он задан
	if limit > 0 && limit < len(news) {
		news = news[:limit]
	}

	// Формируем результат
	result := fmt.Sprintf("Финансовые новости за %s:\n\n", time.Now().Format("02.01.2006"))
	for i, item := range news {
		result += fmt.Sprintf("%d. %s\n", i+1, item.Title)
		result += fmt.Sprintf("   %s\n", item.Description)
		result += fmt.Sprintf("   Источник: %s\n", item.Source)
		result += fmt.Sprintf("   Опубликовано: %s\n", item.PublishedAt.Format("15:04"))
		result += fmt.Sprintf("   URL: %s\n\n", item.URL)
	}

	return mcp.NewToolResultText(result), nil
}

// handleSearchNews обрабатывает запрос на поиск новостей по ключевому слову
func (s *Server) handleSearchNews(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, ok := request.Params.Arguments["keyword"].(string)
	if !ok {
		return mcp.NewToolResultError("параметр keyword должен быть строкой"), nil
	}

	news, err := s.newsService.SearchNewsByKeyword(ctx, keyword)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("не удалось выполнить поиск новостей: %v", err)), nil
	}

	if len(news) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("По запросу '%s' не найдено новостей", keyword)), nil
	}

	// Формируем результат
	result := fmt.Sprintf("Результаты поиска новостей по запросу '%s':\n\n", keyword)
	for i, item := range news {
		result += fmt.Sprintf("%d. %s\n", i+1, item.Title)
		result += fmt.Sprintf("   %s\n", item.Description)
		result += fmt.Sprintf("   Источник: %s\n", item.Source)
		result += fmt.Sprintf("   Опубликовано: %s\n", item.PublishedAt.Format("02.01.2006 15:04"))
		result += fmt.Sprintf("   URL: %s\n\n", item.URL)
	}

	return mcp.NewToolResultText(result), nil
}

// handleGetNewsByTicker обрабатывает запрос на получение новостей по тикеру
func (s *Server) handleGetNewsByTicker(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ticker, ok := request.Params.Arguments["ticker"].(string)
	if !ok {
		return mcp.NewToolResultError("параметр ticker должен быть строкой"), nil
	}

	news, err := s.newsService.GetNewsForTicker(ctx, ticker)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("не удалось получить новости: %v", err)), nil
	}

	if len(news) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("Не найдено новостей, связанных с акцией %s", ticker)), nil
	}

	// Формируем результат
	result := fmt.Sprintf("Новости, связанные с акцией %s:\n\n", ticker)
	for i, item := range news {
		result += fmt.Sprintf("%d. %s\n", i+1, item.Title)
		result += fmt.Sprintf("   %s\n", item.Description)
		result += fmt.Sprintf("   Источник: %s\n", item.Source)
		result += fmt.Sprintf("   Опубликовано: %s\n", item.PublishedAt.Format("02.01.2006 15:04"))
		result += fmt.Sprintf("   URL: %s\n\n", item.URL)
	}

	return mcp.NewToolResultText(result), nil
}

// Обработчики шаблонов

// handleStockAnalysisPrompt обрабатывает запрос на шаблон анализа акции
func (s *Server) handleStockAnalysisPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	ticker, ok := request.Params.Arguments["ticker"]
	if !ok || ticker == "" {
		return nil, fmt.Errorf("требуется параметр ticker")
	}

	// Получаем информацию об акции
	stock, err := s.stockService.GetStockInfo(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить информацию об акции: %w", err)
	}

	// Получаем связанные новости
	news, err := s.newsService.GetNewsForTicker(ctx, ticker)
	if err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось получить новости для акции %s: %v", ticker, err)
		news = []models.News{} // Пустой список, если не удалось получить новости
	}

	// Формируем системное сообщение
	systemMessage := fmt.Sprintf(`Ты - финансовый аналитик, специализирующийся на российском рынке акций. 
Проанализируй акцию %s (%s) на основе предоставленных данных.
Текущая цена: %.2f ₽
Изменение: %.2f ₽ (%.2f%%)
Объем торгов: %d
Дата обновления: %s

Предоставь комплексный анализ акции, включая:
1. Текущее состояние и динамику цены
2. Технический анализ (если возможно)
3. Новостной фон (по предоставленным новостям)
4. Перспективы и возможные сценарии развития`,
		stock.Ticker, stock.Name,
		stock.Price,
		stock.Change, stock.ChangePerc,
		stock.Volume,
		stock.UpdatedAt.Format("2006-01-02 15:04:05"),
	)

	// Формируем контент с новостями
	newsContent := fmt.Sprintf("Связанные новости для акции %s (%s):\n\n", stock.Ticker, stock.Name)
	if len(news) > 0 {
		for i, item := range news {
			newsContent += fmt.Sprintf("%d. %s\n", i+1, item.Title)
			newsContent += fmt.Sprintf("   %s\n", item.Description)
			newsContent += fmt.Sprintf("   Источник: %s, Дата: %s\n\n", item.Source, item.PublishedAt.Format("02.01.2006"))
		}
	} else {
		newsContent += "Новости не найдены.\n"
	}

	return mcp.NewGetPromptResult(
		fmt.Sprintf("Анализ акции %s", ticker),
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleAssistant,
				mcp.NewTextContent(systemMessage),
			),
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(newsContent),
			),
		},
	), nil
}

// handleMarketOverviewPrompt обрабатывает запрос на шаблон обзора рынка
func (s *Server) handleMarketOverviewPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Получаем топ растущих акций
	topGainers, err := s.stockService.GetMOEXTopGainers(ctx, 5)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список растущих акций: %w", err)
	}

	// Получаем топ падающих акций
	topLosers, err := s.stockService.GetMOEXTopLosers(ctx, 5)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список падающих акций: %w", err)
	}

	// Получаем новости за сегодня
	todayNews, err := s.newsService.GetTodayNews(ctx)
	if err != nil {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: не удалось получить новости: %v", err)
		todayNews = []models.News{} // Пустой список, если не удалось получить новости
	}

	// Ограничиваем количество новостей для обзора
	newsLimit := 10
	if len(todayNews) > newsLimit {
		todayNews = todayNews[:newsLimit]
	}

	// Формируем системное сообщение
	systemMessage := `Ты - опытный финансовый аналитик, специализирующийся на российском рынке акций.
Подготовь краткий обзор состояния рынка на сегодня, используя предоставленные данные.
Включи в обзор:
1. Общую оценку настроения рынка
2. Анализ лидеров роста и падения
3. Обзор ключевых новостей и их влияние на рынок
4. Краткий прогноз на ближайшую перспективу`

	// Формируем контент с данными о рынке
	marketContent := "Данные о российском рынке акций (MOEX) на сегодня:\n\n"

	// Добавляем информацию о топ растущих акциях
	marketContent += "Лидеры роста:\n"
	for i, stock := range topGainers {
		marketContent += fmt.Sprintf("%d. %s (%s): %.2f ₽ (%.2f%%)\n",
			i+1, stock.Ticker, stock.Name, stock.Price, stock.ChangePerc)
	}
	marketContent += "\n"

	// Добавляем информацию о топ падающих акциях
	marketContent += "Лидеры падения:\n"
	for i, stock := range topLosers {
		marketContent += fmt.Sprintf("%d. %s (%s): %.2f ₽ (%.2f%%)\n",
			i+1, stock.Ticker, stock.Name, stock.Price, stock.ChangePerc)
	}
	marketContent += "\n"

	// Добавляем информацию о ключевых новостях
	marketContent += "Ключевые новости за сегодня:\n"
	if len(todayNews) > 0 {
		for i, item := range todayNews {
			marketContent += fmt.Sprintf("%d. %s\n", i+1, item.Title)
			marketContent += fmt.Sprintf("   %s\n", item.Description)
			marketContent += fmt.Sprintf("   Источник: %s\n\n", item.Source)
		}
	} else {
		marketContent += "Нет доступных новостей на сегодня.\n"
	}

	return mcp.NewGetPromptResult(
		"Обзор рынка",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleAssistant,
				mcp.NewTextContent(systemMessage),
			),
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(marketContent),
			),
		},
	), nil
}

// handleNewsAnalysisPrompt обрабатывает запрос на шаблон анализа новостей
func (s *Server) handleNewsAnalysisPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Получаем новости за сегодня
	todayNews, err := s.newsService.GetTodayNews(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить новости: %w", err)
	}

	if len(todayNews) == 0 {
		return nil, fmt.Errorf("на сегодня нет доступных финансовых новостей")
	}

	// Формируем системное сообщение
	systemMessage := `Ты - опытный финансовый аналитик.
Проанализируй предоставленные финансовые новости за сегодня и составь краткое резюме.
В своем анализе:
1. Выдели ключевые события и темы
2. Оцени их потенциальное влияние на российский финансовый рынок
3. Отметь, какие компании или секторы экономики могут быть затронуты
4. Предложи возможные торговые идеи на основе новостного фона`

	// Формируем контент с новостями
	newsContent := fmt.Sprintf("Финансовые новости за %s:\n\n", time.Now().Format("02.01.2006"))
	for i, item := range todayNews {
		newsContent += fmt.Sprintf("%d. %s\n", i+1, item.Title)
		newsContent += fmt.Sprintf("   %s\n", item.Description)
		newsContent += fmt.Sprintf("   Источник: %s, Опубликовано: %s\n\n",
			item.Source, item.PublishedAt.Format("15:04"))

		// Добавляем связанные тикеры, если они есть
		if len(item.RelatedTo) > 0 {
			newsContent += fmt.Sprintf("   Связанные компании: %s\n\n",
				formatTickersList(item.RelatedTo))
		}
	}

	return mcp.NewGetPromptResult(
		"Анализ финансовых новостей",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleAssistant,
				mcp.NewTextContent(systemMessage),
			),
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(newsContent),
			),
		},
	), nil
}

// formatTickersList форматирует список тикеров
func formatTickersList(tickers []string) string {
	result := ""
	for i, ticker := range tickers {
		if i > 0 {
			result += ", "
		}
		result += ticker
	}
	return result
}
