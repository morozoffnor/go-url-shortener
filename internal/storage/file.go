package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/morozoffnor/go-url-shortener/internal/auth"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/chargen"
	"github.com/morozoffnor/go-url-shortener/pkg/logger"
	"log"
	"os"
	"sync"
)

type FileStorage struct {
	mu   *sync.Mutex
	cfg  *config.Config
	List []*url
}

func NewFileStorage(cfg *config.Config) *FileStorage {
	u := &FileStorage{
		List: []*url{},
		mu:   &sync.Mutex{},
		cfg:  cfg,
	}
	err := u.LoadFromFile()
	if err != nil {
		logger.Logger.Error("error loading from file", err)
	}
	return u
}

func (s *FileStorage) AddNewURL(ctx context.Context, full string) (string, error) {
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
	_ = s.SaveToFile(newURL)
	s.mu.Unlock()
	return newURL.ShortURL, nil
}

func (s *FileStorage) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
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

func (s *FileStorage) SaveToFile(URLToSave *url) error {
	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Logger.Error("error opening file "+s.cfg.FileStoragePath, err)
		return err
	}
	data, err := json.MarshalIndent(&URLToSave, "", "    ")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return err
}

func (s *FileStorage) LoadFromFile() error {
	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		logger.Logger.Error("error opening file "+s.cfg.FileStoragePath, err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var u url
		err := decoder.Decode(&u)
		if err != nil {
			return err
		}
		s.List = append(s.List, &u)
	}
	return nil
}

func (s *FileStorage) AddBatch(ctx context.Context, urls []BatchInput) ([]BatchOutput, error) {
	if len(urls) < 1 {
		return []BatchOutput{}, nil
	}
	var result []BatchOutput
	for _, v := range urls {
		shortURL, err := s.AddNewURL(ctx, v.OriginalURL)
		if err != nil {
			return nil, err
		}
		result = append(result, BatchOutput{
			ShortURL:      shortURL,
			CorrelationID: v.CorrelationID,
		})
	}
	return result, nil
}

func (s *FileStorage) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]UserURLs, error) {
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

func (s *FileStorage) DeleteURLs(ctx context.Context, userID uuid.UUID, urls URLsForDeletion) {
	input := s.generator(ctx, userID, urls)
	out := s.fanOut(ctx, input)
	in := s.fanIn(ctx, out)
	s.softDeleteURLs(ctx, in)
}

func (s *FileStorage) generator(ctx context.Context, userID uuid.UUID, urls URLsForDeletion) chan DeleteURLItem {
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

func (s *FileStorage) fanOut(ctx context.Context, inputCh <-chan DeleteURLItem) chan string {
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

func (s *FileStorage) fanIn(ctx context.Context, ids ...chan string) chan string {
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

func (s *FileStorage) softDeleteURLs(ctx context.Context, delCh chan string) {
	var idsForDeletion []string
	for item := range delCh {
		idsForDeletion = append(idsForDeletion, item)
	}

	if len(idsForDeletion) == 0 {
		return
	}

	// забираем себе файл, лочим мьютекс
	s.mu.Lock()
	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_TRUNC|os.O_APPEND, 0666)
	if os.IsNotExist(err) {
		logger.Logger.Error(err)
		return
	}
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	defer file.Close()

	// парсим файл в лист
	var list []*url
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var u url
		err := decoder.Decode(&u)
		if err != nil {
			logger.Logger.Error(err)
			return
		}
		list = append(list, &u)
	}

	// ищем айдишники и "удаляем"
	for _, v := range list {
		for _, w := range idsForDeletion {
			if v.UUID == w {
				v.IsDeleted = true
			}
		}
	}

	// очищаем файл
	err = file.Truncate(0)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	// записываем новые урлы в файл
	for _, v := range list {
		data, err := json.MarshalIndent(&v, "", "    ")
		if err != nil {
			logger.Logger.Error(err)
			return
		}
		_, err = file.Write(data)
		if err != nil {
			logger.Logger.Error(err)
			return
		}
	}

	// убираем лок с мьютекса
	s.mu.Unlock()
}
