Отличная идея! Судя по скриншоту `image_dcf07b.png`, вы уже начали структурировать проект `User-service`, выделив директорию `internal` с пакетами `config` и `repository`. Это отличный старт для Clean Architecture.

В Clean Architecture (и DDD) главное правило — **зависимости всегда направлены внутрь**. Внешние слои (HTTP-роутеры, базы данных) зависят от бизнес-логики, а не наоборот.

Давайте разобьем ваш проект на 4 основных слоя:

1. **Domain (Домен)**: Основные сущности и интерфейсы (контракты).
2. **Repository (Репозиторий/Инфраструктура)**: Реализация работы с БД (Redis).
3. **UseCase (Сценарии использования)**: Бизнес-логика.
4. **Delivery/Handler (Слой доставки)**: HTTP роутинг с использованием `go-chi`.

Ниже приведена полная структура и пример реализации.

### 1. Структура проекта

В вашем проекте `User-service` мы создадим такую структуру внутри `internal/`:

```text
User-service/
├── cmd/
│   └── app/
│       └── main.go              # Точка входа
├── internal/
│   ├── config/
│   │   └── config.go            # Ваш файл конфигурации[cite: 1]
│   ├── domain/
│   │   └── user.go              # Сущности и интерфейсы (Core)
│   ├── repository/
│   │   ├── redis_client.go      # Подключение к Redis (ваш код)[cite: 2]
│   │   └── user_redis.go        # Реализация интерфейса репозитория
│   ├── usecase/
│   │   └── user_usecase.go      # Бизнес-логика
│   └── delivery/
│       └── http/
│           └── handler.go       # HTTP обработчики (Chi)
├── .env
├── go.mod
└── go.sum

```

### 2. Domain (Внутренний слой)

Создайте файл `internal/domain/user.go`. Здесь мы описываем сущность и интерфейсы. Этот слой не зависит ни от Redis, ни от HTTP.

```go
package domain

import "context"

// User - основная доменная сущность
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UserRepository - контракт (интерфейс) для работы с БД
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
}

// UserUseCase - контракт бизнес-логики
type UserUseCase interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
}

```

### 3. Repository (Инфраструктура)

Сначала оставим вашу функцию подключения к Redis (из `RedisRepository.go`) в файле `internal/repository/redis_client.go`.

Затем создадим `internal/repository/user_redis.go` для реализации интерфейса `UserRepository`:

```go
package repository

import (
	"context"
	"encoding/json"
	"User-service/internal/domain" // Замените на имя вашего модуля из go.mod
	"github.com/redis/go-redis/v9"
)

type userRedisRepo struct {
	client *redis.Client
}

// NewUserRedisRepository создает новый экземпляр репозитория
func NewUserRedisRepository(client *redis.Client) domain.UserRepository {
	return &userRedisRepo{client: client}
}

func (r *userRedisRepo) Save(ctx context.Context, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	// Сохраняем в Redis без срока действия (0)
	return r.client.Set(ctx, user.ID, data, 0).Err()
}

func (r *userRedisRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	data, err := r.client.Get(ctx, id).Result()
	if err != nil {
		return nil, err // Если ключ не найден, вернется redis.Nil
	}

	var user domain.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

```

### 4. UseCase (Бизнес-логика)

Создайте `internal/usecase/user_usecase.go`. Этот слой управляет потоком данных.

```go
package usecase

import (
	"context"
	"User-service/internal/domain"
)

type userUseCase struct {
	userRepo domain.UserRepository
}

// NewUserUseCase инжектит репозиторий в бизнес-логику
func NewUserUseCase(repo domain.UserRepository) domain.UserUseCase {
	return &userUseCase{
		userRepo: repo,
	}
}

func (u *userUseCase) CreateUser(ctx context.Context, user *domain.User) error {
	// Здесь могут быть бизнес-проверки (например, валидация данных)
	return u.userRepo.Save(ctx, user)
}

func (u *userUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

```

### 5. Delivery (Слой HTTP с Chi)

Создайте `internal/delivery/http/handler.go`. Слой доставки работает с фреймворком `go-chi` и вызывает UseCase.

```go
package http

import (
	"encoding/json"
	"net/http"

	"User-service/internal/domain"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	useCase domain.UserUseCase
}

// RegisterRoutes инициализирует роуты
func RegisterRoutes(r *chi.Mux, uc domain.UserUseCase) {
	handler := &UserHandler{
		useCase: uc,
	}

	r.Route("/users", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/{id}", handler.Get)
	})
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.useCase.CreateUser(r.Context(), &user); err != nil {
		http.Error(w, "Error saving user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.useCase.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

```

### 6. Сборка всего в Main (Dependency Injection)

И наконец, обновите ваш `main.go`. В `main` (или слое `app`) происходит так называемый *Dependency Injection* — мы собираем все кубики вместе.

```go
package main

import (
	"context"
	"log"
	"net/http"

	"User-service/internal/config"
	delivery "User-service/internal/delivery/http"
	"User-service/internal/repository"
	"User-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. Загрузка конфигурации из .env[cite: 1]
	cfg := config.LoadConfig()

	// 2. Инициализация инфраструктуры (Redis)[cite: 2]
	redisClient, err := repository.NewRedisClient(context.Background(), cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Could not connect to redis: %v", err)
	}

	// 3. Сборка слоев (Dependency Injection)
	userRepo := repository.NewUserRedisRepository(redisClient) // Домен зависит от интерфейса
	userUseCase := usecase.NewUserUseCase(userRepo)            // UseCase зависит от репозитория

	// 4. Инициализация Chi Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// 5. Регистрация HTTP обработчиков
	delivery.RegisterRoutes(r, userUseCase)

	// 6. Запуск сервера
	log.Println("Server running on port 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

```

### Почему эта архитектура лучше?

1. **Тестируемость**: Вы легко можете протестировать `UserUseCase`, подменив `UserRepository` с помощью моков (mocking), так как связь идет через интерфейс, а не напрямую к Redis.
2. **Изоляция**: Ваш слой бизнес логики (UseCase) ничего не знает про HTTP-статусы `200` или `404`, а слой роутера (`delivery`) ничего не знает про `redis.Nil`.
3. **Масштабируемость**: Если вы решите перейти с Redis на PostgreSQL, вам нужно будет создать только новый файл репозитория (например `user_postgres.go`) и поменять одну строчку инициализации в `main.go`. Домен и UseCase останутся нетронутыми!