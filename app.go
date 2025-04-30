package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
	"strings"
	"time"
)

var app *tview.Application

var (
	speakersInput *tview.InputField
	sectionsInput *tview.InputField

	presentation   Presentation
	timerText      *tview.TextView
	lastTimeChange int64
	finalTime      int64

	outputHeader string
	outputBody   string
	outputFooter string
)

func InitApp() {
	app = tview.NewApplication()

	inputForm := buildInputForm()

	if err := app.SetRoot(inputForm, true).Run(); err != nil {
		panic(err)
	}
}

func buildInputForm() *tview.Form {
	speakersInput = tview.NewInputField().
		SetLabel("Speakers (comma separated): ").
		SetFieldWidth(40)

	sectionsInput = tview.NewInputField().
		SetLabel("Sections (comma separated): ").
		SetFieldWidth(40)

	form := tview.NewForm().
		AddFormItem(speakersInput).
		AddFormItem(sectionsInput).
		AddButton("Start Presentation", func() {
			speakerNames := parseInput(speakersInput.GetText())
			sectionNames := parseInput(sectionsInput.GetText())

			var speakers []Speaker
			for _, name := range speakerNames {
				speakers = append(speakers, Speaker{Name: Name(name)})
			}

			var sections []Section
			for _, name := range sectionNames {
				sections = append(sections, Section{Name: Name(name), SpeakerSplits: make(map[Name]TotalTimeMs)})
			}

			totalSpeakerSplits := make(map[Name]TotalTimeMs)
			for _, name := range speakers {
				totalSpeakerSplits[name.Name] = 0
			}

			presentation = Presentation{
				Speakers:            speakers,
				Sections:            sections,
				TotalSpeakerSplits:  totalSpeakerSplits,
				CurrentSpeakerIndex: 0,
				CurrentSectionIndex: 0,
				StartTime:           time.Now().UnixMilli(),
				Running:             true,
			}

			lastTimeChange = time.Now().UnixMilli()

			buildTimerScreen()
		})

	form.SetBorder(true).
		SetTitle("Presentation Timer Setup").
		SetTitleAlign(tview.AlignLeft)

	return form
}

func buildTimerScreen() {
	timerText = tview.NewTextView()
	timerText.SetTextAlign(tview.AlignCenter)
	timerText.SetDynamicColors(true)
	timerText.SetBorder(true)
	timerText.SetTitle("Presentation Timer")

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(timerText, 0, 1, true)

	app.SetRoot(layout, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			handleSpacebarSwitch()
			return nil
		}
		return event
	})

	go updateClockTimer()
}

func parseInput(input string) []string {
	parts := strings.Split(input, ",")
	var results []string
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		if cleaned != "" {
			results = append(results, cleaned)
		}
	}
	return results
}

func updateClockTimer() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if presentation.Running {
			elapsedMs := time.Now().UnixMilli() - presentation.StartTime
			elapsed := time.Duration(elapsedMs) * time.Millisecond

			currentSpeaker := presentation.Speakers[presentation.CurrentSpeakerIndex].Name
			currentSection := presentation.Sections[presentation.CurrentSectionIndex].Name

			app.QueueUpdateDraw(func() {
				timerText.SetText(
					"\n\n\n[yellow]Section:[white] " + string(currentSection) + "\n" +
						"[yellow]Speaker:[white] " + string(currentSpeaker) + "\n" +
						"[yellow]Elapsed:[white] " + formatDuration(elapsed),
				)
			})
			finalTime = time.Now().UnixMilli() - presentation.StartTime
		} else {
			app.QueueUpdateDraw(func() {
				timerText.SetText(outputHeader + outputBody + outputFooter)
			})
		}
	}
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return formatTwoDigits(minutes) + ":" + formatTwoDigits(seconds)
}

func formatTwoDigits(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func handleSpacebarSwitch() {
	saveTime()
	presentation.CurrentSpeakerIndex++

	if presentation.CurrentSpeakerIndex >= len(presentation.Speakers) {
		presentation.CurrentSpeakerIndex = 0
		presentation.CurrentSectionIndex++
	}

	if presentation.CurrentSectionIndex >= len(presentation.Sections) {
		presentation.Running = false
		calculateOutput()
	}
}

func saveTime() {
	timeAdded := time.Now().UnixMilli() - lastTimeChange
	currSpeaker := presentation.Speakers[presentation.CurrentSpeakerIndex].Name
	currSection := &presentation.Sections[presentation.CurrentSectionIndex]

	currSection.SpeakerSplits[currSpeaker] = TotalTimeMs(timeAdded)
	currSection.TotalTimeMs += TotalTimeMs(timeAdded)
	presentation.TotalSpeakerSplits[currSpeaker] += TotalTimeMs(timeAdded)

	lastTimeChange = time.Now().UnixMilli()
}

func calculateOutput() {
	outputHeader = "\n\n\n[green]Presentation Complete!\nLet's see the breakdown...\n"

	totalTime := time.Duration(finalTime) * time.Millisecond
	var body strings.Builder

	// Total time
	body.WriteString("\n[yellow]Total time:[white] " + formatDuration(totalTime) + "\n")

	// Speaker split
	body.WriteString("\n[yellow]Total time per speaker:\n")
	for name, total := range presentation.TotalSpeakerSplits {
		body.WriteString("[orange]" + string(name) + ":[white] " + formatDuration(time.Duration(total)*time.Millisecond) + "\n")
	}

	// Section split
	body.WriteString("\n[yellow]Time per section:\n")
	for _, section := range presentation.Sections {
		body.WriteString("[purple]" + string(section.Name) + " [white](" + formatDuration(time.Duration(section.TotalTimeMs)*time.Millisecond) + ")[purple]:\n")
		for name, total := range section.SpeakerSplits {
			body.WriteString("\t[orange]" + string(name) + ":[white] " + formatDuration(time.Duration(total)*time.Millisecond) + "\n")
		}
	}

	outputBody = body.String()

	outputFooter = "\n[green]Press ctrl+c to exit\n(this data will not be saved anywhere)"
}
