package app

const (
	DefaultGasLimit uint64 = 21000
	DefaultGasPrice uint64 = 1
)

func CalculateGasFee(limit, price uint64) uint64 {
	return limit * price
}
