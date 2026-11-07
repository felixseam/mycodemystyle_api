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
	GetByUserName(ctx context.Context, userName string) (*User, error)
	UpdateLogin(ctx context.Context, id int, session string, at time.Time) error
	FindByUUID(ctx context.Context, userUUID string) (*model.UserContext, error)
}

type AuthRepoImpl struct {
	pool *pgxpool.Pool
}

func NewRepoImpl(pool *pgxpool.Pool) *AuthRepoImpl {
	return &AuthRepoImpl{pool: pool}
}

func (r *AuthRepoImpl) GetByUserName(ctx context.Context, userName string) (*User, error) {
	query := `
		SELECT id,
	       user_uuid::text AS user_uuid,
	       user_name,
	       password,
	       login_session,
	       status_id,
		   role_id
		FROM tbl_users
		WHERE user_name = $1
	  	AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, userName)
	if err != nil {
		logger.NewCustomLog("database_error", "auth.GetByUserName query: "+err.Error(), "error")
		return nil, apperr.ErrInternal
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		logger.NewCustomLog("database_error", "auth.GetByUserName scan: "+err.Error(), "error")
		return nil, apperr.ErrInternal
	}

	return &user, nil

}

func (r *AuthRepoImpl) UpdateLogin(ctx context.Context, id int, session string, at time.Time) error {
	query := `
		UPDATE tbl_users
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

func (r *AuthRepoImpl) FindByUUID(ctx context.Context, userUUID string) (*model.UserContext, error) {
	query := `
		SELECT id,
		       user_uuid::text,
		       user_name,
		       login_session,
		       status_id, 
			   role_id
		FROM tbl_users
		WHERE user_uuid = $1
		  AND deleted_at IS NULL
	`

	var u model.UserContext
	err := r.pool.QueryRow(ctx, query, userUUID).
		Scan(&u.ID, &u.UserUUID, &u.UserName, &u.LoginSession, &u.StatusID, &u.RoleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		logger.NewCustomLog("database_error", "auth.FindByUUID: "+err.Error(), "error")
		return nil, apperr.ErrInternal
	}
	return &u, nil
}
