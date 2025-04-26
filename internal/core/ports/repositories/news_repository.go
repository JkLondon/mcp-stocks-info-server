package repositories

import (
	"context"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"time"
)

// NewsRepository определяет интерфейс для работы с финансовыми новостями
type NewsRepository interface {
	// GetNews возвращает новость по ID
	GetNews(ctx context.Context, id string) (*models.News, error)

	// GetNewsByDate возвращает новости за указанную дату
	GetNewsByDate(ctx context.Context, date time.Time) ([]models.News, error)

	// GetNewsForToday возвращает новости за сегодня
	GetNewsForToday(ctx context.Context) ([]models.News, error)

	// GetNewsByKeyword возвращает новости по ключевому слову
	GetNewsByKeyword(ctx context.Context, keyword string) ([]models.News, error)

	// GetNewsByTicker возвращает новости, связанные с указанным тикером
	GetNewsByTicker(ctx context.Context, ticker string) ([]models.News, error)

	// SaveNews сохраняет новость
	SaveNews(ctx context.Context, news *models.News) error

	// SaveNewsCollection сохраняет набор новостей
	SaveNewsCollection(ctx context.Context, newsCollection []models.News) error
}
