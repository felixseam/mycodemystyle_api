package auth

import (
	"context"
	"errors"
	"mycodemystyle_api/internal/apperr"
	"mycodemystyle_api/internal/logger"
	"mycodemystyle_api/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepo interface {
	GetByUserName(ctx context.Context, userName string) (*Member, error)
	UpdateLogin(ctx context.Context, id int, session string, at time.Time) error
	FindByUUID(ctx context.Context, userUUID string) (*model.MemberContext, error)
}

type AuthRepoImpl struct {
	pool *pgxpool.Pool
}

func NewRepoImpl(pool *pgxpool.Pool) *AuthRepoImpl {
	return &AuthRepoImpl{pool: pool}
}

func (r *AuthRepoImpl) GetByUserName(ctx context.Context, userName string) (*Member, error) {
	query := `
		SELECT id,
	       member_uuid::text AS member_uuid,
	       user_name,
	       password,
	       login_session,
	       status_id
		FROM tbl_members
		WHERE user_name = $1
	  	AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, userName)
	if err != nil {
		logger.NewCustomLog("database_error", "auth.GetByUserName query: "+err.Error(), "error")
		return nil, apperr.ErrInternal
	}

	member, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Member])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		logger.NewCustomLog("database_error", "auth.GetByUserName scan: "+err.Error(), "error")
		return nil, apperr.ErrInternal
	}

	return &member, nil

}

func (r *AuthRepoImpl) UpdateLogin(ctx context.Context, id int, session string, at time.Time) error {
	query := `
		UPDATE tbl_members
		SET login_session = $2,
		    last_login    = $3,
		    last_access   = $3
		WHERE id = $1
	`

	if _, err := r.pool.Exec(ctx, query, id, session, at); err != nil {
		logger.NewCustomLog("database_error", "auth.UpdateLogin: "+err.Error(), "error")
		return apperr.ErrInternal
	}
	return nil
}

func (r *AuthRepoImpl) FindByUUID(ctx context.Context, userUUID string) (*model.MemberContext, error) {
	query := `
		SELECT id,
		       member_uuid::text,
		       user_name,
		       login_session,
		       status_id
		FROM tbl_members
		WHERE member_uuid = $1
		  AND deleted_at IS NULL
	`

	var u model.MemberContext
	err := r.pool.QueryRow(ctx, query, userUUID).
		Scan(&u.ID, &u.MemberUUID, &u.UserName, &u.LoginSession, &u.StatusID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		logger.NewCustomLog("database_error", "auth.FindByUUID: "+err.Error(), "error")
		return nil, apperr.ErrInternal
	}
	return &u, nil
}
