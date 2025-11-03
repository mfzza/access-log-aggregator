package accesslog

func (ss Summaries) AddRecord(newRecord *Record) {
	if s, ok := ss[newRecord.Host]; ok {
		s.updateSummary(newRecord)
		ss[newRecord.Host] = s
		return
	}
	ss[newRecord.Host] = Summary{}
}
