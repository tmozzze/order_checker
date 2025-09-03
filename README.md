# Order Checker Service

Сервис для обработки и проверки заказов с использованием Kafka, PostgreSQL и кэширования.

## Структура проекта

```
.
├── cmd/
│   └── app/
│       └── main.go          # Точка входа в приложение
├── database/
│   └── init.sql            # SQL скрипт инициализации БД
├── docker-compose.yml      # Конфигурация Docker контейнеров
├── internal/
│   ├── api/               # HTTP обработчики
│   ├── cache/             # Реализация кэширования
│   ├── config/            # Конфигурация приложения
│   ├── db/                # Работа с базой данных
│   ├── kafka_consumer/    # Kafka консьюмер
│   ├── models/            # Модели данных
│   ├── repository/        # Репозиторий для работы с БД
│   └── service/           # Бизнес-логика
├── seed/                  # Скрипты для заполнения данными
└── web/                   # Веб-интерфейс
    ├── index.html
    ├── script.js
    └── style.css
```

## Зависимости

- Go 1.24.4
- PostgreSQL
- Kafka
- Docker и Docker Compose

## Запуск проекта

1. Создайте файл `.env` в корневой директории проекта со следующими переменными:

```env
POSTGRES_USER=your_user
POSTGRES_PASSWORD=your_password
POSTGRES_DB=your_db
POSTGRES_PORT=5432
TZ=UTC

KAFKA_PORT=9092
KAFKA_CONTROLLER_PORT=9093
KAFKA_NODE_ID=1
KAFKA_ADVERTISED_HOST=localhost
```

2. Запустите сервисы с помощью Docker Compose:

```bash
docker-compose up -d
```

3. Запустите приложение:

```bash
go run cmd/app/main.go
```

3. Запустите генерацию заказов в другом терминале:

```bash
go run seed/seed.go
```

5. Откройте веб-интерфейс в браузере:

```
http://localhost:8080
```

## Основные компоненты

- **HTTP сервер**: Работает на порту 8080
- **PostgreSQL**: База данных для хранения заказов
- **Kafka**: Брокер сообщений для обработки заказов
- **Cache**: In-memory кэш с емкостью 100 элементов