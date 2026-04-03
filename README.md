# Threads

Микросервис с GraphQL API, написанный на Go. 
Приложение поддерживает два режима хранения 
данных: In-Memory и в PostgreSQL.

## Основные технологии и библиотеки

В проекте используются следующие либы:

* **[99designs/gqlgen](https://github.com/99designs/gqlgen)** - кодоген GraphQL
* **[jackc/pgx/v5](https://github.com/jackc/pgx)** - для работы с PostgreSQL запросами
* **[golang-migrate/migrate](https://github.com/golang-migrate/migrate)** - для управления миграциями БД
* **[stretchr/testify](https://github.com/stretchr/testify)** — тесты
* **[pashagolub/pgxmock](https://github.com/pashagolub/pgxmock)** — моки postgres

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

### Пример создания поста
~~~graphql
mutation CreateSinglePost{
  createPost(
    author: "dowtai"
    title: "try"
    content: "wow"
  ){
    id
    author
    title
    commentsAllowed
    createdAt
  }
}
~~~

### Пример запроса создания комментария
~~~graphql
mutation CreateComment{
  createComment(
    postId: "<id существующего поста>"
    author: "dowtai"
    text: "comment text"
  ){
    id
    author
    text
    createdAt
  }
}
~~~

### Пример запроса постов и комментариев к ним:
~~~graphql
query GetPosts{
  posts(limit: 5, offset: 0){
    id
    author
    title
    content
    commentsAllowed
    createdAt
    comments(limit: 3, offset: 2){
      id
      author
      text
      createdAt
    }
  }
}
~~~