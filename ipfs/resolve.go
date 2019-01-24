package ipfs

import (
	"context"
	"gx/ipfs/QmTRhk7cgjUf2gfQ3p2M9KPECNZEW9XUrmHcFCgog4cPgB/go-libp2p-peer"
	ds "gx/ipfs/QmaRb5yNXKonhbkpNxNawoydk4N6es6b4fPj19sjEKsh5D/go-datastore"
	"time"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/namesys"
	nameopts "github.com/ipfs/go-ipfs/namesys/opts"
	ipath "gx/ipfs/QmT3rzed1ppXefourpmoZ7tyVQfsGPQZ1pHDngLmCvXxd3/go-path"
)

// Resolve an IPNS record. This is a multi-step process.
// If the usecache flag is provided we will attempt to load the record from the database. If
// it succeeds we will update the cache in a separate goroutine.
//
// If we need to actually get a record from the network the IPNS namesystem will first check to
// see if it is subscribed to the name with pubsub. If so, it will return cache from the database.
// If not, it will subscribe to the name and proceed to a DHT query to find the record.
// If the DHT query returns nothing it will finally attempt to return from cache.
// All subsequent resolves will return from cache as the pubsub will update the cache in real time
// as new records are published.
func Resolve(n *core.IpfsNode, p peer.ID, timeout time.Duration, usecache bool) (string, error) {
	if usecache {
		pth, err := getFromDatastore(n.Repo.Datastore(), p)
		if err == nil {
			// Update the cache in background
			go func() {
				pth, err := resolve(n, p, timeout)
				if err != nil {
					return
				}
				if err := putToDatastore(n.Repo.Datastore(), p, pth); err != nil {
					log.Error("Error putting IPNS record to datastore: %s", err.Error())
				}
			}()
			return pth.Segments()[1], nil
		}
	}
	pth, err := resolve(n, p, timeout)
	if err != nil {
		// Resolving fail. See if we have it in the db.
		pth, err := getFromDatastore(n.Repo.Datastore(), p)
		if err != nil {
			return "", err
		}
		return pth.Segments()[1], nil
	}
	// Resolving succeeded. Update the cache.
	if err := putToDatastore(n.Repo.Datastore(), p, pth); err != nil {
		log.Error("Error putting IPNS record to datastore: %s", err.Error())
	}
	return pth.Segments()[1], nil
}

func resolve(n *core.IpfsNode, p peer.ID, timeout time.Duration) (ipath.Path, error) {
	cctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// TODO [cp]: we should load the record count from our config and set it here. We'll need a
	// migration for this.
	pth, err := n.Namesys.Resolve(cctx, "/ipns/"+p.Pretty(), nameopts.DhtRecordCount(1))
	if err != nil {
		return pth, err
	}
	return pth, nil
}

func ResolveAltRoot(n *core.IpfsNode, p peer.ID, altRoot string, timeout time.Duration) (string, error) {
	cctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pth, err := n.Namesys.Resolve(cctx, "/ipns/"+p.Pretty()+":"+altRoot)
	if err != nil {
		return "", err
	}
	return pth.Segments()[1], nil
}

func getFromDatastore(datastore ds.Datastore, p peer.ID) (ipath.Path, error) {
	// resolve to what we may already have in the datastore
	data, err := datastore.Get(namesys.IpnsDsKey(p))
	if err != nil {
		if err == ds.ErrNotFound {
			return "", namesys.ErrResolveFailed
		}
		return "", err
	}
	return ipath.ParsePath(string(data))
}

func putToDatastore(datastore ds.Datastore, p peer.ID, pth ipath.Path) error {
	return datastore.Put(namesys.IpnsDsKey(p), []byte(pth.String()))
}
