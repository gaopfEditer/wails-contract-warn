package models

// KLineData K线数据
type KLineData struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// Indicators 技术指标
type Indicators struct {
	MA144    []float64 `json:"ma144"`
	MA10     []float64 `json:"ma10"`
	MA20     []float64 `json:"ma20"`
	MACD     []float64 `json:"macd"`
	Signal   []float64 `json:"signal"`
	Hist     []float64 `json:"hist"`
	BBUpper  []float64 `json:"bbUpper"`
	BBMiddle []float64 `json:"bbMiddle"`
	BBLower  []float64 `json:"bbLower"`
}

// AlertSignal 预警信号
type AlertSignal struct {
	Index     int     `json:"index"`
	Time      int64   `json:"time"`
	Price     float64 `json:"price"`
	Close     float64 `json:"close"`
	LowerBand float64 `json:"lowerBand,omitempty"`
	UpperBand float64 `json:"upperBand,omitempty"`
	Type      string  `json:"type"`     // 信号类型
	Strength  float64 `json:"strength,omitempty"` // 信号强度 0-1
}

