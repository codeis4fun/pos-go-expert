package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	serverStatus := widget.NewLabel("Server Status")
	ranking := widget.NewLabel("Ranking")
	news := widget.NewLabel("News")

	// Shortened Lorem Ipsum text
	loreIpsumShort := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."

	serverStatusLorem := widget.NewLabelWithStyle(loreIpsumShort, fyne.TextAlignCenter, fyne.TextStyle{})
	rankingLorem := widget.NewLabelWithStyle(loreIpsumShort, fyne.TextAlignCenter, fyne.TextStyle{})
	newsLorem := widget.NewLabelWithStyle(loreIpsumShort, fyne.TextAlignCenter, fyne.TextStyle{})

	myWindow := myApp.NewWindow("PristonTale.eu")
	myWindow.Resize(myWindow.Canvas().Size())

	myWindow.SetContent(container.New(layout.NewGridLayout(3),
		container.New(layout.NewVBoxLayout(),
			container.NewVBox(
				container.NewCenter(serverStatus),
				container.NewCenter(serverStatusLorem),
			),
		),
		container.New(layout.NewVBoxLayout(),
			container.NewVBox(
				container.NewCenter(news),
				container.NewCenter(newsLorem),
			),
		),
		container.New(layout.NewVBoxLayout(),
			container.NewVBox(
				container.NewCenter(ranking),
				container.NewCenter(rankingLorem),
			),
		),
	))

	myWindow.ShowAndRun()
}
