package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JkLondon/mcp-stocks-info-server/internal/config"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"github.com/JkLondon/mcp-stocks-info-server/pkg/cache"
)

// NewsAPIClient представляет собой клиент для работы с API новостей
type NewsAPIClient struct {
	baseURL     string
	httpClient  *http.Client
	cache       cache.Cache
	cacheExpiry time.Duration
	apiKey      string
	useCache    bool
	sources     []string
}

// NewNewsAPIClient создает новый клиент для работы с API новостей
func NewNewsAPIClient(cfg *config.Config, cache cache.Cache) *NewsAPIClient {
	return &NewsAPIClient{
		baseURL: cfg.NewsAPI.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.NewsAPI.Timeout,
		},
		cache:       cache,
		cacheExpiry: cfg.Cache.NewsTTL,
		apiKey:      cfg.NewsAPI.APIKey,
		useCache:    cfg.NewsAPI.UseCache,
		sources:     cfg.NewsAPI.Sources,
	}
}

// GetTodayNews получает финансовые новости за сегодняшний день
func (n *NewsAPIClient) GetTodayNews(ctx context.Context) ([]models.News, error) {
	today := time.Now().Format("2006-01-02")
	cacheKey := fmt.Sprintf("news:date:%s", today)

	if n.useCache {
		var cachedNews []models.News
		err := n.cache.Get(ctx, cacheKey, &cachedNews)
		if err == nil && len(cachedNews) > 0 {
			return cachedNews, nil
		}
	}

	// Формируем запрос к API новостей
	apiURL := fmt.Sprintf("%s/everything", n.baseURL)

	// Создаем query-параметры
	params := url.Values{}
	params.Add("q", "финансы OR экономика OR рынок OR биржа OR акции OR MOEX")
	params.Add("from", today)
	params.Add("to", today)
	params.Add("language", "ru")
	params.Add("sortBy", "publishedAt")
	params.Add("apiKey", n.apiKey)

	// Добавляем источники, если они указаны
	if len(n.sources) > 0 {
		params.Add("sources", strings.Join(n.sources, ","))
	}

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	// Выполняем запрос
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API новостей: %s", resp.Status)
	}

	// Читаем и разбираем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var newsResponse struct {
		Status       string `json:"status"`
		TotalResults int    `json:"totalResults"`
		Articles     []struct {
			Source struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"source"`
			Author      string    `json:"author"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			URL         string    `json:"url"`
			URLToImage  string    `json:"urlToImage"`
			PublishedAt time.Time `json:"publishedAt"`
			Content     string    `json:"content"`
		} `json:"articles"`
	}

	if err := json.Unmarshal(body, &newsResponse); err != nil {
		return nil, fmt.Errorf("ошибка при разборе JSON: %w", err)
	}

	// Преобразуем в нашу доменную модель
	var news []models.News
	for _, article := range newsResponse.Articles {
		// Генерируем уникальный ID на основе URL новости
		id := generateNewsID(article.URL)

		// Создаем новость
		newsItem := models.News{
			ID:          id,
			Title:       article.Title,
			Description: article.Description,
			Content:     article.Content,
			URL:         article.URL,
			Source:      article.Source.Name,
			PublishedAt: article.PublishedAt,
			CreatedAt:   time.Now(),
			Tags:        extractTags(article.Title + " " + article.Description),
			RelatedTo:   extractTickers(article.Title + " " + article.Description),
		}

		news = append(news, newsItem)
	}

	// Сохраняем в кэш
	if n.useCache && len(news) > 0 {
		n.cache.Set(ctx, cacheKey, news, n.cacheExpiry)
	}

	return news, nil
}

// GetNewsByKeyword ищет новости по ключевому слову
func (n *NewsAPIClient) GetNewsByKeyword(ctx context.Context, keyword string) ([]models.News, error) {
	if keyword == "" {
		return nil, fmt.Errorf("ключевое слово не может быть пустым")
	}

	cacheKey := fmt.Sprintf("news:keyword:%s", keyword)

	if n.useCache {
		var cachedNews []models.News
		err := n.cache.Get(ctx, cacheKey, &cachedNews)
		if err == nil && len(cachedNews) > 0 {
			return cachedNews, nil
		}
	}

	// Формируем запрос к API новостей
	apiURL := fmt.Sprintf("%s/everything", n.baseURL)

	// Создаем query-параметры
	params := url.Values{}
	params.Add("q", keyword)
	params.Add("language", "ru")
	params.Add("sortBy", "publishedAt")
	params.Add("apiKey", n.apiKey)

	// Добавляем источники, если они указаны
	if len(n.sources) > 0 {
		params.Add("sources", strings.Join(n.sources, ","))
	}

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	// Выполняем запрос
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API новостей: %s", resp.Status)
	}

	// Читаем и разбираем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var newsResponse struct {
		Status       string `json:"status"`
		TotalResults int    `json:"totalResults"`
		Articles     []struct {
			Source struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"source"`
			Author      string    `json:"author"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			URL         string    `json:"url"`
			URLToImage  string    `json:"urlToImage"`
			PublishedAt time.Time `json:"publishedAt"`
			Content     string    `json:"content"`
		} `json:"articles"`
	}

	if err := json.Unmarshal(body, &newsResponse); err != nil {
		return nil, fmt.Errorf("ошибка при разборе JSON: %w", err)
	}

	// Преобразуем в нашу доменную модель
	var news []models.News
	for _, article := range newsResponse.Articles {
		// Генерируем уникальный ID на основе URL новости
		id := generateNewsID(article.URL)

		// Создаем новость
		newsItem := models.News{
			ID:          id,
			Title:       article.Title,
			Description: article.Description,
			Content:     article.Content,
			URL:         article.URL,
			Source:      article.Source.Name,
			PublishedAt: article.PublishedAt,
			CreatedAt:   time.Now(),
			Tags:        extractTags(article.Title + " " + article.Description),
			RelatedTo:   extractTickers(article.Title + " " + article.Description),
		}

		news = append(news, newsItem)
	}

	// Сохраняем в кэш
	if n.useCache && len(news) > 0 {
		n.cache.Set(ctx, cacheKey, news, n.cacheExpiry)
	}

	return news, nil
}

// GetNewsByTicker находит новости, связанные с указанным тикером
func (n *NewsAPIClient) GetNewsByTicker(ctx context.Context, ticker string) ([]models.News, error) {
	if ticker == "" {
		return nil, fmt.Errorf("тикер не может быть пустым")
	}

	cacheKey := fmt.Sprintf("news:ticker:%s", ticker)

	if n.useCache {
		var cachedNews []models.News
		err := n.cache.Get(ctx, cacheKey, &cachedNews)
		if err == nil && len(cachedNews) > 0 {
			return cachedNews, nil
		}
	}

	// Получаем все новости за сегодня
	allNews, err := n.GetTodayNews(ctx)
	if err != nil {
		return nil, err
	}

	// Фильтруем новости, связанные с тикером
	var tickerNews []models.News
	for _, news := range allNews {
		if containsTicker(news, ticker) {
			tickerNews = append(tickerNews, news)
		}
	}

	// Сохраняем в кэш
	if n.useCache && len(tickerNews) > 0 {
		n.cache.Set(ctx, cacheKey, tickerNews, n.cacheExpiry)
	}

	return tickerNews, nil
}

// Вспомогательные функции

// generateNewsID генерирует ID новости на основе URL
func generateNewsID(url string) string {
	// Простой способ - возвращаем последнюю часть URL без расширения
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		lastPart = strings.Split(lastPart, "?")[0] // Убираем параметры
		lastPart = strings.Split(lastPart, "#")[0] // Убираем якорь
		lastPart = strings.Split(lastPart, ".")[0] // Убираем расширение
		return lastPart
	}
	// Если не удалось получить ID из URL, возвращаем timestamp
	return fmt.Sprintf("news_%d", time.Now().Unix())
}

// extractTags извлекает ключевые слова/теги из текста
func extractTags(text string) []string {
	// Простая реализация - ищем ключевые финансовые термины
	keywords := []string{
		"акции", "облигации", "биржа", "MOEX", "инвестиции", "дивиденды",
		"экономика", "финансы", "рынок", "котировки", "банк", "валюта",
	}

	var tags []string
	textLower := strings.ToLower(text)

	for _, keyword := range keywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			tags = append(tags, keyword)
		}
	}

	return tags
}

// extractTickers извлекает тикеры акций из текста
func extractTickers(text string) []string {
	// Список наиболее популярных российских тикеров
	popularTickers := []string{
		"SBER", "GAZP", "LKOH", "GMKN", "ROSN", "NVTK", "TATN",
		"MTSS", "MGNT", "YNDX", "FIVE", "POLY", "ALRS", "VTBR",
	}

	var tickers []string
	textUpper := strings.ToUpper(text)

	for _, ticker := range popularTickers {
		if strings.Contains(textUpper, ticker) {
			tickers = append(tickers, ticker)
		}
	}

	return tickers
}

// containsTicker проверяет, связана ли новость с указанным тикером
func containsTicker(news models.News, ticker string) bool {
	// Проверяем среди связанных тикеров
	for _, t := range news.RelatedTo {
		if strings.EqualFold(t, ticker) {
			return true
		}
	}

	// Проверяем в названии и описании
	tickerUpper := strings.ToUpper(ticker)
	return strings.Contains(strings.ToUpper(news.Title), tickerUpper) ||
		strings.Contains(strings.ToUpper(news.Description), tickerUpper) ||
		strings.Contains(strings.ToUpper(news.Content), tickerUpper)
}
