package mvp

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type Mood int

const (
	MoodNeutral = Mood(iota)
	MoodSuccess
	MoodFailure
	MoodSubtle
)

var _moodStrings = []string{"neutral", "success", "failure"}

func (v Mood) String() string {
	return _moodStrings[v]
}

func ParseMood(s string) (Mood, error) {
	if i := slices.Index(_moodStrings, s); i >= 0 {
		return Mood(i), nil
	} else {
		return MoodNeutral, fmt.Errorf("invalid Mood %q", s)
	}
}

type Msg struct {
	Text     string
	Link     string
	LinkText string
	Mood     Mood
}

func (msg *Msg) Success() bool {
	return msg.Mood == MoodSuccess
}

func (msg *Msg) Failure() bool {
	return msg.Mood == MoodFailure
}

func SubtleMsg(text string) *Msg {
	return &Msg{Text: text, Mood: MoodSubtle}
}
func SuccessMsg(text string) *Msg {
	return &Msg{Text: text, Mood: MoodSuccess}
}
func FailureMsg(text string) *Msg {
	return &Msg{Text: text, Mood: MoodFailure}
}
func NeutralMsg(text string) *Msg {
	return &Msg{Text: text, Mood: MoodNeutral}
}
