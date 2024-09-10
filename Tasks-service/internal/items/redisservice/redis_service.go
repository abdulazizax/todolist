package redisservice

import (
	"context"
	"fmt"
	"log"
	"task-service/internal/items/config"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	redisDb *redis.Client
	logger  *log.Logger
}

func New(redisDb *redis.Client, logger *log.Logger) *RedisService {
	return &RedisService{
		logger:  logger,
		redisDb: redisDb,
	}
}

func NewRedisClient(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURI,
		Password: "",
		DB:       0,
	})

	return rdb
}

func (r *RedisService) StoreEmailAndCode(ctx context.Context, email string, code int) error {
	codeKey := "verification_code:" + email
	err := r.redisDb.Set(ctx, codeKey, code, time.Minute*15).Err()
	if err != nil {
		r.logger.Printf("ERROR WHILE STORING VERIFICATION CODE: %s\n", err.Error())
		return err
	}
	return nil
}

func (r *RedisService) GetCodeByEmail(ctx context.Context, email string) (int, error) {
	codeKey := "verification_code:" + email
	codeStr, err := r.redisDb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		r.logger.Printf("ERROR WHILE GETTING VERIFICATION CODE: %s\n", err.Error())
		return 0, err
	}

	var code int
	_, err = fmt.Sscanf(codeStr, "%d", &code)
	if err != nil {
		r.logger.Printf("ERROR WHILE PARSING VERIFICATION CODE: %s\n", err.Error())
		return 0, err
	}

	return code, nil
}
