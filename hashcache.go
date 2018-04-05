package zecutil

import (
	"bytes"
	"encoding/binary"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const (
	prevoutsHashPersonalization   = "ZcashPrevoutHash"
	sequenceHashPersonalization   = "ZcashSequencHash"
	outputsHashPersonalization    = "ZcashOutputsHash"
	joinsplitsHashPersonalization = "ZcashJSplitsHash"
)

// NewTxSigHashes computes, and returns the cached sighashes of the given
// transaction.
func NewTxSigHashes(tx *MsgTx) (h *txscript.TxSigHashes, err error) {
	h = &txscript.TxSigHashes{}

	//fmt.Printf("IN % x\n", tx.TxIn[0].PreviousOutPoint.Hash.CloneBytes())
	//fmt.Printf("INDEX % x\n", tx.TxIn[0].PreviousOutPoint.Index)

	if h.HashPrevOuts, err = calcHashPrevOuts(tx); err != nil {
		return
	}
	//
	//fmt.Printf("HashPrevOuts: % x\n", h.HashPrevOuts.CloneBytes())
	//os.Exit(1)

	if h.HashSequence, err = calcHashSequence(tx); err != nil {
		return
	}

	if h.HashOutputs, err = calcHashOutputs(tx); err != nil {
		return
	}

	return
}

// calcHashPrevOuts calculates a single hash of all the previous outputs
// (txid:index) referenced within the passed transaction. This calculated hash
// can be re-used when validating all inputs spending segwit outputs, with a
// signature hash type of SigHashAll. This allows validation to re-use previous
// hashing computation, reducing the complexity of validating SigHashAll inputs
// from  O(N^2) to O(N).
func calcHashPrevOuts(tx *MsgTx) (chainhash.Hash, error) {
	var b bytes.Buffer
	for _, in := range tx.TxIn {
		// First write out the 32-byte transaction ID one of whose
		// outputs are being referenced by this input.

		b.Write(in.PreviousOutPoint.Hash[:])

		// Next, we'll encode the index of the referenced output as a
		// little endian integer.
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], in.PreviousOutPoint.Index)
		b.Write(buf[:])
	}

	return blake2bHash(b.Bytes(), []byte(prevoutsHashPersonalization))
}

// calcHashSequence computes an aggregated hash of each of the sequence numbers
// within the inputs of the passed transaction. This single hash can be re-used
// when validating all inputs spending segwit outputs, which include signatures
// using the SigHashAll sighash type. This allows validation to re-use previous
// hashing computation, reducing the complexity of validating SigHashAll inputs
// from O(N^2) to O(N).
func calcHashSequence(tx *MsgTx) (chainhash.Hash, error) {
	var b bytes.Buffer
	for _, in := range tx.TxIn {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], in.Sequence)
		b.Write(buf[:])
	}

	return blake2bHash(b.Bytes(), []byte(sequenceHashPersonalization))
}

// calcHashOutputs computes a hash digest of all outputs created by the
// transaction encoded using the wire format. This single hash can be re-used
// when validating all inputs spending witness programs, which include
// signatures using the SigHashAll sighash type. This allows computation to be
// cached, reducing the total hashing complexity from O(N^2) to O(N).
func calcHashOutputs(tx *MsgTx) (chainhash.Hash, error) {
	var b bytes.Buffer
	for _, out := range tx.TxOut {
		wire.WriteTxOut(&b, 0, 0, out)
	}

	return blake2bHash(b.Bytes(), []byte(outputsHashPersonalization))
}
