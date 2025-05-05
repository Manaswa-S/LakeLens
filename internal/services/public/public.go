package public

import (
	"errors"
	configs "lakelens/internal/config"
	"lakelens/internal/consts/errs"
	"lakelens/internal/dto"
	sqlc "lakelens/internal/sqlc/generate"
	utils "lakelens/internal/utils/common"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type PublicService struct {
	Queries     *sqlc.Queries
	RedisClient *redis.Client
	DB          *pgxpool.Pool
}

func NewPublicService(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *PublicService {
	return &PublicService{
		Queries:     queries,
		RedisClient: redis,
		DB:          db,
	}
}

// NewUser adds a new user to the service.
// Needs email and password, uses dto.NewUser.
// Email and Password conditions can be found in respective Validation functions.
func (s *PublicService) NewUser(ctx *gin.Context, data *dto.NewUser) *errs.Errorf {
	// TODO: need to confirm email, send email
	prblm, emailOk := utils.ValidateEmail(data.Email)
	if !emailOk {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "Invalid email : " + prblm,
			ReturnRaw: true,
		}
	}

	prblm, passOk := utils.ValidatePassword(data.Password)
	if !passOk {
		return &errs.Errorf{
			Type:      errs.ErrInvalidInput,
			Message:   "Invalid password : " + prblm,
			ReturnRaw: true,
		}
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(data.Password), configs.Internal.BcryptPasswordCost)
	if err != nil {
		return &errs.Errorf{
			Type:    errs.ErrInternalServer,
			Message: "Failed to generate bcrypt hash from password : " + err.Error(),
		}
	}

	err = s.Queries.InsertNewUser(ctx, sqlc.InsertNewUserParams{
		Email:    data.Email,
		Password: string(passHash),
	})
	if err != nil {
		errf := errs.Errorf{
			Type:    errs.ErrDBQuery,
			Message: "Failed to insert new user into db : " + err.Error(),
		}

		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.Code == errs.PGErrUniqueViolation {
				errf.Type = errs.ErrConflict
				errf.Message = "User with this email already exists. Log in or use another email."
				errf.ReturnRaw = true
			}
		}
		return &errf
	}

	return nil
}
