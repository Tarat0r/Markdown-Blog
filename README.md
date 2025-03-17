# Markdown-Blog

Этот проект представляет собой веб-приложение для публикации заметок в формате Markdown. Серверная часть реализована с использованием FastAPI, клиентская часть на Flask, а база данных — PostgreSQL. Взаимодействие между клиентом и сервером осуществляется через REST API.

Функционал:

- Создание, чтение, обновление и удаление (CRUD) заметок в формате Markdown.
- Отображение заметок в виде веб-блога с рендерингом Markdown.
- RESTful API для управления заметками.
- Хранение данных в базе PostgreSQL.
- Покрытие серверной и клиентской части модульными тестами.

## Database

![DB_schema](/database/schema.png)

## Resources

<details>
<summary><code>GET</code> <code><b>/notes</b></code> <code>(Получение списка заметок)</code></summary>

#### Request

```bash
curl -X GET http://localhost:8080/notes \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$"
```

#### Response

```json
[
    {
        "path": "file_path/file_name_1.md",
        "hash": "hash sha256"
    },
    {
        "path": "file_path/file_name_2.md",
        "hash": "hash sha256"
    }
]
```

</details>

<details>
<summary><code>GET</code> <code><b>/notes/{path}</b></code> <code>(Получение конкретной заметки)</code></summary>

#### Request

```bash
curl -X GET http://localhost:8080/notes/{path} \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$" \
     -H "Content-Type: application/json" \
     -d '{
           "path": "file_path/file_name.md"
         }'
```

#### Response

```json
{
    "path": "file_path/file_name.md",
    "content": "html_text",
    "hash": "hash sha256",
    "updated_at": "time" 
}
```

</details>

<details>
<summary><code>POST</code> <code><b>/notes</b></code> <code>(Создание новой заметки)</code></summary>

#### Request

```bash
curl -X POST http://localhost:8080/notes \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$" \
     -H "Content-Type: application/json" \
     -d '{
           "path": "file_path/file_name.md",
           "content": "md_text"
         }'
```

#### Response

```json
{
    "path": "file_path/file_name.md",
    "hash": "hash sha256"
}
```

</details>

<details>
<summary><code>PUT</code> <code><b>/notes/{path}</b></code> <code>(Обновление заметки)</code></summary>

#### Request

```bash
curl -X PUT http://localhost:8080/notes/{path} \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$" \
     -H "Content-Type: application/json" \
     -d '{
           "path": "file_path/file_name.md",
           "content": "md_text"
         }'
```

#### Response

```json
{
    "path": "file_path/file_name.md",
    "hash": "hash sha256"
}
```

</details>

<details>
<summary><code>DELETE</code> <code><b>/notes/{path}</b></code> <code>(Удаление заметки)</code></summary>

#### Request

```bash
curl -X DELETE http://localhost:8080/notes/{path} \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$" \
     -H "Content-Type: application/json" \
     -d '{
           "path": "file_path/file_name.md"
         }'
```

#### Response

```json
{
    "path": "file_path/file_name.md"
}
```

</details>
