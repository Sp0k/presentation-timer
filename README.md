# Presentation Timer (Terminal App)

A terminal-based Go application to help practice and time presentations with multiple speakers and sections.  
Built using [tview](https://github.com/rivo/tview) and [tcell](https://github.com/gdamore/tcell).

---

## ✨ Features

- ⏱️ Tracks total presentation time
- 🎙️ Tracks time per speaker
- 📚 Tracks time per section
- 🔍 Tracks how each speaker contributed within each section
- 🧠 Unlimited number of speakers and sections
- 🧼 Clean, centered terminal UI
- 🔁 Use `SPACEBAR` to switch between speakers and sections
- ✅ Displays a detailed time breakdown at the end

---

## 🛠️ Installation

### Go

```
go install github.com/Sp0k/presentation-timer
```

### From source

```
git clone https://github.com/Sp0k/presentation-timer
cd presentation-timer
go install
```

### Binary

You can also download a binary from the [release](https://github.com/Sp0k/presentation-timer/releases)

---

## 🧑‍💻 Usage

### 1. Run the app

```bash
go run main.go app.go models.go
```

> [!NOTE]
> Make sure Go modules are initialized, or run go mod init presentation-timer if needed.

### 2. Enter your speakers and sections

- Use **comma-separated** lists for both fields
- Example:
  - Speakers: `Alice, Bob, Charlie`
  - Sections: `Intro, Methods, Conclusion`

### 3. Start the presentation

- Click **Start Presentation**
- The timer screen will begin
- It shows:
  - Current speaker
  - Current section
  - Elapsed time

### 4. Control the flow

- Press the `SPACEBAR` to:
  - Move to the next speaker
  - Automatically move to the next section after the last speaker
  - End the presentation after the last section

### 5. Get your results

- At the end, you’ll see:
  - ✅ Total elapsed time
  - 🧑‍🤝‍🧑 Time per speaker
  - 📚 Time per section, including how much each speaker contributed

> [!WARNING]
> This data is not saved anywhere. You must manually record it if needed

---

## 📦 Dependencies

- [Go](https://go.dev/) 1.18+
- [tview](https://github.com/rivo/tview)
- [tcell](https://github.com/gdamore/tcell)

Install them with:

```bash
go get github.com/rivo/tview
go get github.com/gdamore/tcell/v2
```

---

## 📁 Project Structure

```
presentation-timer/
├── main.go        # Entry point
├── app.go         # All UI and timer logic
├── models.go      # Data models: Speaker, Section, Presentation
├── go.mod         # (Optional) Go module file
```
