package byzcoin

/*
The `NewProtocol` method is used to define the protocol and to register
the handlers that will be called if a certain type of message is received.
The handlers will be treated according to their signature.

The protocol-file defines the actions that the protocol needs to do in each
step. The root-node will call the `Start`-method of the protocol. Each
node will only use the `Handle`-methods, and not call `Start` again.
*/

import (
	"go.dedis.ch/cothority/v3/skipchain"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"golang.org/x/xerrors"
)




func init() {
	_, err := onet.GlobalProtocolRegister(Name, NewProtocol)
	if err != nil {
		panic(err)
	}
}

// TemplateProtocol holds the state of a given protocol.
//
// For this example, it defines a channel that will receive the number
// of children. Only the root-node will write to the channel.
type RollupTxProtocol struct {
	*onet.TreeNodeInstance

	//Boilerplate code
	announceChan chan announceWrapper
	repliesChan  chan []replyWrapper
	ChildCount   chan int

	//From CollectTxProtocol
	TxsChan           chan []ClientTransaction
	CommonVersionChan chan Version
	SkipchainID       skipchain.SkipBlockID
	LatestID          skipchain.SkipBlockID
	MaxNumTxs         int
	getTxs            getTxsCallback
	Finish            chan bool
	closing           chan bool
	version           int


	//Create channel for txs in leader instead
	/*
	requestChan       chan structCollectTxRequest
	responseChan      chan structCollectTxResponse
	*/





}

// Check that *TemplateProtocol implements onet.ProtocolInstance
var _ onet.ProtocolInstance = (*RollupTxProtocol)(nil)

// NewProtocol initialises the structure for use in one round
func NewProtocol(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	t := &RollupTxProtocol{
		TreeNodeInstance: n,
		ChildCount:       make(chan int),
	}
	if err := n.RegisterChannels(&t.announceChan, &t.repliesChan); err != nil {
		return nil, err
	}
	return t, nil
}

// Start sends the Announce-message to all children
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
	log.LLvl1(p.ServerIdentity(), "STARTING txRollupProtocol")
	return p.SendTo(p.TreeNode(), &Announce{"cothority rulez!"})
}

// Dispatch implements the main logic of the protocol. The function is only
// called once. The protocol is considered finished when Dispatch returns and
// Done is called.
func (p *RollupTxProtocol) Dispatch() error {
	defer p.Done()
	ann := <-p.announceChan
	if p.IsLeaf() {
		return p.SendToParent(&Reply{1, "hello world"})
	}
	p.SendToChildren(&ann.Announce)


	replies := <-p.repliesChan
	n := 1
	for _, c := range replies {
		log.LLvl1("REPLY", c.Message)
		n += c.ChildrenCount
	}

	if !p.IsRoot() {
		return p.SendToParent(&Reply{n, ""})
	}

	p.ChildCount <- n
	return nil
}
