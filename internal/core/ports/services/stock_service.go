package services

import (
	"context"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"time"
)

// StockService определяет интерфейс сервиса для работы с акциями
type StockService interface {
	// GetStockInfo возвращает информацию о котировке акции
	GetStockInfo(ctx context.Context, ticker string) (*models.Stock, error)

	// GetMultipleStocks возвращает информацию о нескольких акциях
	GetMultipleStocks(ctx context.Context, tickers []string) ([]models.Stock, error)

	// GetStockQuote возвращает детальные данные по акции за указанную дату
	GetStockQuote(ctx context.Context, ticker string, date time.Time) (*models.StockQuote, error)

	// GetStockHistoricalData возвращает историю котировок акции за период
	GetStockHistoricalData(ctx context.Context, ticker string, startDate, endDate time.Time) ([]models.StockQuote, error)

	// GetMOEXTopGainers возвращает топ растущих акций на MOEX
	GetMOEXTopGainers(ctx context.Context, limit int) ([]models.Stock, error)

	// GetMOEXTopLosers возвращает топ падающих акций на MOEX
	GetMOEXTopLosers(ctx context.Context, limit int) ([]models.Stock, error)

	// GetMOEXTopVolume возвращает акции с наибольшим объемом торгов на MOEX
	GetMOEXTopVolume(ctx context.Context, limit int) ([]models.Stock, error)

	// SearchStocks ищет акции по названию или тикеру
	SearchStocks(ctx context.Context, query string) ([]models.Stock, error)

	// RefreshStockData запускает обновление данных по котировкам
	RefreshStockData(ctx context.Context) error
}
