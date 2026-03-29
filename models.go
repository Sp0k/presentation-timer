package main

type Name string
type TotalTimeMs int64

type Speaker struct {
	Name
	TotalTimeMs
}

type Section struct {
	Name
	TotalTimeMs
	AssignedSpeakers []Name
	SpeakerSplits map[Name]TotalTimeMs
}

type Presentation struct {
	Speakers            []Speaker
	Sections            []Section
	AssignedSections		bool
	TotalSpeakerSplits  map[Name]TotalTimeMs
	CurrentSpeakerIndex int
	CurrentSectionIndex int
	StartTime           int64
	Running             bool
	PausedTimeMs        TotalTimeMs
}
