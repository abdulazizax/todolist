package msgbroker

import (
	"context"
	"log/slog"
	"sync"
	"task-service/internal/items/service"

	pb "task-service/genproto/auth"

	"github.com/segmentio/kafka-go"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type MsgBroker struct {
	service *service.Service
	readers *MsgBrokers
	logger  *slog.Logger
	wg      *sync.WaitGroup
}

func New(service *service.Service, logger *slog.Logger, readers *MsgBrokers, wg *sync.WaitGroup) *MsgBroker {
	return &MsgBroker{
		service: service,
		readers: readers,
		logger:  logger,
		wg:      wg,
	}
}

func (m *MsgBroker) StartToConsume(ctx context.Context) {
	m.wg.Add(4)

	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go m.consumeMessages(consumerCtx, m.readers.Register, "register")
	go m.consumeMessages(consumerCtx, m.readers.CreateAdmin, "create_admin")
	go m.consumeMessages(consumerCtx, m.readers.DeleteUser, "delete_user")
	go m.consumeMessages(consumerCtx, m.readers.UpdatePassword, "update_password")

	<-consumerCtx.Done()
	m.logger.Info("All consumers have stopped")
}

func (m *MsgBroker) consumeMessages(ctx context.Context, reader *kafka.Reader, logPrefix string) {
	defer m.wg.Done()
	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Context done, stopping consumer", "consumer", logPrefix)
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				m.logger.Error("Error reading message", "error", err, "topic", logPrefix)
				return
			}

			var response proto.Message
			var errUnmarshal error

			switch logPrefix {
			case "register":
				var req pb.RegisterRequest
				errUnmarshal = protojson.Unmarshal(msg.Value, &req)
				// response, err = m.service.Register(ctx, &req)
			case "create_admin":
				var req pb.CreateAdminRequest
				errUnmarshal = protojson.Unmarshal(msg.Value, &req)
				// response, err = m.service.CreateAdmin(ctx, &req)
			case "delete_user":
				var req pb.DeleteUserRequest
				errUnmarshal = protojson.Unmarshal(msg.Value, &req)
				// response, err = m.service.DeleteUser(ctx, &req)
			case "update_password":
				var req pb.UpdateUserPasswordRequest
				errUnmarshal = protojson.Unmarshal(msg.Value, &req)
				// response, err = m.service.UpdatePasswordService(ctx, &req)
			}

			if errUnmarshal != nil {
				m.logger.Error("Error while unmarshaling data", "error", errUnmarshal)
				continue
			}

			if err != nil {
				m.logger.Error("Failed in %s: %s\n", logPrefix, err.Error())
				continue
			}

			_, err = proto.Marshal(response)
			if err != nil {
				m.logger.Error("Failed to marshal response", "error", err)
				continue
			}

			m.logger.Info("Successfully processed message", "topic", logPrefix)
		}
	}
}
