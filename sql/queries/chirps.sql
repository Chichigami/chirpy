-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ListChirpsASC :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: ListChirpsDESC :many
SELECT * FROM chirps
ORDER BY created_at DESC;

-- name: GetAllChirpsFromAuthorASC :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: GetAllChirpsFromAuthorDESC :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at DESC;


-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;