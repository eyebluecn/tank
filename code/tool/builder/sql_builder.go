package builder

type OrderPair struct {
	Key   string
	Value string
}

type WherePair struct {
	Query string
	Args  []interface{}
}

func (this *WherePair) And(where *WherePair) *WherePair {
	if this.Query == "" {
		return where
	} else {
		return &WherePair{Query: this.Query + " AND " + where.Query, Args: append(this.Args, where.Args...)}
	}

}

func (this *WherePair) Or(where *WherePair) *WherePair {
	if this.Query == "" {
		return where
	} else {
		return &WherePair{Query: this.Query + " OR " + where.Query, Args: append(this.Args, where.Args...)}
	}

}
