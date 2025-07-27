package config

type ServerCfg struct {
	AddrServer    string //Адресс:порт сервера
	LogLevel      string //Уровень логирования
	StoreInterval int    //Интеврал записи в файл в секундах
	FileStorage   string //Путь к файлу записи
	Restore       bool   //Признак, надо ли загружать данные из файла, 1 - надо, 0 - нет
}

type ServerOption func(*ServerCfg)

func NewServerCfg(opts ...ServerOption) *ServerCfg {
	cfg := &ServerCfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
