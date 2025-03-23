# Markdown-Blog

This project is a web application for publishing notes in Markdown format. The server part is implemented using Go, the client part is on Flask (Web-client) and Python (PC-client), and the database is PostgreSQL. Interaction between the clients and the server is carried out via REST API.

Functionality:

- Create, read, update and delete (CRUD) notes in Markdown format.
- Display notes as a web blog with Markdown rendering.
- RESTful API for managing notes.
- Storing data in a PostgreSQL database.
- Covering the server and client parts with tests.

## Database

![DB_schema](/database/schema.png)

## Resources

<details>
<summary><code>GET</code> <code><b>/notes</b></code> <code>(Getting a list of notes)</code></summary>

#### Request

```bash
curl -X GET http://localhost:8080/notes \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$"
```

#### Response

```json
[
    {
        "id": 1,
        "path": "file_path/file_name_1.md",
        "hash": "hash sha256"
    },
    {
        "id": 2,
        "path": "file_path/file_name_2.md",
        "hash": "hash sha256"
    }
]
```

</details>

<details>
<summary><code>GET</code> <code><b>/notes/{id}</b></code> <code>(Getting a specific note)</code></summary>

#### Request

```bash
curl -X GET http://localhost:8080/notes/{id} \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$" \
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
<summary><code>GET</code> <code><b>/images/{hash}</b></code> <code>(Getting a specific image)</code></summary>

#### Request

```bash
curl -X GET http://localhost:8080/images/{hash} \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$" \
```

#### Response

```http
HTTP/1.1 200 OK
Content-Type: image/jpeg
Content-Length: 13422
```

</details>

<details>
<summary><code>POST</code> <code><b>/notes</b></code> <code>(Create a new note)</code></summary>

#### Request

```bash
curl -X POST http://localhost:8080/notes \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$" \
     -H "Content-Type: multipart/form-data" \
     -F "metadata={
         \"path\": \"notes/note1.md\",
         \"images\": [
             {\"path\": \"images/img1.jpg\"},
             {\"path\": \"images/img2.jpg\"}
         ]
     };type=application/json" \
     -F "markdown=@note1.md" \
     -F "image=@img1.jpg" \
     -F "image=@img2.jpg"


```

#### Response

```json
{
    "id": 1,
    "path": "file_path/file_name.md",
    "hash": "hash sha256"
}
```

</details>

<details>
<summary><code>PUT</code> <code><b>/notes/{id}</b></code> <code>(Updating note)</code></summary>

#### Request

```bash
curl -X PUT http://localhost:8080/notes/{id} \
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
<summary><code>DELETE</code> <code><b>/notes/{id}</b></code> <code>(Delete note)</code></summary>

#### Request

```bash
curl -X DELETE http://localhost:8080/notes/{id} \
     -H "API_TOKEN_FORMAT: ^[a-zA-Z0-9]{64}$" \
```

#### Response

```json
{
    "path": "file_path/file_name.md"
}
```

</details>

## TODO

- [x] GET /notes
- [x] GET /notes/{id}
- [x] GET /images/{hash}
- [x] POST /notes
- [x] PUT /notes/{id}
- [x] DELETE /notes/{id}
