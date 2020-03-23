package byzcoin

import (
	"time"

	"go.dedis.ch/cothority/v3/skipchain"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
	"golang.org/x/xerrors"
)



func init() {
	network.RegisterMessages(CollectTxRequest{}, CollectTxResponse{})
}

// CollectTxProtocol is a protocol for collecting pending transactions.
type RollupTxProtocol struct {
	*onet.TreeNodeInstance
	TxsChan           chan []ClientTransaction
	CommonVersionChan chan Version
	SkipchainID       skipchain.SkipBlockID
	LatestID          skipchain.SkipBlockID
	MaxNumTxs         int

	getTxs            getTxsCallback
	Finish            chan bool
	closing           chan bool
	version           int

	requestChan       chan structCollectTxRequest
	responseChan      chan structCollectTxResponse

}

// CollectTxRequest is the request message that asks the receiver to send their
// pending transactions back to the leader.
type RollupTxRequest struct {
	SkipchainID skipchain.SkipBlockID
	LatestID    skipchain.SkipBlockID
	MaxNumTxs   int
	Version     int
}

// CollectTxResponse is the response message that contains all the pending
// transactions on the node.
type RollupTxResponse struct {
	Txs            []ClientTransaction
	ByzcoinVersion Version
}

type structRollupTxRequest struct {
	*onet.TreeNode
	CollectTxRequest
}

type structRollupTxResponse struct {
	*onet.TreeNode
	CollectTxResponse
}

// NewCollectTxProtocol is used for registering the protocol.
func NewRollupTxProtocol(getTxs getTxsCallback) func(*onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	return func(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
		c := &RollupTxProtocol{
			TreeNodeInstance: node,
			// If we do not buffer this channel then the protocol
			// might be blocked from stopping when the receiver
			// stops reading from this channel.
			TxsChan:           make(chan []ClientTransaction, len(node.List())),
			CommonVersionChan: make(chan Version, len(node.List())),
			MaxNumTxs:         defaultMaxNumTxs,
			getTxs:            getTxs,
			Finish:            make(chan bool),
			closing:           make(chan bool),
			version:           1,
		}
		if err := node.RegisterChannels(&c.requestChan, &c.responseChan); err != nil {
			return c, xerrors.Errorf("registering channels: %v", err)
		}
		return c, nil
	}
}

// Start starts the protocol, it should only be called on the root node.
func (p *RollupTxProtocol) Start() error {
	if !p.IsRoot() {
		return xerrors.New("only the root should call start")
	}
	if len(p.SkipchainID) == 0 {
		return xerrors.New("missing skipchain ID")
	}
	if len(p.LatestID) == 0 {
		return xerrors.New("missing latest skipblock ID")
	}
	log.LLvl1(p.ServerIdentity(), "starting rollup tx protocol")

	req := &CollectTxRequest{
		SkipchainID: p.SkipchainID,
		LatestID:    p.LatestID,
		MaxNumTxs:   p.MaxNumTxs,
		Version:     p.version,
	}
	// send to myself and the children
	if err := p.SendTo(p.TreeNode(), req); err != nil {
		return xerrors.Errorf("sending msg: %v", err)
	}
	// do not return an error if we fail to send to some children
	if errs := p.SendToChildrenInParallel(req); len(errs) > 0 {
		for _, err := range errs {
			log.Error(p.ServerIdentity(), err)
		}
	}
	return nil
}

// Dispatch runs the protocol.
func (p *RollupTxProtocol) Dispatch() error {
	defer p.Done()

	var req structCollectTxRequest
	select {
	case req = <-p.requestChan:
	case <-p.Finish:
		return nil
	case <-time.After(time.Second):
		// This timeout checks whether the root started the protocol,
		// it is not like our usual timeout that detect failures.
		//return xerrors.New("did not receive request")
	case <-p.closing:
		return xerrors.New("closing down system")
	}



	maxOut := -1
	if req.Version >= 1 {
		// Leader with older version will send a maximum value of 0 which
		// is the default value as the field is unknown.
		maxOut = req.MaxNumTxs
	}

	//TODO : should work for all skipchain id and not only the one set by the protocol. How to get the skipchain id without getting it from the request message?
	//TODO : how to get the last block hash without the request? Use the service from the tx processor?
	//TODO: send the result of the callback to the root

	resp := &CollectTxResponse{
		Txs:            p.getTxs(req.ServerIdentity, p.Roster(), req.SkipchainID, req.LatestID, maxOut),
		ByzcoinVersion: p.getByzcoinVersion(),
	}
	log.LLvl1(p.ServerIdentity(), "sends back", len(resp.Txs), "transactions")
	if p.IsRoot() {
		if err := p.SendTo(p.TreeNode(), resp); err != nil {
			return xerrors.Errorf("sending msg: %v", err)
		}
	} else {
		if err := p.SendToParent(resp); err != nil {
			return xerrors.Errorf("sending msg: %v", err)
		}
	}

	// wait for the results to come back and write to the channel
	defer close(p.TxsChan)
	if p.IsRoot() {
		vb := newVersionBuffer(len(p.Children()) + 1)

		leaderVersion := p.getByzcoinVersion()
		vb.add(p.ServerIdentity(), leaderVersion)

		finish := false

		for i := 0; i < len(p.List()) && !finish; i++ {
			select {
			case resp := <-p.responseChan:
				vb.add(resp.ServerIdentity, resp.ByzcoinVersion)

				// If more than the limit is sent, we simply drop all of them
				// as the conode is not behaving correctly.
				if p.version == 0 || len(resp.Txs) <= p.MaxNumTxs {
					p.TxsChan <- resp.Txs
				}
			case <-p.Finish:
				finish = true
			case <-p.closing:
				finish = true
			}
		}

		if vb.hasThresholdFor(leaderVersion) {
			p.CommonVersionChan <- leaderVersion
		}
	}
	return nil
}

// Shutdown closes the closing channel to abort any waiting on messages.
func (p *RollupTxProtocol) Shutdown() error {
	close(p.closing)
	return nil
}

func (p *RollupTxProtocol) getByzcoinVersion() Version {
	srv := p.Host().Service(ServiceName)
	if srv == nil {
		panic("Byzcoin should always be available as a service for this protocol")
	}

	return srv.(*Service).GetProtocolVersion()
}

