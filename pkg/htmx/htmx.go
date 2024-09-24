package htmx

import "fmt"

type HtmxMessage struct {
	CurrentUrl  *string `json:"HX-Current-URL"`
	Request     *string `json:"HX-Request"`
	Target      *string `json:"HX-Target"`
	Trigger     *string `json:"HX-Trigger"`
	TriggerName *string `json:"HX-TriggerName"`
	ActionType  *int    `json:"actionType"`
}

type HtmxMessageI struct {
	Header HtmxMessage `json:"HEADERS"`
	Any    interface{}
}

func (h *HtmxMessage) String() string {
	return fmt.Sprintf("CurrentUrl: %v, Request: %v, Target: %v, Trigger: %v, TriggerName: %v, ActionType: %d\n", h.CurrentUrl, h.Request, h.Target, h.Trigger, h.TriggerName, h.ActionType)
}
