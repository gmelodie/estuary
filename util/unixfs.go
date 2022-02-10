package util

import (
	"errors"
	"io"

	"github.com/ipfs/go-cidutil"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs"
	"github.com/ipfs/go-unixfs/importer/balanced"
	ihelper "github.com/ipfs/go-unixfs/importer/helpers"
	mh "github.com/multiformats/go-multihash"
)

var DefaultHashFunction = uint64(mh.SHA2_256)

func ImportFile(dserv ipld.DAGService, fi io.Reader) (ipld.Node, error) {
	prefix, err := merkledag.PrefixForCidVersion(1)
	if err != nil {
		return nil, err
	}
	prefix.MhType = DefaultHashFunction

	spl := chunker.NewSizeSplitter(fi, 1024*1024)
	dbp := ihelper.DagBuilderParams{
		Maxlinks:  1024,
		RawLeaves: true,

		CidBuilder: cidutil.InlineBuilder{
			Builder: prefix,
			Limit:   32,
		},

		Dagserv: dserv,
	}

	db, err := dbp.New(spl)
	if err != nil {
		return nil, err
	}

	return balanced.Layout(db)
}

// TryExtractFSNode wraps around FSNodeFromBytes to cast a
// ipld node (ProtoNode) into an unixfs node
// Returns a unixfs node when successful
// Returns an error when the node type is not supported
func TryExtractFSNode(linkNode ipld.Node) (*unixfs.FSNode, error) {
	switch linkNode := linkNode.(type) {
	case *merkledag.ProtoNode:
		n, err := unixfs.FSNodeFromBytes(linkNode.Data())
		if err != nil {
			return nil, err
		}
		return n, nil // success!
	// case *merkledag.RawNode:
	// TODO: something here?
	default:
		return nil, errors.New("unsupported node type while extracting unixfs node")
	}
}
