package logger

import (
	"encoding/json"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"log"
	"os"
	"time"
)

/*import "os"

type Storage struct {
	file *os.File
}

func NewStore(filename string) (*Storage, error) {
	// открываем файл для записи
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Storage{file: file}, nil
}

func (s *Storage) Close() error {
	// закрываем файл
	return s.file.Close()
}
*/

type FileStorage2 struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

type FileStorage struct {
	memStore      *repository.MemStore
	storeInterval int
}

func NewFileStorage(store *repository.MemStore, sendInt int) *FileStorage {
	return &FileStorage{memStore: store, storeInterval: sendInt}
}

func (storage *FileStorage) Save(fname string) error {
	// сериализуем структуру в JSON формат
	var writeText []byte
	for key, value := range storage.memStore.GetGauges() {
		send := models.Metrics{
			ID:    key,
			MType: "gauge",
			Value: &value,
		}
		body, err := json.Marshal(send)
		if err != nil {
			log.Printf("Error marshalling json: %s\n", err)
			continue
		}
		writeText = append(writeText, body...)
		writeText = append(writeText, '\n')
	}
	for key, value := range storage.memStore.GetCounters() {
		send := models.Metrics{
			ID:    key,
			MType: "counter",
			Delta: &value,
		}
		body, err := json.Marshal(send)
		if err != nil {
			log.Printf("Error marshalling json: %s\n", err)
			continue
		}
		writeText = append(writeText, body...)
		writeText = append(writeText, '\n')
	}
	/*data, err := json.MarshalIndent(storage.memStore.GetGauges(), "", "   ")
	if err != nil {
		return err
	}*/
	// сохраняем данные в файл
	return os.WriteFile(fname, writeText, 0666)
}

func (storage *FileStorage) Write(fname string) {
	/*settings := logger.FileStorage{
		Port: 4000,
		Host: `localhost`,
	}*/
	log.Printf("Запись в файл 0")
	for {
		log.Printf("Запись в файл 1")
		ticker := time.NewTicker(time.Duration(5) * time.Second)
		for range ticker.C {
			log.Printf("Запись в файл 2")
			if err := storage.Save(fname /*config.FileStorage*/); err != nil {
			}

		}
	}
}
