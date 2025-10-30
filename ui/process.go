package ui

import (
	"fmt"
	"sort"

	"github.com/rivo/tview"

	"github.com/SwarnenduG07/wtop/types"
)

func (d *Dashboard) updateProcessTable(snap *snapshot) {
	table := d.processTable
	table.Clear()

	_, _, width, _ := table.GetInnerRect()
	if width <= 0 {
		width = 120
	}

	// Build a map of PID -> GPU info (gpu index and memory used) so we can mark
	// processes that have GPU attachments in the main process table.
	gpuMap := map[int]struct {
		Index int
		Mem   float64
	}{}
	if snap != nil && snap.GPUProcesses != nil {
		for gidx, procs := range snap.GPUProcesses {
			for _, p := range procs {
				gpuMap[p.PID] = struct {
					Index int
					Mem   float64
				}{Index: gidx, Mem: p.MemoryUsed}
			}
		}
	}

	type columnDef struct {
		header string
		cell   func(*types.ProcessInfo) *tview.TableCell
	}

	columns := []columnDef{
		{
			header: "PID",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				return tview.NewTableCell(fmt.Sprintf("%6d", info.PID)).
					SetAlign(tview.AlignRight).
					SetTextColor(d.theme.Foreground)
			},
		},
		{
			header: "USER",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				return tview.NewTableCell(truncateLabel(info.User, 12)).
					SetTextColor(d.theme.Foreground)
			},
		},
		{
			header: "CPU%",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				return tview.NewTableCell(fmt.Sprintf("%5.1f", info.CPUPercent)).
					SetAlign(tview.AlignRight).
					SetTextColor(usageColor(info.CPUPercent, d.theme))
			},
		},
		{
			header: "MEM%",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				return tview.NewTableCell(fmt.Sprintf("%5.1f", info.MemPercent)).
					SetAlign(tview.AlignRight).
					SetTextColor(usageColor(float64(info.MemPercent), d.theme))
			},
		},
		{
			header: "GPU",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				// check gpuMap for this PID
				pid := int(info.PID)
				if g, ok := gpuMap[pid]; ok {
					return tview.NewTableCell(fmt.Sprintf("%d:%.0fMB", g.Index, g.Mem)).
						SetAlign(tview.AlignRight).
						SetTextColor(d.theme.Accent)
				}
				return tview.NewTableCell("").SetTextColor(d.theme.Muted)
			},
		},
		{
			header: "STATE",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				return tview.NewTableCell(info.Status).
					SetAlign(tview.AlignCenter).
					SetTextColor(d.theme.Muted)
			},
		},
	}

	// If wide enough, show GPU column indicating GPU index and memory used
	if width >= 100 {
		columns = append(columns, columnDef{
			header: "GPU",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				if snap == nil || snap.GPUProcesses == nil {
					return tview.NewTableCell("")
				}
				pid := int(info.PID)
				for gpuIdx, procs := range snap.GPUProcesses {
					for _, p := range procs {
						if p.PID == pid {
							return tview.NewTableCell(fmt.Sprintf("G%d:%4.0fMB", gpuIdx, p.MemoryUsed)).
								SetAlign(tview.AlignRight).
								SetTextColor(d.theme.Accent)
						}
					}
				}
				return tview.NewTableCell("")
			},
		})
	}

	if width >= 90 {
		columns = append(columns, columnDef{
			header: "THR",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				return tview.NewTableCell(fmt.Sprintf("%3d", info.Threads)).
					SetAlign(tview.AlignRight).
					SetTextColor(d.theme.Foreground)
			},
		})
	}
	if width >= 110 {
		columns = append(columns, columnDef{
			header: "PRI",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				return tview.NewTableCell(fmt.Sprintf("%3d", info.Priority)).
					SetAlign(tview.AlignRight).
					SetTextColor(d.theme.Foreground)
			},
		})
	}
	if width >= 120 {
		columns = append(columns, columnDef{
			header: "NI",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				return tview.NewTableCell(fmt.Sprintf("%3d", info.Nice)).
					SetAlign(tview.AlignRight).
					SetTextColor(d.theme.Foreground)
			},
		})
	}
	if width >= 140 {
		columns = append(columns,
			columnDef{
				header: "VIRT",
				cell: func(info *types.ProcessInfo) *tview.TableCell {
					return tview.NewTableCell(formatBytes(float64(info.VirtMem))).
						SetAlign(tview.AlignRight).
						SetTextColor(d.theme.Muted)
				},
			},
			columnDef{
				header: "RES",
				cell: func(info *types.ProcessInfo) *tview.TableCell {
					return tview.NewTableCell(formatBytes(float64(info.ResMem))).
						SetAlign(tview.AlignRight).
						SetTextColor(d.theme.Muted)
				},
			})
	}

	cmdWidth := clampInt(width-(len(columns)*10), 16, 48)
	columns = append(columns,
		columnDef{
			header: "TIME",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				runtime := snap.Timestamp.Unix() - info.CreateTime/1000
				if runtime < 0 {
					runtime = 0
				}
				return tview.NewTableCell(formatProcessRuntime(runtime)).
					SetAlign(tview.AlignRight).
					SetTextColor(d.theme.Muted)
			},
		},
		columnDef{
			header: "COMMAND",
			cell: func(info *types.ProcessInfo) *tview.TableCell {
				command := info.Command
				if command == "" {
					command = info.Name
				}
				return tview.NewTableCell(truncateLabel(command, cmdWidth)).
					SetTextColor(d.theme.Foreground)
			},
		})

	for col, def := range columns {
		cell := tview.NewTableCell(fmt.Sprintf("[::b]%s", def.header)).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetTextColor(d.theme.Accent).
			SetBackgroundColor(d.theme.Background)
		table.SetCell(0, col, cell)
	}

	if snap == nil || len(snap.Processes) == 0 {
		table.SetCell(1, 0, tview.NewTableCell("[yellow]no process data available[-]").
			SetSelectable(false))
		table.Select(0, 0)
		return
	}

	procs := make([]*types.ProcessInfo, len(snap.Processes))
	copy(procs, snap.Processes)

	switch d.sortMode {
	case SortByMemory:
		sort.SliceStable(procs, func(i, j int) bool {
			return procs[i].MemPercent > procs[j].MemPercent
		})
	case SortByTime:
		sort.SliceStable(procs, func(i, j int) bool {
			return procs[i].CreateTime < procs[j].CreateTime
		})
	default:
		sort.SliceStable(procs, func(i, j int) bool {
			return procs[i].CPUPercent > procs[j].CPUPercent
		})
	}

	_, _, _, height := table.GetInnerRect()
	maxRows := height - 1
	if maxRows < 5 {
		maxRows = 25
	}
	if maxRows > len(procs) {
		maxRows = len(procs)
	}

	for i := 0; i < maxRows; i++ {
		proc := procs[i]
		row := i + 1
		for col, def := range columns {
			table.SetCell(row, col, def.cell(proc))
		}
	}

	title := fmt.Sprintf(" Processes · sort: %s ", d.sortMode.String())
	table.SetTitle(title)
	table.SetTitleColor(d.theme.Accent)

	if rowCount := table.GetRowCount(); rowCount > 1 {
		currentRow, currentCol := table.GetSelection()
		if currentRow <= 0 || currentRow >= rowCount {
			table.Select(1, 0)
		} else {
			table.Select(currentRow, currentCol)
		}
	}
}
