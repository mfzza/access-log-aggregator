package accesslog

func (ss *Summaries) AddRecord(newRecord *Record) {
	// looking for same host
	for i := range *ss {
		if (*ss)[i].host == newRecord.Host {
			(*ss)[i].updateSummary(newRecord)
			return
		}
	}
	// proceed to append
	// TODO: handle error
	s, _ := NewSummary(newRecord.Host)
	*ss = append(*ss, *s)
}
