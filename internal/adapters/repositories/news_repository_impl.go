package repositories

import (
	"context"
	"fmt"
	"github.com/JkLondon/mcp-stocks-info-server/internal/adapters/repositories/apis"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/domain/models"
	"github.com/JkLondon/mcp-stocks-info-server/internal/core/ports/repositories"
	"github.com/JkLondon/mcp-stocks-info-server/pkg/cache"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewsRepositoryImpl реализация интерфейса NewsRepository
type NewsRepositoryImpl struct {
	db          *mongo.Collection
	cache       cache.Cache
	newsAPI     *apis.NewsAPIClient
	cacheExpiry time.Duration
	useCache    bool
}

// NewNewsRepository создает новый экземпляр репозитория для работы с новостями
func NewNewsRepository(
	db *mongo.Database,
	cache cache.Cache,
	newsAPI *apis.NewsAPIClient,
	cacheExpiry time.Duration,
	useCache bool,
) repositories.NewsRepository {
	return &NewsRepositoryImpl{
		db:          db.Collection("news"),
		cache:       cache,
		newsAPI:     newsAPI,
		cacheExpiry: cacheExpiry,
		useCache:    useCache,
	}
}

// GetNews возвращает новость по ID
func (r *NewsRepositoryImpl) GetNews(ctx context.Context, id string) (*models.News, error) {
	cacheKey := fmt.Sprintf("news:%s", id)

	// Проверяем кэш, если включено использование кэша
	if r.useCache {
		var cachedNews models.News
		err := r.cache.Get(ctx, cacheKey, &cachedNews)
		if err == nil && cachedNews.ID != "" {
			return &cachedNews, nil
		}
	}

	// Ищем в базе данных
	var news models.News
	err := r.db.FindOne(ctx, bson.M{"_id": id}).Decode(&news)
	if err == nil {
		// Сохраняем в кэш
		if r.useCache {
			r.cache.Set(ctx, cacheKey, news, r.cacheExpiry)
		}
		return &news, nil
	}

	return nil, fmt.Errorf("новость с ID %s не найдена", id)
}

// GetNewsByDate возвращает новости за указанную дату
func (r *NewsRepositoryImpl) GetNewsByDate(ctx context.Context, date time.Time) ([]models.News, error) {
	// Нормализуем дату, отбрасывая время
	startDate := date.Truncate(24 * time.Hour)
	endDate := startDate.Add(24 * time.Hour)

	cacheKey := fmt.Sprintf("news:date:%s", startDate.Format("2006-01-02"))

	// Проверяем кэш, если включено использование кэша
	if r.useCache {
		var cachedNews []models.News
		err := r.cache.Get(ctx, cacheKey, &cachedNews)
		if err == nil && len(cachedNews) > 0 {
			return cachedNews, nil
		}
	}

	// Ищем в базе данных
	cursor, err := r.db.Find(ctx, bson.M{
		"published_at": bson.M{
			"$gte": startDate,
			"$lt":  endDate,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска в базе данных: %w", err)
	}
	defer cursor.Close(ctx)

	var news []models.News
	if err = cursor.All(ctx, &news); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	// Если нашли новости в базе, возвращаем их
	if len(news) > 0 {
		// Сохраняем в кэш
		if r.useCache {
			r.cache.Set(ctx, cacheKey, news, r.cacheExpiry)
		}
		return news, nil
	}

	// Если не нашли в базе, и сегодняшний день, делаем запрос к NewsAPI
	if startDate.Year() == time.Now().Year() && startDate.Month() == time.Now().Month() && startDate.Day() == time.Now().Day() {
		return r.fetchTodayNewsFromAPI(ctx)
	}

	// Для исторических дат просто возвращаем пустой результат
	return []models.News{}, nil
}

// GetNewsForToday возвращает новости за сегодня
func (r *NewsRepositoryImpl) GetNewsForToday(ctx context.Context) ([]models.News, error) {
	// Используем метод GetNewsByDate с сегодняшней датой
	return r.GetNewsByDate(ctx, time.Now())
}

// GetNewsByKeyword возвращает новости по ключевому слову
func (r *NewsRepositoryImpl) GetNewsByKeyword(ctx context.Context, keyword string) ([]models.News, error) {
	if keyword == "" {
		return nil, fmt.Errorf("ключевое слово не может быть пустым")
	}

	cacheKey := fmt.Sprintf("news:keyword:%s", keyword)

	// Проверяем кэш, если включено использование кэша
	if r.useCache {
		var cachedNews []models.News
		err := r.cache.Get(ctx, cacheKey, &cachedNews)
		if err == nil && len(cachedNews) > 0 {
			return cachedNews, nil
		}
	}

	// Ищем в базе данных
	// Для простоты используем поиск по title и description
	// Для более точного поиска можно использовать полнотекстовый индекс
	cursor, err := r.db.Find(ctx, bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": keyword, "$options": "i"}},
			{"description": bson.M{"$regex": keyword, "$options": "i"}},
			{"content": bson.M{"$regex": keyword, "$options": "i"}},
			{"tags": keyword},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска в базе данных: %w", err)
	}
	defer cursor.Close(ctx)

	var news []models.News
	if err = cursor.All(ctx, &news); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	// Если нашли новости в базе, возвращаем их
	if len(news) > 0 {
		// Сохраняем в кэш
		if r.useCache {
			r.cache.Set(ctx, cacheKey, news, r.cacheExpiry)
		}
		return news, nil
	}

	// Если не нашли в базе, делаем запрос к NewsAPI
	return r.fetchNewsByKeywordFromAPI(ctx, keyword)
}

// GetNewsByTicker возвращает новости, связанные с указанным тикером
func (r *NewsRepositoryImpl) GetNewsByTicker(ctx context.Context, ticker string) ([]models.News, error) {
	if ticker == "" {
		return nil, fmt.Errorf("тикер не может быть пустым")
	}

	cacheKey := fmt.Sprintf("news:ticker:%s", ticker)

	// Проверяем кэш, если включено использование кэша
	if r.useCache {
		var cachedNews []models.News
		err := r.cache.Get(ctx, cacheKey, &cachedNews)
		if err == nil && len(cachedNews) > 0 {
			return cachedNews, nil
		}
	}

	// Ищем в базе данных
	// Используем поле related_to для поиска связанных с тикером новостей
	cursor, err := r.db.Find(ctx, bson.M{
		"$or": []bson.M{
			{"related_to": ticker},
			{"title": bson.M{"$regex": ticker, "$options": "i"}},
			{"description": bson.M{"$regex": ticker, "$options": "i"}},
			{"content": bson.M{"$regex": ticker, "$options": "i"}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска в базе данных: %w", err)
	}
	defer cursor.Close(ctx)

	var news []models.News
	if err = cursor.All(ctx, &news); err != nil {
		return nil, fmt.Errorf("ошибка декодирования результатов: %w", err)
	}

	// Если нашли новости в базе, возвращаем их
	if len(news) > 0 {
		// Сохраняем в кэш
		if r.useCache {
			r.cache.Set(ctx, cacheKey, news, r.cacheExpiry)
		}
		return news, nil
	}

	// Если не нашли в базе, делаем запрос к NewsAPI по ключевому слову (тикеру)
	return r.fetchNewsByKeywordFromAPI(ctx, ticker)
}

// SaveNews сохраняет новость
func (r *NewsRepositoryImpl) SaveNews(ctx context.Context, news *models.News) error {
	if news == nil {
		return fmt.Errorf("новость не может быть nil")
	}

	// Проверяем, существует ли новость с таким ID
	var existingNews models.News
	err := r.db.FindOne(ctx, bson.M{"_id": news.ID}).Decode(&existingNews)
	if err == nil {
		// Обновляем существующую
		_, err = r.db.ReplaceOne(ctx, bson.M{"_id": news.ID}, news)
	} else {
		// Вставляем новую
		_, err = r.db.InsertOne(ctx, news)
	}

	if err != nil {
		return fmt.Errorf("ошибка сохранения в базу данных: %w", err)
	}

	// Обновляем кэш
	if r.useCache {
		cacheKey := fmt.Sprintf("news:%s", news.ID)
		r.cache.Set(ctx, cacheKey, news, r.cacheExpiry)
	}

	return nil
}

// SaveNewsCollection сохраняет набор новостей
func (r *NewsRepositoryImpl) SaveNewsCollection(ctx context.Context, newsCollection []models.News) error {
	for _, news := range newsCollection {
		err := r.SaveNews(ctx, &news)
		if err != nil {
			return err
		}
	}
	return nil
}

// Вспомогательные методы

// fetchTodayNewsFromAPI получает новости за сегодня из NewsAPI
func (r *NewsRepositoryImpl) fetchTodayNewsFromAPI(ctx context.Context) ([]models.News, error) {
	// Делаем запрос к NewsAPI
	news, err := r.newsAPI.GetTodayNews(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных из NewsAPI: %w", err)
	}

	// Сохраняем полученные новости в базу данных
	for i := range news {
		r.SaveNews(ctx, &news[i])
	}

	// Обновляем кэш
	if r.useCache && len(news) > 0 {
		today := time.Now().Format("2006-01-02")
		cacheKey := fmt.Sprintf("news:date:%s", today)
		r.cache.Set(ctx, cacheKey, news, r.cacheExpiry)
	}

	return news, nil
}

// fetchNewsByKeywordFromAPI получает новости по ключевому слову из NewsAPI
func (r *NewsRepositoryImpl) fetchNewsByKeywordFromAPI(ctx context.Context, keyword string) ([]models.News, error) {
	// Делаем запрос к NewsAPI
	news, err := r.newsAPI.GetNewsByKeyword(ctx, keyword)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных из NewsAPI: %w", err)
	}

	// Сохраняем полученные новости в базу данных
	for i := range news {
		r.SaveNews(ctx, &news[i])
	}

	// Обновляем кэш
	if r.useCache && len(news) > 0 {
		cacheKey := fmt.Sprintf("news:keyword:%s", keyword)
		r.cache.Set(ctx, cacheKey, news, r.cacheExpiry)
	}

	return news, nil
}
