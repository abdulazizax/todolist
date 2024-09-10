package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type (
	Config struct {
		Server   ServerConfig
		Postgres PostgresConfig
		Mongo    MongoConfig
		JWT      JWTConfig
		Email    EmailConfig
		Kafka    string
		RedisURI string
	}
	JWTConfig struct {
		SecretKey string
	}
	ServerConfig struct {
		Port string
	}
	PostgresConfig struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
	}
	MongoConfig struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
	}
	EmailConfig struct {
		SmtpHost string
		SmtpPort int
		SmtpUser string
		SmtpPass string
	}
)

func (c *Config) Load() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	c.Server.Port = ":" + os.Getenv("SERVER_PORT")

	c.Postgres.Host = os.Getenv("POSTGRES_DB_HOST")
	c.Postgres.Port = os.Getenv("POSTGRES_DB_PORT")
	c.Postgres.User = os.Getenv("POSTGRES_DB_USER")
	c.Postgres.Password = os.Getenv("POSTGRES_DB_PASSWORD")
	c.Postgres.DBName = os.Getenv("POSTGRES_DB_NAME")

	c.Mongo.Host = os.Getenv("MONGO_DB_HOST")
	c.Mongo.Port = os.Getenv("MONGO_DB_PORT")
	c.Mongo.User = os.Getenv("MONGO_DB_USER")
	c.Mongo.Password = os.Getenv("MONGO_DB_PASSWORD")
	c.Mongo.DBName = os.Getenv("MONGO_DB_NAME")

	c.JWT.SecretKey = os.Getenv("JWT_SECRET_KEY")

	c.Email.SmtpHost = os.Getenv("SMTP_HOST")
	c.Email.SmtpPort = smtpPort
	c.Email.SmtpUser = os.Getenv("SMTP_USER")
	c.Email.SmtpPass = os.Getenv("SMTP_PASS")

	c.RedisURI = os.Getenv("REDIS_URI")
	c.Kafka = os.Getenv("KAFKA_URI")

	return nil
}

func New() (*Config, error) {
	var config Config
	if err := config.Load(); err != nil {
		return nil, err
	}
	return &config, nil
}
