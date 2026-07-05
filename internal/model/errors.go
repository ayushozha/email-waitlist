package model

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
	ErrReferralCodeTaken = errors.New("referral code already taken")
	ErrSlugTaken         = errors.New("project slug already exists")
)

// isUniqueViolation reports whether err is a Postgres unique_violation (23505)
// on the named constraint. Matching by constraint name lets callers tell a
// duplicate email apart from a position-counter race or referral collision.
func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraint
}
