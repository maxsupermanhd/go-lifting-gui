package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/maxsupermanhd/go-lifting"
)

type inputData struct {
	Version              string
	VersionInitialized   bool `json:"-"`
	Structure            string
	StructureInitialized bool `json:"-"`
	X, Z                 string
	XInitialized         bool `json:"-"`
	ZInitialized         bool `json:"-"`
}

func inputsToDatapoints(input []inputData) []lifting.Datapoint {
	output := []lifting.Datapoint{}
	for _, i := range input {
		is := lifting.DesertPyramid
		for _, s := range lifting.Structures {
			if s.String() == i.Structure {
				is = s
				break
			}
		}
		iv := lifting.MC_NEWEST
		for _, s := range lifting.Versions {
			if s.String() == i.Version {
				iv = s
				break
			}
		}
		ix, _ := strconv.Atoi(i.X)
		iz, _ := strconv.Atoi(i.Z)
		output = append(output, lifting.NewDP(is, iv, ix, iz))
	}
	return output
}
func main() {
	// d := []lifting.Datapoint{
	// 	lifting.NewDP(lifting.Shipwreck, lifting.MC_1_17, -4914, 86450),
	// 	lifting.NewDP(lifting.Shipwreck, lifting.MC_1_17, 80990, 416),
	// 	lifting.NewDP(lifting.Shipwreck, lifting.MC_1_17, -48884, 8587),
	// 	lifting.NewDP(lifting.SwampHut, lifting.MC_1_17, 14356, 119044),
	// 	lifting.NewDP(lifting.SwampHut, lifting.MC_1_17, -66110, 92835),
	// }
	// prog := make(chan lifting.LiftingProgress, 2)
	// go func() {
	// 	for p := range prog {
	// 		log.Printf("%#v", p)
	// 	}
	// }()
	// structureSeeds := lifting.LiftStructures(d, prog, nil)
	// close(prog)
	// fmt.Printf("Got %d structure seeds: %v\n", len(structureSeeds), structureSeeds)
	// worldSeeds := []int64{}
	// for _, v := range structureSeeds {
	// 	worldSeeds = append(worldSeeds, lifting.StructureSeedToWorldSeeds(v)...)
	// }
	// fmt.Printf("Got %d world seeds: %v\n", len(worldSeeds), worldSeeds)

	versionList := []string{}
	for _, v := range lifting.Versions {
		versionList = append(versionList, v.String())
	}
	structuresList := []string{}
	for _, v := range lifting.Structures {
		structuresList = append(structuresList, v.String())
	}

	tolift := []inputData{}

	intValidator := validation.NewRegexp(`^-?\d+$`, "Must be integer")
	a := app.New()
	w := a.NewWindow("Lifting")
	w.Resize(fyne.NewSize(650, 800))

	// delete, version, x, z, structure

	t := widget.NewTable(func() (int, int) {
		return 1 + len(tolift), 5
	}, func() fyne.CanvasObject {
		return container.NewMax(
			widget.NewLabel("template1"),
			widget.NewButton("template2", func() {}),
			widget.NewSelect([]string{}, func(s string) {}),
			widget.NewEntry())
	}, func(cell widget.TableCellID, obj fyne.CanvasObject) {
		wl := obj.(*fyne.Container).Objects[0].(*widget.Label)
		wb := obj.(*fyne.Container).Objects[1].(*widget.Button)
		ws := obj.(*fyne.Container).Objects[2].(*widget.Select)
		we := obj.(*fyne.Container).Objects[3].(*widget.Entry)
		wl.Hide()
		wb.Hide()
		ws.Hide()
		we.Hide()
		dataIndex := cell.Row - 1
		switch cell.Col {
		case 0:
			if cell.Row == 0 {
				wl.SetText("Delete")
				wl.Show()
			} else {
				wb.SetText("Delete")
				wb.OnTapped = func() {
					tolift = append(tolift[:dataIndex], tolift[dataIndex+1:]...)
				}
				wb.Show()
			}
		case 1:
			if cell.Row == 0 {
				wl.SetText("Version")
				wl.Show()
			} else {
				ws.Options = versionList
				ws.OnChanged = func(s string) {
					tolift[dataIndex].Version = s
				}
				if !tolift[dataIndex].VersionInitialized {
					ws.Selected = tolift[dataIndex].Version
					tolift[dataIndex].VersionInitialized = true
				}
				ws.Show()
			}
		case 2:
			if cell.Row == 0 {
				wl.SetText("Structure")
				wl.Show()
			} else {
				ws.Options = structuresList
				ws.OnChanged = func(s string) {
					tolift[dataIndex].Structure = s
				}
				if !tolift[dataIndex].StructureInitialized {
					ws.Selected = tolift[dataIndex].Structure
					tolift[dataIndex].StructureInitialized = true
				}
				ws.Show()
			}
		case 3:
			if cell.Row == 0 {
				wl.SetText("Chunk X")
				wl.Show()
			} else {
				we.Validator = intValidator
				we.OnChanged = func(s string) {
					tolift[dataIndex].X = s
				}
				if !tolift[dataIndex].XInitialized {
					we.Text = tolift[dataIndex].X
					tolift[dataIndex].XInitialized = true
				}
				we.Show()
			}
		case 4:
			if cell.Row == 0 {
				wl.SetText("Chunk Z")
				wl.Show()
			} else {
				we.Validator = intValidator
				we.OnChanged = func(s string) {
					tolift[dataIndex].Z = s
				}
				if !tolift[dataIndex].ZInitialized {
					we.Text = tolift[dataIndex].Z
					tolift[dataIndex].ZInitialized = true
				}
				we.Show()
			}
		}
	})

	prog := make(chan lifting.LiftingProgress, 2)
	lstop := make(chan struct{})
	foundStructureSeeds := []int64{}
	foundWorldSeeds := []int64{}

	var startLiftingBtn *widget.Button
	var stopLiftingBtn *widget.Button
	startLiftingBtn = widget.NewButton("Start lifting", func() {
		dp := inputsToDatapoints(tolift)
		foundStructureSeeds = []int64{}
		startLiftingBtn.Disable()
		stopLiftingBtn.Enable()
		go func() {
			lifting.LiftStructures(dp, prog, lstop)
			startLiftingBtn.Enable()
			stopLiftingBtn.Disable()
		}()
	})
	stopLiftingBtn = widget.NewButton("Stop lifting", func() {
		lstop <- struct{}{}
		startLiftingBtn.Enable()
	})
	stopLiftingBtn.Disable()
	addInputBtn := widget.NewButton("Add structure", func() {
		tolift = append(tolift, inputData{
			Version:              lifting.MC_NEWEST.String(),
			VersionInitialized:   false,
			Structure:            lifting.Shipwreck.String(),
			StructureInitialized: false,
			X:                    "0",
			Z:                    "0",
			XInitialized:         false,
			ZInitialized:         false,
		})
		t.Refresh()
	})
	saveInputsBtn := widget.NewButton("Save structures", func() {
		dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if uc != nil {
				b, err := json.MarshalIndent(tolift, "", "\t")
				if err != nil {
					log.Println(err)
				}
				uc.Write(b)
				uc.Close()
			}
		}, w).Show()
	})
	loadInputsBtn := widget.NewButton("Load structures", func() {
		dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if uc != nil {
				b, err := io.ReadAll(uc)
				if err != nil {
					log.Println(err)
				}
				err = json.Unmarshal(b, &tolift)
				if err != nil {
					log.Println(err)
				}
			}
		}, w).Show()
	})
	btnbox := container.NewMax(container.NewHBox(startLiftingBtn, stopLiftingBtn, addInputBtn, saveInputsBtn, loadInputsBtn))

	progressLower := widget.NewProgressBar()
	progressHigher := widget.NewProgressBar()
	progressStructureSeedsCount := widget.NewLabel("Found structure seeds: 0")
	saveStructureSeedsBtn := widget.NewButton("Save structure seeds", func() {
		dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if uc != nil {
				fmt.Fprint(uc, "Structure seeds from go-lifting-gui\n\n")
				b, err := json.Marshal(tolift)
				if err != nil {
					log.Println(err)
				}
				fmt.Fprint(uc, "Input data is: ", string(b), "\n\n")
				for _, v := range foundStructureSeeds {
					fmt.Fprintln(uc, v)
				}
				uc.Close()
			}
		}, w).Show()
	})
	css := container.NewHBox(progressStructureSeedsCount, saveStructureSeedsBtn)
	progressWorldSeedsCount := widget.NewLabel("Found world seeds: 0")
	saveWorldSeedsBtn := widget.NewButton("Save world seeds", func() {
		dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if uc != nil {
				fmt.Fprint(uc, "World seeds from go-lifting-gui\n\n")
				b, err := json.Marshal(tolift)
				if err != nil {
					log.Println(err)
				}
				fmt.Fprint(uc, "Input data is: ", string(b), "\n\n")
				for _, v := range foundWorldSeeds {
					fmt.Fprintln(uc, v)
				}
				uc.Close()
			}
		}, w).Show()
	})
	cws := container.NewHBox(progressWorldSeedsCount, saveWorldSeedsBtn)

	go func() {
		for p := range prog {
			progressLower.SetValue(p.LowerProgress)
			progressLower.Refresh()
			progressLower.TextFormatter = func() string {
				return fmt.Sprintf("Lower: %d/%d (%.0f%%)", p.LowerCurrent, p.LowerMax, p.LowerProgress*100)
			}
			progressHigher.SetValue(p.HigherProgress)
			progressHigher.Refresh()
			progressHigher.TextFormatter = func() string {
				return fmt.Sprintf("Higher: %d/%d (%.0f%%)", p.HigherCurrent, p.HigherMax, p.HigherProgress*100)
			}
			foundStructureSeeds = append(foundStructureSeeds, p.Found...)
			for _, v := range p.Found {
				foundWorldSeeds = append(foundWorldSeeds, lifting.StructureSeedToWorldSeeds(v)...)
			}
			progressStructureSeedsCount.SetText(fmt.Sprint("Found structure seeds: ", len(foundStructureSeeds)))
			progressWorldSeedsCount.SetText(fmt.Sprint("Found world seeds: ", len(foundWorldSeeds)))
		}
	}()

	content := container.NewBorder(nil, container.NewVBox(btnbox, progressLower, progressHigher, css, cws), nil, nil, t)
	// content := container.NewVSplit(tc, startLiftingBtn)
	// content.SetOffset(0.8)
	w.SetContent(content)
	w.ShowAndRun()
}
