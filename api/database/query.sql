-------------------
-- Users Queries --
-------------------

-- name: TokenExists :one
SELECT COUNT(*) FROM users
WHERE api_token = $1;

-- name: GetIDByToken :one
SELECT id FROM users
WHERE api_token = $1;

-- name: SetTestToken :exec
INSERT INTO users (api_token, name, email)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-------------------
-- Notes Queries --
-------------------

-- name: CreateNote :one
INSERT INTO notes (user_id, path, content, content_md, hash) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING *;

-- name: GetNoteByID :one
SELECT * FROM notes WHERE id = $1;

-- name: GetNoteByPath :many
SELECT id, path FROM notes WHERE user_id = $1 and path LIKE $2;

-- name: GetNoteByPathAndID :one
SELECT * FROM notes 
WHERE user_id = $1 and path = $2 and id = $3;

-- name: ListNotesByUser :many
SELECT id, path, hash FROM notes WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateNote :exec
UPDATE notes SET content = $4, hash = $5, content_md = $6
WHERE user_id = $1 and path = $2 and id = $3;

-- name: DeleteNote :one
DELETE FROM notes WHERE user_id = $1 and id = $2
RETURNING path;

--------------------
-- Images Queries --
--------------------

-- name: UploadImage :one
INSERT INTO images (hash) 
VALUES ($1) 
RETURNING id;

-- name: GetImageByID :one
SELECT * FROM images WHERE id = $1;

-- name: GetImageByHash :one
SELECT * FROM images WHERE hash = $1;

-- name: UserCanAccessImageByHash :one
SELECT 1
FROM notes_images ni
JOIN notes n ON n.id = ni.note_id
JOIN images i ON i.id = ni.image_id
WHERE n.user_id = $1
  AND i.hash = $2
LIMIT 1;

-- name: DeleteImage :exec
DELETE FROM images WHERE id = $1;

------------------------------------------------
-- Many-to-Many Relationship (Notes & Images) --
------------------------------------------------

-- name: GetNoteImage :one
SELECT * FROM notes_images
WHERE note_id = $1 AND image_id = $2;

-- name: LinkImageToNote :exec
INSERT INTO notes_images (note_id, image_id) 
VALUES ($1, $2);

-- name: GetImagesForNote :many
SELECT i.* 
FROM images i
JOIN notes_images ni ON i.id = ni.image_id
WHERE ni.note_id = $1;

-- name: GetNotesForImage :many
SELECT n.* 
FROM notes n
JOIN notes_images ni ON n.id = ni.note_id
WHERE ni.image_id = $1;

-- name: UnlinkImageFromNote :exec
DELETE FROM notes_images WHERE note_id = $1 AND image_id = $2;

-- name: UnlinkOldImagesFromNote :exec
DELETE FROM notes_images ni
USING images i
WHERE ni.image_id = i.id
  AND ni.note_id = $1
  AND i.hash NOT IN (SELECT UNNEST(sqlc.arg(hashes)::text[]));


