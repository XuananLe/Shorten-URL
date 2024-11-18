package services

import (
	"context"
	"fmt"
	"shorten-url/backend/pkg/utils"
	"github.com/google/uuid"
)

type User struct {
	UserID string `json:"userId"`
}


func (s *UrlService) CreateUser(userIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	err = s.postgresClient.Queries.InsertUser(context.Background(), utils.ConvertFromUuidPg(userID))

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

