package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/switfs/filwallet/chain"
	"github.com/switfs/filwallet/client"
	"github.com/switfs/filwallet/modules/buildmessage"
)

// Transfer Post
func (w *Wallet) Transfer(c *gin.Context) {
	param := client.TransferRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("Transfer: BindJSON", "err", err)
		ReturnError(c, ParamErr)
		return
	}

	msg, err := buildmessage.NewTransferMessage(w.node.Api, param.BaseParams, param.From, param.To, param.Amount)
	if err != nil {
		log.Warnw("Transfer: NewTransferMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	myMsg, err := chain.EncodeMessage(msg, nil)
	if err != nil {
		log.Warnw("Transfer: EncodeMessage", "err", err)
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, myMsg)
	return
}
