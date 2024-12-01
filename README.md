# Songs Library API

Этот проект представляет собой реализацию онлайн библиотеки песен с использованием Go.

## Переменные окружения

Перед запуском сервиса, создайте файл `.env` и укажите следующие переменные:

- `ENV`: среда выполнения приложения (`local`, `dev`, `prod`).
- `EXTERNAL_API`: URL внешнего API для получения дополнительной информации о песнях.
- `PAGE_SIZE_LIMIT`: максимальное количество записей на страницу для пагинации.
- `DB_HOST`: хост базы данных PostgreSQL.
- `DB_PORT`: порт базы данных PostgreSQL.
- `DB_USER`: пользователь базы данных PostgreSQL.
- `DB_PASSWORD`: пароль для доступа к базе данных PostgreSQL.
- `DB_NAME`: имя базы данных PostgreSQL.
- `DB_SSLMODE`: режим SSL для подключения к базе данных PostgreSQL.
- `ADDRESS`: адрес, на котором будет запущен сервис (например, `localhost:8080`).
- `USER`: имя пользователя для базовой авторизации.
- `USER_PASSWORD`: пароль для базовой авторизации.
- `TIMEOUT`: тайм-аут для запросов.
- `IDLE_TIMEOUT`: тайм-аут ожидания для неактивных соединений.

### Пример `.env` файла:

```plaintext
ENV=local
EXTERNAL_API=http://localhost:8081
PAGE_SIZE_LIMIT=20

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=12345
DB_NAME=postgres
DB_SSLMODE=disable

ADDRESS=localhost:8080
USER=user
USER_PASSWORD=password
TIMEOUT=5s
IDLE_TIMEOUT=60s
```

### Миграции базы данных:

После того, как переменные окружения указаны, выполните миграции для создания структуры базы данных. Для этого используйте команду:
```sh
go run ./cmd/migrator/main.go -migrations-path=migrations
```

## Запуск сервиса:
После применения миграций, запустите сам сервис:
```sh
go run ./cmd/songs-lib/main.go
```

## API Документация
Полная документация API доступна [здесь](swagger/swagger.yaml).

Эндпоинты
* GET /songs: Получение данных библиотеки с фильтрацией по полям и пагинацией.

* POST /songs: Добавление новой песни.

* GET /songs/{id}: Получение текста песни с пагинацией по куплетам.

* PATCH /songs/{id}: Обновление данных песни.

* DELETE /songs/{id}: Удаление песни.

## Примеры запросов
### Добавление новой песни
```sh
curl -X POST http://localhost:8080/songs -d '{
  "group": "Muse",
  "song": "Supermassive Black Hole"
}'
```

### Получение списка песен с фильтрацией
```sh
curl -X GET "http://localhost:8080/songs?group=Muse&page=1&per_page=10"
```

*todo: запуск через docker-compose, чтобы не приходилось дважды писать go run*