package model

import (
	jsonutil "github.com/cpusoft/goutil/jsonutil"
)

// error: simple error message;  detail: describe error;
// stage: generate error stage : 'parsevalidate'/'chainvalidate'
//({'fail': 'EE Certificate is about to expire',
//'detail': 'Expiration Time: ' + roa['ee']['valto'],
//'stage': 'parsevalidate'})
type StateMsg struct {
	Stage  string `json:"stage"`
	Fail   string `json:"fail"`
	Detail string `json:"detail"`
}

func (s *StateMsg) Equal(n *StateMsg) bool {
	if s.Fail == n.Fail && s.Detail == n.Detail && s.Stage == n.Stage {
		return true
	}
	return false
}

type StateModel struct {
	//valid/invalid/unknown(default)
	State    string     `json:"state"`
	Errors   []StateMsg `json:"errors"`
	Warnings []StateMsg `json:"warnings"`
}

// get stateModel from json string,
//and clear stage(chainvalidate/parsevalidate). if stage is "",then not clear
func GetStateModelAndResetStage(state string, clearStage string) (stateModel StateModel) {
	if len(state) == 0 {
		return NewStateModel()
	}

	jsonutil.UnmarshalJson(state, &stateModel)
	if len(clearStage) > 0 {
		stateModel.ClearStage(clearStage)
	}
	return stateModel
}
func NewStateModel() StateModel {
	st := StateModel{}
	st.State = "unknown"
	st.Errors = make([]StateMsg, 0)
	st.Warnings = make([]StateMsg, 0)

	return st
}

func (s *StateModel) JudgeState() {
	if len(s.Errors) > 0 {
		s.State = "invalid"
		return
	}
	if len(s.Warnings) > 0 {
		s.State = "warning"
		return
	}
	s.State = "valid"
}
func (s *StateModel) ClearStage(stage string) {
	if len(stage) == 0 {
		return
	}
	newErrors := make([]StateMsg, 0)
	newWarnings := make([]StateMsg, 0)
	for i := range s.Errors {
		if s.Errors[i].Stage != stage {
			newErrors = append(newErrors, s.Errors[i])
		}
	}
	for i := range s.Warnings {
		if s.Warnings[i].Stage != stage {
			newWarnings = append(newWarnings, s.Warnings[i])
		}
	}
	s.Errors = newErrors
	s.Warnings = newWarnings
}
func (s *StateModel) AddError(stateMsg *StateMsg) {
	for i := range s.Errors {
		if s.Errors[i].Equal(stateMsg) {
			return
		}
	}
	s.Errors = append(s.Errors, *stateMsg)
}
func (s *StateModel) AddWarning(stateMsg *StateMsg) {
	for i := range s.Warnings {
		if s.Warnings[i].Equal(stateMsg) {
			return
		}
	}
	s.Warnings = append(s.Warnings, *stateMsg)
}
