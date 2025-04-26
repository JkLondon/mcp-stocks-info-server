package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/JkLondon/mcp-stocks-info-server/internal/config"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"github.com/JkLondon/mcp-stocks-info-server/pkg/cache"
)

// MOEXAPIClient представляет собой клиент для работы с API MOEX
type MOEXAPIClient struct {
	baseURL     string
	httpClient  *http.Client
	cache       cache.Cache
	cacheExpiry time.Duration
	apiKey      string
	useCache    bool
}

// NewMOEXAPIClient создает новый клиент для работы с API MOEX
func NewMOEXAPIClient(cfg *config.Config, cache cache.Cache) *MOEXAPIClient {
	return &MOEXAPIClient{
		baseURL: cfg.MOEX.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.MOEX.Timeout,
		},
		cache:       cache,
		cacheExpiry: cfg.Cache.StocksTTL,
		apiKey:      cfg.MOEX.APIKey,
		useCache:    cfg.MOEX.UseCache,
	}
}

// GetStock получает информацию о котировке акции по тикеру
func (m *MOEXAPIClient) GetStock(ctx context.Context, ticker string) (*models.Stock, error) {
	cacheKey := fmt.Sprintf("moex:stock:%s", ticker)

	if m.useCache {
		var cachedStock models.Stock
		err := m.cache.Get(ctx, cacheKey, &cachedStock)
		if err == nil && cachedStock.Ticker != "" {
			return &cachedStock, nil
		}
	}

	// URL для API MOEX (пример)
	url := fmt.Sprintf("%s/securities/%s.json", m.baseURL, ticker)
	if m.apiKey != "" {
		url += fmt.Sprintf("?apikey=%s", m.apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API MOEX: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Предполагаем, что ответ имеет структуру JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, fmt.Errorf("ошибка при разборе JSON: %w", err)
	}

	// Преобразование данных в модель Stock (зависит от формата ответа MOEX API)
	stock := parseStockFromResponse(responseData, ticker)

	// Сохраняем в кэш
	if m.useCache {
		m.cache.Set(ctx, cacheKey, stock, m.cacheExpiry)
	}

	return stock, nil
}

// GetStocks получает информацию о нескольких акциях
func (m *MOEXAPIClient) GetStocks(ctx context.Context, tickers []string) ([]models.Stock, error) {
	var stocks []models.Stock

	for _, ticker := range tickers {
		stock, err := m.GetStock(ctx, ticker)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения информации о %s: %w", ticker, err)
		}
		stocks = append(stocks, *stock)
	}

	return stocks, nil
}

// GetTopGainers возвращает топ растущих акций
func (m *MOEXAPIClient) GetTopGainers(ctx context.Context, limit int) ([]models.Stock, error) {
	cacheKey := fmt.Sprintf("moex:top_gainers:%d", limit)

	if m.useCache {
		var cachedStocks []models.Stock
		err := m.cache.Get(ctx, cacheKey, &cachedStocks)
		if err == nil && len(cachedStocks) > 0 {
			return cachedStocks, nil
		}
	}

	// URL для API MOEX (пример)
	url := fmt.Sprintf("%s/securities/topgainers.json?limit=%d", m.baseURL, limit)
	if m.apiKey != "" {
		url += fmt.Sprintf("&apikey=%s", m.apiKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API MOEX: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Предполагаем, что ответ имеет структуру JSON с массивом акций
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, fmt.Errorf("ошибка при разборе JSON: %w", err)
	}

	// Парсим данные о топовых акциях (зависит от формата ответа MOEX API)
	stocks := parseStocksFromResponse(responseData)

	// Сохраняем в кэш
	if m.useCache {
		m.cache.Set(ctx, cacheKey, stocks, m.cacheExpiry)
	}

	return stocks, nil
}

// Вспомогательные функции для парсинга ответов API

// parseStockFromResponse преобразует JSON-ответ в модель Stock
func parseStockFromResponse(data map[string]interface{}, ticker string) *models.Stock {
	// Примечание: реальный парсинг зависит от структуры ответа MOEX API
	// Это упрощенный пример

	stock := &models.Stock{
		Ticker:    ticker,
		UpdatedAt: time.Now(),
	}

	// Пытаемся извлечь данные из ответа
	if securities, ok := data["securities"].(map[string]interface{}); ok {
		if data, ok := securities["data"].([]interface{}); ok && len(data) > 0 {
			if stockData, ok := data[0].([]interface{}); ok && len(stockData) > 5 {
				// Предполагаем, что данные имеют определенную структуру
				if name, ok := stockData[2].(string); ok {
					stock.Name = name
				}

				if priceStr, ok := stockData[3].(string); ok {
					if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
						stock.Price = price
					}
				}

				if changeStr, ok := stockData[4].(string); ok {
					if change, err := strconv.ParseFloat(changeStr, 64); err == nil {
						stock.Change = change
					}
				}

				if changePercStr, ok := stockData[5].(string); ok {
					if changePerc, err := strconv.ParseFloat(changePercStr, 64); err == nil {
						stock.ChangePerc = changePerc
					}
				}
			}
		}
	}

	return stock
}

// parseStocksFromResponse преобразует JSON-ответ в слайс моделей Stock
func parseStocksFromResponse(data map[string]interface{}) []models.Stock {
	// Примечание: реальный парсинг зависит от структуры ответа MOEX API
	// Это упрощенный пример

	var stocks []models.Stock

	if securities, ok := data["securities"].(map[string]interface{}); ok {
		if columns, ok := securities["columns"].([]interface{}); ok {
			// Определяем индексы нужных столбцов
			tickerIdx, nameIdx, priceIdx, changeIdx, changePercIdx := -1, -1, -1, -1, -1
			for i, col := range columns {
				colName, ok := col.(string)
				if !ok {
					continue
				}

				switch colName {
				case "SECID":
					tickerIdx = i
				case "SHORTNAME":
					nameIdx = i
				case "LAST":
					priceIdx = i
				case "CHANGE":
					changeIdx = i
				case "LASTTOPREVPRICE":
					changePercIdx = i
				}
			}

			// Парсим данные о акциях
			if data, ok := securities["data"].([]interface{}); ok {
				for _, item := range data {
					stockData, ok := item.([]interface{})
					if !ok || len(stockData) <= max(tickerIdx, nameIdx, priceIdx, changeIdx, changePercIdx) {
						continue
					}

					stock := models.Stock{
						UpdatedAt: time.Now(),
					}

					if ticker, ok := stockData[tickerIdx].(string); ok {
						stock.Ticker = ticker
					}

					if name, ok := stockData[nameIdx].(string); ok {
						stock.Name = name
					}

					if priceVal, ok := stockData[priceIdx].(float64); ok {
						stock.Price = priceVal
					} else if priceStr, ok := stockData[priceIdx].(string); ok {
						if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
							stock.Price = price
						}
					}

					if changeVal, ok := stockData[changeIdx].(float64); ok {
						stock.Change = changeVal
					} else if changeStr, ok := stockData[changeIdx].(string); ok {
						if change, err := strconv.ParseFloat(changeStr, 64); err == nil {
							stock.Change = change
						}
					}

					if changePercVal, ok := stockData[changePercIdx].(float64); ok {
						stock.ChangePerc = changePercVal
					} else if changePercStr, ok := stockData[changePercIdx].(string); ok {
						if changePerc, err := strconv.ParseFloat(changePercStr, 64); err == nil {
							stock.ChangePerc = changePerc
						}
					}

					stocks = append(stocks, stock)
				}
			}
		}
	}

	return stocks
}

// max возвращает максимальное значение из чисел
func max(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}

	max := nums[0]
	for _, num := range nums[1:] {
		if num > max {
			max = num
		}
	}

	return max
}
