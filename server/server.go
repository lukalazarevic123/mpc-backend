package server

import (
	"mpc-backend/config"

	"github.com/rs/zerolog/log"
)

type Server struct {
	handler *Handler
}

func NewServer(conf config.Configuration) error {
	handler := NewHandler(conf)

	if err := handler.Run(); err != nil {
		log.Fatal().Err(err)
		return err
	}

	return nil
}
