# Система аналитики личных трат (CashApp)

Микросервисное приложение для учёта и анализа личных финансов.  
Система позволяет фиксировать личные траты, управлять бюджетами, получать аналитические отчёты и использовать **Google Таблицы как основной пользовательский интерфейс** через HTTP API.

Проект реализован в рамках итоговой работы и соответствует требованиям по gRPC, чистой архитектуре, JWT-безопасности и контейнеризации.

---

### Назначение проекта

Цель проекта — предоставить простой и удобный способ:
- учитывать личные расходы;
- контролировать бюджеты по категориям;
- анализировать траты за выбранный период;
- использовать Google Sheets в качестве UI без отдельного фронтенда.

Пользователь взаимодействует с системой через Google Таблицы, а все операции выполняются через публичный HTTP API.

---

### Архитектурная идея

Проект построен как микросервисное приложение:

- **Gateway** — HTTP-шлюз:
    - принимает запросы от клиентов и Google Таблиц;
    - проверяет JWT-токен;
    - преобразует HTTP-запросы в gRPC-вызовы;
    - возвращает результат клиенту.

- **Ledger** — основной бизнес-сервис:
    - реализует бизнес-логику;
    - хранит транзакции и бюджеты;
    - проверяет превышение лимитов;
    - формирует отчёты;
    - работает через gRPC.

- **Auth** — сервис аутентификации:
    - регистрация пользователей;
    - логин;
    - генерация JWT access-токенов.

- **Google Таблицы**:
    - используются как UI;
    - через Google Apps Script отправляют запросы в Gateway;
    - получают данные отчётов обратно.

---

### Используемые технологии

- **Go 1.25**
- **gRPC** — взаимодействие микросервисов
- **HTTP / REST** — внешний API
- **PostgreSQL** — основное хранилище данных
- **Redis** — кэширование
- **JWT (JSON Web Token)** — безопасность
- **Docker / Docker Compose** — локальное и публичное развёртывание
- **Goose** — миграции базы данных
- **Google Apps Script** — UI и интеграция с Google Таблицами

---

### Безопасность

- Все защищённые эндпоинты требуют JWT-токен.
- Токен передаётся в заголовке:

---

## Запуск
### Применение миграций
```
make migrate-up
```

### Запуск сервисов
```
make compose-up
```

### Остановка
```
make compose-down
```

### Генерация protobuf
gRPC-контракты описаны в ```.proto``` файле и генерируются для сервисов.
```
make proto
```

---
## Примеры запросов

### Регистрация
Запрос
```
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```
Ответ
```
{
  "user_id": "b2c1a4e2-8f4d-4a8d-9c1b-0a9c6c9a1234"
}
```

### Вход
Запрос
```
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```
Ответ
```
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Бюджеты
Создать / обновить бюджет
```
curl -X POST http://localhost:8080/api/budgets \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "food",
    "limit": 15000
  }'
```
Ответ
```
{
  "category": "food",
  "limit": 15000,
  "period": "fixed"
}
```

### Получить список бюджетов
Запрос
```
curl http://localhost:8080/api/budgets \
  -H "Authorization: Bearer <TOKEN>"
```
Ответ
```
[
  {
    "category": "food",
    "limit": 15000,
    "period": "fixed"
  }
]
```

### Транзакции
Добавить транзакцию
```
curl -X POST http://localhost:8080/api/transactions \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1500,
    "category": "food",
    "description": "Lunch",
    "date": "2025-12-19T12:30:00+03:00"
  }'
```
Ответ
```
{
  "id": 1,
  "amount": 1500,
  "category": "food",
  "description": "Lunch",
  "date": "2025-12-19T12:30:00+03:00"
}
```

### Получить список транзакций
```
curl http://localhost:8080/api/transactions \
  -H "Authorization: Bearer <TOKEN>"
```
Ответ
```
[
  {
    "id": 1,
    "amount": 1500,
    "category": "food",
    "description": "Lunch",
    "date": "2025-12-19T12:30:00+03:00"
  }
]
```

### Отчёты
Сводный отчёт по расходам за период
```
curl "http://localhost:8080/api/reports/summary?from=2025-12-01&to=2025-12-31" \
  -H "Authorization: Bearer <TOKEN>"
```
Ответ
```
{
  "food": 2000,
  "transport": 2000
}
```