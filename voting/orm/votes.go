package orm

import (
	"fmt"
	"time"
)

type Vote struct {
	Id        int `gorm:"primary_key" json:"id"`
	VoteId    int `gorm:"type:integer;unique;not null" json:"voteId"`
	VotingId  int `gorm:"type:integer;not null" json:"votingId"`
	OptionId  int `gorm:"type:integer;not null" json:"optionId"`
	CreatedAt time.Time
}

func (model *Vote) FindBy(field string, value any) (bool, error) {
	err := db.First(&model, fmt.Sprintf("%v = ?", field), value)
	if err.Error != nil {
		if IfEmpty(err.Error.Error()) {
			return false, nil
		}
		return false, err.Error
	}
	return true, nil
}

func (model *Vote) Create() error {
	err := db.Create(&model)
	if err.Error != nil {
		return err.Error
	}
	return nil
}

func (model *Vote) Delete() error {
	err := db.Delete(&model)
	if err.Error != nil {
		return err.Error
	}
	return nil
}

type Result struct {
	Count int
	Vid   string
	Oid   string
}

func FindAllVotes() (results []Result) {
	db.Raw("" +
		"select count(*) as count, v.voting_id vid, v.option_id oid " +
		"from votes v " +
		"group by v.voting_id, v.option_id;").Scan(&results)

	return results
}
