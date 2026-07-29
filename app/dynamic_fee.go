package app

import "sync"

const (
	BaseGasPrice uint64 = 1
	MaxGasPrice  uint64 = 100
)

type DynamicFeeManager struct {
	mu           sync.RWMutex
	PendingTx    uint64
}

var NodeDynamicFee = &DynamicFeeManager{}

func (d *DynamicFeeManager) SetPending(count uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.PendingTx = count
}

func (d *DynamicFeeManager) GasPrice() uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	switch {
	case d.PendingTx < 100:
		return BaseGasPrice
	case d.PendingTx < 500:
		return 2
	case d.PendingTx < 1000:
		return 5
	case d.PendingTx < 5000:
		return 10
	default:
		if MaxGasPrice < 20 {
			return MaxGasPrice
		}
		return 20
	}
}

func (d *DynamicFeeManager) CalculateFee(gasLimit uint64) uint64 {
	return gasLimit * d.GasPrice()
}
