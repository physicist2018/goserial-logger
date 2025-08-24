# data-logger

WEB-приложение для управления логированием данных, поступающих с
последовательного интерфейса. При запуске она подключается к порту и слушает данные,
сохраняя их в БД. Предоставляет пользователю WEB интерфейс для управления логированием.

Программа написана для запуска под управлением ОС Windows 7 32 битной версии,
поэтому максимальная версия go 1.20


## Установка и запуск

Для установки и запуска приложения выполните следующие команды:

```bash
git clone https://github.com/physicist2018/goserial-logger.git
cd goserial-logger
go mod tidy
go build
```

Программа принимает следующие опции:

```bash
Usage: ./data-logger [options]
Options:
  -com string
    	COM port name (default "/dev/ttyUSB0")
  -db string
    	SQLite database file path (default "data/experiments.db")
  -port int
    	Server port number (default 5000)
```

Должна быть запущена на устройстве-носителе, к которому подключена последовательная коммуникация.
