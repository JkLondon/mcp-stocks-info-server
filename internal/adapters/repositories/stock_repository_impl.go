package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/JkLondon/mcp-stocks-info-server/internal/adapters/repositories/apis"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/ports/repositories"
	"github.com/JkLondon/mcp-stocks-info-server/pkg/cache"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// StockRepositoryImpl реализация интерфейса StockRepository
type StockRepositoryImpl struct {
	db          *mongo.Collection
	cache       cache.Cache
	moexAPI     *apis.MOEXAPIClient
	cacheExpiry time.Duration
	useCache    bool
}

// NewStockRepository создает новый экземпляр репозитория для работы с акциями
func NewStockRepository(
	db *mongo.Database,
	cache cache.Cache,
	moexAPI *apis.MOEXAPIClient,
	cacheExpiry time.Duration,
	useCache bool,
) repositories.StockRepository {
	return &StockRepositoryImpl{
		db:          db.Collection("stocks"),
		cache:       cache,
		moexAPI:     moexAPI,
		cacheExpiry: cacheExpiry,
		useCache:    useCache,
	}
}

// GetStock возвращает информацию об акции по тикеру
func (r *StockRepositoryImpl) GetStock(ctx context.Context, ticker string) (*models.Stock, error) {
	cacheKey := fmt.Sprintf("stock:%s", ticker)

	// Проверяем кэш, если включено использование кэша
	if r.useCache {
		var cachedStock models.Stock
		err := r.cache.Get(ctx, cacheKey, &cachedStock)
		if err == nil && cachedStock.Ticker != "" {
			return &cachedStock, nil
		}
	}

	// Ищем в базе данных
	var stock models.Stock
	err := r.db.FindOne(ctx, bson.M{"ticker": ticker}).Decode(&stock)
	if err == nil {
		// Сохраняем в кэш
		if r.useCache {
			r.cache.Set(ctx, cacheKey, stock, r.cacheExpiry)
		}
		return &stock, nil
	}

	// Если не нашли в базе, делаем запрос к MOEX API
	stock, err = r.fetchStockFromAPI(ctx, ticker)
	if err != nil {
		return nil, err
	}

	// Сохраняем в базу данных
	_, err = r.db.InsertOne(ctx, stock)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения в базу данных: %w", err)
	}

	// Сохраняем в кэш
	if r.useCache {
		r.cache.Set(ctx, cacheKey, stock, r.cacheExpiry)
	}

	return &stock, nil
}

// GetStocks возвращает список акций по указанным тикерам
func (r *StockRepositoryImpl) GetStocks(ctx context.Context, tickers []string) ([]models.Stock, error) {
	if len(tickers) == 0 {
		// Возвращаем все акции
		return r.getAllStocks(ctx)
	}

	var stocks []models.Stock
	for _, ticker := range tickers {
		stock, err := r.GetStock(ctx, ticker)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения информации о %s: %w", ticker, err)
		}
		stocks = append(stocks, *stock)
	}

	return stocks, nil
}

// GetStockQuote возвращает детальные котировки акции за указанную дату
func (r *StockRepositoryImpl) GetStockQuote(ctx context.Context, ticker string, date time.Time) (*models.StockQuote, error) {
	cacheKey := fmt.Sprintf("stock_quote:%s:%s", ticker, date.Format("2006-01-02"))

	// Проверяем кэш, если включено использование кэша
	if r.useCache {
		var cachedQuote models.StockQuote
		err := r.cache.Get(ctx, cacheKey, &cachedQuote)
		if err == nil && cachedQuote.Ticker != "" {
			return &cachedQuote, nil
		}
	}

	// Ищем в базе данных
	var quote models.StockQuote
	err := r.db.FindOne(ctx, bson.M{
		"ticker": ticker,
		"date": bson.M{
			"$gte": date.Truncate(24 * time.Hour),
			"$lt":  date.Add(24 * time.Hour).Truncate(24 * time.Hour),
		},
	}).Decode(&quote)
	if err == nil {
		// Сохраняем в кэш
		if r.useCache {
			r.cache.Set(ctx, cacheKey, quote, r.cacheExpiry)
		}
		return &quote, nil
	}

	// Если не нашли в базе, делаем запрос к MOEX API
	// Этот метод нужно реализовать в MOEX API клиенте
	// Для примера просто создаем заглушку

	// Заглушка для примера
	quote = models.StockQuote{
		Ticker: ticker,
		Date:   date,
	}

	// Заполняем данными из обычной акции
	stock, err := r.GetStock(ctx, ticker)
	if err != nil {
		return nil, err
	}

	quote.Open = stock.Price - stock.Change
	quote.Close = stock.Price
	quote.High = stock.Price + (stock.Change * 0.1)
	quote.Low = stock.Price - (stock.Change * 0.1)
	quote.Volume = stock.Volume

	// Сохраняем в базу данных
	_, err = r.db.InsertOne(ctx, quote)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения в базу данных: %w", err)
	}

	// Сохраняем в кэш
	if r.useCache {
		r.cache.Set(ctx, cacheKey, quote, r.cacheExpiry)
	}

	return &quote, nil
}

// GetStockHistory возвращает исторические данные по акции за период
func (r *StockRepositoryImpl) GetStockHistory(ctx context.Context, ticker string, startDate, endDate time.Time) ([]models.StockQuote, error) {
	cacheKey := fmt.Sprintf("stock_history:%s:%s:%s", ticker, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Проверяем кэш, если включено использование кэша
	if r.useCache {
		var cachedHistory []models.StockQuote
		err := r.cache.Get(ctx, cacheKey, &cachedHistory)
		if err == nil && len(cachedHistory) > 0 {
			return cachedHistory, nil
		}
	}

	// Ищем в базе данных
	cursor, err := r.db.Find(ctx, bson.M{
		"ticker": ticker,
		"date": bson.M{
			"$gte": startDate.Truncate(24 * time.Hour),
			"$lte": endDate.Add(24 * time.Hour).Truncate(24 * time.Hour),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска в базе данных: %w", err)
	}
	defer cursor.Close(ctx)

	var history []models.StockQuote
	if err = cursor.All(ctx, &history); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	// Если нашли историю в базе, возвращаем ее
	if len(history) > 0 {
		// Сохраняем в кэш
		if r.useCache {
			r.cache.Set(ctx, cacheKey, history, r.cacheExpiry)
		}
		return history, nil
	}

	// Если не нашли в базе, делаем запрос к MOEX API
	// Этот метод нужно реализовать в MOEX API клиенте
	// Для примера просто создаем заглушку

	// Генерируем историю для примера
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		// Пропускаем выходные
		if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
			quote, err := r.GetStockQuote(ctx, ticker, currentDate)
			if err == nil {
				history = append(history, *quote)
			}
		}
		currentDate = currentDate.Add(24 * time.Hour)
	}

	// Сохраняем в кэш
	if r.useCache && len(history) > 0 {
		r.cache.Set(ctx, cacheKey, history, r.cacheExpiry)
	}

	return history, nil
}

// SaveStock сохраняет информацию об акции
func (r *StockRepositoryImpl) SaveStock(ctx context.Context, stock *models.Stock) error {
	if stock == nil {
		return fmt.Errorf("акция не может быть nil")
	}

	// Обновляем время
	stock.UpdatedAt = time.Now()

	// Ищем существующую акцию
	var existingStock models.Stock
	err := r.db.FindOne(ctx, bson.M{"ticker": stock.Ticker}).Decode(&existingStock)
	if err == nil {
		// Обновляем существующую
		_, err = r.db.ReplaceOne(ctx, bson.M{"ticker": stock.Ticker}, stock)
	} else {
		// Вставляем новую
		_, err = r.db.InsertOne(ctx, stock)
	}

	if err != nil {
		return fmt.Errorf("ошибка сохранения в базу данных: %w", err)
	}

	// Обновляем кэш
	if r.useCache {
		cacheKey := fmt.Sprintf("stock:%s", stock.Ticker)
		r.cache.Set(ctx, cacheKey, stock, r.cacheExpiry)
	}

	return nil
}

// SaveStockQuote сохраняет котировки акции
func (r *StockRepositoryImpl) SaveStockQuote(ctx context.Context, quote *models.StockQuote) error {
	if quote == nil {
		return fmt.Errorf("котировка не может быть nil")
	}

	// Ищем существующую котировку
	var existingQuote models.StockQuote
	err := r.db.FindOne(ctx, bson.M{
		"ticker": quote.Ticker,
		"date": bson.M{
			"$gte": quote.Date.Truncate(24 * time.Hour),
			"$lt":  quote.Date.Add(24 * time.Hour).Truncate(24 * time.Hour),
		},
	}).Decode(&existingQuote)
	if err == nil {
		// Обновляем существующую
		_, err = r.db.ReplaceOne(ctx, bson.M{
			"ticker": quote.Ticker,
			"date": bson.M{
				"$gte": quote.Date.Truncate(24 * time.Hour),
				"$lt":  quote.Date.Add(24 * time.Hour).Truncate(24 * time.Hour),
			},
		}, quote)
	} else {
		// Вставляем новую
		_, err = r.db.InsertOne(ctx, quote)
	}

	if err != nil {
		return fmt.Errorf("ошибка сохранения в базу данных: %w", err)
	}

	// Обновляем кэш
	if r.useCache {
		cacheKey := fmt.Sprintf("stock_quote:%s:%s", quote.Ticker, quote.Date.Format("2006-01-02"))
		r.cache.Set(ctx, cacheKey, quote, r.cacheExpiry)
	}

	return nil
}

// SaveStockQuotes сохраняет список котировок акций
func (r *StockRepositoryImpl) SaveStockQuotes(ctx context.Context, quotes []models.StockQuote) error {
	for _, quote := range quotes {
		err := r.SaveStockQuote(ctx, &quote)
		if err != nil {
			return err
		}
	}
	return nil
}

// Вспомогательные методы

// getAllStocks возвращает все акции
func (r *StockRepositoryImpl) getAllStocks(ctx context.Context) ([]models.Stock, error) {
	cacheKey := "all_stocks"

	// Проверяем кэш, если включено использование кэша
	if r.useCache {
		var cachedStocks []models.Stock
		err := r.cache.Get(ctx, cacheKey, &cachedStocks)
		if err == nil && len(cachedStocks) > 0 {
			return cachedStocks, nil
		}
	}

	// Ищем в базе данных
	cursor, err := r.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска в базе данных: %w", err)
	}
	defer cursor.Close(ctx)

	var stocks []models.Stock
	if err = cursor.All(ctx, &stocks); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	// Если нашли акции в базе, возвращаем их
	if len(stocks) > 0 {
		// Сохраняем в кэш
		if r.useCache {
			r.cache.Set(ctx, cacheKey, stocks, r.cacheExpiry)
		}
		return stocks, nil
	}

	// Если не нашли в базе, делаем запрос к MOEX API
	// Этот метод нужно реализовать в MOEX API клиенте
	stocks, err = r.fetchAllStocksFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	// Сохраняем в базу данных
	for _, stock := range stocks {
		_, err = r.db.InsertOne(ctx, stock)
		if err != nil {
			return nil, fmt.Errorf("ошибка сохранения в базу данных: %w", err)
		}
	}

	// Сохраняем в кэш
	if r.useCache && len(stocks) > 0 {
		r.cache.Set(ctx, cacheKey, stocks, r.cacheExpiry)
	}

	return stocks, nil
}

// fetchStockFromAPI получает информацию об акции из MOEX API
func (r *StockRepositoryImpl) fetchStockFromAPI(ctx context.Context, ticker string) (models.Stock, error) {
	// Делаем запрос к MOEX API
	stockPtr, err := r.moexAPI.GetStock(ctx, ticker)
	if err != nil {
		return models.Stock{}, fmt.Errorf("ошибка получения данных из MOEX API: %w", err)
	}

	return *stockPtr, nil
}

// fetchAllStocksFromAPI получает список всех акций из MOEX API
func (r *StockRepositoryImpl) fetchAllStocksFromAPI(ctx context.Context) ([]models.Stock, error) {
	// Список популярных российских тикеров
	defaultTickers := []string{
		"SBER", "GAZP", "LKOH", "GMKN", "ROSN", "NVTK", "TATN",
		"MTSS", "MGNT", "YNDX", "FIVE", "POLY", "ALRS", "VTBR",
	}

	return r.moexAPI.GetStocks(ctx, defaultTickers)
}
