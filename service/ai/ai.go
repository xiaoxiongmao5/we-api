package ai

type DoCompletions interface {
	Do(req OpenAiReq) (*OpenAiRes[OpenAiChoice], error)
	DoStream(req OpenAiReq, resChan chan *OpenAiRes[OpenAiChoiceStream], errChan chan error)
}
