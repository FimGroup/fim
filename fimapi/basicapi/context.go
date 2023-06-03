package basicapi

type Ctx struct {
	Tracing struct {
		TraceId string
		//SpanId has two formats - 1. current-parent pair(e.g. spanId:parentSpanId)  2. tier level(e.g. lv1_seq.lv2_seq.lv3_seq)
		SpanId string
	}
}
