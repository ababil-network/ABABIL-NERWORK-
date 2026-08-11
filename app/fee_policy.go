package app

import "errors"

var ErrZeroTransactionFee = errors.New("transaction fee must be greater than zero")

// CalculateFinalNativeFee returns the final transaction fee in the
// smallest native ABABIL unit using the validated reference price.
func CalculateFinalNativeFee() (uint64, error) {
	referencePrice, err := NodeReferencePrice.Price()
	if err != nil {
		return 0, err
	}

	return NodeDynamicFee.CalculateNativeFee(referencePrice)
}
