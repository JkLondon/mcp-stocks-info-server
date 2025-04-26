package repositories

import (
	"context"
	"time"

	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
)

// StockRepository определяет интерфейс для работы с акциями
type StockRepository interface {
	// GetStock возвращает информацию об акции по тикеру
	GetStock(ctx context.Context, ticker string) (*models.Stock, error)

	// GetStocks возвращает список акций по указанным тикерам
	GetStocks(ctx context.Context, tickers []string) ([]models.Stock, error)

	// GetStockQuote возвращает детальные котировки акции за указанную дату
	GetStockQuote(ctx context.Context, ticker string, date time.Time) (*models.StockQuote, error)

	// GetStockHistory возвращает исторические данные по акции за период
	GetStockHistory(ctx context.Context, ticker string, startDate, endDate time.Time) ([]models.StockQuote, error)

	// SaveStock сохраняет информацию об акции
	SaveStock(ctx context.Context, stock *models.Stock) error

	// SaveStockQuote сохраняет котировки акции
	SaveStockQuote(ctx context.Context, quote *models.StockQuote) error

	// SaveStockQuotes сохраняет список котировок акций
	SaveStockQuotes(ctx context.Context, quotes []models.StockQuote) error
}
