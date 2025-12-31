-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),
		NOW(),
		NOW(),
    $1
)
RETURNING *;

-- name: GetUserById :one
select * from users where id=$1;

-- name: DeleteUsers :exec
delete from users;

