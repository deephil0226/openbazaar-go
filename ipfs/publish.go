package ipfs

import (
	"context"
	"errors"

	"github.com/ipfs/go-ipfs/core"
	ipath "gx/ipfs/QmT3rzed1ppXefourpmoZ7tyVQfsGPQZ1pHDngLmCvXxd3/go-path"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("ipfs")

var pubErr = errors.New(`Name publish failed`)

// Publish a signed IPNS record to our Peer ID
func Publish(n *core.IpfsNode, hash string) error {
	err := n.Namesys.Publish(context.Background(), n.PrivateKey, ipath.FromString("/ipfs/"+hash))
	if err == nil {
		log.Infof("Published %s to IPNS", hash)
		return nil
	} else {
		return pubErr
	}
}
