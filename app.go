package main

import (
	"fmt"
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

	paused bool
)

func InitApp() {
	app = tview.NewApplication()

	inputForm := buildInputForm()

	if err := app.SetRoot(inputForm, true).Run(); err != nil {
		panic(err)
	}

	paused = false
}

func buildInputForm() *tview.Form {
	speakersInput = tview.NewInputField().
		SetLabel("Speakers (comma separated): ").
		SetFieldWidth(40)

	sectionsInput = tview.NewInputField().
		SetLabel("Sections (comma separated): ").
		SetFieldWidth(40)

	assignmentCheckbox := tview.NewCheckbox().
		SetLabel("Assign speakers per section: ").
		SetChecked(false)

	form := tview.NewForm().
		AddFormItem(speakersInput).
		AddFormItem(sectionsInput).
		AddFormItem(assignmentCheckbox).
		AddButton("Start Presentation", func() {
			speakerNames := parseInput(speakersInput.GetText())
			sectionNames := parseInput(sectionsInput.GetText())
			assignSections := assignmentCheckbox.IsChecked()

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
				AssignedSections: 	 assignSections,
				TotalSpeakerSplits:  totalSpeakerSplits,
				CurrentSpeakerIndex: 0,
				CurrentSectionIndex: 0,
				StartTime:           time.Now().UnixMilli(),
				Running:             true,
			}

			lastTimeChange = time.Now().UnixMilli()

			if presentation.AssignedSections {
				buildAssignmentScreen()
			} else {
				for i := range presentation.Sections {
					for _, speaker := range presentation.Speakers {
						presentation.Sections[i].AssignedSpeakers =
							append(presentation.Sections[i].AssignedSpeakers, speaker.Name)
					}
				}
				startPresentation()
			}
		})

	form.SetBorder(true).
		SetTitle("Presentation Timer Setup").
		SetTitleAlign(tview.AlignLeft)

	return form
}

func buildAssignmentScreen() {
	form := tview.NewForm()

	checkboxes := make([][]*tview.Checkbox, len(presentation.Sections))

	for i, section := range presentation.Sections {
		form.AddTextView(
			fmt.Sprintf("section-label-%d", i),
			fmt.Sprintf("Section: %s", section.Name),
			40,
			1,
			false,
			false,
		)

		checkboxes[i] = make([]*tview.Checkbox, len(presentation.Speakers))

		for j, speaker := range presentation.Speakers {
			cb := tview.NewCheckbox().
				SetLabel("  " + string(speaker.Name)).
				SetChecked(false)

			checkboxes[i][j] = cb
			form.AddFormItem(cb)
		}
	}

	form.AddButton("Start Timer", func() {
		for i := range presentation.Sections {
			presentation.Sections[i].AssignedSpeakers = nil

			for j, speaker := range presentation.Speakers {
				if checkboxes[i][j].IsChecked() {
					presentation.Sections[i].AssignedSpeakers =
						append(presentation.Sections[i].AssignedSpeakers, speaker.Name)
				}
			}

			if len(presentation.Sections[i].AssignedSpeakers) == 0 {
				showAssignmentError("Each section must have at least one assigned speaker.")
				return
			}
		}

		startPresentation()
	})

	form.AddButton("Back", func() {
		app.SetRoot(buildInputForm(), true)
	})

	form.SetBorder(true).
		SetTitle("Assign Speakers to Sections").
		SetTitleAlign(tview.AlignLeft)

	app.SetRoot(form, true)
}

func showAssignmentError(msg string) {
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			buildAssignmentScreen()
		})

	app.SetRoot(modal, true)
}

func startPresentation() {
	presentation.CurrentSpeakerIndex = 0
	presentation.CurrentSectionIndex = 0
	presentation.StartTime = time.Now().UnixMilli()
	presentation.Running = true
	lastTimeChange = presentation.StartTime
	paused = false

	buildTimerScreen()
}

func buildTimerScreen() {
	timerText = tview.NewTextView()
	timerText.SetTextAlign(tview.AlignCenter)
	timerText.SetDynamicColors(true)
	timerText.SetBorder(true)
	timerText.SetTitle(" Presentation Timer ")

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(timerText, 0, 1, true)

	app.SetRoot(layout, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case ' ':
				if !presentation.Running && !paused {
					app.Stop()
					return nil
				}
				handleSpacebarSwitch()
			case 'p':
				if presentation.Running {
					saveTime()
					presentation.Running = false
					paused = true
					presentation.PausedTimeMs = TotalTimeMs(time.Now().UnixMilli())
				} else {
					resumeTime := time.Now().UnixMilli()
					presentation.StartTime += (resumeTime - int64(presentation.PausedTimeMs))
					lastTimeChange = resumeTime
					presentation.Running = true
					paused = false
				}
			case 'n':
				saveTime()
				presentation.CurrentSpeakerIndex = (presentation.CurrentSpeakerIndex + 1) % len(presentation.Speakers)
			case 'N':
				saveTime()
				presentation.CurrentSpeakerIndex = (presentation.CurrentSpeakerIndex - 1) % len(presentation.Speakers)
			}
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

func currentSectionSpeakers() []Name {
	if presentation.CurrentSectionIndex >= len(presentation.Sections) {
		return nil
	}
	return presentation.Sections[presentation.CurrentSectionIndex].AssignedSpeakers
}

func currentSpeakerName() Name {
	speakers := currentSectionSpeakers()
	if len(speakers) == 0 {
		return ""
	}
	if presentation.CurrentSpeakerIndex >= len(speakers) {
		presentation.CurrentSpeakerIndex = 0
	}
	return speakers[presentation.CurrentSpeakerIndex]
}

func updateClockTimer() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if presentation.Running {
			elapsedMs := time.Now().UnixMilli() - presentation.StartTime
			elapsed := time.Duration(elapsedMs) * time.Millisecond

			currentSpeaker := currentSpeakerName()
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
			if paused {
				elapsedMS := int64(presentation.PausedTimeMs) - presentation.StartTime
				elapsed := time.Duration(elapsedMS) * time.Millisecond
				app.QueueUpdateDraw(func() {
					timerText.SetText(
						"\n\n\n[red]PAUSED\n[yellow]Total:[white] " + formatDuration(elapsed),
						)
				})
			} else {
				app.QueueUpdateDraw(func() {
					timerText.SetText(outputHeader + outputBody + outputFooter)
				})
			}
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

	speakers := currentSectionSpeakers()
	if len(speakers) == 0 {
		return
	}

	presentation.CurrentSpeakerIndex++

	if presentation.CurrentSpeakerIndex >= len(speakers) {
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
	currSpeaker := currentSpeakerName()
	if currSpeaker == "" {
		return
	}
	currSection := &presentation.Sections[presentation.CurrentSectionIndex]

	currSection.SpeakerSplits[currSpeaker] += TotalTimeMs(timeAdded)
	currSection.TotalTimeMs += TotalTimeMs(timeAdded)
	presentation.TotalSpeakerSplits[currSpeaker] += TotalTimeMs(timeAdded)

	lastTimeChange = time.Now().UnixMilli()
}

func calculateOutput() {
	outputHeader = "\n\n\n[green]Presentation Complete!\nLet's see the breakdown...\n"

	totalTime := time.Duration(finalTime) * time.Millisecond
	var body strings.Builder
	var txtFileOutput strings.Builder

	// Total time
	body.WriteString("\n[yellow]Total time:[white] " + formatDuration(totalTime) + "\n")
	txtFileOutput.WriteString("Total time: " + formatDuration(totalTime) + "\n")

	// Speaker split
	body.WriteString("\n[yellow]Total time per speaker:\n")
	txtFileOutput.WriteString("\nTotal time per speaker:\n")
	for name, total := range presentation.TotalSpeakerSplits {
		body.WriteString("[orange]" + string(name) + ":[white] " + formatDuration(time.Duration(total)*time.Millisecond) + "\n")
		txtFileOutput.WriteString("\t" + string(name) + ": " + formatDuration(time.Duration(total)*time.Millisecond) + "\n")
	}

	// Section split
	body.WriteString("\n[yellow]Time per section:\n")
	txtFileOutput.WriteString("\nTime per section:\n")
	for _, section := range presentation.Sections {
		body.WriteString("[purple]" + string(section.Name) + " [white](" + formatDuration(time.Duration(section.TotalTimeMs)*time.Millisecond) + ")[purple]:\n")
		txtFileOutput.WriteString("\t" + string(section.Name) + " (" + formatDuration(time.Duration(section.TotalTimeMs)*time.Millisecond) + "):\n")
		for name, total := range section.SpeakerSplits {
			body.WriteString("\t[orange]" + string(name) + ":[white] " + formatDuration(time.Duration(total)*time.Millisecond) + "\n")
			txtFileOutput.WriteString("\t\t" + string(name) + ": " + formatDuration(time.Duration(total)*time.Millisecond) + "\n")
		}
	}

	outputBody = body.String()
	savedPath := SavePresentationData(txtFileOutput.String())

	outputFooter = fmt.Sprintf("\n[green]Press ctrl+c to exit\nYour data was saved at: " + savedPath)
}
