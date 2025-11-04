package accesslog

func (ss Summaries) AddRecord(newRecord *Record) {
	s, ok := ss[newRecord.Host]
	if !ok {
		ss[newRecord.Host] = Summary{}
	}
	s.updateSummary(newRecord)
	ss[newRecord.Host] = s
}
