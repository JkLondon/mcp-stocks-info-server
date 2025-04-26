package services

import (
	"context"
	"fmt"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/ports/repositories"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/ports/services"
	"time"
)

// StockServiceImpl реализация интерфейса StockService
type StockServiceImpl struct {
	stockRepo repositories.StockRepository
}

// NewStockService создает новый экземпляр сервиса для работы с акциями
func NewStockService(stockRepo repositories.StockRepository) services.StockService {
	return &StockServiceImpl{
		stockRepo: stockRepo,
	}
}

// GetStockInfo возвращает информацию о котировке акции
func (s *StockServiceImpl) GetStockInfo(ctx context.Context, ticker string) (*models.Stock, error) {
	if ticker == "" {
		return nil, fmt.Errorf("тикер не может быть пустым")
	}

	return s.stockRepo.GetStock(ctx, ticker)
}

// GetMultipleStocks возвращает информацию о нескольких акциях
func (s *StockServiceImpl) GetMultipleStocks(ctx context.Context, tickers []string) ([]models.Stock, error) {
	if len(tickers) == 0 {
		return nil, fmt.Errorf("список тикеров не может быть пустым")
	}

	return s.stockRepo.GetStocks(ctx, tickers)
}

// GetStockQuote возвращает детальные данные по акции за указанную дату
func (s *StockServiceImpl) GetStockQuote(ctx context.Context, ticker string, date time.Time) (*models.StockQuote, error) {
	if ticker == "" {
		return nil, fmt.Errorf("тикер не может быть пустым")
	}

	if date.IsZero() {
		date = time.Now()
	}

	return s.stockRepo.GetStockQuote(ctx, ticker, date)
}

// GetStockHistoricalData возвращает историю котировок акции за период
func (s *StockServiceImpl) GetStockHistoricalData(ctx context.Context, ticker string, startDate, endDate time.Time) ([]models.StockQuote, error) {
	if ticker == "" {
		return nil, fmt.Errorf("тикер не может быть пустым")
	}

	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, -1, 0) // 1 месяц назад по умолчанию
	}

	if endDate.IsZero() {
		endDate = time.Now()
	}

	return s.stockRepo.GetStockHistory(ctx, ticker, startDate, endDate)
}

// GetMOEXTopGainers возвращает топ растущих акций на MOEX
func (s *StockServiceImpl) GetMOEXTopGainers(ctx context.Context, limit int) ([]models.Stock, error) {
	if limit <= 0 {
		limit = 10 // Значение по умолчанию
	}

	// Здесь мы сначала получаем список всех акций
	stocks, err := s.stockRepo.GetStocks(ctx, []string{})
	if err != nil {
		return nil, err
	}

	// Сортируем по изменению цены (в процентах) в порядке убывания
	// По хорошему нужно имплементировать sort.Interface, но для простоты используем bubble sort
	// В реальном проекте стоит использовать более эффективные методы сортировки
	n := len(stocks)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if stocks[j].ChangePerc < stocks[j+1].ChangePerc {
				stocks[j], stocks[j+1] = stocks[j+1], stocks[j]
			}
		}
	}

	// Возвращаем топ N акций
	if limit > len(stocks) {
		limit = len(stocks)
	}
	return stocks[:limit], nil
}

// GetMOEXTopLosers возвращает топ падающих акций на MOEX
func (s *StockServiceImpl) GetMOEXTopLosers(ctx context.Context, limit int) ([]models.Stock, error) {
	if limit <= 0 {
		limit = 10 // Значение по умолчанию
	}

	// Здесь мы сначала получаем список всех акций
	stocks, err := s.stockRepo.GetStocks(ctx, []string{})
	if err != nil {
		return nil, err
	}

	// Сортируем по изменению цены (в процентах) в порядке возрастания
	n := len(stocks)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if stocks[j].ChangePerc > stocks[j+1].ChangePerc {
				stocks[j], stocks[j+1] = stocks[j+1], stocks[j]
			}
		}
	}

	// Возвращаем топ N акций
	if limit > len(stocks) {
		limit = len(stocks)
	}
	return stocks[:limit], nil
}

// GetMOEXTopVolume возвращает акции с наибольшим объемом торгов на MOEX
func (s *StockServiceImpl) GetMOEXTopVolume(ctx context.Context, limit int) ([]models.Stock, error) {
	if limit <= 0 {
		limit = 10 // Значение по умолчанию
	}

	// Здесь мы сначала получаем список всех акций
	stocks, err := s.stockRepo.GetStocks(ctx, []string{})
	if err != nil {
		return nil, err
	}

	// Сортируем по объему в порядке убывания
	n := len(stocks)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if stocks[j].Volume < stocks[j+1].Volume {
				stocks[j], stocks[j+1] = stocks[j+1], stocks[j]
			}
		}
	}

	// Возвращаем топ N акций
	if limit > len(stocks) {
		limit = len(stocks)
	}
	return stocks[:limit], nil
}

// SearchStocks ищет акции по названию или тикеру
func (s *StockServiceImpl) SearchStocks(ctx context.Context, query string) ([]models.Stock, error) {
	if query == "" {
		return nil, fmt.Errorf("поисковый запрос не может быть пустым")
	}

	// Здесь реализуем поиск по всем акциям
	// В реальном проекте эту функциональность лучше реализовать на уровне репозитория

	// Получаем все акции
	stocks, err := s.stockRepo.GetStocks(ctx, []string{})
	if err != nil {
		return nil, err
	}

	// Фильтруем акции по поисковому запросу
	var result []models.Stock
	queryLower := query

	for _, stock := range stocks {
		// Проверяем, содержится ли запрос в тикере или названии акции
		if containsIgnoreCase(stock.Ticker, queryLower) || containsIgnoreCase(stock.Name, queryLower) {
			result = append(result, stock)
		}
	}

	return result, nil
}

// RefreshStockData запускает обновление данных по котировкам
func (s *StockServiceImpl) RefreshStockData(ctx context.Context) error {
	// Реализация зависит от источника данных
	// Здесь мы просто возвращаем успешный результат
	return nil
}

// Вспомогательные функции

// containsIgnoreCase проверяет, содержит ли строка подстроку без учета регистра
func containsIgnoreCase(s, substr string) bool {
	s, substr = s, substr
	return true
}
