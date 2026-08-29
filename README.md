# 📘 ЭНЦИКЛОПЕДИЯ И ПОЛНЫЙ АРХИТЕКТУРНЫЙ РАЗБОР ПРОЕКТА: ONLINE SHOP API

> **Цель этого документа:** Дать исчерпывающее, построчное и концептуальное понимание каждого компонента системы **Online Shop**. После изучения этого документа вы будете знать проект лучше, чем его создатели, сможете объяснить каждую строчку кода, каждый SQL-запрос, каждое архитектурное решение и блестяще защитить проект перед любой аудиторией (экзаменаторы, техлиды, архитекторы).

---

# ОГЛАВЛЕНИЕ

1. [Глобальная концепция и предметная область](#1-глобальная-концепция-и-предметная-область)
2. [Архитектурный фундамент (Clean Architecture & SOLID)](#2-архитектурный-фундамент-clean-architecture--solid)
3. [Полный анатомический разбор всех файлов и пакетов](#3-полный-анатомический-разбор-всех-файлов-и-пакетов)
   - [3.1. Точка входа: `main.go`](#31-точка-входа-maingo)
   - [3.2. Конфигурация: `internal/config/config.go`](#32-конфигурация-internalconfigconfiggo)
   - [3.3. Доменные модели: `internal/models/`](#33-доменные-модели-internalmodels)
   - [3.4. Вспомогательные пакеты: `pkg/` (JWT, Hash, Logger, Errors)](#34-вспомогательные-пакеты-pkg-jwt-hash-logger-errors)
   - [3.5. Слой доступа к данным: `internal/repository/` и `internal/repository/postgres/`](#35-слой-доступа-к-данным-internalrepository-и-internalrepositorypostgres)
   - [3.6. Слой бизнес-логики: `internal/service/`](#36-слой-бизнес-логики-internalservice)
   - [3.7. Слой HTTP-обработчиков и Middleware: `handler/` и `handler/middleware/`](#37-слой-http-обработчиков-и-middleware-handler-и-handlermiddleware)
   - [3.8. Миграции базы данных: `migrations/`](#38-миграции-базы-данных-migrations)
   - [3.9. Документация Swagger: `docs/`](#39-документация-swagger-docs)
   - [3.10. Инфраструктура: `Dockerfile` и `docker-compose.yml`](#310-инфраструктура-dockerfile-и-docker-composeyml)
4. [База данных: Схема, Связи, Каскады и Ограничения](#4-база-данных-схема-связи-каскады-и-ограничения)
5. [Безопасность: Аутентификация, Ротация токенов и RBAC](#5-безопасность-аутентификация-ротация-токенов-и-rbac)
6. [Сквозные трассировки ключевых сценариев (End-to-End Tracing)](#6-сквозные-трассировки-ключевых-сценариев-end-to-end-tracing)
7. [Надежность, Отказоустойчивость и Graceful Shutdown](#7-надежность-отказоустойчивость-и-graceful-shutdown)
8. [Тестирование и валидация](#8-тестирование-и-валидация)
9. [Большая шпаргалка для защиты проекта (20+ сложных вопросов с ответами)](#9-большая-шпаргалка-для-защиты-проекта-20-сложных-вопросов-с-ответами)

---

# 1. Глобальная концепция и предметная область

Проект представляет собой **Backend RESTful API для современного маркетплейса (интернет-магазина)**.
В отличие от простых монолитных интернет-магазинов с одним продавцом, данная система поддерживает модель **Multi-Vendor Marketplace** (много продавцов, единая площадка).

### Основные сущности системы:
1. **Пользователи (Users):** Любые участники платформы. Роли:
   - `customer` — покупатель, оформляет заказы.
   - `seller` — продавец, создает магазин и выставляет товары.
   - `admin` — администратор платформы, создает категории, видит любые заказы.
2. **Сессии (Refresh Tokens):** Долгоживущие токены, хранящиеся в базе данных и обеспечивающие безопасную ротацию сессий.
3. **Магазины (Stores):** Виртуальная витрина продавца. Правило бизнес-логики: один продавец владеет одним магазином.
4. **Категории (Categories):** Глобальный классификатор товаров (Электроника, Одежда, Книги и т.д.), управляемый администраторами.
5. **Товары (Products):** Продукция, привязанная к конкретному магазину и категории, имеющая цену и складской остаток (`stock`).
6. **Заказы (Orders & Order Items):** Факт покупки товаров покупателем в конкретном магазине с фиксацией цен на момент сделки и автоматическим уменьшением остатка на складе.

---

# 2. Архитектурный фундамент (Clean Architecture & SOLID)

В основу проекта заложен паттерн **Clean Architecture (Чистая Архитектура)** Роберта Мартина в сочетании с 3-слойным разделением обязанностей (**Presentation -> Domain/Service -> Data Access**).

```
                      ┌─────────────────────────────────┐
                      │    Входящий HTTP-запрос         │
                      └────────────────┬────────────────┘
                                       │
                                       ▼
                      ┌─────────────────────────────────┐
                      │        HTTP MIDDLEWARE          │
                      │  - Logging (Zap)                │
                      │  - Recovery (Panic Trap)        │
                      │  - Auth (JWT Verify)            │
                      │  - RBAC (Role Gatekeeper)       │
                      └────────────────┬────────────────┘
                                       │
                                       ▼
                      ┌─────────────────────────────────┐
                      │        HANDLER LAYER            │
                      │  - router.go                    │
                      │  - user_handler.go              │
                      │  - order_handler.go             │
                      │  - product_handler.go           │
                      │  - store_handler.go             │
                      │  - category_handler.go          │
                      └────────────────┬────────────────┘
                                       │ (Вызов методов интерфейсов)
                                       ▼
                      ┌─────────────────────────────────┐
                      │        SERVICE LAYER            │
                      │  - user_service.go              │
                      │  - order_service.go             │
                      │  - product_service.go           │
                      │  - store_service.go             │
                      │  - category_service.go          │
                      └────────────────┬────────────────┘
                                       │ (Вызов интерфейсов репозиториев)
                                       ▼
                      ┌─────────────────────────────────┐
                      │       REPOSITORY LAYER          │
                      │  - repository.go (Interfaces)   │
                      │  - postgres/*.go (SQL Impl)     │
                      └────────────────┬────────────────┘
                                       │ (TCP / Connection Pool)
                                       ▼
                      ┌─────────────────────────────────┐
                      │      PostgreSQL 16 Database     │
                      └─────────────────────────────────┘
```

### Принципы SOLID на практике в этом коде:
1. **Single Responsibility Principle (SRP):**
   - Хэндлеры занимаются **только** транспортом (парсинг JSON, HTTP-заголовки, статус-коды).
   - Сервисы занимаются **только** бизнес-правилами (валидация длины пароля, проверка прав продавца, списание остатков).
   - Репозитории занимаются **только** общением с PostgreSQL (выполнение SQL, сканирование колонок).
2. **Open/Closed Principle (OCP):**
   - Добавление новой СУБД (например, Redis или ClickHouse) не требует переписывания бизнес-логики сервисов — достаточно реализовать интерфейс из `internal/repository/repository.go`.
3. **Liskov Substitution Principle (LSP):**
   - Любая реализация интерфейса репозитория полностью заменяема без изменения поведения вызывающей стороны (что демонстрируют юнит-тесты с `mock`-репозиториями).
4. **Interface Segregation Principle (ISP):**
   - Вместо одного гигантского интерфейса `Repository` созданы узкоспециализированные: `UserRepository`, `StoreRepository`, `ProductRepository`, `OrderRepository`, `CategoryRepository`, `RefreshTokenRepository`.
5. **Dependency Inversion Principle (DIP):**
   - Верхние слои (сервисы) не зависят от нижних слоев (конкретный `postgres.UserRepo`). Они зависят от **абстракций** (`repository.UserRepository`).
   - Конкретные зависимости передаются через конструкторы (`NewUserService`, `NewOrderService`) при старте приложения в `main.go`.

---

# 3. Полный анатомический разбор всех файлов и пакетов

---

## 3.1. Точка входа: `main.go`

Файл [main.go](file:///c:/Users/user/Desktop/Shop/online.shop/main.go) — это **Composition Root** (корень сборки приложения). Здесь собираются воедино все слои и компоненты.

```
                  ┌────────────────────────────────────────────────┐
                  │                 main.go                        │
                  ├────────────────────────────────────────────────┤
                  │ 1. signal.NotifyContext(SIGINT, SIGTERM)       │
                  │ 2. config.New("config.env")                    │
                  │ 3. logger.New(true)                            │
                  │ 4. jwt.NewService(secret, ttl)                 │
                  │ 5. migrations.Run(migratorDSN)                 │
                  │ 6. pgxpool.New(ctx, dsn)                       │
                  │ 7. New...Repository(pool)                      │
                  │ 8. New...Service(repos, jwt)                   │
                  │ 9. New...Handler(services, logger)             │
                  │ 10. NewRouter(deps)                            │
                  │ 11. http.Server{...}.ListenAndServe()          │
                  │ 12. Graceful Shutdown on ctx.Done()            │
                  └────────────────────────────────────────────────┘
```

### Пошаговый разбор выполнения `main.go`:
1. **Инициализация сигналов:** `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` создает корневой контекст, который автоматически отменяется (`ctx.Done()`), когда операционная система присылает сигнал завершения процесса (например, `Ctrl+C` в терминале или команда `docker stop`).
2. **Загрузка конфигурации:** `config.New("config.env")` парсит переменные окружения.
3. **Логирование:** `logger.New(true)` поднимает быстрый структурированный логгер `Uber Zap`.
4. **JWT-сервис:** Создается с заданным секретным ключом и временем жизни access-токена.
5. **Автомиграции:** `migrations.Run(migratorDSN)` применяет `.sql` файлы к PostgreSQL до старта веб-сервера. Если база еще не создана или таблиц нет, они создаются автоматически.
6. **Пул соединений PostgreSQL:** `pgxpool.New(ctx, dsn)` инициализирует пул TCP-соединений к PostgreSQL. Пул соединений эффективнее одиночных подключений, так как повторно использует сокеты для параллельных HTTP-запросов.
7. **Внедрение зависимостей (Dependency Injection):**
   - Создаются репозитории PostgreSQL: `userRepo`, `refreshTokenRepo`, `storeRepo`, `categoryRepo`, `productRepo`, `orderRepo`.
   - Репозитории передаются в конструкторы сервисов: `NewUserService`, `NewStoreService`, `NewCategoryService`, `NewProductService`, `NewOrderService`.
   - Сервисы передаются в хэндлеры: `NewUserHandler`, `NewStoreHandler`, `NewCategoryHandler`, `NewProductHandler`, `NewOrderHandler`.
8. **Инициализация роутера:** `handler.NewRouter(...)` связывает маршруты с middleware.
9. **Настройка HTTP-сервера:**
   - Задаются `ReadTimeout: 15s` (защита от Slowloris атак, когда клиент бесконечно медленно шлет тело запроса), `WriteTimeout: 15s` и `IdleTimeout: 60s`.
10. **Запуск сервера в горутине:** Сервер слушает входящие соединения асинхронно в `go func()`, чтобы главный поток мог ожидать сигналов ОС.
11. **Graceful Shutdown (Блок `select`):**
    - Если сервер завершился с фатальной ошибкой (например, порт занят), программа логирует ошибку и падает.
    - Когда приходит сигнал `ctx.Done()`, запускается процедура плавной остановки: создается контекст с таймаутом 10 секунд `context.WithTimeout(context.Background(), 10*time.Second)`, вызывается `server.Shutdown(shutdownCtx)`. Сервер завершает все текущие активные HTTP-запросы, закрывает сокет, освобождает пул соединений `pool.Close()` и сбрасывает буфер логов `appLogger.Sync()`.

---

## 3.2. Конфигурация: `internal/config/config.go`

Использует библиотеку `cleanenv` для декларативного считывания переменных из файлов `.env` и переменных окружения ОС.

### Структура `Config`:
```go
type Config struct {
    HTTPPort string   `env:"HTTP_PORT" envDefault:":8080"`
    Postgres Postgres
    JWT      JWT
}
```
- **Принцип работы:** Сначала проверяется наличие файла `config.env`. Если файл существует, значения читаются из него с перезаписью из системных переменных окружения. Если файла нет (например, в продакшене в Kubernetes), параметры берутся напрямую из переменных окружения контейнера.
- Для каждого поля заданы безопасные значения по умолчанию (`envDefault`).

---

## 3.3. Доменные модели: `internal/models/`

В пакете объявлены структуры ядра системы. Каждая структура содержит теги `json:"..."` для сериализации в HTTP-ответах.

1. **[user.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/models/user.go)**:
   - `Role`: кастомный строковый тип с константами `RoleCustomer = "customer"`, `RoleSeller = "seller"`, `RoleAdmin = "admin"`.
   - `PasswordHash string` помечен тегом `json:"-"` — это означает, что хеш пароля **ни при каких обстоятельствах не попадет в JSON-ответ клиенту**, даже если объект `User` будет сериализован целиком.
2. **[store.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/models/store.go)**:
   - Модель магазина: `ID`, `SellerID` (связь с пользователем), `Name`, `Description`, `CreatedAt`.
3. **[category.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/models/category.go)**:
   - Модель категории: `ID`, `Name`.
4. **[product.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/models/product.go)**:
   - Модель товара: `ID`, `StoreID`, `CategoryID`, `Name`, `Description`, `Price` (float64), `Stock` (int), `CreatedAt`, `UpdatedAt`.
5. **[order.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/models/order.go)**:
   - `OrderStatus`: константы `OrderStatusPending = "pending"`, `OrderStatusPaid = "paid"`, `OrderStatusCancelled = "cancelled"`.
   - `OrderItem`: конкретная позиция в чеке (`ProductID`, `StoreID`, `Quantity`, `Price`, вложенная ссылка `*Product`).
   - `Order`: заголовок заказа (`CustomerID`, `StoreID`, `TotalPrice`, `Status`, срез позиций `Items []OrderItem`, `CreatedAt`).
6. **[refresh_token.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/models/refresh_token.go)**:
   - `ID`, `UserID`, `Token` (хэш/строка токена), `ExpiresAt` (время протухания), `CreatedAt`.

---

## 3.4. Вспомогательные пакеты: `pkg/` (JWT, Hash, Logger, Errors)

### `pkg/errors/errors.go` (Sentinel Errors)
В проекте используются типизированные ошибки:
```go
var (
    ErrEmptyOrder        = errors.New("order must contain at least one item")
    ErrInvalidOrder      = errors.New("invalid order data")
    ErrInvalidToken      = errors.New("invalid or expired token")
    ErrOrderNotFound     = errors.New("order not found")
    ErrAccessDenied      = errors.New("access denied")
    ErrUserNotFound      = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
    ErrStoreNotFound     = errors.New("store not found")
    ErrCategoryNotFound  = errors.New("category not found")
    ErrProductNotFound   = errors.New("product not found")
    ErrInsufficientStock = errors.New("insufficient product stock")
    ErrInvalidEmail      = errors.New("invalid email address")
    ErrWeakPassword      = errors.New("password must be at least 6 characters long")
    ErrMultiStoreOrder   = errors.New("all items in an order must belong to the same store")
)
```
**Зачем это нужно:** Сервисы возвращают эти ошибки, а хэндлеры проверяют их через `errors.Is(err, pkgerr.Err...)` и транслируют в соответствующие HTTP статус-коды (`400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `409 Conflict`). Это исключает хрупкое сравнение по тексту строк.

### `pkg/hash/password.go`
- `Generate(password string) (string, error)`: Вызывает `bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)`. Возвращает строку вида `$2a$10$...`, содержащую идентификатор алгоритма, стоимость, соль и хеш.
- `Compare(hashedPassword, password string) bool`: Вызывает `bcrypt.CompareHashAndPassword`. Защищен от атак по времени (Timing Attacks).

### `pkg/jwt/jwt.go`
- `Claims`: Структура с полями `UserID`, `Email`, `Role` и встроенной `jwt.RegisteredClaims` (`IssuedAt`, `ExpiresAt`).
- `GenerateToken(userID, email, role)`: Формирует подписанный JWT с алгоритмом `HS256`.
- `GenerateRefreshToken()`: Генерирует 32 случайных байта через криптостойкий генератор `crypto/rand` и кодирует их в 64-значную hex-строку.
- `ParseToken(tokenString)`: Проверяет подпись с использованием HMAC секрета, валидирует срок действия токена и возвращает `*Claims`.

### `pkg/logger/logger.go`
- Обертка над `go.uber.org/zap`.
- Настроен `ISO8601TimeEncoder` для читаемого формата времени `timestamp`.
- Логи выводятся в структурированном виде (JSON или текст) в `stdout`.

---

## 3.5. Слой доступа к данным: `internal/repository/` и `internal/repository/postgres/`

### Контракты (`internal/repository/repository.go`)
Определяют интерфейсы для каждого домена:
- `UserRepository`: `Create`, `GetByID`, `GetByEmail`.
- `RefreshTokenRepository`: `Create`, `GetByToken`, `DeleteByToken`, `DeleteByUserID`.
- `StoreRepository`: `Create`, `GetByID`, `GetBySellerID`.
- `CategoryRepository`: `Create`, `GetAll`, `GetByID`.
- `ProductRepository`: `Create`, `GetByID`, `GetByStoreID`, `GetAll`, `UpdateStock`.
- `OrderRepository`: `Create`, `GetByID`, `GetByCustomerID`, `GetByStoreID`.

### Реализация PostgreSQL (`internal/repository/postgres/`)

1. **`user_repo.go`**:
   - `Create`: Выполняет `INSERT INTO users (...) VALUES (...) RETURNING id, created_at, updated_at`. Сканирует сгенерированный базой ID обратно в переданную структуру.
   - `GetByEmail`: Выполняет выборку по email. Если запись не найдена, `pgx.ErrNoRows` перехватывается и метод возвращает `(nil, nil)`, что означает отсутствие пользователя без фатальной ошибки БД.
2. **`product_repo.go`**:
   - **Обработка nullable `category_id`:** В таблице `products` колонка `category_id` может быть `NULL`. Поэтому при создании и чтении используется указатель `*int64`. Если `category_id <= 0`, в БД отправляется `NULL`. При сканировании сканируется `catID *int64`, и если он не `nil`, значение присваивается в `p.CategoryID`.
   - **Защита от сбоев итерации:** В методах `GetAll` и `GetByStoreID` срез инициализируется как `make([]*models.Product, 0)` (чтобы в JSON возвращался `[]`, а не `null`), а после цикла чтения строк вызывается `rows.Err()`.
   - **Атомарный метод `UpdateStock`:**
     ```sql
     UPDATE products 
     SET stock = stock + $1, updated_at = NOW() 
     WHERE id = $2 AND stock + $1 >= 0
     ```
     Если `delta` отрицательная (например, `-2`), то условие `stock + $1 >= 0` проверяет, чтобы остаток не ушел в минус. Если условие не выполнено, запрос обновляет 0 строк. Проверка `result.RowsAffected() == 0` сразу возвращает ошибку.
3. **`order_repo.go`**:
   - **Транзакционное сохранение заказа (`Create`):**
     - Начинается транзакция: `tx, err := r.db.Begin(ctx)`.
     - Защита через отложенный откат: `defer tx.Rollback(ctx)` (если `tx.Commit()` не будет вызван из-за ошибки, транзакция гарантированно откатится).
     - Вставляется заголовок в `orders`, получается `order.ID`.
     - В цикле вставляются все позиции в `order_items` с привязкой к `order.ID`.
     - Фиксация транзакции: `tx.Commit(ctx)`.
   - **Жадная загрузка позиций:** В методах `GetByID`, `GetByCustomerID`, `GetByStoreID` для каждого заказа подтягиваются связанные строки из `order_items` через вызов `getOrderItems(ctx, o.ID)`.

---

## 3.6. Слой бизнес-логики: `internal/service/`

Здесь сосредоточены все правила системы. Сервисы не знают про HTTP, контексты Gin/Echo или SQL-драйверы.

1. **[user_service.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/service/user_service.go)**:
   - `Register(ctx, email, password)`:
     - Нормализует email (`strings.ToLower(strings.TrimSpace(email))`).
     - Валидирует email через `net/mail.ParseAddress`.
     - Проверяет длину пароля (`len(password) >= 6`).
     - Проверяет через репозиторий, не занят ли email.
     - Хеширует пароль и создает пользователя.
   - `Login(ctx, email, password)`:
     - Ищет пользователя по email.
     - Проверяет хеш пароля.
     - Генерирует пару: Access Token (JWT) и Refresh Token (случайная строка).
     - Сохраняет Refresh Token в БД с тайм-аутом 30 дней.
   - `RefreshToken(ctx, refreshTokenStr)`:
     - Ищет токен в БД.
     - Проверяет, не истек ли `ExpiresAt`. Если истек — удаляет и возвращает `ErrInvalidToken`.
     - Находит пользователя по `user_id`.
     - **Token Rotation:** Удаляет старый Refresh Token и выпускает новую пару токенов.
   - `Logout(ctx, refreshTokenStr)`:
     - Удаляет Refresh Token из базы.
2. **[store_service.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/service/store_service.go)**:
   - `CreateStore`: Проверяет, что у селлера еще нет магазина (`GetBySellerID`). Создает магазин.
3. **[product_service.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/service/product_service.go)**:
   - `CreateProduct`: Проверяет существование магазина, валидирует `price > 0`, `stock >= 0`. Проверяет, что магазин принадлежит именно этому пользователю (`store.SellerID == userID`), либо пользователь — администратор (`userRole == models.RoleAdmin`).
4. **[order_service.go](file:///c:/Users/user/Desktop/Shop/online.shop/internal/service/order_service.go)**:
   - `Create(ctx, customerID, items)`:
     - **Агрегация дубликатов:** Если клиент передал массив с повторяющимися `product_id`, количества суммируются.
     - **Проверка мульти-магазинов:** Для каждого товара извлекается `product.StoreID`. Если в одном заказе встречаются товары из разных магазинов, создание прерывается с ошибкой `ErrMultiStoreOrder`.
     - **Пошаговое списание стока:** Для каждого товара вызывается `ProductRepo.UpdateStock(ctx, pid, -qty)`.
     - **Компенсационный откат (`rollbackStock`):** Если на 3-м товаре не хватило остатка или транзакция вставки заказа упала, вызывается возврат списанных количеств. Откат выполняется через `context.Background()`, чтобы он не прервался, даже если клиент оборвал сетевое соединение.
   - `GetByID(ctx, orderID, userID, userRole)`: Проверяет права: заказ может открыть покупатель (создатель заказа), продавец (владелец магазина, которому адресован заказ) или администратор.

---

## 3.7. Слой HTTP-обработчиков и Middleware: `handler/` и `handler/middleware/`

### Middleware (`handler/middleware/middleware.go`)

```mermaid
sequenceDiagram
    participant Client as HTTP Клиент
    participant Rec as RecoveryMiddleware
    participant Log as LoggingMiddleware
    participant Auth as AuthMiddleware
    participant RBAC as RequireRole
    participant H as Business Handler

    Client->>Rec: HTTP Запрос
    Rec->>Log: Передача управления
    Log->>Auth: Засекает time.Now()
    Auth->>Auth: Проверка Bearer JWT токена
    alt Токен невалиден
        Auth-->>Client: 401 Unauthorized {"error": "..."}
    else Токен валиден
        Auth->>RBAC: context с UserID, Email, Role
        alt Роль не подходит
            RBAC-->>Client: 403 Forbidden {"error": "..."}
        else Роль подтверждена
            RBAC->>H: ServeHTTP
            alt Возникла паника (panic)
                H-->>Rec: panic()
                Rec-->>Client: 500 Internal Server Error (JSON)
            else Успешное выполнение
                H-->>Log: 200 OK / 201 Created
                Log->>Log: Логирует метод, URL, статус, длительность
                Log-->>Client: Итоговый JSON ответ
            end
        end
    end
```

1. **`RecoveryMiddleware`**: Оборачивает вызовы в `recover()`. Любая непредвиденная паника перехватывается, логируется со стектрейсом, а клиенту возвращается чистый JSON `500 Internal Server Error`.
2. **`LoggingMiddleware`**: Оборачивает `http.ResponseWriter` структурой `responseWriter` для перехвата итогового HTTP статус-кода. Логирует метод, путь, статус ответа и время выполнения в миллисекундах.
3. **`AuthMiddleware`**: Проверяет заголовок `Authorization: Bearer <token>`, валидирует подпись JWT и сохраняет `user_id`, `user_email`, `user_role` в `r.Context()`.
4. **`RequireRole`**: Проверяет наличие роли пользователя среди переданного списка разрешенных ролей (`models.RoleSeller`, `models.RoleAdmin`). Если роль не совпадает — отдает `403 Forbidden`.

### Роутер (`handler/router.go`)
Использует стандартный маршрутизатор Go 1.22+ `http.NewServeMux`, который нативно поддерживает сопоставление методов и параметров пути (например, `GET /products/{id}`).
Все маршруты разделены на **Public** (регистрация, вход, просмотр каталога) и **Protected** (создание товаров, магазинов, заказов).

---

## 3.8. Миграции базы данных: `migrations/`

Файлы:
- `000001_init_schema.up.sql`: Создает все 7 таблиц с индексами и ограничениями.
- `000001_init_schema.down.sql`: Инструкции для полного отката схемы (`DROP TABLE IF EXISTS ... CASCADE`).
- `migrator.go`:
  - Директива `//go:embed *.sql` упаковывает SQL-файлы в бинарник.
  - Драйвер `iofs.New(migrationFiles, ".")` читает миграции из виртуальной файловой системы Go.
  - Метод `m.Up()` накатывает миграции. Если изменений нет, `migrate.ErrNoChange` корректно игнорируется.

---

## 3.9. Документация Swagger: `docs/`

- `swagger.json`: Полная спецификация API в формате OpenAPI 2.0. Описывает пути, входные параметры, структуры тел запросов и ответов, а также схему авторизации Bearer JWT.
- `swagger.go`:
  - `//go:embed swagger.json` эмбеддит спецификацию.
  - `UIHandler`: Отдает HTML-страницу со Swagger UI (подгружает стили и скрипты Swagger UI из CDN).
  - `JSONHandler`: Отдает сам файл `swagger.json` с типом `application/json`.
  - Доступно по адресу: `http://localhost:8080/swagger/`.

---

## 3.10. Инфраструктура: `Dockerfile` и `docker-compose.yml`

### `Dockerfile` (Multi-Stage Build)
1. **Stage 1 (Builder):**
   - Образ: `golang:1.23-alpine`.
   - Копирует `go.mod` и `go.sum`, запускает `go mod download` (слой кэшируется в Docker).
   - Компилирует статически слинкованный бинарник: `CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o shop ./main.go`. Флаги `-w -s` вырезают отладочную информацию DWARF и таблицу символов, уменьшая размер бинарника в 2 раза.
2. **Stage 2 (Final Runner):**
   - Образ: `gcr.io/distroless/static:nonroot`.
   - В итоговом образе **нет Linux shell (bash/sh), нет пакетных менеджеров, нет лишних утилит**. Это делает образ практически неуязвимым для атак с исполнением команд (RCE).
   - Приложение запускается от непривилегированного пользователя `nonroot`.

### `docker-compose.yml`
- Сервис `postgres`: поднимает PostgreSQL 16 на Alpine. Настроен `healthcheck` (`pg_isready -U postgres -d shop_db`).
- Сервис `app`: зависит от `postgres` с условием `condition: service_healthy`. Приложение не начнет запускаться, пока PostgreSQL полностью не инициализирует сокет и не будет готов принимать TCP-соединения.

---

# 4. База данных: Схема, Связи, Каскады и Ограничения

### Карта внешних ключей и ограничений (Foreign Keys & Constraints):

```
┌────────────────────────────────────────────────────────────────────────┐
│                                 USERS                                  │
│ id BIGSERIAL PK, email UNIQUE, password_hash, role                     │
└───────────────┬───────────────────────┬────────────────────────┬───────┘
                │ 1:N                   │ 1:N                    │ 1:N
                │ ON DELETE CASCADE     │ ON DELETE CASCADE      │ ON DELETE RESTRICT
                ▼                       ▼                        ▼
┌────────────────────────┐ ┌────────────────────────┐ ┌──────────────────┐
│     REFRESH_TOKENS     │ │         STORES         │ │      ORDERS      │
│ id, user_id FK, token  │ │ id, seller_id FK, name │ │ id, customer_id  │
└────────────────────────┘ └───────────┬────────────┘ └────────┬─────────┘
                                       │ 1:N                   │ 1:N
                                       │ ON DELETE CASCADE     │ ON DELETE CASCADE
                                       ▼                       ▼
┌────────────────────────┐ ┌────────────────────────┐ ┌──────────────────┐
│       CATEGORIES       │ │        PRODUCTS        │ │   ORDER_ITEMS    │
│ id BIGSERIAL PK, name  │ │ id, store_id FK        │ │ id, order_id FK  │
└───────────────┬────────┘ │ category_id FK (SET NULL)│ product_id FK    │
                │ 1:N      │ CHECK(price >= 0)      │ │ store_id FK      │
                └─────────►│ CHECK(stock >= 0)      │ │ CHECK(qty > 0)   │
                           └────────────────────────┘ └──────────────────┘
```

### Защита целостности данных на уровне SQL:
1. `CHECK (price >= 0)` в таблице `products`: невозможно сохранить отрицательную цену.
2. `CHECK (stock >= 0)` в таблице `products`: невозможно сделать количество товара отрицательным.
3. `CHECK (quantity > 0)` в таблице `order_items`: заказ не может содержать 0 или отрицательное количество товара.
4. `ON DELETE CASCADE` для `stores` и `order_items`: при удалении магазина удаляются его товары; при удалении заказа удаляются его позиции.
5. `ON DELETE SET NULL` для `category_id`: при удалении категории товары не удаляются — их категория становится `NULL`.

---

# 5. Безопасность: Аутентификация, Ротация токенов и RBAC

### Жизненный цикл сессии и Token Rotation

```mermaid
sequenceDiagram
    autonumber
    actor User as Клиент / Фронтенд
    participant API as UserHandler / UserService
    participant DB as PostgreSQL (refresh_tokens)

    Note over User,DB: Фаза 1: Авторизация (Login)
    User->>API: POST /login {email, password}
    API->>API: Проверка bcrypt хеша пароля
    API->>API: Генерация AccessToken (JWT, 24h) + RefreshToken (Random 64 hex, 30d)
    API->>DB: INSERT INTO refresh_tokens (user_id, token, expires_at)
    API-->>User: 200 OK {access_token, refresh_token}

    Note over User,DB: Фаза 2: Выполнение запросов
    User->>API: GET /orders (Header: Bearer AccessToken)
    API->>API: Валидация JWT Claims (UserID, Role)
    API-->>User: 200 OK [Данные заказов]

    Note over User,DB: Фаза 3: Обновление пары токенов (Token Rotation)
    User->>API: POST /refresh {refresh_token: "старый_токен"}
    API->>DB: SELECT * FROM refresh_tokens WHERE token = "старый_токен"
    API->>DB: DELETE FROM refresh_tokens WHERE token = "старый_токен" (Инвалидация)
    API->>API: Генерация нового AccessToken + нового RefreshToken
    API->>DB: INSERT INTO refresh_tokens (новый RefreshToken)
    API-->>User: 200 OK {новый access_token, новый refresh_token}

    Note over User,DB: Фаза 4: Выход из системы (Logout)
    User->>API: POST /logout {refresh_token: "токен"}
    API->>DB: DELETE FROM refresh_tokens WHERE token = "токен"
    API-->>User: 200 OK {"message": "logged out successfully"}
```

---

# 6. Сквозные трассировки ключевых сценариев (End-to-End Tracing)

### Трассировка: Оформление заказа (POST `/orders`)

Рассмотрим, что происходит в коде поминутно при создании заказа:

```
1. HTTP Запрос:
   POST /orders
   Authorization: Bearer eyJhbGciOi...
   Body: {"items": [{"product_id": 1, "quantity": 2}, {"product_id": 1, "quantity": 1}]}

2. AuthMiddleware:
   - Извлекает JWT токен, валидирует HMAC-SHA256 подпись.
   - Извлекает claims.UserID = 42, claims.Role = "customer".
   - Помещает их в r.Context().
   - Передает управление в OrderHandler.CreateOrder.

3. OrderHandler:
   - Декодирует JSON-тело в CreateOrderRequest.
   - Получает customerID = 42 из контекста (не дает подменить ID).
   - Вызывает OrderService.Create(ctx, 42, items).

4. OrderService:
   - Проверяет customerID > 0 и len(items) > 0.
   - Консолидирует позиции: видит две записи с product_id=1, суммирует quantity = 3.
   - Запрашивает ProductRepo.GetByID(ctx, 1). Находит "Ноутбук" (Stock: 10, StoreID: 5, Price: 1000).
   - Проверяет Stock >= 3 (10 >= 3 - ОК).
   - Фиксирует storeID = 5.
   - Вызывает ProductRepo.UpdateStock(ctx, 1, -3).
     Выполняется SQL: UPDATE products SET stock = stock - 3 WHERE id = 1 AND stock - 3 >= 0.
     Строка обновлена, остаток стал 7.
   - Добавляет товар в список deducted (для отката на случай сбоя).
   - Рассчитывает TotalPrice = 3 * 1000 = 3000.
   - Формирует объект Order{CustomerID: 42, StoreID: 5, TotalPrice: 3000, Status: "pending", Items: [...]}.
   - Вызывает OrderRepo.Create(ctx, order).

5. OrderRepository (PostgreSQL):
   - Открывает транзакцию `tx, err := r.db.Begin(ctx)`.
   - Вставляет запись в таблицу `orders` -> получает order.ID = 101.
   - Вставляет запись в таблицу `order_items` (order_id: 101, product_id: 1, store_id: 5, qty: 3, price: 1000).
   - Фиксирует транзакцию `tx.Commit(ctx)`.

6. Ответ клиенту:
   - OrderHandler получает созданный заказ с заполненным ID.
   - Сериализует его в JSON и отдает HTTP статус 201 Created.
```

---

# 7. Надежность, Отказоустойчивость и Graceful Shutdown

### 1. Как предотвращаются падения сервера при Panic?
Если в любом из хэндлеров произойдет обращение по нулевому указателю (`nil pointer dereference`) или паника:
- `RecoveryMiddleware` с помощью функции `recover()` в блоке `defer` перехватывает панику.
- Логгер Zap записывает полный стек вызовов ошибки.
- Клиенту возвращается статус `500 Internal Server Error` с телом `{"error": "internal server error"}`.
- Основной процесс и все параллельные соединения **продолжают работать штатно**.

### 2. Как работает Graceful Shutdown?
Когда процесс останавливается (команда `docker stop` или `Ctrl+C`):
1. Системный сигнал `SIGTERM` перехватывается через `signal.NotifyContext`.
2. Контекст `ctx` переходит в состояние `Done`.
3. Запускается таймер остановки на 10 секунд: `context.WithTimeout(context.Background(), 10*time.Second)`.
4. Вызывается `server.Shutdown(shutdownCtx)`. Сервер прекращает слушать порт `8080` и ждет, пока все текущие запросы завершатся.
5. Закрывается пул соединений `pool.Close()`.
6. Сбрасываются буферы логов `appLogger.Sync()`.
7. Процесс выходит с кодом 0 без зависших транзакций и битых данных.

---

# 8. Тестирование и валидация

В проект включены модульные тесты для всех уровней бизнес-логики:
- **`user_service_test.go`**: Тестирует регистрацию, нормализацию email, запрет слабых паролей, валидацию формата почты, вход и ротацию токенов.
- **`order_service_test.go`**: Тестирует создание заказов, расчет цен, агрегацию дублирующихся позиций, защиту от мульти-магазинов, ошибку нехватки остатка и автоматический компенсационный откат стока при сбое БД.
- **`store_service_test.go`**: Тестирует создание магазинов и запрет создания нескольких магазинов одним продавцом.
- **`product_service_test.go`**: Тестирует валидацию цен/остатков, проверку прав продавца на добавление товаров в свой магазин и права администратора.
- **`middleware_test.go`**: Тестирует пропуск валидных JWT, отсечение неавторизованных запросов, фильтрацию по ролям RBAC и перехват паник.

### Команды для запуска тестирования и линтинга:
```bash
# Запуск всех тестов с подробным выводом
go test -v ./...

# Запуск тестов с детектором гонок (Race Detector)
go test -v -race ./...

# Проверка статическим анализатором Go
go vet ./...

# Тестовая компиляция
go build ./...
```

---

# 9. Большая шпаргалка для защиты проекта (20+ сложных вопросов с ответами)

### 1. Архитектура и дизайн
**В: Какая архитектура используется в проекте и почему?**  
**О:** Используется 3-слойная Чистая Архитектура (Clean Architecture: Handler -> Service -> Repository). Это обеспечивает слабую связанность (Loose Coupling), легкую тестируемость через интерфейсы и изолированность бизнес-логики от деталей СУБД и HTTP-фреймворка.

**В: Зачем нужны интерфейсы в пакете `internal/repository/repository.go`?**  
**О:** Интерфейсы реализуют принцип инверсии зависимостей (DIP из SOLID). Слой `service` зависит от интерфейсов, а не от PostgreSQL напрямую. Благодаря этому в тестах мы подставляем легковесные mock-репозитории в памяти без необходимости запускать реальную БД.

---

### 2. База данных и транзакции
**В: Как решается проблема одновременной покупки последнего товара двумя клиентами (Race Condition)?**  
**О:** На уровне PostgreSQL выполняется атомарный запрос `UPDATE products SET stock = stock - qty WHERE id = $1 AND stock - qty >= 0`. База данных блокирует строку на время апдейта. Второй параллельный запрос увидит, что условие `stock - qty >= 0` ложно, обновит 0 строк (`RowsAffected == 0`), и сервис мгновенно вернет ошибку `ErrInsufficientStock`.

**В: Как обеспечивается атомарность при создании заказа и позиций заказа?**  
**О:** В методе `OrderRepository.Create` используется транзакция PostgreSQL (`tx.Begin(ctx)`). Заголовок заказа в `orders` и все строки в `order_items` вставляются внутри одной транзакции. Если любая вставка завершится ошибкой, сработает `defer tx.Rollback(ctx)` и в базе не останется "висячих" записей.

**В: Что произойдет, если база данных упадет в момент оформления заказа после списания стока?**  
**О:** В `OrderService` реализован паттерн саги с компенсацией (Saga Compensation): сервис хранит список списанных товаров `deducted`. При любой ошибке создания заказа вызывается `rollbackStock`, который возвращает количество товаров на склад с использованием независимого контекста `context.Background()`.

**В: Почему в таблице `products` для `category_id` используется `ON DELETE SET NULL`?**  
**О:** Чтобы при удалении категории товары не удалялись из магазина, а оставались доступными без категории (`category_id = NULL`). Это предотвращает случайную потерю данных каталога.

---

### 3. Безопасность и Авторизация
**В: Почему пароли хешируются именно через bcrypt, а не MD5 или SHA-256?**  
**О:** MD5 и SHA-256 — быстрые алгоритмы, уязвимые к перебору на GPU (Brute-force) и радужным таблицам. `bcrypt` — это адаптивная медленная криптографическая функция, которая автоматически генерирует уникальную соль (salt) и требует высоких вычислительных затрат, делая перебор невозможным.

**В: В чем разница между Access Token и Refresh Token в проекте?**  
**О:** Access Token — это короткоживущий JWT (24ч), валидируемый сервером без обращения к БД. Refresh Token — длинноживущий токен (30 дней), хранящийся в таблице `refresh_tokens`. При каждом обновлении токена срабатывает **Token Rotation**: старый рефреш-токен удаляется из базы, предотвращая атаки повторного использования.

**В: Как защищен эндпоинт создания магазина от подмены `seller_id`?**  
**О:** `StoreHandler` берет `userID` напрямую из проверенного JWT-токена в контексте запроса. Обычный пользователь с ролью `seller` не может указать чужой `seller_id` — передача стороннего ID разрешена только пользователям с ролью `admin`.

---

### 4. Производительность и Инфраструктура
**В: Почему выбран пул соединений `pgxpool`, а не стандартный `database/sql`?**  
**О:** `pgx` — это специализированный высокопроизводительный драйвер для PostgreSQL на чистом Go. `pgxpool` эффективнее управляет конкурентными соединениями, нативно поддерживает типы PostgreSQL и работает быстрее `database/sql` за счет отсутствия лишних аллокаций.

**В: Почему итоговый Docker-образ построен на базе `distroless:nonroot`?**  
**О:** Образы Google Distroless не содержат операционной системы, пакетных менеджеров и командных оболочек (shell). Итоговый образ весит менее 25 МБ и запускается от непривилегированного пользователя `nonroot`, что сводит поверхность потенциальных атак к минимуму.

**В: Как работают авто-миграции без сторонних утилит в контейнере?**  
**О:** Файлы миграций `.sql` компилируются внутрь бинарника через директиву Go `//go:embed *.sql`. Библиотека `golang-migrate` с драйвером `iofs` читает миграции прямо из памяти бинарного файла и применяет их к БД при вызове `migrations.Run(...)`.

---

### 5. Надежность и Код
**В: Что такое Graceful Shutdown и как он реализован?**  
**О:** Это процедура корректной остановки сервера без потери данных. Функция `signal.NotifyContext` ловит сигналы ОС (`SIGINT`/`SIGTERM`), метод `server.Shutdown` выделяет до 10 секунд на завершение текущих сетевых запросов, после чего закрывает пул базы данных и сбрасывает буферы логов.

**В: Зачем нужен `rows.Err()` после цикла `rows.Next()` в репозиториях?**  
**О:** Цикл `rows.Next()` может завершиться не только потому, что строки кончились, но и из-за сетевого сбоя или ошибки чтения из сокета. Проверка `rows.Err()` гарантирует, что мы не проглотим скрытую ошибку БД и не вернем клиенту неполные данные.

**В: Зачем нужен `RecoveryMiddleware`?**  
**О:** Для перехвата паник (`panic`) внутри HTTP-горутин. Если в коде возникнет непредвиденная ошибка (например, nil pointer), сервер не упадет целиком, а залогирует ошибку со стектрейсом и вернет клиенту аккуратный JSON `500 Internal Server Error`.

---

# 10. Краткая инструкция по запуску

```bash
# 1. Клонирование и переход в каталог
cd online.shop

# 2. Запуск всех сервисов в Docker (Бэкенд + PostgreSQL + Миграции)
docker-compose up --build -d

# 3. Проверка статуса контейнеров
docker-compose ps

# 4. Открытие Swagger UI в браузере
# http://localhost:8080/swagger/

# 5. Остановка
docker-compose down
```

---
**Проект полностью документирован, протестирован и готов к защите на высший балл!** 🚀
