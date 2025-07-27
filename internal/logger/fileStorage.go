package logger

import (
	"bytes"
	"encoding/json"
	models "github.com/PavlovAndre/go-metrics-and-alerting.git/internal/model"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/repository"
	"log"
	"os"
	"time"
)

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
		if len(writeText) > 0 {
			writeText = append(writeText, '\n')
		}
		writeText = append(writeText, body...)
		//writeText = append(writeText, '\n')
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
		if len(writeText) > 0 {
			writeText = append(writeText, '\n')
		}
		writeText = append(writeText, body...)
	}
	// сохраняем данные в файл
	return os.WriteFile(fname, writeText, 0666)
}

func (storage *FileStorage) Write(fname string) {
	log.Printf("Запись в файл 0")
	for {
		log.Printf("Запись в файл 1")
		ticker := time.NewTicker(time.Duration(5) * time.Second)
		for range ticker.C {
			log.Printf("Запись в файл 2")
			if err := storage.Save(fname /*config.FileStorage*/); err != nil {
				log.Printf("Write fileStorage %s", err)
			}

		}
	}
}

func (storage *FileStorage) WriteEnd(fname string) {
	log.Printf("Запись в файл перед завершением")
	if err := storage.Save(fname /*config.FileStorage*/); err != nil {
		log.Printf("Write fileStorage %s", err)
	}

}

func (storage *FileStorage) Read(fname string) {
	log.Printf("Чтение из файла")
	data, err := os.ReadFile(fname)
	if err != nil {
		log.Printf("Read fileStorage %s", err)
		return
	}
	var req models.Metrics
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("Error unmarshalling json: %s\n", err)
			return
		}
		log.Printf("result")

		//Выполняем инкремент значения counter
		if req.MType == "counter" {
			if req.Delta == nil {
				return
			}
			storage.memStore.AddCounter(req.ID, *req.Delta)
			log.Printf("result:")
		}

		if req.MType == "gauge" {
			if req.Value == nil {
				return
			}
			storage.memStore.SetGauge(req.ID, *req.Value)
			log.Printf("result: ")
		}
	}
}
