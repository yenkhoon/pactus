package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pactus-project/pactus/types/account"
	"github.com/pactus-project/pactus/www/capnp"
)

// GetAccountHandler returns a handler to get account by address.
func (s *Server) GetAccountHandler(w http.ResponseWriter, r *http.Request) {
	b := s.capnp.GetAccount(s.ctx, func(p capnp.PactusServer_getAccount_Params) error {
		vars := mux.Vars(r)
		return p.SetAddress(vars["address"])
	})

	a, err := b.Struct()
	if err != nil {
		s.writeError(w, err)
		return
	}

	res, _ := a.Result()
	d, _ := res.Data()
	acc, err := account.FromBytes(d)
	if err != nil {
		s.writeError(w, err)
		return
	}

	tm := newTableMaker()
	tm.addRowAccAddress("Address", acc.Address().String())
	tm.addRowInt("Number", int(acc.Number()))
	tm.addRowInt("Sequence", int(acc.Sequence()))
	tm.addRowAmount("Balance", acc.Balance())
	tm.addRowBytes("Hash", acc.Hash().Bytes())
	tm.addRowBytes("Data", d)

	s.writeHTML(w, tm.html())
}
