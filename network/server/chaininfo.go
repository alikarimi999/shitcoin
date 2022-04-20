package server

import (
	"github.com/labstack/echo/v4"
)

func (s *Server) peers(ctx echo.Context) error {
	for _, p := range s.Ch.Peers {
		ctx.JSONPretty(200, p, " ")
	}
	return nil

}

func (s *Server) block(ctx echo.Context) error {
	hash := ctx.Param("hash")
	if hash != "" {
		block, err := s.Ch.DB.GetBlockH([]byte(hash), nil)
		if err != nil {
			return err
		}
		ctx.JSONPretty(200, block, " ")
	}
	return nil
}
