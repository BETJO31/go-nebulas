// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/nebulasio/go-nebulas/common/trie"
	"github.com/nebulasio/go-nebulas/core/state"
	"github.com/nebulasio/go-nebulas/net"
	"github.com/nebulasio/go-nebulas/util/byteutils"

	pb "github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/neblet/pb"
	"github.com/nebulasio/go-nebulas/storage"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/logging"
	"github.com/stretchr/testify/assert"
)

const (
	BlockInterval        = 5
	AcceptedNetWorkDelay = 2
)

var (
	stor, _ = storage.NewMemoryStorage()
)

var (
	MockDynasty = []string{
		"1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
		"2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
		"333cb3ed8c417971845382ede3cf67a0a96270c05fe2f700",
		"48f981ed38910f1232c1bab124f650c482a57271632db9e3",
		"59fc526072b09af8a8ca9732dae17132c4e9127e43cf2232",
		"75e4e5a71d647298b88928d8cb5da43d90ab1a6c52d0905f",
		"7da9dabedb4c6e121146fb4250a9883d6180570e63d6b080",
		"98a3eed687640b75ec55bf5c9e284371bdcaeab943524d51",
		"a8f1f53952c535c6600c77cf92b65e0c9b64496a8a328569",
		"b040353ec0f2c113d5639444f7253681aecda1f8b91f179f",
		"b414432e15f21237013017fa6ee90fc99433dec82c1c8370",
		"b49f30d0e5c9c88cade54cd1adecf6bc2c7e0e5af646d903",
		"b7d83b44a3719720ec54cdb9f54c0202de68f1ebcb927b4f",
		"ba56cc452e450551b7b9cffe25084a069e8c1e94412aad22",
		"c5bcfcb3fa8250be4f2bf2b1e70e1da500c668377ba8cd4a",
		"c79d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe13a833d",
		"d1abde197e97398864ba74511f02832726edad596775420a",
		"d86f99d97a394fa7a623fdf84fdc7446b99c3cb335fca4bf",
		"e0f78b011e639ce6d8b76f97712118f3fe4a12dd954eba49",
		"f38db3b6c801dddd624d6ddc2088aa64b5a24936619e4848",
		"fc751b484bd5296f8d267a8537d33f25a848f7f7af8cfcf6",
	}
)

// MockGenesisConf return mock genesis conf
func MockGenesisConf() *corepb.Genesis {
	return &corepb.Genesis{
		Meta: &corepb.GenesisMeta{ChainId: 100},
		Consensus: &corepb.GenesisConsensus{
			Dpos: &corepb.GenesisConsensusDpos{
				Dynasty: MockDynasty,
			},
		},
		TokenDistribution: []*corepb.GenesisTokenDistribution{
			&corepb.GenesisTokenDistribution{
				Address: "1a263547d167c74cf4b8f9166cfa244de0481c514a45aa2c",
				Value:   "10000000000000000000000",
			},
			&corepb.GenesisTokenDistribution{
				Address: "2fe3f9f51f9a05dd5f7c5329127f7c917917149b4e16b0b8",
				Value:   "10000000000000000000000",
			},
		},
	}
}

type mockConsensusState struct {
	votes      *trie.BatchTrie
	delegates  *trie.BatchTrie
	candidates *trie.BatchTrie
}

func newMockConsensusState() (*mockConsensusState, error) {
	votes, err := trie.NewBatchTrie(nil, stor)
	if err != nil {
		return nil, err
	}
	delegates, err := trie.NewBatchTrie(nil, stor)
	if err != nil {
		return nil, err
	}
	candidates, err := trie.NewBatchTrie(nil, stor)
	if err != nil {
		return nil, err
	}
	return &mockConsensusState{
		votes:      votes,
		delegates:  delegates,
		candidates: candidates,
	}, nil
}

func (cs *mockConsensusState) BeginBatch() {}
func (cs *mockConsensusState) Commit()     {}
func (cs *mockConsensusState) RollBack()   {}

func (cs *mockConsensusState) RootHash() (byteutils.Hash, error)    { return nil, nil }
func (cs *mockConsensusState) Clone() (state.ConsensusState, error) { return cs, nil }

func (cs *mockConsensusState) Proposer() byteutils.Hash { return nil }
func (cs *mockConsensusState) TimeStamp() int64         { return 0 }
func (cs *mockConsensusState) NextConsensusState(int64, state.WorldState) (state.ConsensusState, error) {
	return cs, nil
}

func (cs *mockConsensusState) GetMintCnt(int64, byteutils.Hash) (int64, error) { return 0, nil }
func (cs *mockConsensusState) PutMintCnt(int64, byteutils.Hash, int64) error   { return nil }

func (cs *mockConsensusState) CandidatesRoot() byteutils.Hash {
	return cs.candidates.RootHash()
}
func (cs *mockConsensusState) Candidates() ([]byteutils.Hash, error) {
	members := []byteutils.Hash{}
	iter, err := cs.candidates.Iterator(nil)
	if err != nil && err != storage.ErrKeyNotFound {
		return nil, err
	}
	if err != nil {
		return members, nil
	}
	exist, err := iter.Next()
	for exist {
		members = append(members, iter.Value())
		exist, err = iter.Next()
	}
	return members, nil
}
func (cs *mockConsensusState) HasCandidate(candidate byteutils.Hash) (bool, error) {
	_, err := cs.candidates.Get(candidate)
	if err != nil && err != storage.ErrKeyNotFound {
		return false, err
	}
	if err == storage.ErrKeyNotFound {
		return false, nil
	}
	return true, nil
}
func (cs *mockConsensusState) AddCandidate(candidate byteutils.Hash) error {
	_, err := cs.candidates.Put(candidate, candidate)
	return err
}
func (cs *mockConsensusState) DelCandidate(candidate byteutils.Hash) error {
	_, err := cs.candidates.Del(candidate)
	return err
}

func (cs *mockConsensusState) GetVote(voter byteutils.Hash) (byteutils.Hash, error) {
	return cs.votes.Get(voter)
}
func (cs *mockConsensusState) AddVote(voter byteutils.Hash, votee byteutils.Hash) error {
	_, err := cs.votes.Put(voter, votee)
	return err
}
func (cs *mockConsensusState) DelVote(voter byteutils.Hash) error {
	_, err := cs.votes.Del(voter)
	return err
}
func (cs *mockConsensusState) IterVote() (state.Iterator, error) {
	return cs.votes.Iterator(nil)
}

func (cs *mockConsensusState) HasDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) (bool, error) {
	key := append(delegatee, delegator...)
	_, err := cs.delegates.Get(key)
	if err != nil && err != storage.ErrKeyNotFound {
		return false, err
	}
	if err == storage.ErrKeyNotFound {
		return false, nil
	}
	return true, nil
}
func (cs *mockConsensusState) AddDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	key := append(delegatee, delegator...)
	_, err := cs.delegates.Put(key, delegator)
	return err
}
func (cs *mockConsensusState) DelDelegate(delegator byteutils.Hash, delegatee byteutils.Hash) error {
	key := append(delegatee, delegator...)
	_, err := cs.delegates.Del(key)
	return err
}
func (cs *mockConsensusState) IterDelegate(delegatee byteutils.Hash) (state.Iterator, error) {
	return cs.delegates.Iterator(delegatee)
}

func (cs *mockConsensusState) Dynasty() ([]byteutils.Hash, error)     { return nil, nil }
func (cs *mockConsensusState) DynastyRoot() byteutils.Hash            { return nil }
func (cs *mockConsensusState) NextDynasty() ([]byteutils.Hash, error) { return nil, nil }
func (cs *mockConsensusState) NextDynastyRoot() byteutils.Hash        { return nil }

type mockConsensus struct {
	chain *BlockChain
}

func (c mockConsensus) Setup(neb Neblet) error {
	c.chain = neb.BlockChain()
	return nil
}

func (c mockConsensus) Start() {}
func (c mockConsensus) Stop()  {}

func (c mockConsensus) FastVerifyBlock(block *Block) error {
	block.miner = block.Coinbase()
	return nil
}

func (c mockConsensus) VerifyBlock(block *Block, parent *Block) error {
	block.miner = block.Coinbase()
	return nil
}

func (c mockConsensus) ForkChoice() error {
	return nil
}

func (c mockConsensus) UpdateLIB() {}

func (c mockConsensus) SuspendMining() {}
func (c mockConsensus) ResumeMining()  {}
func (c mockConsensus) Pending() bool  { return false }

func (c mockConsensus) EnableMining(passphrase string) error { return nil }
func (c mockConsensus) DisableMining() error                 { return nil }
func (c mockConsensus) Enable() bool                         { return true }

func (c mockConsensus) CheckTimeout(block *Block) bool {
	return time.Now().Unix()-block.Timestamp() > AcceptedNetWorkDelay
}
func (c mockConsensus) NewState(byteutils.Hash, storage.Storage) (state.ConsensusState, error) {
	return newMockConsensusState()
}
func (c mockConsensus) GenesisConsensusState(*BlockChain, *corepb.Genesis) (state.ConsensusState, error) {
	return newMockConsensusState()
}

type mockManager struct{}

func (m mockManager) NewAccount([]byte) (*Address, error) { return nil, nil }
func (m mockManager) Accounts() []*Address                { return nil }

func (m mockManager) Unlock(addr *Address, passphrase []byte, expire time.Duration) error { return nil }
func (m mockManager) Lock(addr *Address) error                                            { return nil }

func (m mockManager) SignBlock(addr *Address, block *Block) error                        { return nil }
func (m mockManager) SignTransaction(*Address, *Transaction) error                       { return nil }
func (m mockManager) SignTransactionWithPassphrase(*Address, *Transaction, []byte) error { return nil }

func (m mockManager) Update(*Address, []byte, []byte) error   { return nil }
func (m mockManager) Load([]byte, []byte) (*Address, error)   { return nil, nil }
func (m mockManager) Import([]byte, []byte) (*Address, error) { return nil, nil }
func (m mockManager) Delete(*Address, []byte) error           { return nil }

var (
	received = []byte{}
)

type mockNetService struct{}

func (n mockNetService) Start() error { return nil }
func (n mockNetService) Stop()        {}

func (n mockNetService) Node() *net.Node { return nil }

func (n mockNetService) Sync(net.Serializable) error { return nil }

func (n mockNetService) Register(...*net.Subscriber)   {}
func (n mockNetService) Deregister(...*net.Subscriber) {}

func (n mockNetService) Broadcast(name string, msg net.Serializable, priority int) {}
func (n mockNetService) Relay(name string, msg net.Serializable, priority int)     {}
func (n mockNetService) SendMsg(name string, msg []byte, target string, priority int) error {
	received = msg
	return nil
}

func (n mockNetService) SendMessageToPeers(messageName string, data []byte, priority int, filter net.PeerFilterAlgorithm) []string {
	return make([]string, 0)
}
func (n mockNetService) SendMessageToPeer(messageName string, data []byte, priority int, peerID string) error {
	return nil
}

func (n mockNetService) ClosePeer(peerID string, reason error) {}

func (n mockNetService) BroadcastNetworkID([]byte) {}

func (n mockNetService) BuildRawMessageData([]byte, string) []byte { return nil }

type mockNeb struct {
	config    *nebletpb.Config
	chain     *BlockChain
	ns        net.Service
	am        Manager
	genesis   *corepb.Genesis
	storage   storage.Storage
	consensus Consensus
	emitter   *EventEmitter
}

func (n *mockNeb) Genesis() *corepb.Genesis {
	return n.genesis
}

func (n *mockNeb) Config() *nebletpb.Config {
	return n.config
}

func (n *mockNeb) Storage() storage.Storage {
	return n.storage
}

func (n *mockNeb) EventEmitter() *EventEmitter {
	return n.emitter
}

func (n *mockNeb) Consensus() Consensus {
	return n.consensus
}

func (n *mockNeb) BlockChain() *BlockChain {
	return n.chain
}

func (n *mockNeb) NetService() net.Service {
	return n.ns
}

func (n *mockNeb) AccountManager() Manager {
	return n.am
}

func (n *mockNeb) StartActiveSync() {}

func testNeb(t *testing.T) *mockNeb {
	storage, _ := storage.NewMemoryStorage()
	eventEmitter := NewEventEmitter(1024)
	var consensus mockConsensus
	var am mockManager
	var ns mockNetService
	neb := &mockNeb{
		genesis:   MockGenesisConf(),
		config:    &nebletpb.Config{Chain: &nebletpb.ChainConfig{ChainId: MockGenesisConf().Meta.ChainId}},
		storage:   storage,
		emitter:   eventEmitter,
		consensus: consensus,
		am:        am,
		ns:        ns,
	}
	chain, err := NewBlockChain(neb)
	assert.Nil(t, err)
	chain.bkPool.RegisterInNetwork(ns)
	neb.chain = chain
	consensus.Setup(neb)
	chain.Setup(neb)
	return neb
}

func TestBlock(t *testing.T) {
	type fields struct {
		header       *BlockHeader
		miner        *Address
		height       uint64
		transactions Transactions
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"full struct",
			fields{
				&BlockHeader{
					hash:          []byte("124546"),
					parentHash:    []byte("344543"),
					stateRoot:     []byte("43656"),
					txsRoot:       []byte("43656"),
					eventsRoot:    []byte("43656"),
					consensusRoot: []byte("43656"),
					nonce:         3546456,
					coinbase:      &Address{[]byte("hello")},
					timestamp:     time.Now().Unix(),
					chainID:       100,
				},
				&Address{[]byte("hello")},
				1,
				Transactions{
					&Transaction{
						[]byte("123452"),
						&Address{[]byte("1335")},
						&Address{[]byte("1245")},
						util.NewUint128(),
						456,
						1516464510,
						&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hello")},
						1,
						util.NewUint128(),
						util.NewUint128(),
						uint8(keystore.SECP256K1),
						nil,
					},
					&Transaction{
						[]byte("123455"),
						&Address{[]byte("1235")},
						&Address{[]byte("1425")},
						util.NewUint128(),
						446,
						1516464511,
						&corepb.Data{Type: TxPayloadBinaryType, Payload: []byte("hllo")},
						2,
						util.NewUint128(),
						util.NewUint128(),
						uint8(keystore.SECP256K1),
						nil,
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				header:       tt.fields.header,
				miner:        tt.fields.miner,
				height:       tt.fields.height,
				transactions: tt.fields.transactions,
			}
			proto, _ := b.ToProto()
			ir, _ := pb.Marshal(proto)
			nb := new(Block)
			pb.Unmarshal(ir, proto)
			nb.FromProto(proto)
			b.header.timestamp = nb.header.timestamp
			if !reflect.DeepEqual(*b.header, *nb.header) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.header, *nb.header)
			}
			if !reflect.DeepEqual(*b.transactions[0], *nb.transactions[0]) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.transactions[0], *nb.transactions[0])
			}
			if !reflect.DeepEqual(*b.transactions[1], *nb.transactions[1]) {
				t.Errorf("Transaction.Serialize() = %v, want %v", *b.transactions[1], *nb.transactions[1])
			}
		})
	}
}

func TestBlock_LinkParentBlock(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	genesis := bc.genesisBlock
	assert.Equal(t, genesis.Height(), uint64(1))
	block1 := &Block{
		header: &BlockHeader{
			hash:          []byte("124546"),
			parentHash:    GenesisHash,
			stateRoot:     []byte("43656"),
			txsRoot:       []byte("43656"),
			eventsRoot:    []byte("43656"),
			consensusRoot: []byte("43656"),
			nonce:         3546456,
			coinbase:      &Address{[]byte("hello")},
			timestamp:     BlockInterval,
			chainID:       100,
		},
		transactions: []*Transaction{},
	}
	assert.Equal(t, block1.Height(), uint64(0))
	assert.Equal(t, block1.LinkParentBlock(bc, genesis), nil)
	assert.Equal(t, block1.Height(), uint64(2))
	assert.Equal(t, block1.ParentHash(), genesis.Hash())
	block2 := &Block{
		header: &BlockHeader{
			hash:          []byte("124546"),
			parentHash:    []byte("344543"),
			stateRoot:     []byte("43656"),
			txsRoot:       []byte("43656"),
			eventsRoot:    []byte("43656"),
			consensusRoot: []byte("43656"),
			nonce:         3546456,
			coinbase:      &Address{[]byte("hello")},
			timestamp:     BlockInterval * 2,
			chainID:       100,
		},
		transactions: []*Transaction{},
	}
	assert.Equal(t, block2.LinkParentBlock(bc, genesis), ErrLinkToWrongParentBlock)
	assert.Equal(t, block2.Height(), uint64(0))
}

func TestBlock_CollectTransactions(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain

	tail := bc.tailBlock

	ks := keystore.DefaultKS
	priv := secp256k1.GeneratePrivateKey()
	pubdata, _ := priv.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata)
	ks.SetKey(from.String(), priv, []byte("passphrase"))
	ks.Unlock(from.String(), []byte("passphrase"), time.Second*60*60*24*365)

	key, _ := ks.GetUnlocked(from.String())
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signature.InitSign(key.(keystore.PrivateKey))

	priv1 := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	to, _ := NewAddressFromPublicKey(pubdata1)
	priv2 := secp256k1.GeneratePrivateKey()
	pubdata2, _ := priv2.PublicKey().Encoded()
	coinbase, _ := NewAddressFromPublicKey(pubdata2)

	block0, _ := NewBlock(bc.ChainID(), from, tail)
	block0.header.timestamp = BlockInterval
	block0.SetMiner(from)
	block0.Seal()
	//bc.BlockPool().push(block0)
	bc.SetTailBlock(block0)

	block, _ := NewBlock(bc.ChainID(), coinbase, block0)
	block.header.timestamp = BlockInterval * 2

	tx1 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	tx3 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 0, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx3.Sign(signature)
	tx4 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 4, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx4.Sign(signature)
	tx5 := NewTransaction(bc.ChainID(), from, to, util.NewUint128FromInt(1), 3, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx5.Sign(signature)
	tx6 := NewTransaction(bc.ChainID()+1, from, to, util.NewUint128FromInt(1), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx6.Sign(signature)

	assert.Nil(t, bc.txPool.Push(tx1))
	assert.Nil(t, bc.txPool.Push(tx2))
	assert.Nil(t, bc.txPool.Push(tx3))
	assert.Nil(t, bc.txPool.Push(tx4))
	assert.Nil(t, bc.txPool.Push(tx5))
	assert.NotNil(t, bc.txPool.Push(tx6), ErrInvalidChainID)

	assert.Equal(t, len(block.transactions), 0)
	assert.Equal(t, bc.txPool.cache.Len(), 5)
	block.CollectTransactions(time.Now().Unix() + 2)
	assert.Equal(t, len(block.transactions), 4)
	assert.Equal(t, block.txPool.cache.Len(), 0)

	assert.Equal(t, block.Sealed(), false)
	balance, err := block.GetBalance(block.header.coinbase.address)
	assert.Nil(t, err)
	assert.Equal(t, balance.Cmp(util.NewUint128().Int), 1)
	block.SetMiner(coinbase)
	block.Seal()
	assert.Equal(t, block.Sealed(), true)
	assert.Equal(t, block.transactions[0], tx1)
	assert.Equal(t, block.transactions[1], tx2)
	accStateRoot, err := block.worldState.AccountsRoot()
	assert.Nil(t, err)
	txsRoot, err := block.worldState.TxsRoot()
	assert.Nil(t, err)
	assert.Equal(t, block.StateRoot().Equals(accStateRoot), true)
	assert.Equal(t, block.TxsRoot().Equals(txsRoot), true)
	balance, err = block.GetBalance(block.header.coinbase.address)
	assert.Nil(t, err)
	// balance > BlockReward (BlockReward + gas)
	//gas, _ := bc.EstimateGas(tx1)
	logging.CLog().Info(balance.String())
	logging.CLog().Info(BlockReward.String())
	assert.NotEqual(t, balance.Cmp(BlockReward.Int), 0)
	// mock net message
	block, _ = mockBlockFromNetwork(block)
	assert.Equal(t, block.LinkParentBlock(bc, bc.tailBlock), nil)
	block.SetMiner(coinbase)
	assert.Nil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
}

func TestBlock_fetchEvents(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain

	tail := bc.tailBlock
	events := []*state.Event{
		&state.Event{Topic: "chain.block", Data: "hello"},
		&state.Event{Topic: "chain.tx", Data: "hello"},
		&state.Event{Topic: "chain.block", Data: "hello"},
		&state.Event{Topic: "chain.block", Data: "hello"},
	}
	tx := &Transaction{hash: []byte("tx")}
	for _, event := range events {
		assert.Nil(t, tail.worldState.RecordEvent(tx.Hash(), event))
	}
	es, err := tail.FetchEvents(tx.Hash())
	assert.Nil(t, err)
	for idx, event := range es {
		assert.Equal(t, events[idx], event)
	}
}

func TestSerializeTxByHash(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	block := bc.tailBlock
	tx := NewTransaction(bc.ChainID(), mockAddress(), mockAddress(), util.NewUint128(), 1, TxPayloadBinaryType, []byte(""), TransactionGasPrice, TransactionMaxGas)
	hash, err := HashTransaction(tx)
	assert.Nil(t, err)
	tx.hash = hash
	block.acceptTransaction(tx)
	msg, err := block.SerializeTxByHash(hash)
	assert.Nil(t, err)
	bytes, err := pb.Marshal(msg)
	assert.Nil(t, err)
	msg2, err := tx.ToProto()
	assert.Nil(t, err)
	bytes2, err := pb.Marshal(msg2)
	assert.Equal(t, bytes, bytes2)
}

func TestBlockSign(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	block := bc.tailBlock
	ks := keystore.DefaultKS
	signature, _ := crypto.NewSignature(keystore.SECP256K1)
	signer := mockAddress()
	key, _ := ks.GetUnlocked(signer.String())
	signature.InitSign(key.(keystore.PrivateKey))
	assert.Nil(t, block.Sign(signature))
	assert.Equal(t, block.Alg(), uint8(keystore.SECP256K1))
	assert.Equal(t, block.Signature(), block.header.sign)
}

func TestGivebackInvalidTx(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	from := mockAddress()
	ks := keystore.DefaultKS
	tx := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	tx.Sign(signature)
	assert.Nil(t, bc.txPool.Push(tx))
	assert.Equal(t, len(bc.txPool.all), 1)
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	block.CollectTransactions(time.Now().Unix() + 2)
	timer := time.NewTimer(time.Second).C
	<-timer
	assert.Equal(t, len(bc.txPool.all), 1)
}

func TestRecordEvent(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	txHash := []byte("hello")
	assert.Nil(t, bc.tailBlock.RecordEvent(txHash, TopicSendTransaction, "world"))
	events, err := bc.tailBlock.FetchEvents(txHash)
	assert.Nil(t, err)
	assert.Equal(t, len(events), 1)
	assert.Equal(t, events[0].Topic, TopicSendTransaction)
	assert.Equal(t, events[0].Data, "world")
}

func TestBlockVerifyIntegrity(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, nil), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), nil), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	tx2.hash[0]++
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	assert.NotNil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
}

func TestBlockVerifyIntegrityDup(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, nil), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), nil), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx1)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	assert.Equal(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()), ErrSmallTransactionNonce)
}

func TestBlockVerifyExecution(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, nil), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), nil), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 3, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	assert.Nil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
	header1, err := block.worldState.AccountsRoot()
	assert.Nil(t, err)
	assert.Equal(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()), ErrLargeTransactionNonce)
	header2, err := block.worldState.AccountsRoot()
	assert.Nil(t, err)
	assert.Equal(t, header1, header2)
}

func TestBlockVerifyState(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(0, nil), ErrInvalidChainID)
	bc.tailBlock.header.hash[0] = 1
	assert.Equal(t, bc.tailBlock.VerifyIntegrity(bc.ChainID(), nil), ErrInvalidBlockHash)
	ks := keystore.DefaultKS
	from := mockAddress()
	key, err := ks.GetUnlocked(from.String())
	assert.Nil(t, err)
	signature, err := crypto.NewSignature(keystore.SECP256K1)
	assert.Nil(t, err)
	signature.InitSign(key.(keystore.PrivateKey))
	block, err := bc.NewBlock(from)
	assert.Nil(t, err)
	tx1 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 1, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx1.Sign(signature)
	tx2 := NewTransaction(bc.ChainID(), from, from, util.NewUint128(), 2, TxPayloadBinaryType, []byte("nas"), TransactionGasPrice, util.NewUint128FromInt(200000))
	tx2.Sign(signature)
	block.transactions = append(block.transactions, tx1)
	block.transactions = append(block.transactions, tx2)
	block.miner = from
	block.Seal()
	block.Sign(signature)
	assert.Nil(t, block.VerifyIntegrity(bc.ChainID(), bc.ConsensusHandler()))
	block.header.stateRoot[0]++
	assert.NotNil(t, block.VerifyExecution(bc.tailBlock, bc.ConsensusHandler()))
}

func TestBlock_String(t *testing.T) {
	neb := testNeb(t)
	bc := neb.chain
	bc.genesisBlock.miner = nil
	logging.CLog().Info(bc.genesisBlock)
	assert.NotNil(t, bc.genesisBlock.String())
}
