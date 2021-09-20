package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type animation struct {
	start    time.Time
	duration time.Duration
}
type C = layout.Context
type D = layout.Dimensions

func main() {
	go func() {
		// create new window
		window := app.NewWindow(
			app.Title("Egg timer"),
			app.Size(unit.Dp(400), unit.Dp(600)),
		)
		if err := draw(window); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

// animate starts an animation at the current frame which will last for the
// provided duration.
func (a *animation) animate(gtx C, duration time.Duration) {
	a.start = gtx.Now
	a.duration = duration
	op.InvalidateOp{}.Add(gtx.Ops)
}

// stop ends the animation immediately.
func (a *animation) stop() {
	a.duration = time.Duration(0)
}

// progress returns whether the animation is currently running and (if so) how
// far through the animation it is.
func (a animation) progress(gtx layout.Context) (animating bool, progress float32) {
	if gtx.Now.After(a.start.Add(a.duration)) {
		return false, 0
	}
	op.InvalidateOp{}.Add(gtx.Ops)
	return true, float32(gtx.Now.Sub(a.start)) / float32(a.duration)
}

func draw(window *app.Window) error {

	// the operations from the UI
	var operations op.Ops

	// startButton is a clickable widget
	var startBtn widget.Clickable

	//boilDurationInput is a field to input boil duration
	var boilDurationInput widget.Editor
	// track the progress of the boil animation
	var anim animation

	// th defines the material design style
	theme := material.NewTheme(gofont.Collection())

	for {
		select {
		// listen for events in the window.
		case events := <-window.Events():

			// detect what type of event
			switch events := events.(type) {

			// this is sent when the application should re-render.
			case system.FrameEvent:
				gtx := layout.NewContext(&operations, events)
				boiling, progress := anim.progress(gtx)
				if startBtn.Clicked() {
					// Start (or stop) the boil
					if boiling {
						anim.stop()
					} else {
						// Read from the input box
						inputString := boilDurationInput.Text()
						inputString = strings.TrimSpace(inputString)
						inputFloat, _ := strconv.ParseFloat(inputString, 32)
						anim.animate(gtx, time.Duration(inputFloat)*time.Second)
					}
				}

				// flexbox layout concept
				layout.Flex{
					Axis: layout.Vertical,
					// Empty space is left at the start, i.e. at the top
					Spacing: layout.SpaceStart,
				}.Layout(gtx,

					//the egg
					layout.Rigid(
						func(gtx C) D {
							//draw a cusom path, shaped like an egg
							var eggPath clip.Path
							op.Offset(f32.Pt(200, 150)).Add(gtx.Ops)
							eggPath.Begin(gtx.Ops)

							//rotate from 0 to 360 degress
							for degress := 0.0; degress <= 360; degress++ {

								//convert degress to radians
								rad := degress / 360 * 2 * math.Pi
								// Trig gives the distance in X and Y direction
								cosT := math.Cos(rad)
								sinT := math.Sin(rad)
								// Constants to define the eggshape
								a := 110.0
								b := 150.0
								d := 20.0
								// The x/y coordinates
								x := a * cosT
								y := -(math.Sqrt(b*b-d*d*cosT*cosT) + d*sinT) * sinT
								// Finally the point on the outline
								p := f32.Pt(float32(x), float32(y))
								// Draw the line to this point
								eggPath.LineTo(p)
							}
							// Close the path
							eggPath.Close()

							// Get hold of the actual clip
							eggArea := clip.Outline{Path: eggPath.End()}.Op()

							// Fill the shape
							color := color.NRGBA{R: 255, G: uint8(239 * (1 - progress)), B: uint8(174 * (1 - progress)), A: 255}
							paint.FillShape(gtx.Ops, color, eggArea)

							dimensions := image.Point{Y: 335}
							return layout.Dimensions{Size: dimensions}
						},
					),
					//the inputbox
					layout.Rigid(
						func(gtx C) D {
							//define characteristics of the input box
							boilDurationInput.SingleLine = true
							boilDurationInput.Alignment = text.Middle

							// Count down the text when boiling
							if boiling && progress < 1 {
								boilRemain := (1 - progress) * float32(anim.duration.Seconds())
								// Format to 1 decimal.
								// Using the good old multiply-by-10-divide-by-10 trick to get rounded values with 1 decimal
								inputStr := fmt.Sprintf("%.1f", math.Round(float64(boilRemain)*10)/10)
								boilDurationInput.SetText(inputStr)
							}

							//Define insets
							margins := layout.Inset{
								Top:    unit.Dp(0),
								Bottom: unit.Dp(40),
								Right:  unit.Dp(170),
								Left:   unit.Dp(170),
							}

							//define borders
							border := widget.Border{
								Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
								CornerRadius: unit.Dp(3),
								Width:        unit.Dp(2),
							}

							// ... and material design
							input := material.Editor(theme, &boilDurationInput, "sec")

							return margins.Layout(gtx,
								func(gtx C) D {
									return border.Layout(gtx, input.Layout)
								},
							)
						},
					),
					//the progress bar
					layout.Rigid(
						func(gtx C) D {
							bar := material.ProgressBar(theme, progress)
							if boiling && progress < 1 {
								op.InvalidateOp{At: gtx.Now.Add(time.Second / 25)}.Add(&operations)
							}
							return bar.Layout(gtx)
						},
					),
					//the button
					layout.Rigid(
						func(gtx C) D {
							// defining a set of margins
							margins := layout.Inset{
								Top:    unit.Dp(25),
								Bottom: unit.Dp(25),
								Right:  unit.Dp(35),
								Left:   unit.Dp(35),
							}
							// Then we lay out within those margins ...
							return margins.Layout(gtx,
								// create a button
								func(gtx C) D {
									var text string
									if !boiling {
										text = "Start"
									}
									if boiling && progress < 1 {
										text = "Stop"
									}
									if boiling && progress >= 1 {
										text = "Finished"
									}
									btn := material.Button(theme, &startBtn, text)
									return btn.Layout(gtx)
								},
							)
						},
					),
				)
				events.Frame(gtx.Ops)

			// this is sent when the application is closed.
			case system.DestroyEvent:
				return events.Err

			}

		}
	}
}
