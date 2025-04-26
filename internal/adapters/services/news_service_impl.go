package services

import (
	"context"
	"fmt"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/ports/repositories"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/ports/services"
	"time"
)

// NewsServiceImpl реализация интерфейса NewsService
type NewsServiceImpl struct {
	newsRepo repositories.NewsRepository
}

// NewNewsService создает новый экземпляр сервиса для работы с новостями
func NewNewsService(newsRepo repositories.NewsRepository) services.NewsService {
	return &NewsServiceImpl{
		newsRepo: newsRepo,
	}
}

// GetNewsById возвращает новость по ID
func (s *NewsServiceImpl) GetNewsById(ctx context.Context, id string) (*models.News, error) {
	if id == "" {
		return nil, fmt.Errorf("id новости не может быть пустым")
	}

	return s.newsRepo.GetNews(ctx, id)
}

// GetNewsByDate возвращает новости за указанную дату
func (s *NewsServiceImpl) GetNewsByDate(ctx context.Context, date time.Time) ([]models.News, error) {
	if date.IsZero() {
		date = time.Now()
	}

	return s.newsRepo.GetNewsByDate(ctx, date)
}

// GetTodayNews возвращает новости за сегодняшний день
func (s *NewsServiceImpl) GetTodayNews(ctx context.Context) ([]models.News, error) {
	return s.newsRepo.GetNewsForToday(ctx)
}

// GetRecentNews возвращает последние новости
func (s *NewsServiceImpl) GetRecentNews(ctx context.Context, limit int) ([]models.News, error) {
	if limit <= 0 {
		limit = 10 // Значение по умолчанию
	}

	// Получаем новости за сегодня
	news, err := s.newsRepo.GetNewsForToday(ctx)
	if err != nil {
		return nil, err
	}

	// Сортируем по дате публикации в порядке убывания (от новых к старым)
	// По хорошему нужно имплементировать sort.Interface, но для простоты используем bubble sort
	n := len(news)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if news[j].PublishedAt.Before(news[j+1].PublishedAt) {
				news[j], news[j+1] = news[j+1], news[j]
			}
		}
	}

	// Возвращаем топ N новостей
	if limit > len(news) {
		limit = len(news)
	}
	return news[:limit], nil
}

// SearchNewsByKeyword ищет новости по ключевому слову
func (s *NewsServiceImpl) SearchNewsByKeyword(ctx context.Context, keyword string) ([]models.News, error) {
	if keyword == "" {
		return nil, fmt.Errorf("ключевое слово не может быть пустым")
	}

	return s.newsRepo.GetNewsByKeyword(ctx, keyword)
}

// GetNewsForTicker возвращает новости, связанные с указанным тикером
func (s *NewsServiceImpl) GetNewsForTicker(ctx context.Context, ticker string) ([]models.News, error) {
	if ticker == "" {
		return nil, fmt.Errorf("тикер не может быть пустым")
	}

	return s.newsRepo.GetNewsByTicker(ctx, ticker)
}

// GetNewsForMultipleTickers возвращает новости, связанные с несколькими тикерами
func (s *NewsServiceImpl) GetNewsForMultipleTickers(ctx context.Context, tickers []string) ([]models.News, error) {
	if len(tickers) == 0 {
		return nil, fmt.Errorf("список тикеров не может быть пустым")
	}

	// Получаем все новости за сегодня
	allNews, err := s.newsRepo.GetNewsForToday(ctx)
	if err != nil {
		return nil, err
	}

	// Фильтруем новости, связанные с тикерами
	var result []models.News
	seen := make(map[string]bool) // Для предотвращения дублей

	for _, news := range allNews {
		for _, ticker := range tickers {
			if containsTickerInNews(news, ticker) {
				if !seen[news.ID] {
					result = append(result, news)
					seen[news.ID] = true
				}
				break
			}
		}
	}

	return result, nil
}

// RefreshNews запускает обновление новостей
func (s *NewsServiceImpl) RefreshNews(ctx context.Context) error {
	// Реализация зависит от источника данных
	// Здесь мы просто возвращаем успешный результат
	return nil
}

// Вспомогательные функции

// containsTickerInNews проверяет, содержится ли тикер в новости
func containsTickerInNews(news models.News, ticker string) bool {
	// Проверяем в списке связанных тикеров
	for _, t := range news.RelatedTo {
		if t == ticker {
			return true
		}
	}

	// Проверяем в названии и описании
	return containsIgnoreCase(news.Title, ticker) ||
		containsIgnoreCase(news.Description, ticker) ||
		containsIgnoreCase(news.Content, ticker)
}
