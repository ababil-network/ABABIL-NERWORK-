package app

import (
	"errors"
)

var (
	ErrMempoolAdmissionDuplicateHash  = errors.New("transaction already exists in mempool")
	ErrMempoolAdmissionDuplicateNonce = errors.New("sender nonce already exists in mempool")
	ErrMempoolAdmissionSenderLimit    = errors.New("sender mempool limit reached")
	ErrMempoolAdmissionFull           = errors.New("mempool capacity reached")
)

const (
	NormalSenderPendingLimit uint64 = 256

	Congestion95SenderPendingLimit  uint64 = 128
	Congestion96SenderPendingLimit  uint64 = 102
	Congestion97SenderPendingLimit  uint64 = 76
	Congestion98SenderPendingLimit  uint64 = 51
	Congestion99SenderPendingLimit  uint64 = 25
	Congestion100SenderPendingLimit uint64 = 16
)

// SenderPendingLimit returns the node-local pending transaction limit
// for one sender at the current network load.
//
// This is a mempool admission policy only.
// It does not change transaction validity or consensus rules.
func SenderPendingLimit(load uint64) uint64 {
	switch {
	case load < FeeLoadCongestionEndPercent:
		return NormalSenderPendingLimit

	case load == 95:
		return Congestion95SenderPendingLimit

	case load == 96:
		return Congestion96SenderPendingLimit

	case load == 97:
		return Congestion97SenderPendingLimit

	case load == 98:
		return Congestion98SenderPendingLimit

	case load == 99:
		return Congestion99SenderPendingLimit

	default:
		return Congestion100SenderPendingLimit
	}
}

// AdmitTransaction performs the node-local mempool admission checks.
//
// Consensus validation must happen before calling this function.
// This function only decides whether an already-valid transaction
// should enter this node's mempool.
func (m *Mempool) AdmitTransaction(tx Transaction) error {
	if m == nil {
		return ErrMempoolAdmissionFull
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.Transactions) >= MaxMempoolTransactions {
		return ErrMempoolAdmissionFull
	}

	load := NodeDynamicFee.Load()
	senderLimit := SenderPendingLimit(load)

	senderPending := uint64(0)

	for _, existing := range m.Transactions {
		if existing.Hash == tx.Hash {
			return ErrMempoolAdmissionDuplicateHash
		}

		if existing.From == tx.From {
			if existing.Nonce == tx.Nonce {
				return ErrMempoolAdmissionDuplicateNonce
			}

			senderPending++
		}
	}

	if senderPending >= senderLimit {
		return ErrMempoolAdmissionSenderLimit
	}

	m.Transactions = append(m.Transactions, tx)
	m.sortByPriorityLocked()

	return nil
}
