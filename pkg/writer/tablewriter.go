package writer

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

//NodeWrite
func NodeWrite(data [][]string, resourceType string, outType bool) {
	//var table *tablewriter.Table

	table := table(outType)
	switch {
	case resourceType == "cpu":
		table.SetHeader([]string{"NODE", "CPU USE", "CPU REQ", "CPU REQ(%)", "CPU LIM", "CPU LIM(%)", "CPU Capacity"})
		for _, i := range data {
			table.Append(i)
		}
		table.Render()
	case resourceType == "memory":
		table.SetHeader([]string{"NODE", "MEM USE", "MEM REQ", "MEM REQ(%)", "MEM LIM", "MEM LIM(%)", "MEM Capacity"})
		for _, i := range data {
			table.Append(i)
		}
		table.Render()
	case resourceType == "pod":
		table.SetHeader([]string{"NODE", "Pod Allocated", "Pod Capacity", "Pod(%)"})
		for _, i := range data {
			table.Append(i)
		}
		table.Render()
	default:
		table.SetHeader([]string{"NODE", "CPU USE", "CPU REQ", "CPU REQ(%)", "CPU LIM", "CPU LIM(%)", "CPU Capacity",
			"MEM USE", "MEM REQ", "MEM REQ(%)", "MEM LIM", "MEM LIM(%)", "MEM Capacity",
			"Pod Allocated", "Pod Capacity", "Pod(%)"})
		for _, i := range data {
			table.Append(i)
		}
		table.Render()
	}

}

//PodWrite
func PodWrite(data [][]string, resourceType string, outType bool) {
	//var table *tablewriter.Table

	table := table(outType)
	switch {
	case resourceType == "cpu":
		table.SetHeader([]string{"NAMESPACE", "POD NAME",
			"CPU USE", "CPU USE(%)", "CPU REQ", "CPU LIM"})
		for _, i := range data {
			table.Append(i)
		}
		table.Render()
	case resourceType == "memory":
		table.SetHeader([]string{"NAMESPACE", "POD NAME",
			"MEM USE", "MEM USE(%)", "MEM REQ", "MEM LIM"})
		for _, i := range data {
			table.Append(i)
		}
		table.Render()
	default:
		table.SetHeader([]string{"NAMESPACE", "POD NAME",
			"CPU USE", "CPU USE(%)", "CPU REQ", "CPU LIM",
			"MEM USE", "MEM USE(%)", "MEM REQ", "MEM LIM"})
		for _, i := range data {
			table.Append(i)
		}
		table.Render()
	}

}

//table
func table(outType bool) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	if outType {
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t") // pad with tabs
		table.SetNoWhiteSpace(true)
		return table
	}
	return table
}
