package accesslog

func (s *Summary) updateSummary(newRecord *Record) {
	s.avgDuration = (s.avgDuration*float64(s.requestTotal) + newRecord.Duration) / float64(s.requestTotal+1)

	if newRecord.StatusCode >= 200 && newRecord.StatusCode < 300 {
		s.request2xx++
	}

	s.requestTotal++
}

func (ss Summaries) AddRecord(newRecord *Record) {
	s, ok := ss[newRecord.Host]
	if !ok {
		ss[newRecord.Host] = Summary{}
	}
	s.updateSummary(newRecord)
	ss[newRecord.Host] = s
}
