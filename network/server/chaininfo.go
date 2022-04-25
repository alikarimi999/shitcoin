package server

import (
	"encoding/hex"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (s *Server) peers(ctx echo.Context) error {
	s.PeerSet.Mu.Lock()
	defer s.PeerSet.Mu.Unlock()
	for _, p := range s.PeerSet.Peers {
		ctx.JSONPretty(200, p, " ")
	}
	return nil

}

func (s *Server) block(ctx echo.Context) error {

	hash := ctx.QueryParam("hash")
	if hash != "" {
		xh, err := hex.DecodeString(hash)
		if err != nil {
			return err
		}
		block, err := s.Ch.DB.GetBlockH(xh, nil)
		if err != nil {
			return err
		}
		ctx.JSONPretty(200, block, " ")
	}

	i := ctx.QueryParam("index")
	index, err := strconv.ParseUint(i, 10, 64)
	if err != nil {
		return err
	}
	block, err := s.Ch.DB.GetBlockI(index, nil)
	if err != nil {
		return err
	}
	ctx.JSONPretty(200, block, " ")

	return nil
}
