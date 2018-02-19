package omnicon

import (
	"fmt"
	"time"

	"github.com/dedis/cothority/cosi/protocol"
	"github.com/dedis/kyber"
	"github.com/dedis/kyber/sign/cosi"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
)

// ProtocolBFTCoSi contains the state used in the execution of the BFTCoSi
// protocol. It is also known as OmniCon, which is described in the OmniLedger
// paper - https://eprint.iacr.org/2017/406
type ProtocolBFTCoSi struct {
	// the node we are represented-in
	*onet.TreeNodeInstance
	// Msg is the message that will be signed by cosigners
	Msg []byte
	// Data is used for verification only, not signed
	Data []byte
	// FinalSignature is output of the protocol, for the caller to read
	FinalSignatureChan chan FinalSignature
	// CreateProtocol TODO
	CreateProtocol protocol.CreateProtocolFunction
	// Timeout define the timeout duration
	Timeout time.Duration
	// prepCosiProtoName is the cosi protocol name for the prepare phase
	prepCosiProtoName string
	// commitCosiProtoName is the cosi protocol name for the commit phase
	commitCosiProtoName string
	// prepSigChan is the channel for reading the prepare phase signature
	prepSigChan chan []byte
	// publics
	publics []kyber.Point
}

// FinalSignature holds the message Msg and its signature
type FinalSignature struct {
	Msg []byte
	Sig []byte
}

type phase int

const (
	phasePrep phase = iota
	phaseCommit
)

// Start begins the BFTCoSi protocol by starting the prepare cosi.
func (bft *ProtocolBFTCoSi) Start() error {
	// prepare phase (part 1)
	log.Lvl3("Starting prepare phase")
	prepProto, err := bft.initCosiProtocol(phasePrep)
	if err != nil {
		return err
	}

	err = prepProto.Start()
	if err != nil {
		return err
	}

	go func() {
		bft.prepSigChan <- <-prepProto.FinalSignature
	}()

	return nil
}

func (bft *ProtocolBFTCoSi) initCosiProtocol(phase phase) (*protocol.CoSiRootNode, error) {
	var name string
	if phase == phasePrep {
		name = bft.prepCosiProtoName
	} else if phase == phaseCommit {
		name = bft.commitCosiProtoName
	} else {
		return nil, fmt.Errorf("invalid phase %v", phase)
	}

	pi, err := bft.CreateProtocol(name, bft.Tree())
	if err != nil {
		return nil, err
	}
	cosiProto := pi.(*protocol.CoSiRootNode)
	cosiProto.CreateProtocol = bft.CreateProtocol
	// We set it to n / 10 to have every sub-leader manage 10 nodes.
	// This setting is bad if there are thousands of nodes as the root
	// would need to manage hundres of sub-leaders.
	cosiProto.NSubtrees = len(bft.List()) / 10
	cosiProto.Msg = bft.Msg
	cosiProto.Data = bft.Data

	return cosiProto, nil
}

// Dispatch is the main logic of the BFTCoSi protocol. It runs two CoSi
// protocols as the prepare and the commit phase of PBFT. Concretely, it does:
// 1, wait for the prepare phase to finish
// 2, check the signature
// 3, if it is, start the commit phase,
//    otherwise send an empty signature
// 4, wait for the commit phase to finish
// 5, send the final signature
func (bft *ProtocolBFTCoSi) Dispatch() error {

	if !bft.IsRoot() {
		return fmt.Errorf("non-root should not start this protocol")
	}

	// prepare phase (part 2)
	prepSig := <-bft.prepSigChan
	suite := bft.Suite().(cosi.Suite)
	nbrFault := (len(bft.List())-1)/3 - 1
	if nbrFault < 0 {
		nbrFault = 0
	}
	err := cosi.Verify(suite, bft.publics, bft.Msg, prepSig, cosi.NewThresholdPolicy(nbrFault))
	if err != nil {
		log.Lvl2("Signature verification failed on root during the prepare phase with error:", err)
		bft.FinalSignatureChan <- FinalSignature{nil, nil}
		return nil
	}
	log.Lvl3("Finished prepare phase")

	// commit phase
	log.Lvl3("Starting commit phase")
	commitProto, err := bft.initCosiProtocol(phaseCommit)
	if err != nil {
		return err
	}

	err = commitProto.Start()
	if err != nil {
		return err
	}

	commitSig := <-commitProto.FinalSignature
	log.Lvl3("Finished commit phase")

	bft.FinalSignatureChan <- FinalSignature{bft.Msg, commitSig}
	return nil
}

// NewBFTCoSiProtocol TODO
func NewBFTCoSiProtocol(n *onet.TreeNodeInstance, prepCosiProtoName, commitCosiProtoName string) (*ProtocolBFTCoSi, error) {
	publics := make([]kyber.Point, n.Tree().Size())
	for i, node := range n.Tree().List() {
		publics[i] = node.ServerIdentity.Public
	}
	return &ProtocolBFTCoSi{
		TreeNodeInstance: n,
		// we do not have Msg to make the protocol fail if it's not set
		Data: make([]byte, 0),
		// the caller also needs to make FinalSignatureChan
		prepCosiProtoName:   prepCosiProtoName,
		commitCosiProtoName: commitCosiProtoName,
		Timeout:             protocol.DefaultProtocolTimeout * 2, // TODO not used
		prepSigChan:         make(chan []byte, 0),
		publics:             publics,
	}, nil
}

func makeProtocols(vf, ack protocol.VerificationFn, protoName string) map[string]onet.NewProtocol {

	protocolMap := make(map[string]onet.NewProtocol)

	prepCosiProtoName := protoName + "_cosi_prep"
	prepCosiSubProtoName := protoName + "_subcosi_prep"
	commitCosiProtoName := protoName + "_cosi_commit"
	commitCosiSubProtoName := protoName + "_subcosi_commit"

	bftProto := func(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
		return NewBFTCoSiProtocol(n, prepCosiProtoName, commitCosiProtoName)
	}
	protocolMap[protoName] = bftProto

	prepCosiProto := func(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
		return protocol.NewProtocol(n, vf, prepCosiSubProtoName)
	}
	protocolMap[prepCosiProtoName] = prepCosiProto

	prepCosiSubProto := func(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
		return protocol.NewSubProtocol(n, vf)
	}
	protocolMap[prepCosiSubProtoName] = prepCosiSubProto

	commitCosiProto := func(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
		return protocol.NewProtocol(n, ack, commitCosiSubProtoName)
	}
	protocolMap[commitCosiProtoName] = commitCosiProto

	commitCosiSubProto := func(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
		return protocol.NewSubProtocol(n, ack)
	}
	protocolMap[commitCosiSubProtoName] = commitCosiSubProto

	return protocolMap
}

// GlobalInitBFTCoSiProtocol creates and registers the protocols required to run
// BFTCoSi globally.
func GlobalInitBFTCoSiProtocol(vf, ack protocol.VerificationFn, protoName string) error {
	protocolMap := makeProtocols(vf, ack, protoName)
	for protoName, proto := range protocolMap {
		if _, err := onet.GlobalProtocolRegister(protoName, proto); err != nil {
			return err
		}
	}
	return nil
}

// InitBFTCoSiProtocol creates and registers the protocols required to run
// BFTCoSi to the context c.
func InitBFTCoSiProtocol(c *onet.Context, vf, ack protocol.VerificationFn, protoName string) error {
	protocolMap := makeProtocols(vf, ack, protoName)
	for protoName, proto := range protocolMap {
		if _, err := c.ProtocolRegister(protoName, proto); err != nil {
			return err
		}
	}
	return nil
}
