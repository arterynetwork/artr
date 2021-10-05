package types

func NewRecord() Record {
	return Record{}
}

func (x Record) IsEmpty() bool {
	return x.Requests == nil && x.NextAccrue == nil
}
