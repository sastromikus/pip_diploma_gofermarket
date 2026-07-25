# Gophermart

Накопительная система лояльности интернет-магазина на Go.

Сервис позволяет пользователям регистрироваться, загружать номера заказов, отслеживать их обработку во внешней системе начислений, получать бонусные баллы и списывать их в счёт новых заказов.

## Основные возможности

- регистрация пользователей;
- вход по логину и паролю;
- хранение паролей в виде bcrypt-хешей;
- аутентификация через подписанную HTTP cookie;
- загрузка номера заказа;
- проверка номера заказа по алгоритму Луна;
- защита от повторной загрузки заказа тем же пользователем;
- запрет привязки одного заказа к разным пользователям;
- получение списка заказов пользователя;
- фоновый опрос внешней системы начислений;
- обработка статусов `NEW`, `PROCESSING`, `INVALID`, `PROCESSED`;
- учёт начисленных бонусных баллов;
- получение текущего и уже использованного баланса;
- списание бонусов в счёт нового заказа;
- история списаний;
- автоматическое применение миграций PostgreSQL;
- штатное завершение HTTP-сервера и фонового worker-процесса.

## Технологии

- Go 1.25;
- `net/http`;
- Chi;
- PostgreSQL;
- `pgx/v5` и `pgxpool`;
- `golang-migrate`;
- bcrypt;
- HMAC-SHA256;
- `slog`;
- GitHub Actions.

## Архитектура

Приложение разделено на несколько уровней:

```text
HTTP request
    ↓
handler / middleware
    ↓
service — бизнес-логика
    ↓
repository — PostgreSQL
```

Взаимодействие с системой начислений выполняется отдельным фоновым компонентом:

```text
PostgreSQL: NEW / PROCESSING orders
    ↓
AccrualWorker
    ↓
External Accrual HTTP API
    ↓
Update order status and accrual
```

## Структура проекта

```text
cmd/gophermart/             точка входа основного HTTP-сервера
cmd/accrual/                бинарные файлы учебной системы начислений
internal/accrual/           HTTP-клиент системы начислений
internal/auth/              подпись и проверка аутентификационной cookie
internal/config/            флаги и переменные окружения
internal/handler/           HTTP handlers, router и middleware
internal/luhn/              проверка номеров заказов по алгоритму Луна
internal/model/             модели пользователей, заказов, баланса и списаний
internal/repository/        PostgreSQL repositories и запуск миграций
internal/service/           бизнес-логика и фоновый AccrualWorker
migrations/                 SQL-миграции PostgreSQL
.github/workflows/          интеграционные и статические проверки
```

## Конфигурация

Настройки задаются через переменные окружения или флаги командной строки.

| Назначение | Флаг | Переменная окружения | Значение по умолчанию |
|---|---|---|---|
| Адрес HTTP-сервера | `-a` | `RUN_ADDRESS` | `localhost:8080` |
| PostgreSQL DSN | `-d` | `DATABASE_URI` | отсутствует |
| Адрес системы начислений | `-r` | `ACCRUAL_SYSTEM_ADDRESS` | отсутствует |
| Секрет подписи cookie | `-s` | `AUTH_SECRET` | `dev-secret` |

Переменные окружения используются как значения по умолчанию для флагов. Переданный флаг имеет приоритет.

`DATABASE_URI` обязателен. Без него приложение завершает работу.

Значение `dev-secret` предназначено только для локальной разработки.

## Требования

- Go 1.25.3 или совместимая версия;
- PostgreSQL;
- запущенная система начислений — для автоматического обновления заказов.

## Сборка

Из корня проекта:

```bash
go build -o gophermart ./cmd/gophermart
```

Windows:

```powershell
go build -o gophermart.exe ./cmd/gophermart
```

## Запуск PostgreSQL

Пример строки подключения:

```text
postgres://postgres:password@127.0.0.1:5432/gophermart?sslmode=disable
```

База данных должна быть создана заранее. Таблицы и индексы приложение создаёт автоматически через миграции при запуске.

## Запуск системы начислений

В архиве проекта находятся готовые бинарные файлы учебной системы начислений для некоторых платформ.

Пример запуска на Linux:

```bash
./cmd/accrual/accrual_linux_amd64 \
  -a=127.0.0.1:8081 \
  -d='postgres://postgres:password@127.0.0.1:5432/gophermart?sslmode=disable'
```

Параметры конкретного бинарного файла можно посмотреть через `-h`.

## Запуск Gophermart

Через флаги:

```bash
./gophermart \
  -a=127.0.0.1:8080 \
  -d='postgres://postgres:password@127.0.0.1:5432/gophermart?sslmode=disable' \
  -r=http://127.0.0.1:8081 \
  -s='replace-with-a-random-secret'
```

Через переменные окружения:

```bash
export RUN_ADDRESS=127.0.0.1:8080
export DATABASE_URI='postgres://postgres:password@127.0.0.1:5432/gophermart?sslmode=disable'
export ACCRUAL_SYSTEM_ADDRESS='http://127.0.0.1:8081'
export AUTH_SECRET='replace-with-a-random-secret'

./gophermart
```

Windows PowerShell:

```powershell
$env:RUN_ADDRESS = "127.0.0.1:8080"
$env:DATABASE_URI = "postgres://postgres:password@127.0.0.1:5432/gophermart?sslmode=disable"
$env:ACCRUAL_SYSTEM_ADDRESS = "http://127.0.0.1:8081"
$env:AUTH_SECRET = "replace-with-a-random-secret"

go run ./cmd/gophermart
```

## HTTP API

| Метод | Путь | Назначение | Авторизация |
|---|---|---|---|
| `POST` | `/api/user/register` | Регистрация | нет |
| `POST` | `/api/user/login` | Вход | нет |
| `POST` | `/api/user/orders` | Загрузка заказа | да |
| `GET` | `/api/user/orders` | Список заказов | да |
| `GET` | `/api/user/balance` | Текущий баланс | да |
| `POST` | `/api/user/balance/withdraw` | Списание баллов | да |
| `GET` | `/api/user/withdrawals` | История списаний | да |

После регистрации или входа сервер устанавливает cookie `user_id`. Для последующих запросов клиент должен передавать эту cookie.

## Регистрация

```bash
curl -i \
  -c cookies.txt \
  -X POST http://127.0.0.1:8080/api/user/register \
  -H 'Content-Type: application/json' \
  -d '{"login":"alex","password":"strong-password"}'
```

Возможные ответы:

- `200 OK` — пользователь зарегистрирован;
- `400 Bad Request` — некорректное тело запроса;
- `409 Conflict` — логин уже занят;
- `500 Internal Server Error` — внутренняя ошибка.

## Вход

```bash
curl -i \
  -c cookies.txt \
  -X POST http://127.0.0.1:8080/api/user/login \
  -H 'Content-Type: application/json' \
  -d '{"login":"alex","password":"strong-password"}'
```

Возможные ответы:

- `200 OK` — вход выполнен;
- `400 Bad Request` — некорректный запрос;
- `401 Unauthorized` — неверный логин или пароль.

## Загрузка заказа

Номер передаётся как `text/plain` и проверяется по алгоритму Луна.

```bash
curl -i \
  -b cookies.txt \
  -X POST http://127.0.0.1:8080/api/user/orders \
  -H 'Content-Type: text/plain' \
  --data '12345678903'
```

Возможные ответы:

- `202 Accepted` — новый заказ принят;
- `200 OK` — заказ уже был загружен этим пользователем;
- `409 Conflict` — заказ принадлежит другому пользователю;
- `422 Unprocessable Entity` — номер не прошёл проверку Луна;
- `401 Unauthorized` — отсутствует или недействительна cookie.

## Получение заказов

```bash
curl -i \
  -b cookies.txt \
  http://127.0.0.1:8080/api/user/orders
```

Пример ответа:

```json
[
  {
    "number": "12345678903",
    "status": "PROCESSED",
    "accrual": 500,
    "uploaded_at": "2026-07-25T12:00:00Z"
  }
]
```

Заказы возвращаются от новых к старым.

Если заказов нет, сервер отвечает `204 No Content`.

## Статусы заказов

| Статус | Значение |
|---|---|
| `NEW` | заказ загружен и ожидает обработки |
| `PROCESSING` | заказ зарегистрирован или обрабатывается системой начислений |
| `INVALID` | система начислений отклонила заказ |
| `PROCESSED` | заказ обработан, начисление сохранено |

## Баланс

```bash
curl -i \
  -b cookies.txt \
  http://127.0.0.1:8080/api/user/balance
```

Пример ответа:

```json
{
  "current": 500,
  "withdrawn": 100
}
```

- `current` — доступный баланс;
- `withdrawn` — сумма всех произведённых списаний.

## Списание бонусов

```bash
curl -i \
  -b cookies.txt \
  -X POST http://127.0.0.1:8080/api/user/balance/withdraw \
  -H 'Content-Type: application/json' \
  -d '{"order":"2377225624","sum":100}'
```

Возможные ответы:

- `200 OK` — списание выполнено;
- `402 Payment Required` — недостаточно баллов;
- `422 Unprocessable Entity` — некорректный номер заказа;
- `400 Bad Request` — некорректная сумма или JSON;
- `401 Unauthorized` — пользователь не авторизован.

Списание выполняется внутри PostgreSQL-транзакции. Строка пользователя блокируется через `SELECT ... FOR UPDATE`, после чего баланс повторно вычисляется и создаётся запись о списании. Это предотвращает одновременное расходование одного и того же баланса несколькими запросами.

## История списаний

```bash
curl -i \
  -b cookies.txt \
  http://127.0.0.1:8080/api/user/withdrawals
```

Пример ответа:

```json
[
  {
    "order": "2377225624",
    "sum": 100,
    "processed_at": "2026-07-25T12:15:00Z"
  }
]
```

Списания возвращаются от новых к старым. Если списаний нет, сервер отвечает `204 No Content`.

## Авторизация

Пароли пользователей хешируются с помощью bcrypt и не сохраняются в открытом виде.

После регистрации или входа сервер формирует токен вида:

```text
<user-id>:<HMAC-SHA256-signature>
```

Токен сохраняется в cookie `user_id` с флагом `HttpOnly`. Middleware проверяет подпись токена и наличие пользователя в базе данных перед вызовом защищённого handler.

Текущая реализация cookie предназначена прежде всего для учебного проекта. Для публичного production-развёртывания дополнительно следует настроить HTTPS и параметры `Secure`, `SameSite` и срок действия cookie.

## Интеграция с системой начислений

Фоновый `AccrualWorker` раз в секунду выбирает до 10 заказов со статусами `NEW` или `PROCESSING`.

Пакет обрабатывается worker pool из четырёх goroutine. Для каждого заказа вызывается:

```http
GET /api/orders/{number}
```

Поддерживается обработка:

- `200 OK` — статус заказа обновляется;
- `204 No Content` — заказ пока отсутствует в системе начислений;
- `429 Too Many Requests` — worker приостанавливает запросы на время из `Retry-After`;
- сетевых и неожиданных HTTP-ошибок.

`Retry-After` поддерживает как количество секунд, так и HTTP-date.

## PostgreSQL и миграции

Для соединения используется concurrency-safe `pgxpool`.

При старте приложение:

1. проверяет наличие `DATABASE_URI`;
2. находит каталог `migrations` относительно рабочей директории или бинарного файла;
3. применяет миграции через `golang-migrate`;
4. создаёт пул PostgreSQL;
5. запускает фонового worker и HTTP-сервер.

Схема содержит таблицы:

- `users`;
- `orders`;
- `withdrawals`.

Основные ограничения базы данных:

- уникальный логин;
- уникальный номер заказа;
- внешний ключ заказа и списания на пользователя;
- допустимые статусы заказа ограничены `CHECK` constraint;
- сумма списания должна быть положительной;
- добавлены индексы для выборки заказов, статусов и истории списаний.

## Graceful shutdown

Приложение обрабатывает `SIGINT` и `SIGTERM`.

При завершении:

- отменяется контекст фонового worker;
- ожидается завершение его goroutine;
- HTTP-сервер получает до пяти секунд на завершение активных запросов;
- закрывается пул PostgreSQL.

## Тестирование

Unit-тесты находятся в пакетах:

- `internal/luhn`;
- `internal/service`.

Запуск всех доступных Go-тестов:

```bash
go test ./...
```

С detector состояния гонки:

```bash
go test -race ./...
```

Статический анализ:

```bash
go vet ./...
```

## Continuous Integration

В репозитории настроены два workflow GitHub Actions.

### Интеграционный автотест

Workflow `.github/workflows/gophermart.yml`:

- запускается на push и pull request;
- поднимает PostgreSQL service container;
- собирает Gophermart;
- запускает учебную систему начислений;
- выполняет официальный интеграционный `gophermarttest`.

### Статическая проверка

Workflow `.github/workflows/statictest.yml` запускает специальный анализатор Яндекс Практикума через `go vet`.

## Ограничения

- HTTP-сервер не настраивает TLS самостоятельно;
- cookie не содержит production-параметров `Secure`, `SameSite` и времени истечения;
- значение `dev-secret` нельзя использовать вне локальной разработки;
- система начислений является внешним компонентом и не входит в исходный код основного сервиса;
- миграции хранятся как внешние файлы и должны присутствовать рядом с проектом или бинарным файлом в поддерживаемой структуре каталогов;
- README описывает текущую реализацию репозитория, а не промышленную систему лояльности.
