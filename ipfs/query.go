package ipfs

import (
	"context"
	"errors"
	peer "gx/ipfs/QmTRhk7cgjUf2gfQ3p2M9KPECNZEW9XUrmHcFCgog4cPgB/go-libp2p-peer"
	routing "gx/ipfs/Qmdfkd5HZgR2xc38TTb2afbM8nVHM8X1UowL5o7QFVb8uc/go-libp2p-kad-dht"

	"github.com/ipfs/go-ipfs/core"
)

func Query(n *core.IpfsNode, peerID string) ([]peer.ID, error) {
	dht, ok := n.Routing.(*routing.IpfsDHT)
	if !ok {
		return nil, errors.New("routing is not type IpfsDHT")
	}
	id, err := peer.IDB58Decode(peerID)
	if err != nil {
		return nil, err
	}

	ch, err := dht.GetClosestPeers(context.Background(), string(id))
	if err != nil {
		return nil, err
	}
	var closestPeers []peer.ID
	events := make(chan struct{})
	go func() {
		defer close(events)
		for p := range ch {
			closestPeers = append(closestPeers, p)
		}
	}()
	<-events
	return closestPeers, nil
}
