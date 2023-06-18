package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/switfs/filwallet/build"
	"github.com/switfs/filwallet/client"
)

// Status Get
func (w *Wallet) Status(c *gin.Context) {
	ReturnOk(c, client.StatusInfo{
		Lock:    w.lock,
		Offline: w.offline,
		Version: build.Version(),
	})
}
