package models

const (
	TypeSimpleUtterance = "SimpleUtterance"
)

// Структура для приема метрики
type Metrics struct {
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // Параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge
	//Hash  string   `json:"hash,omitempty"`
}

// Структура для отправки
type ResponseBody struct {
	Status  string  `json:"status"` // Успешный или не успешный результат
	ID      string  `json:"id,omitempty"`
	Message string  `json:"message,omitempty"`
	Delta   int64   `json:"delta,omitempty"` // Новое значение метрики в случае передачи counter
	Value   float64 `json:"value,omitempty"` // Новое значение метрики в случае передачи gauge
}

// Request описывает запрос пользователя.
/*type Request struct {
	Request SimpleUtterance `json:"request"`
	Version string          `json:"version"`
}

// SimpleUtterance описывает команду, полученную в запросе типа SimpleUtterance.
type SimpleUtterance struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// Response описывает ответ сервера.
// см. https://yandex.ru/dev/dialogs/alice/doc/response.html
type Response struct {
	Response ResponsePayload `json:"response"`
	Version  string          `json:"version"`
}

// ResponsePayload описывает ответ, который нужно озвучить.
type ResponsePayload struct {
	Text string `json:"text"`
}
*/
