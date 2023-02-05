package request

type VotingRequest struct {
	VoteId   int `json:"voteId"`   // уникальный идентификатор голоса
	VotingId int `json:"votingId"` // уникальный идентификатор голосования (либералы, консерваторы)
	OptionId int `json:"optionId"` // уникальный идентификатор варианта голосования (За трампа)
}
