package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"

	"log"
	"os"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

// in seconds
var gifDuration = 30

// this is used to set the initial offset of the text from the left side of the image
var textOriginOffset = 50

// width and height of the overall gif
var imageHeight = 150
var imageWidth = 530 + textOriginOffset

// font size and label size for the text
var fontSize = 25
var labelSize = 15

// space between the value and label
var valLabelPadding = 20

// spacing to the next group
var groupPadding = 50

func main() {
	startTime := time.Now()

	// if no date, or time is provided, exit
	if len(os.Args) < 2 {
		fmt.Println("Please provide a date in the format YYYY-MM-DD")
		os.Exit(1)
	}

	// Set the target date to count down to os.Args[1] and convert it to a time.Time
	inputDate, err := time.Parse("2006-01-02", os.Args[1])
	if err != nil {
		fmt.Println("Please provide a date in the format YYYY-MM-DD")
		os.Exit(1)
	}

	// Set the target time to count down to os.Args[2] and convert it to a time.Time
	inputTime, err := time.Parse("15:04:05", os.Args[2])
	if err != nil {
		fmt.Println("Please provide a time in the format HH:MM:SS")
		os.Exit(1)
	}
	now := time.Now()

	targetDate := time.Date(
		inputDate.Year(),
		inputDate.Month(),
		inputDate.Day(),
		inputTime.Hour(),
		inputTime.Minute(),
		inputTime.Second(), 0,
		time.UTC)

	currentDate := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(), 0,
		time.UTC)

	if targetDate.Before(currentDate) {
		targetDate = currentDate
	}

	duration := targetDate.Sub(currentDate)

	// Create a new GIF
	anim := gif.GIF{}

	// use getDaysHoursMinutesSeconds to get the total days, hours, minutes, and seconds
	totalDays, totalHours, totalMinutes, totalSeconds := getDaysHoursMinutesSeconds(duration)

	println("Start Date: ", fmt.Sprintf("%d-%02d-%02d", currentDate.Year(), currentDate.Month(), currentDate.Day()), fmt.Sprintf("%d:%02d", currentDate.Hour(), currentDate.Minute()))
	println("End Date: ", fmt.Sprintf("%d-%02d-%02d", targetDate.Year(), targetDate.Month(), targetDate.Day()), fmt.Sprintf("%d:%02d", targetDate.Hour(), targetDate.Minute()))
	println(fmt.Sprintf("Days: %d, Hours: %d, Minutes: %d, Seconds: %d", totalDays, totalHours, totalMinutes, totalSeconds))

	var fontData = goregular.TTF

	myFont, err := truetype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}

	// Create the context for drawing text on the image
	c := freetype.NewContext()
	c.SetDPI(150)
	c.SetFont(myFont)
	c.SetHinting(font.HintingFull)

	type Item struct {
		Label string
		Value int
	}

	// Loop through the duration, subtracting 1 second each time
	for i := 0; i < gifDuration; i++ {

		black := color.RGBA{0, 0, 0, 255}

		expired := duration <= 0

		if expired {
			// turn black to light gray
			black = color.RGBA{210, 210, 210, 255}
		}

		// Create a new image for the frame
		img := image.NewPaletted(image.Rect(0, 0, imageWidth, imageHeight), color.Palette{
			color.White,
			black,
		})

		c.SetClip(img.Bounds())
		c.SetDst(img)
		c.SetSrc(image.NewUniform(color.Black))

		// subtract 1 second from the duration
		duration = duration - time.Second
		totalDays, totalHours, totalMinutes, totalSeconds := getDaysHoursMinutesSeconds(duration)

		items := []Item{
			{Label: "days", Value: totalDays},
			{Label: "hours", Value: totalHours},
			{Label: "minutes", Value: totalMinutes},
			{Label: "seconds", Value: totalSeconds},
		}

		xOffset := textOriginOffset

		for _, item := range items {
			valueText := fmt.Sprintf("%d", item.Value)

			yOffset := 50

			// get width of valueText
			drawGroup(xOffset, yOffset, c, item.Label, valueText, myFont)

			xOffset += getLength(item.Label, fontSize, myFont)/2 + groupPadding

		}

		// Add the frame to the GIF
		anim.Image = append(anim.Image, img)
		anim.Delay = append(anim.Delay, 100)

		if expired {
			break
		}
	}

	// Create a new file for the GIF
	f, err := os.Create("countdown.gif")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Encode the GIF and write it to the file
	gif.EncodeAll(f, &anim)

	elapsed := time.Since(startTime)
	fmt.Printf("Total execution time: %s\n", elapsed)
}

/////////////////////////
// Helper Functions
/////////////////////////

func drawGroup(xOffset int, yOffset int, c *freetype.Context, labelText string, valueText string, myFont *truetype.Font) {
	drawText(xOffset, yOffset+fontSize+labelSize+valLabelPadding, c, labelText, labelSize)

	labelWidth := getLength(labelText, labelSize, myFont) / 2
	valueWidth := getLength(valueText, fontSize, myFont) / 2

	localOffset := labelWidth - valueWidth

	drawText(xOffset+localOffset, yOffset+fontSize, c, valueText, fontSize)
}

func drawText(xOffset int, yOffset int, c *freetype.Context, text string, size int) {
	pt := freetype.Pt(xOffset, yOffset)

	c.SetFontSize(float64(size))
	c.DrawString(text, pt)
}

func getLength(text string, size int, myFont *truetype.Font) int {
	localFace := truetype.NewFace(myFont, &truetype.Options{
		Size:    float64(size),
		DPI:     150,
		Hinting: font.HintingFull,
	})

	d := &font.Drawer{
		Dst:  nil,
		Src:  image.NewUniform(color.Black),
		Face: localFace,
		Dot:  fixed.P(0, 0),
	}

	return d.MeasureString(text).Ceil()
}

func getDaysHoursMinutesSeconds(duration time.Duration) (int, int, int, int) {
	totalDays := int(duration.Hours() / 24)
	if totalDays < 0 {
		totalDays = 0
	}
	totalHours := int(duration.Hours()) % 24
	if totalHours < 0 {
		totalHours = 0
	}
	totalMinutes := int(duration.Minutes()) % 60
	if totalMinutes < 0 {
		totalMinutes = 0
	}
	totalSeconds := int(duration.Seconds()) % 60
	if totalSeconds < 0 {
		totalSeconds = 0
	}

	return totalDays, totalHours, totalMinutes, totalSeconds
}
