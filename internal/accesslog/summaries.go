package accesslog

func (s *summary) updateSummary(newRecord *Record) {
	s.durationTotal = s.durationTotal + newRecord.Duration

	if newRecord.StatusCode >= 200 && newRecord.StatusCode < 300 {
		s.request2xx++
	}

	s.requestTotal++
}

func (ss Summaries) Aggregate(rawRecord []byte) error {
	newRecord, err := NewRecord(rawRecord)
	if err != nil {
		return err
	}

	s, ok := ss[newRecord.Host]
	if !ok {
		s = summary{}
	}
	// wether it new or not, it still need to update
	s.updateSummary(newRecord)
	ss[newRecord.Host] = s
	return nil
}
