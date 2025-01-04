package main

import (
	"fmt"
)

func displayDosHeaderDetails(ui *MyAppUI, dosHeader *DOSHeader) {

	details := fmt.Sprintf(
		`DOS Header:
	e_magic: %#x
	e_cblp: %#x
	e_cp: %#x
	e_crlc: %#x
	e_cparhdr: %#x
	e_minalloc: %#x
	e_maxalloc: %#x
	e_ss: %#x
	e_sp: %#x
	e_csum: %#x
	e_ip: %#x
	e_cs: %#x
	e_lfarlc: %#x
	e_ovno: %#x
	e_res: %#x
	e_oemid: %#x
	e_oeminfo: %#x
	e_res2: %#x    
	e_ifanew: %#x`,
		dosHeader.E_magic,
		dosHeader.E_cblp,
		dosHeader.E_cp,
		dosHeader.E_crlc,
		dosHeader.E_cparhdr,
		dosHeader.E_minalloc,
		dosHeader.E_maxalloc,
		dosHeader.E_ss,
		dosHeader.E_sp,
		dosHeader.E_csum,
		dosHeader.E_ip,
		dosHeader.E_cs,
		dosHeader.E_lfarlc,
		dosHeader.E_ovno,
		dosHeader.E_res,
		dosHeader.E_oemid,
		dosHeader.E_oeminfo,
		dosHeader.E_res2,
		dosHeader.E_ifanew,
	)

	// Display the details in the right pane
	ui.rightPane.SetText(details)
}
