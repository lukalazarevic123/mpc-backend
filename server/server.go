package server

import (
	"fmt"
	"mpc-backend/config"
	crud "mpc-backend/core"
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

	crudHandler := crud.NewCRUD(masterDb)

	handler := NewHandler(conf, crudHandler)

	if err := handler.Run(); err != nil {
		log.Fatal().Err(err)
		return err
	}

	return nil
}
