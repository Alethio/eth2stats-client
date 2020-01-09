package types

type ChainHead struct {
	HeadSlot           uint64
	HeadBlockRoot      string
	FinalizedSlot      uint64
	FinalizedBlockRoot string
	JustifiedSlot      uint64
	JustifiedBlockRoot string
}
