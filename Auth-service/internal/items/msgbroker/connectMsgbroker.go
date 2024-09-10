package msgbroker

import (
	"auth-service/internal/items/config"

	"github.com/segmentio/kafka-go"
)

type MsgBrokers struct {
	Register       *kafka.Reader
	CreateAdmin    *kafka.Reader
	DeleteUser     *kafka.Reader
	UpdatePassword *kafka.Reader
}

func InitMessageBroker(config *config.Config) *MsgBrokers {
	readers := map[string]*kafka.Reader{
		"register": kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{config.Kafka},
			Topic:   "register",
			GroupID: "auth_service",
		}),
		"create_admin": kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{config.Kafka},
			Topic:   "create_admin",
			GroupID: "auth_service",
		}),
		"delete_user": kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{config.Kafka},
			Topic:   "delete_user",
			GroupID: "auth_service",
		}),
		"update_password": kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{config.Kafka},
			Topic:   "update_password",
			GroupID: "auth_service",
		}),
	}

	return &MsgBrokers{
		Register:       readers["register"],
		CreateAdmin:    readers["create_admin"],
		DeleteUser:     readers["delete_user"],
		UpdatePassword: readers["update_password"],
	}
}
