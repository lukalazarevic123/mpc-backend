package server

import (
	"fmt"
	"mpc-backend/config"
	"mpc-backend/db"

	"github.com/rs/zerolog/log"
)

type Server struct {
	handler *Handler
}

func NewServer(conf config.Configuration) error {
	masterDb, err := db.NewMasterDb(conf.DbConfig)
	if err != nil {
		return fmt.Errorf("could not connect to db: %w", err)
	}

	log.Print(masterDb)

	handler := NewHandler(conf)

	if err := handler.Run(); err != nil {
		log.Fatal().Err(err)
		return err
	}

	return nil
}
