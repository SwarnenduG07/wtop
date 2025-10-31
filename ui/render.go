package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	refreshInterval   = 2 * time.Second
	maxProcessEntries = 256
	historySize       = 180
)

type SortMode int

const (
	SortByCPU SortMode = iota
	SortByMemory
	SortByTime
)

func (s SortMode) String() string {
	switch s {
	case SortByMemory:
		return "Mem"
	case SortByTime:
		return "Time"
	default:
		return "CPU"
	}
}

type netRates struct {
	Up    float64
	Down  float64
	Valid bool
}

type Dashboard struct {
	app *tview.Application

	root        *tview.Flex
	leftFlex    *tview.Flex
	rightFlex   *tview.Flex
	mainFlex    *tview.Flex
	cpuFlex     *tview.Flex
	memDiskFlex *tview.Flex
	gpuFlex     *tview.Flex
	header      *tview.TextView
	lastRates   netRates

	cpuView      *tview.TextView
	memoryView   *tview.TextView
	diskView     *tview.TextView
	gpuView      *tview.TextView
	processTable *tview.Table
	footer       *tview.TextView

	refreshInterval time.Duration
	ticker          *time.Ticker
	stopCh          chan struct{}

	sortMode     SortMode
	lastSnapshot *snapshot

	prevNetSent  uint64
	prevNetRecv  uint64
	prevSnapshot time.Time

	cpuHistory   *sparkHistory
	memHistory   *sparkHistory
	swapHistory  *sparkHistory
	diskHistory  *sparkHistory
	netUpHistory *sparkHistory
	netDnHistory *sparkHistory
	gpuHistory   map[int]*sparkHistory

	lastLayoutWidth int
}

func NewDashboard() *Dashboard {
	app := tview.NewApplication()
	dash := &Dashboard{
		app:             app,
		refreshInterval: refreshInterval,
		stopCh:          make(chan struct{}),
		sortMode:        SortByCPU,
	}

	dash.header = dash.newSection(" SUMMARY ")
	dash.header.SetWrap(false)

	dash.cpuView = dash.newSection(" CPU ")
	dash.cpuView.SetWrap(false)

	dash.memoryView = dash.newSection(" MEMORY ")
	dash.memoryView.SetWrap(false)

	dash.diskView = dash.newSection(" DISKS ")
	dash.diskView.SetWrap(false)

	dash.gpuView = dash.newSection(" GPU ")
	dash.gpuView.SetWrap(false)

	dash.processTable = tview.NewTable().SetBorders(false)
	dash.processTable.SetBackgroundColor(tcell.ColorBlack)
	dash.processTable.SetTitle(" Processes ")
	dash.processTable.SetTitleColor(tcell.ColorLightCyan)
	dash.processTable.SetBorder(true)
	dash.processTable.SetBorderColor(tcell.ColorDarkSlateGray)
	dash.processTable.SetSelectable(true, false)
	dash.processTable.SetFixed(1, 0)
	dash.processTable.SetSelectedStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorLightCyan).
		Bold(true))

	dash.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false).
		SetWrap(false)
	dash.footer.SetBackgroundColor(tcell.ColorDimGray)

	dash.cpuFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(dash.cpuView, 0, 1, false)

	dash.memDiskFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(dash.memoryView, 0, 1, false).
		AddItem(dash.diskView, 0, 1, false)

	dash.gpuFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(dash.gpuView, 0, 1, false)

	dash.leftFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(dash.cpuFlex, 0, 1, false).
		AddItem(dash.memDiskFlex, 0, 1, false).
		AddItem(dash.gpuFlex, 0, 2, false)

	dash.rightFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(dash.processTable, 0, 1, true)

	dash.mainFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(dash.leftFlex, 0, 1, false).
		AddItem(dash.rightFlex, 0, 1, false)

	dash.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(dash.header, 3, 0, false).
		AddItem(dash.mainFlex, 0, 1, false).
		AddItem(dash.footer, 2, 0, false)

	dash.root.SetBackgroundColor(tcell.ColorBlack)
	dash.header.SetBackgroundColor(tcell.ColorBlack)
	dash.cpuView.SetBackgroundColor(tcell.ColorBlack)
	dash.memoryView.SetBackgroundColor(tcell.ColorBlack)
	dash.diskView.SetBackgroundColor(tcell.ColorBlack)
	dash.gpuView.SetBackgroundColor(tcell.ColorBlack)

	dash.cpuHistory = newSparkHistory(historySize)
	dash.memHistory = newSparkHistory(historySize)
	dash.swapHistory = newSparkHistory(historySize)
	dash.diskHistory = newSparkHistory(historySize)
	dash.netUpHistory = newSparkHistory(historySize)
	dash.netDnHistory = newSparkHistory(historySize)
	dash.gpuHistory = make(map[int]*sparkHistory)

	dash.app.SetRoot(dash.root, true)
	dash.app.EnableMouse(true)
	dash.bindKeys()
	dash.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, _ := screen.Size()
		dash.reflowLayout(width)
		return false
	})

	return dash
}

func (d *Dashboard) Run() error {
	initial, err := collectSnapshot(maxProcessEntries)
	if err == nil {
		d.lastSnapshot = initial
		d.applySnapshot(initial, false)
	} else {
		d.header.SetText(fmt.Sprintf("[red]failed to gather metrics: %v[-]", err))
	}

	d.ticker = time.NewTicker(d.refreshInterval)
	go d.updateLoop()

	d.app.SetFocus(d.processTable)
	errRun := d.app.Run()
	d.stop()
	return errRun
}

func (d *Dashboard) stop() {
	if d.ticker != nil {
		d.ticker.Stop()
	}
	select {
	case <-d.stopCh:

	default:
		close(d.stopCh)
	}
}

func (d *Dashboard) updateLoop() {
	for {
		select {
		case <-d.stopCh:
			return
		case <-d.ticker.C:
			snap, err := collectSnapshot(maxProcessEntries)
			if err != nil {
				d.app.QueueUpdateDraw(func() {
					d.footer.SetText(fmt.Sprintf("[red]metrics error: %v[-]", err))
				})
				continue
			}
			d.app.QueueUpdateDraw(func() {
				d.applySnapshot(snap, true)
			})
		}
	}
}

func (d *Dashboard) applySnapshot(snap *snapshot, fromLoop bool) {
	if snap == nil {
		return
	}
	d.lastSnapshot = snap
	rates := d.computeNetworkRates(snap, fromLoop)
	d.lastRates = rates
	d.recordHistory(snap, rates)
	d.updateHeader(snap, rates)
	d.updateCPU(snap)
	d.updateMemory(snap)
	d.updateGPU(snap)
	d.updateProcessTable(snap)
	d.updateFooter(snap, rates)
}

func (d *Dashboard) recordHistory(snap *snapshot, rates netRates) {
	if snap == nil {
		return
	}
	if d.cpuHistory != nil {
		d.cpuHistory.Push(snap.TotalCPU)
	}
	if snap.Memory != nil && d.memHistory != nil {
		d.memHistory.Push(snap.Memory.UsedPercent)
	}
	if d.swapHistory != nil {
		if snap.Swap != nil && snap.Swap.Total > 0 {
			d.swapHistory.Push(snap.Swap.UsedPercent)
		} else {
			d.swapHistory.Push(0)
		}
	}
	if d.diskHistory != nil && snap.Disk != nil && snap.Disk.Total > 0 {
		percent := (float64(snap.Disk.Used) / float64(snap.Disk.Total)) * 100
		d.diskHistory.Push(percent)
	}
	if rates.Valid {
		if d.netUpHistory != nil {
			d.netUpHistory.Push(rates.Up)
		}
		if d.netDnHistory != nil {
			d.netDnHistory.Push(rates.Down)
		}
	}
	if len(snap.GPUInfos) > 0 {
		for _, gpu := range snap.GPUInfos {
			hist := d.gpuHistory[gpu.Index]
			if hist == nil {
				hist = newSparkHistory(historySize)
				d.gpuHistory[gpu.Index] = hist
			}
			hist.Push(gpu.Utilization)
		}
	}
}

func (d *Dashboard) computeNetworkRates(snap *snapshot, fromLoop bool) netRates {
	if snap == nil {
		return netRates{}
	}
	if d.prevSnapshot.IsZero() || !fromLoop {
		d.prevNetSent = snap.NetBytesSent
		d.prevNetRecv = snap.NetBytesRecv
		d.prevSnapshot = snap.Timestamp
		return netRates{Valid: false}
	}

	elapsed := snap.Timestamp.Sub(d.prevSnapshot).Seconds()
	if elapsed <= 0 {
		return netRates{Valid: false}
	}

	up := 0.0
	if snap.NetBytesSent >= d.prevNetSent {
		up = float64(snap.NetBytesSent-d.prevNetSent) / elapsed
	}
	down := 0.0
	if snap.NetBytesRecv >= d.prevNetRecv {
		down = float64(snap.NetBytesRecv-d.prevNetRecv) / elapsed
	}

	d.prevNetSent = snap.NetBytesSent
	d.prevNetRecv = snap.NetBytesRecv
	d.prevSnapshot = snap.Timestamp

	return netRates{Up: up, Down: down, Valid: true}
}

func (d *Dashboard) cycleSortMode() {
	switch d.sortMode {
	case SortByCPU:
		d.sortMode = SortByMemory
	case SortByMemory:
		d.sortMode = SortByTime
	default:
		d.sortMode = SortByCPU
	}
	if d.lastSnapshot != nil {
		d.updateProcessTable(d.lastSnapshot)
		d.updateFooter(d.lastSnapshot, d.lastRates)
	}
}

func (d *Dashboard) newSection(title string) *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)
	tv.SetTitle(title)
	tv.SetBorder(true)
	tv.SetBorderColor(tcell.ColorDarkSlateGray)
	tv.SetTitleColor(tcell.ColorLightCyan)
	tv.SetBackgroundColor(tcell.ColorBlack)
	return tv
}

func (d *Dashboard) reflowLayout(width int) {
	if width <= 0 {
		return
	}
	if width == d.lastLayoutWidth {
		return
	}
	d.lastLayoutWidth = width

	if width < 100 {
		d.mainFlex.SetDirection(tview.FlexRow)
		d.mainFlex.ResizeItem(d.leftFlex, 0, 2)
		d.mainFlex.ResizeItem(d.rightFlex, 0, 3)
	} else {
		d.mainFlex.SetDirection(tview.FlexColumn)
		d.mainFlex.ResizeItem(d.leftFlex, 0, 1)
		d.mainFlex.ResizeItem(d.rightFlex, 0, 1)
	}

	if width < 80 {
		d.footer.SetWrap(true)
	} else {
		d.footer.SetWrap(false)
	}
}
