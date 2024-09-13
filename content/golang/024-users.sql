-- name: InsertUser :one
insert into users (email, display_name, password_hash) values ($1, $2, $3) returning *;

-- name: GetUserByEmail :one
select * from users where email = $1;
