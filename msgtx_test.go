// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcwire_test

import (
	"bytes"
	"github.com/conformal/btcwire"
	"github.com/davecgh/go-spew/spew"
	"io"
	"reflect"
	"testing"
)

// TestTx tests the MsgTx API.
func TestTx(t *testing.T) {
	pver := btcwire.ProtocolVersion

	// Block 100000 hash.
	hashStr := "3ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"
	hash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
	}

	// Ensure the command is expected value.
	wantCmd := "tx"
	msg := btcwire.NewMsgTx()
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgAddr: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	// Num addresses (varInt) + max allowed addresses.
	wantPayload := uint32(1000 * 1000)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Ensure we get the same transaction output point data back out.
	prevOutIndex := uint32(1)
	prevOut := btcwire.NewOutPoint(hash, prevOutIndex)
	if !prevOut.Hash.IsEqual(hash) {
		t.Errorf("NewOutPoint: wrong hash - got %v, want %v",
			spew.Sprint(&prevOut.Hash), spew.Sprint(hash))
	}
	if prevOut.Index != prevOutIndex {
		t.Errorf("NewOutPoint: wrong index - got %v, want %v",
			prevOut.Index, prevOutIndex)
	}

	// Ensure we get the same transaction input back out.
	sigScript := []byte{0x04, 0x31, 0xdc, 0x00, 0x1b, 0x01, 0x62}
	txIn := btcwire.NewTxIn(prevOut, sigScript)
	if !reflect.DeepEqual(&txIn.PreviousOutpoint, prevOut) {
		t.Errorf("NewTxIn: wrong prev outpoint - got %v, want %v",
			spew.Sprint(&txIn.PreviousOutpoint),
			spew.Sprint(prevOut))
	}
	if !bytes.Equal(txIn.SignatureScript, sigScript) {
		t.Errorf("NewTxIn: wrong signature script - got %v, want %v",
			spew.Sdump(txIn.SignatureScript),
			spew.Sdump(sigScript))
	}

	// Ensure we get the same transaction output back out.
	txValue := int64(5000000000)
	pkScript := []byte{
		0x41, // OP_DATA_65
		0x04, 0xd6, 0x4b, 0xdf, 0xd0, 0x9e, 0xb1, 0xc5,
		0xfe, 0x29, 0x5a, 0xbd, 0xeb, 0x1d, 0xca, 0x42,
		0x81, 0xbe, 0x98, 0x8e, 0x2d, 0xa0, 0xb6, 0xc1,
		0xc6, 0xa5, 0x9d, 0xc2, 0x26, 0xc2, 0x86, 0x24,
		0xe1, 0x81, 0x75, 0xe8, 0x51, 0xc9, 0x6b, 0x97,
		0x3d, 0x81, 0xb0, 0x1c, 0xc3, 0x1f, 0x04, 0x78,
		0x34, 0xbc, 0x06, 0xd6, 0xd6, 0xed, 0xf6, 0x20,
		0xd1, 0x84, 0x24, 0x1a, 0x6a, 0xed, 0x8b, 0x63,
		0xa6, // 65-byte signature
		0xac, // OP_CHECKSIG
	}
	txOut := btcwire.NewTxOut(txValue, pkScript)
	if txOut.Value != txValue {
		t.Errorf("NewTxOut: wrong pk script - got %v, want %v",
			txOut.Value, txValue)

	}
	if !bytes.Equal(txOut.PkScript, pkScript) {
		t.Errorf("NewTxOut: wrong pk script - got %v, want %v",
			spew.Sdump(txOut.PkScript),
			spew.Sdump(pkScript))
	}

	// Ensure transaction inputs are added properly.
	msg.AddTxIn(txIn)
	if !reflect.DeepEqual(msg.TxIn[0], txIn) {
		t.Errorf("AddTxIn: wrong transaction input added - got %v, want %v",
			spew.Sprint(msg.TxIn[0]), spew.Sprint(txIn))
	}

	// Ensure transaction outputs are added properly.
	msg.AddTxOut(txOut)
	if !reflect.DeepEqual(msg.TxOut[0], txOut) {
		t.Errorf("AddTxIn: wrong transaction output added - got %v, want %v",
			spew.Sprint(msg.TxOut[0]), spew.Sprint(txOut))
	}

	// Ensure the copy produced an identical transaction message.
	newMsg := msg.Copy()
	if !reflect.DeepEqual(newMsg, msg) {
		t.Errorf("Copy: mismatched tx messages - got %v, want %v",
			spew.Sdump(newMsg), spew.Sdump(msg))
	}

	return
}

// TestTxSha tests the ability to generate the hash of a transaction accurately.
func TestTxSha(t *testing.T) {
	// Hash of first transaction from block 113875.
	hashStr := "f051e59b5e2503ac626d03aaeac8ab7be2d72ba4b7e97119c5852d70d52dcb86"
	wantHash, err := btcwire.NewShaHashFromStr(hashStr)
	if err != nil {
		t.Errorf("NewShaHashFromStr: %v", err)
		return
	}

	// First transaction from block 113875.
	msgTx := btcwire.NewMsgTx()
	txIn := btcwire.TxIn{
		PreviousOutpoint: btcwire.OutPoint{
			Hash:  btcwire.ShaHash{},
			Index: 0xffffffff,
		},
		SignatureScript: []byte{0x04, 0x31, 0xdc, 0x00, 0x1b, 0x01, 0x62},
		Sequence:        0xffffffff,
	}
	txOut := btcwire.TxOut{
		Value: 5000000000,
		PkScript: []byte{
			0x41, // OP_DATA_65
			0x04, 0xd6, 0x4b, 0xdf, 0xd0, 0x9e, 0xb1, 0xc5,
			0xfe, 0x29, 0x5a, 0xbd, 0xeb, 0x1d, 0xca, 0x42,
			0x81, 0xbe, 0x98, 0x8e, 0x2d, 0xa0, 0xb6, 0xc1,
			0xc6, 0xa5, 0x9d, 0xc2, 0x26, 0xc2, 0x86, 0x24,
			0xe1, 0x81, 0x75, 0xe8, 0x51, 0xc9, 0x6b, 0x97,
			0x3d, 0x81, 0xb0, 0x1c, 0xc3, 0x1f, 0x04, 0x78,
			0x34, 0xbc, 0x06, 0xd6, 0xd6, 0xed, 0xf6, 0x20,
			0xd1, 0x84, 0x24, 0x1a, 0x6a, 0xed, 0x8b, 0x63,
			0xa6, // 65-byte signature
			0xac, // OP_CHECKSIG
		},
	}
	msgTx.AddTxIn(&txIn)
	msgTx.AddTxOut(&txOut)
	msgTx.LockTime = 0

	// Ensure the hash produced is expected.
	txHash, err := msgTx.TxSha()
	if err != nil {
		t.Errorf("TxSha: %v", err)
	}
	if !txHash.IsEqual(wantHash) {
		t.Errorf("TxSha: wrong hash - got %v, want %v",
			spew.Sprint(txHash), spew.Sprint(wantHash))
	}
}

// TestTxWire tests the MsgTx wire encode and decode for various numbers
// of transaction inputs and outputs and protocol versions.
func TestTxWire(t *testing.T) {
	// Empty tx message.
	noTx := btcwire.NewMsgTx()
	noTx.Version = 1
	noTxEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x00,                   // Varint for number of input transactions
		0x00,                   // Varint for number of output transactions
		0x00, 0x00, 0x00, 0x00, // Lock time
	}

	tests := []struct {
		in   *btcwire.MsgTx // Message to encode
		out  *btcwire.MsgTx // Expected decoded message
		buf  []byte         // Wire encoding
		pver uint32         // Protocol version for wire encoding
	}{
		// Latest protocol version with no transactions.
		{
			noTx,
			noTx,
			noTxEncoded,
			btcwire.ProtocolVersion,
		},

		// Latest protocol version with multiple transactions.
		{
			multiTx,
			multiTx,
			multiTxEncoded,
			btcwire.ProtocolVersion,
		},

		// Protocol version BIP0035Version with no transactions.
		{
			noTx,
			noTx,
			noTxEncoded,
			btcwire.BIP0035Version,
		},

		// Protocol version BIP0035Version with multiple transactions.
		{
			multiTx,
			multiTx,
			multiTxEncoded,
			btcwire.BIP0035Version,
		},

		// Protocol version BIP0031Version with no transactions.
		{
			noTx,
			noTx,
			noTxEncoded,
			btcwire.BIP0031Version,
		},

		// Protocol version BIP0031Version with multiple transactions.
		{
			multiTx,
			multiTx,
			multiTxEncoded,
			btcwire.BIP0031Version,
		},

		// Protocol version NetAddressTimeVersion with no transactions.
		{
			noTx,
			noTx,
			noTxEncoded,
			btcwire.NetAddressTimeVersion,
		},

		// Protocol version NetAddressTimeVersion with multiple transactions.
		{
			multiTx,
			multiTx,
			multiTxEncoded,
			btcwire.NetAddressTimeVersion,
		},

		// Protocol version MultipleAddressVersion with no transactions.
		{
			noTx,
			noTx,
			noTxEncoded,
			btcwire.MultipleAddressVersion,
		},

		// Protocol version MultipleAddressVersion with multiple transactions.
		{
			multiTx,
			multiTx,
			multiTxEncoded,
			btcwire.MultipleAddressVersion,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode the message to wire format.
		var buf bytes.Buffer
		err := test.in.BtcEncode(&buf, test.pver)
		if err != nil {
			t.Errorf("BtcEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("BtcEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the message from wire format.
		var msg btcwire.MsgTx
		rbuf := bytes.NewBuffer(test.buf)
		err = msg.BtcDecode(rbuf, test.pver)
		if err != nil {
			t.Errorf("BtcDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&msg, test.out) {
			t.Errorf("BtcDecode #%d\n got: %s want: %s", i,
				spew.Sdump(&msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestTxWireErrors performs negative tests against wire encode and decode
// of MsgTx to confirm error paths work correctly.
func TestTxWireErrors(t *testing.T) {
	// Use protocol version 60002 specifically here instead of the latest
	// because the test data is using bytes encoded with that protocol
	// version.
	pver := uint32(60002)

	tests := []struct {
		in       *btcwire.MsgTx // Value to encode
		buf      []byte         // Wire encoding
		pver     uint32         // Protocol version for wire encoding
		max      int            // Max size of fixed buffer to induce errors
		writeErr error          // Expected write error
		readErr  error          // Expected read error
	}{
		// Force error in version.
		{multiTx, multiTxEncoded, pver, 0, io.ErrShortWrite, io.EOF},
		// Force error in number of transaction inputs.
		{multiTx, multiTxEncoded, pver, 4, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block hash.
		{multiTx, multiTxEncoded, pver, 5, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block hash.
		{multiTx, multiTxEncoded, pver, 5, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block output index.
		{multiTx, multiTxEncoded, pver, 37, io.ErrShortWrite, io.EOF},
		// Force error in transaction input signature script length.
		{multiTx, multiTxEncoded, pver, 41, io.ErrShortWrite, io.EOF},
		// Force error in transaction input signature script.
		{multiTx, multiTxEncoded, pver, 42, io.ErrShortWrite, io.EOF},
		// Force error in transaction input sequence.
		{multiTx, multiTxEncoded, pver, 49, io.ErrShortWrite, io.EOF},
		// Force error in number of transaction outputs.
		{multiTx, multiTxEncoded, pver, 53, io.ErrShortWrite, io.EOF},
		// Force error in transaction output value.
		{multiTx, multiTxEncoded, pver, 54, io.ErrShortWrite, io.EOF},
		// Force error in transaction output pk script length.
		{multiTx, multiTxEncoded, pver, 62, io.ErrShortWrite, io.EOF},
		// Force error in transaction output pk script.
		{multiTx, multiTxEncoded, pver, 63, io.ErrShortWrite, io.EOF},
		// Force error in transaction output lock time.
		{multiTx, multiTxEncoded, pver, 130, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := test.in.BtcEncode(w, test.pver)
		if err != test.writeErr {
			t.Errorf("BtcEncode #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Decode from wire format.
		var msg btcwire.MsgTx
		r := newFixedReader(test.max, test.buf)
		err = msg.BtcDecode(r, test.pver)
		if err != test.readErr {
			t.Errorf("BtcDecode #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestTxSerialize tests MsgTx serialize and deserialize.
func TestTxSerialize(t *testing.T) {
	noTx := btcwire.NewMsgTx()
	noTx.Version = 1
	noTxEncoded := []byte{
		0x01, 0x00, 0x00, 0x00, // Version
		0x00,                   // Varint for number of input transactions
		0x00,                   // Varint for number of output transactions
		0x00, 0x00, 0x00, 0x00, // Lock time
	}

	tests := []struct {
		in  *btcwire.MsgTx // Message to encode
		out *btcwire.MsgTx // Expected decoded message
		buf []byte         // Serialized data
	}{
		// No transactions.
		{
			noTx,
			noTx,
			noTxEncoded,
		},

		// Multiple transactions.
		{
			multiTx,
			multiTx,
			multiTxEncoded,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Serialize the transaction.
		var buf bytes.Buffer
		err := test.in.Serialize(&buf)
		if err != nil {
			t.Errorf("Serialize #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("Serialize #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Deserialize the transaction.
		var tx btcwire.MsgTx
		rbuf := bytes.NewBuffer(test.buf)
		err = tx.Deserialize(rbuf)
		if err != nil {
			t.Errorf("Deserialize #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(&tx, test.out) {
			t.Errorf("Deserialize #%d\n got: %s want: %s", i,
				spew.Sdump(&tx), spew.Sdump(test.out))
			continue
		}
	}
}

// TestTxSerializeErrors performs negative tests against wire encode and decode
// of MsgTx to confirm error paths work correctly.
func TestTxSerializeErrors(t *testing.T) {
	tests := []struct {
		in       *btcwire.MsgTx // Value to encode
		buf      []byte         // Serialized data
		max      int            // Max size of fixed buffer to induce errors
		writeErr error          // Expected write error
		readErr  error          // Expected read error
	}{
		// Force error in version.
		{multiTx, multiTxEncoded, 0, io.ErrShortWrite, io.EOF},
		// Force error in number of transaction inputs.
		{multiTx, multiTxEncoded, 4, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block hash.
		{multiTx, multiTxEncoded, 5, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block hash.
		{multiTx, multiTxEncoded, 5, io.ErrShortWrite, io.EOF},
		// Force error in transaction input previous block output index.
		{multiTx, multiTxEncoded, 37, io.ErrShortWrite, io.EOF},
		// Force error in transaction input signature script length.
		{multiTx, multiTxEncoded, 41, io.ErrShortWrite, io.EOF},
		// Force error in transaction input signature script.
		{multiTx, multiTxEncoded, 42, io.ErrShortWrite, io.EOF},
		// Force error in transaction input sequence.
		{multiTx, multiTxEncoded, 49, io.ErrShortWrite, io.EOF},
		// Force error in number of transaction outputs.
		{multiTx, multiTxEncoded, 53, io.ErrShortWrite, io.EOF},
		// Force error in transaction output value.
		{multiTx, multiTxEncoded, 54, io.ErrShortWrite, io.EOF},
		// Force error in transaction output pk script length.
		{multiTx, multiTxEncoded, 62, io.ErrShortWrite, io.EOF},
		// Force error in transaction output pk script.
		{multiTx, multiTxEncoded, 63, io.ErrShortWrite, io.EOF},
		// Force error in transaction output lock time.
		{multiTx, multiTxEncoded, 130, io.ErrShortWrite, io.EOF},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Serialize the transaction.
		w := newFixedWriter(test.max)
		err := test.in.Serialize(w)
		if err != test.writeErr {
			t.Errorf("Serialize #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Deserialize the transaction.
		var tx btcwire.MsgTx
		r := newFixedReader(test.max, test.buf)
		err = tx.Deserialize(r)
		if err != test.readErr {
			t.Errorf("Deserialize #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}

// TestTxOverflowErrors performs tests to ensure deserializing transactions
// which are intentionally crafted to use large values for the variable number
// of inputs and outputs are handled properly.  This could otherwise potentially
// be used as an attack vector.
func TestTxOverflowErrors(t *testing.T) {
	// Use protocol version 70001 and transaction version 1 specifically
	// here instead of the latest values because the test data is using
	// bytes encoded with those versions.
	pver := uint32(70001)
	txVer := uint32(1)

	tests := []struct {
		buf     []byte // Wire encoding
		pver    uint32 // Protocol version for wire encoding
		version uint32 // Transaction version
		err     error  // Expected error
	}{
		// Transaction that claims to have ~uint64(0) inputs.
		{
			[]byte{
				0x00, 0x00, 0x00, 0x01, // Version
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // Varint for number of input transactions
			}, pver, txVer, &btcwire.MessageError{},
		},

		// Transaction that claims to have ~uint64(0) outputs.
		{
			[]byte{
				0x00, 0x00, 0x00, 0x01, // Version
				0x00, // Varint for number of input transactions
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // Varint for number of output transactions
			}, pver, txVer, &btcwire.MessageError{},
		},

		// Transaction that has an input with a signature script that
		// claims to have ~uint64(0) length.
		{
			[]byte{
				0x00, 0x00, 0x00, 0x01, // Version
				0x01, // Varint for number of input transactions
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Previous output hash
				0xff, 0xff, 0xff, 0xff, // Prevous output index
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // Varint for length of signature script
			}, pver, txVer, &btcwire.MessageError{},
		},

		// Transaction that has an output with a public key script
		// that claims to have ~uint64(0) length.
		{
			[]byte{
				0x00, 0x00, 0x00, 0x01, // Version
				0x01, // Varint for number of input transactions
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Previous output hash
				0xff, 0xff, 0xff, 0xff, // Prevous output index
				0x00,                   // Varint for length of signature script
				0xff, 0xff, 0xff, 0xff, // Sequence
				0x01,                                           // Varint for number of output transactions
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Transaction amount
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, // Varint for length of public key script
			}, pver, txVer, &btcwire.MessageError{},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Decode from wire format.
		var msg btcwire.MsgTx
		r := bytes.NewBuffer(test.buf)
		err := msg.BtcDecode(r, test.pver)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("BtcDecode #%d wrong error got: %v, want: %v",
				i, err, reflect.TypeOf(test.err))
			continue
		}

		// Decode from wire format.
		r = bytes.NewBuffer(test.buf)
		err = msg.Deserialize(r)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Deserialize #%d wrong error got: %v, want: %v",
				i, err, reflect.TypeOf(test.err))
			continue
		}
	}
}

// TestTxSerializeSize performs tests to ensure the serialize size for various
// transactions is accurate.
func TestTxSerializeSize(t *testing.T) {
	// Empty tx message.
	noTx := btcwire.NewMsgTx()
	noTx.Version = 1

	tests := []struct {
		in   *btcwire.MsgTx // Tx to encode
		size int            // Expected serialized size
	}{
		// No inputs or outpus.
		{noTx, 10},

		// Transcaction with an input and an output.
		{multiTx, 134},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		serializedSize := test.in.SerializeSize()
		if serializedSize != test.size {
			t.Errorf("MsgTx.SerializeSize: #%d got: %d, want: %d", i,
				serializedSize, test.size)
			continue
		}
	}
}

// multiTx is a MsgTx with an input and output and used in various tests.
var multiTx = &btcwire.MsgTx{
	Version: 1,
	TxIn: []*btcwire.TxIn{
		{
			PreviousOutpoint: btcwire.OutPoint{
				Hash:  btcwire.ShaHash{},
				Index: 0xffffffff,
			},
			SignatureScript: []byte{
				0x04, 0x31, 0xdc, 0x00, 0x1b, 0x01, 0x62,
			},
			Sequence: 0xffffffff,
		},
	},
	TxOut: []*btcwire.TxOut{
		{
			Value: 0x12a05f200,
			PkScript: []byte{
				0x41, // OP_DATA_65
				0x04, 0xd6, 0x4b, 0xdf, 0xd0, 0x9e, 0xb1, 0xc5,
				0xfe, 0x29, 0x5a, 0xbd, 0xeb, 0x1d, 0xca, 0x42,
				0x81, 0xbe, 0x98, 0x8e, 0x2d, 0xa0, 0xb6, 0xc1,
				0xc6, 0xa5, 0x9d, 0xc2, 0x26, 0xc2, 0x86, 0x24,
				0xe1, 0x81, 0x75, 0xe8, 0x51, 0xc9, 0x6b, 0x97,
				0x3d, 0x81, 0xb0, 0x1c, 0xc3, 0x1f, 0x04, 0x78,
				0x34, 0xbc, 0x06, 0xd6, 0xd6, 0xed, 0xf6, 0x20,
				0xd1, 0x84, 0x24, 0x1a, 0x6a, 0xed, 0x8b, 0x63,
				0xa6, // 65-byte signature
				0xac, // OP_CHECKSIG
			},
		},
	},
	LockTime: 0,
}

// multiTxEncoded is the wire encoded bytes for multiTx using protocol version
// 60002 and is used in the various tests.
var multiTxEncoded = []byte{
	0x01, 0x00, 0x00, 0x00, // Version
	0x01, // Varint for number of input transactions
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Previous output hash
	0xff, 0xff, 0xff, 0xff, // Prevous output index
	0x07,                                     // Varint for length of signature script
	0x04, 0x31, 0xdc, 0x00, 0x1b, 0x01, 0x62, // Signature script
	0xff, 0xff, 0xff, 0xff, // Sequence
	0x01,                                           // Varint for number of output transactions
	0x00, 0xf2, 0x05, 0x2a, 0x01, 0x00, 0x00, 0x00, // Transaction amount
	0x43, // Varint for length of pk script
	0x41, // OP_DATA_65
	0x04, 0xd6, 0x4b, 0xdf, 0xd0, 0x9e, 0xb1, 0xc5,
	0xfe, 0x29, 0x5a, 0xbd, 0xeb, 0x1d, 0xca, 0x42,
	0x81, 0xbe, 0x98, 0x8e, 0x2d, 0xa0, 0xb6, 0xc1,
	0xc6, 0xa5, 0x9d, 0xc2, 0x26, 0xc2, 0x86, 0x24,
	0xe1, 0x81, 0x75, 0xe8, 0x51, 0xc9, 0x6b, 0x97,
	0x3d, 0x81, 0xb0, 0x1c, 0xc3, 0x1f, 0x04, 0x78,
	0x34, 0xbc, 0x06, 0xd6, 0xd6, 0xed, 0xf6, 0x20,
	0xd1, 0x84, 0x24, 0x1a, 0x6a, 0xed, 0x8b, 0x63,
	0xa6,                   // 65-byte signature
	0xac,                   // OP_CHECKSIG
	0x00, 0x00, 0x00, 0x00, // Lock time
}
