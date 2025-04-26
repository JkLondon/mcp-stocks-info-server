package services

import (
	"context"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"time"
)

// NewsService определяет интерфейс сервиса для работы с финансовыми новостями
type NewsService interface {
	// GetNewsById возвращает новость по ID
	GetNewsById(ctx context.Context, id string) (*models.News, error)

	// GetNewsByDate возвращает новости за указанную дату
	GetNewsByDate(ctx context.Context, date time.Time) ([]models.News, error)

	// GetTodayNews возвращает новости за сегодняшний день
	GetTodayNews(ctx context.Context) ([]models.News, error)

	// GetRecentNews возвращает последние новости
	GetRecentNews(ctx context.Context, limit int) ([]models.News, error)

	// SearchNewsByKeyword ищет новости по ключевому слову
	SearchNewsByKeyword(ctx context.Context, keyword string) ([]models.News, error)

	// GetNewsForTicker возвращает новости, связанные с указанным тикером
	GetNewsForTicker(ctx context.Context, ticker string) ([]models.News, error)

	// GetNewsForMultipleTickers возвращает новости, связанные с несколькими тикерами
	GetNewsForMultipleTickers(ctx context.Context, tickers []string) ([]models.News, error)

	// RefreshNews запускает обновление новостей
	RefreshNews(ctx context.Context) error
}
