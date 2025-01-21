// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const checkAccountExistsByEmail = `-- name: CheckAccountExistsByEmail :one
SELECT EXISTS(
    SELECT 1 FROM accounts WHERE email = $1
) AS exists
`

func (q *Queries) CheckAccountExistsByEmail(ctx context.Context, email string) (bool, error) {
	row := q.db.QueryRow(ctx, checkAccountExistsByEmail, email)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT email, password, role
FROM accounts
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (Accounts, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i Accounts
	err := row.Scan(&i.Email, &i.Password, &i.Role)
	return i, err
}

const insertAccount = `-- name: InsertAccount :exec
INSERT INTO "accounts" ("email", "password", "role")
VALUES ($1, $2, $3)
`

type InsertAccountParams struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     NullRoles `json:"role"`
}

func (q *Queries) InsertAccount(ctx context.Context, arg InsertAccountParams) error {
	_, err := q.db.Exec(ctx, insertAccount, arg.Email, arg.Password, arg.Role)
	return err
}

const insertUserAction = `-- name: InsertUserAction :exec
INSERT INTO "user_actions" ("email", "created_at", "updated_at", "login_at", "logout_at")
VALUES ($1, NOW(), NOW(), NULL, NULL)
`

func (q *Queries) InsertUserAction(ctx context.Context, email pgtype.Text) error {
	_, err := q.db.Exec(ctx, insertUserAction, email)
	return err
}

const insertUserInfo = `-- name: InsertUserInfo :exec
INSERT INTO "user_infos" ("email", "name", "sex", "birth_day")
VALUES ($1, $2, $3, $4)
`

type InsertUserInfoParams struct {
	Email    pgtype.Text `json:"email"`
	Name     string      `json:"name"`
	Sex      NullSex     `json:"sex"`
	BirthDay string      `json:"birth_day"`
}

func (q *Queries) InsertUserInfo(ctx context.Context, arg InsertUserInfoParams) error {
	_, err := q.db.Exec(ctx, insertUserInfo,
		arg.Email,
		arg.Name,
		arg.Sex,
		arg.BirthDay,
	)
	return err
}

const updateAction = `-- name: UpdateAction :execresult
UPDATE "user_actions"
SET 
    login_at = CASE WHEN $1::timestamptz IS NOT NULL THEN $1::timestamptz ELSE login_at END,
    logout_at = CASE WHEN $2::timestamptz IS NOT NULL THEN $2::timestamptz ELSE logout_at END,
    updated_at = now()
WHERE email = $3::text
RETURNING 
    login_at AS login_at,
    logout_at AS logout_at,
    email AS email
`

type UpdateActionParams struct {
	Column1 pgtype.Timestamptz `json:"column_1"`
	Column2 pgtype.Timestamptz `json:"column_2"`
	Column3 string             `json:"column_3"`
}

func (q *Queries) UpdateAction(ctx context.Context, arg UpdateActionParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, updateAction, arg.Column1, arg.Column2, arg.Column3)
}
