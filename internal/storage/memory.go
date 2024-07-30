package storage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/morozoffnor/go-url-shortener/internal/auth"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/chargen"
	"log"
	"sync"
)

type MemoryStorage struct {
	mu   *sync.Mutex
	cfg  *config.Config
	List []*url
}

func NewMemoryStorage(cfg *config.Config) *MemoryStorage {
	u := &MemoryStorage{
		List: []*url{},
		mu:   &sync.Mutex{},
		cfg:  cfg,
	}
	return u
}

func (s *MemoryStorage) AddNewURL(ctx context.Context, full string) (string, error) {
	if len(full) < 1 {
		return "", errors.New("blank URL")
	}
	for _, v := range s.List {
		if v.OriginalURL == full {
			return v.ShortURL, nil
		}
	}
	newURL := &url{
		UUID:        uuid.NewString(),
		ShortURL:    chargen.CreateRandomCharSeq(),
		OriginalURL: full,
		UserID:      ctx.Value(auth.ContextUserID).(uuid.UUID).String(),
		IsDeleted:   false,
	}
	s.mu.Lock()
	s.List = append(s.List, newURL)
	s.mu.Unlock()
	return newURL.ShortURL, nil
}

func (s *MemoryStorage) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
	if len(shortURL) < 1 {
		return "", false, errors.New("no short URL provided")
	}
	for _, v := range s.List {
		if v.ShortURL == shortURL {
			return v.OriginalURL, v.IsDeleted, nil
		}
	}
	return "", false, errors.New("there is no such URL")
}

func (s *MemoryStorage) AddBatch(ctx context.Context, urls []BatchInput) ([]BatchOutput, error) {
	if len(urls) < 1 {
		return []BatchOutput{}, nil
	}
	var result []BatchOutput
	for _, v := range urls {
		shortURL, err2 := s.AddNewURL(ctx, v.OriginalURL)
		if err2 != nil {
			return nil, err2
		}
		result = append(result, BatchOutput{
			ShortURL:      shortURL,
			CorrelationID: v.CorrelationID,
		})
	}
	return result, nil
}

func (s *MemoryStorage) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]UserURLs, error) {
	if len(userID) == 0 {
		return nil, nil
	}
	var result []UserURLs
	for _, v := range s.List {
		if v.UserID == userID.String() {
			var u UserURLs
			u.ShortURL = s.cfg.ResultAddr + "/" + v.ShortURL
			u.OriginalURL = v.OriginalURL

			result = append(result, u)

		}
	}
	return result, nil
}

func (s *MemoryStorage) DeleteURLs(ctx context.Context, userID uuid.UUID, urls URLsForDeletion) {
	input := s.generator(ctx, userID, urls)
	out := s.fanOut(ctx, input)
	in := s.fanIn(ctx, out)
	s.softDeleteURLs(ctx, in)
}

func (s *MemoryStorage) generator(ctx context.Context, userID uuid.UUID, urls URLsForDeletion) chan DeleteURLItem {
	inputCh := make(chan DeleteURLItem)

	// наполняем канал айтемами
	go func() {
		defer close(inputCh)
		for _, v := range urls {
			item := DeleteURLItem{
				UserID:   userID,
				ShortURL: v,
			}
			log.Print("gen", item)
			select {
			case <-ctx.Done():
				return
			case inputCh <- item:
			}
		}
	}()
	return inputCh
}

func (s *MemoryStorage) fanOut(ctx context.Context, inputCh <-chan DeleteURLItem) chan string {
	outCh := make(chan string)

	// распределяем работу: ищем айдишники урлов в базе
	go func() {
		defer close(outCh)

		for item := range inputCh {
			var id string
			for _, v := range s.List {
				if item.ShortURL == v.ShortURL && item.UserID.String() == v.UserID {
					id = v.UUID
				}
			}
			if id == "" {
				continue
			}
			log.Print("fanOut", "sent id")
			select {
			case <-ctx.Done():
				return
			case outCh <- id:
			}
		}
	}()

	return outCh
}

func (s *MemoryStorage) fanIn(ctx context.Context, ids ...chan string) chan string {
	delCh := make(chan string)

	var wg sync.WaitGroup

	// собираем полученные айдишники в один канал
	for _, ch := range ids {
		wg.Add(1)
		log.Print("fanIn", " collected id")
		go func() {
			defer wg.Done()

			for item := range ch {
				select {
				case <-ctx.Done():
					return
				case delCh <- item:
				}
			}
		}()
	}

	// ждём выполнения и закрываем канал
	go func() {
		wg.Wait()
		close(delCh)
	}()

	return delCh
}

func (s *MemoryStorage) softDeleteURLs(ctx context.Context, delCh chan string) {
	var idsForDeletion []string
	for item := range delCh {
		idsForDeletion = append(idsForDeletion, item)
	}

	if len(idsForDeletion) == 0 {
		return
	}

	// лочим мьютекс
	s.mu.Lock()

	// ищем айдишники и "удаляем"
	for _, v := range s.List {
		for _, w := range idsForDeletion {
			if v.UUID == w {
				v.IsDeleted = true
			}
		}
	}

	// убираем лок с мьютекса
	s.mu.Unlock()
}
