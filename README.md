# Threads

Микросервис с GraphQL API, написанный на Go. 
Приложение поддерживает два режима хранения 
данных: In-Memory и в PostgreSQL.

## Основные технологии и библиотеки

В проекте используются следующие либы:

* **[99designs/gqlgen](https://github.com/99designs/gqlgen)** - кодоген GraphQL
* **[jackc/pgx/v5](https://github.com/jackc/pgx)** - для работы с PostgreSQL запросами
* **[golang-migrate/migrate](https://github.com/golang-migrate/migrate)** - для управления миграциями БД
* **[stretchr/testify](https://github.com/stretchr/testify)** - тесты
* **[pashagolub/pgxmock](https://github.com/pashagolub/pgxmock)** - моки postgres

## Запуск без базы данных

~~~bash
docker-compose --profile inmemory up -d --build
~~~

## Запуск с PostgreSQL

~~~bash
docker-compose --profile postgres up -d --build
~~~

### Требования
* Установленный [Docker](https://docs.docker.com/get-docker/)
* Установленный [Docker Compose](https://docs.docker.com/compose/install/)

### Инструкция

1. Склонируйте репозиторий
~~~bash
git clone https://github.com/Dowtai/threads.git
cd threads
~~~

2. Запустите проект с профилем `postgres` (или `inmemory`
если достаточно хранения данных в оперативной памяти):
~~~bash
docker-compose --profile postgres up -d --build
~~~

3. Для остановки сервиса и базы данных выполните:
~~~bash
docker-compose --profile postgres down
~~~
или
~~~bash
docker-compose --profile inmemory down
~~~

## Использование API

После успешного запуска сервис будет доступен по адресу:
**http://localhost:8080/**

Перейдите по этой ссылке в браузере, чтобы открыть **GraphQL Playground**.

### Получение списка постов

Возвращает список всех постов. 
Можно сразу запросить вложенные комментарии 
(также с пагинацией).

```graphql
query GetPosts {
  posts(limit: 10, offset: 0) {
    id
    author
    title
    content
    commentsAllowed
    createdAt
    comments(limit: 5, offset: 0) {
      id
      author
      text
      createdAt
      replies(limit: 2, offset: 1) {
        id
        author
        text
        createdAt
      }
    }
  }
}
```

### Получение конкретного поста по ID

```graphql
query GetSinglePost {
  post(id: "UUID-вашего-поста") {
    id
    title
    content
    createdAt
  }
}
```

### Создание поста

Создает новый пост. 
Поле `commentsAllowed` по умолчанию 
равно `true`, но его можно передать явно.

```graphql
mutation CreatePost {
  createPost(
    author: "user_123"
    title: "Мой первый пост"
    content: "Текст поста..."
    commentsAllowed: true
  ) {
    id
    title
    createdAt
  }
}
```

### Создание комментария или ответа

Если не передавать `parentId`, то 
комментарий будет привязан к самому посту. 
Если дополнительно передать parentId 
(ID другого комментария), он станет ответом.

```graphql
mutation CreateComment {
  createComment(
    postId: "UUID-поста"
    author: "user_456"
    text: "Отличный пост!"
  ) {
    id
    text
    createdAt
  }
}
```

### Включение/отключение комментариев

Позволяет автору поста закрыть 
или открыть возможность комментирования.

```graphql
mutation ToggleComments {
  updateCommentsAllowed(
    postId: "UUID-поста"
    author: "user_123"
    allowed: false
  ) {
    id
    commentsAllowed
  }
}
```

### Подписка на новые комментарии у поста

Используется для получения обновлений 
в реальном времени по протоколу WebSocket.

```graphql
subscription OnCommentAdded {
  commentAdded(postId: "UUID-поста") {
    id
    author
    text
    createdAt
  }
}
```