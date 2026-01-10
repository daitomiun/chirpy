-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    gen_random_uuid(),
		NOW(),
		NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetUserById :one
select * from users where id=$1;

-- name: DeleteUsers :exec
delete from users;

-- name: GetUserByEmail :one
select * from users where email=$1;

-- name: UpdateUser :one
update users set
    email = $1,
    password = $2,
    updated_at = NOW()
where id = $3
returning *;


