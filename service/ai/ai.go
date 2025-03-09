package ai

type DoCompletions interface {
	Do(req OpenAiReq) (*OpenAiRes[Choice], error)
	DoStream(req OpenAiReq, resChan chan *OpenAiRes[ChoiceStream], errChan chan error)
}
