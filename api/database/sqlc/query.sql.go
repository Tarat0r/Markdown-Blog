// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package db

import (
	"context"
)

const createNote = `-- name: CreateNote :one

INSERT INTO notes (user_id, path, content, hash) 
VALUES ($1, $2, $3, $4) 
RETURNING id, user_id, path, content, hash, created_at, updated_at
`

type CreateNoteParams struct {
	UserID  int32  `json:"user_id"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Hash    string `json:"hash"`
}

// -----------------
// Notes Queries --
// -----------------
func (q *Queries) CreateNote(ctx context.Context, arg CreateNoteParams) (Note, error) {
	row := q.db.QueryRow(ctx, createNote,
		arg.UserID,
		arg.Path,
		arg.Content,
		arg.Hash,
	)
	var i Note
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Path,
		&i.Content,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteImage = `-- name: DeleteImage :exec
DELETE FROM images WHERE id = $1
`

func (q *Queries) DeleteImage(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteImage, id)
	return err
}

const deleteNote = `-- name: DeleteNote :one
DELETE FROM notes WHERE user_id = $1 and id = $2
RETURNING path
`

type DeleteNoteParams struct {
	UserID int32 `json:"user_id"`
	ID     int32 `json:"id"`
}

func (q *Queries) DeleteNote(ctx context.Context, arg DeleteNoteParams) (string, error) {
	row := q.db.QueryRow(ctx, deleteNote, arg.UserID, arg.ID)
	var path string
	err := row.Scan(&path)
	return path, err
}

const getIDByToken = `-- name: GetIDByToken :one
SELECT id FROM users
WHERE api_token = $1
`

func (q *Queries) GetIDByToken(ctx context.Context, apiToken string) (int32, error) {
	row := q.db.QueryRow(ctx, getIDByToken, apiToken)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const getImageByHash = `-- name: GetImageByHash :one
SELECT id, hash, uploaded_at FROM images WHERE hash = $1
`

func (q *Queries) GetImageByHash(ctx context.Context, hash string) (Image, error) {
	row := q.db.QueryRow(ctx, getImageByHash, hash)
	var i Image
	err := row.Scan(&i.ID, &i.Hash, &i.UploadedAt)
	return i, err
}

const getImageByID = `-- name: GetImageByID :one
SELECT id, hash, uploaded_at FROM images WHERE id = $1
`

func (q *Queries) GetImageByID(ctx context.Context, id int32) (Image, error) {
	row := q.db.QueryRow(ctx, getImageByID, id)
	var i Image
	err := row.Scan(&i.ID, &i.Hash, &i.UploadedAt)
	return i, err
}

const getImagesForNote = `-- name: GetImagesForNote :many
SELECT i.id, i.hash, i.uploaded_at 
FROM images i
JOIN notes_images ni ON i.id = ni.image_id
WHERE ni.note_id = $1
`

func (q *Queries) GetImagesForNote(ctx context.Context, noteID int32) ([]Image, error) {
	rows, err := q.db.Query(ctx, getImagesForNote, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Image
	for rows.Next() {
		var i Image
		if err := rows.Scan(&i.ID, &i.Hash, &i.UploadedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNoteByID = `-- name: GetNoteByID :one
SELECT id, user_id, path, content, hash, created_at, updated_at FROM notes WHERE id = $1
`

func (q *Queries) GetNoteByID(ctx context.Context, id int32) (Note, error) {
	row := q.db.QueryRow(ctx, getNoteByID, id)
	var i Note
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Path,
		&i.Content,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getNoteByPath = `-- name: GetNoteByPath :many
SELECT id, path FROM notes WHERE user_id = $1 and path LIKE $2
`

type GetNoteByPathParams struct {
	UserID int32  `json:"user_id"`
	Path   string `json:"path"`
}

type GetNoteByPathRow struct {
	ID   int32  `json:"id"`
	Path string `json:"path"`
}

func (q *Queries) GetNoteByPath(ctx context.Context, arg GetNoteByPathParams) ([]GetNoteByPathRow, error) {
	rows, err := q.db.Query(ctx, getNoteByPath, arg.UserID, arg.Path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetNoteByPathRow
	for rows.Next() {
		var i GetNoteByPathRow
		if err := rows.Scan(&i.ID, &i.Path); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNoteByPathAndID = `-- name: GetNoteByPathAndID :one
SELECT id, user_id, path, content, hash, created_at, updated_at FROM notes 
WHERE user_id = $1 and path = $2 and id = $3
`

type GetNoteByPathAndIDParams struct {
	UserID int32  `json:"user_id"`
	Path   string `json:"path"`
	ID     int32  `json:"id"`
}

func (q *Queries) GetNoteByPathAndID(ctx context.Context, arg GetNoteByPathAndIDParams) (Note, error) {
	row := q.db.QueryRow(ctx, getNoteByPathAndID, arg.UserID, arg.Path, arg.ID)
	var i Note
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Path,
		&i.Content,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getNoteImage = `-- name: GetNoteImage :one

SELECT note_id, image_id FROM notes_images
WHERE note_id = $1 AND image_id = $2
`

type GetNoteImageParams struct {
	NoteID  int32 `json:"note_id"`
	ImageID int32 `json:"image_id"`
}

// ----------------------------------------------
// Many-to-Many Relationship (Notes & Images) --
// ----------------------------------------------
func (q *Queries) GetNoteImage(ctx context.Context, arg GetNoteImageParams) (NotesImage, error) {
	row := q.db.QueryRow(ctx, getNoteImage, arg.NoteID, arg.ImageID)
	var i NotesImage
	err := row.Scan(&i.NoteID, &i.ImageID)
	return i, err
}

const getNotesForImage = `-- name: GetNotesForImage :many
SELECT n.id, n.user_id, n.path, n.content, n.hash, n.created_at, n.updated_at 
FROM notes n
JOIN notes_images ni ON n.id = ni.note_id
WHERE ni.image_id = $1
`

func (q *Queries) GetNotesForImage(ctx context.Context, imageID int32) ([]Note, error) {
	rows, err := q.db.Query(ctx, getNotesForImage, imageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Note
	for rows.Next() {
		var i Note
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Path,
			&i.Content,
			&i.Hash,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const linkImageToNote = `-- name: LinkImageToNote :exec
INSERT INTO notes_images (note_id, image_id) 
VALUES ($1, $2)
`

type LinkImageToNoteParams struct {
	NoteID  int32 `json:"note_id"`
	ImageID int32 `json:"image_id"`
}

func (q *Queries) LinkImageToNote(ctx context.Context, arg LinkImageToNoteParams) error {
	_, err := q.db.Exec(ctx, linkImageToNote, arg.NoteID, arg.ImageID)
	return err
}

const listNotesByUser = `-- name: ListNotesByUser :many
SELECT id, path, hash FROM notes WHERE user_id = $1 ORDER BY created_at DESC
`

type ListNotesByUserRow struct {
	ID   int32  `json:"id"`
	Path string `json:"path"`
	Hash string `json:"hash"`
}

func (q *Queries) ListNotesByUser(ctx context.Context, userID int32) ([]ListNotesByUserRow, error) {
	rows, err := q.db.Query(ctx, listNotesByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListNotesByUserRow
	for rows.Next() {
		var i ListNotesByUserRow
		if err := rows.Scan(&i.ID, &i.Path, &i.Hash); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const tokenExists = `-- name: TokenExists :one

SELECT COUNT(*) FROM users
WHERE api_token = $1
`

// -----------------
// Users Queries --
// -----------------
func (q *Queries) TokenExists(ctx context.Context, apiToken string) (int64, error) {
	row := q.db.QueryRow(ctx, tokenExists, apiToken)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const unlinkImageFromNote = `-- name: UnlinkImageFromNote :exec
DELETE FROM notes_images WHERE note_id = $1 AND image_id = $2
`

type UnlinkImageFromNoteParams struct {
	NoteID  int32 `json:"note_id"`
	ImageID int32 `json:"image_id"`
}

func (q *Queries) UnlinkImageFromNote(ctx context.Context, arg UnlinkImageFromNoteParams) error {
	_, err := q.db.Exec(ctx, unlinkImageFromNote, arg.NoteID, arg.ImageID)
	return err
}

const unlinkOldImagesFromNote = `-- name: UnlinkOldImagesFromNote :exec
DELETE FROM notes_images ni
USING images i
WHERE ni.image_id = i.id
  AND ni.note_id = $1
  AND i.hash NOT IN (SELECT UNNEST($2::text[]))
`

type UnlinkOldImagesFromNoteParams struct {
	NoteID int32    `json:"note_id"`
	Hashes []string `json:"hashes"`
}

func (q *Queries) UnlinkOldImagesFromNote(ctx context.Context, arg UnlinkOldImagesFromNoteParams) error {
	_, err := q.db.Exec(ctx, unlinkOldImagesFromNote, arg.NoteID, arg.Hashes)
	return err
}

const updateNote = `-- name: UpdateNote :exec
UPDATE notes SET content = $4, hash = $5
WHERE user_id = $1 and path = $2 and id = $3
`

type UpdateNoteParams struct {
	UserID  int32  `json:"user_id"`
	Path    string `json:"path"`
	ID      int32  `json:"id"`
	Content string `json:"content"`
	Hash    string `json:"hash"`
}

func (q *Queries) UpdateNote(ctx context.Context, arg UpdateNoteParams) error {
	_, err := q.db.Exec(ctx, updateNote,
		arg.UserID,
		arg.Path,
		arg.ID,
		arg.Content,
		arg.Hash,
	)
	return err
}

const uploadImage = `-- name: UploadImage :one

INSERT INTO images (hash) 
VALUES ($1) 
RETURNING id
`

// ------------------
// Images Queries --
// ------------------
func (q *Queries) UploadImage(ctx context.Context, hash string) (int32, error) {
	row := q.db.QueryRow(ctx, uploadImage, hash)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const userCanAccessImageByHash = `-- name: UserCanAccessImageByHash :one
SELECT 1
FROM notes_images ni
JOIN notes n ON n.id = ni.note_id
JOIN images i ON i.id = ni.image_id
WHERE n.user_id = $1
  AND i.hash = $2
LIMIT 1
`

type UserCanAccessImageByHashParams struct {
	UserID int32  `json:"user_id"`
	Hash   string `json:"hash"`
}

func (q *Queries) UserCanAccessImageByHash(ctx context.Context, arg UserCanAccessImageByHashParams) (int32, error) {
	row := q.db.QueryRow(ctx, userCanAccessImageByHash, arg.UserID, arg.Hash)
	var column_1 int32
	err := row.Scan(&column_1)
	return column_1, err
}
