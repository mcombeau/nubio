package nubio

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

const (
	marginSide            = 50
	fontSize              = 10
	fontSizeHeading       = 15
	fontSizeTitle         = 20
	colLeftSize           = 60
	a4WidthPt, a4HeightPt = 595.28, 842.89
)

func ExportPDF(w io.Writer, p *Profile) error {
	pdf := fpdf.New("P", "pt", "A4", "")
	pdf.SetCreationDate(time.Now())
	pdf.SetLang("en")
	pdf.SetAuthor(p.Name, true)
	pdf.SetTitle("Curriculum Vitae - "+p.Name, true)
	pdf.SetFont("Arial", "", fontSize)

	pdf.SetTopMargin(50)
	pdf.SetLeftMargin(marginSide)
	pdf.SetRightMargin(marginSide)
	pdf.SetTextColor(10, 10, 10)

	// Define footer.
	pdf.AliasNbPages("{max_page}")
	pdf.SetFooterFuncLpi(func(isLastPage bool) {
		pdf.SetFontStyle("")
		pdf.SetFontSize(fontSize)
		pdf.SetTextColor(50, 50, 50)
		txt := fmt.Sprintf("Page %d/{max_page}", pdf.PageCount())
		pdf.Text(marginSide, a4HeightPt-3*fontSize, txt)
	})

	// Append title (name).
	pdf.AddPage()
	pdf.SetFontSize(fontSizeTitle)
	pdf.SetFontStyle("B")
	pdf.MultiCell(0, fontSizeTitle, p.Name, "", "C", false)

	// Append horizontal line below title.
	pdf.Ln(fontSizeTitle)
	pdf.Rect(marginSide, pdf.GetY(), a4WidthPt-2*marginSide, 0.5, "F")

	// Append work experiences.
	pdf.Ln(24)
	writeHeading(pdf, "Experiences")
	for _, v := range p.Experiences {
		pdf.Ln(16)
		pdf.SetFontSize(fontSize)
		pdf.SetFontStyle("B")
		pdf.SetTextColor(30, 30, 30)
		pdf.MultiCell(0, fontSize, v.Title+" at "+v.Organization, "", "", false)
		pdf.Ln(6)
		pdf.SetFontStyle("")
		pdf.SetTextColor(50, 50, 50)
		pdf.MultiCell(0, fontSize, v.Description, "", "", false)
		pdf.Ln(6)

		writeKV(pdf, "Duration", v.From+" to "+v.To)
		writeKV(pdf, "Location", v.Location)
		writeKV(pdf, "Skills", strings.Join(v.Skills, ", "))
	}

	// Append skills.
	pdf.AddPage()
	writeHeading(pdf, "Skills")
	for _, v := range p.Skills {
		pdf.Ln(16)
		pdf.SetFontSize(fontSize)
		pdf.SetFontStyle("B")
		pdf.SetTextColor(10, 10, 10)
		pdf.MultiCell(0, fontSize, v.Title, "", "", false)
		pdf.Ln(8)
		pdf.SetFontStyle("")
		pdf.SetTextColor(80, 80, 80)
		pdf.MultiCell(0, fontSize, strings.Join(v.Tools, ", "), "", "", false)
	}

	// Append languages.
	pdf.Ln(24)
	writeHeading(pdf, "Languages")
	pdf.Ln(16)
	for _, v := range p.Languages {
		writeKV(pdf, v.Label, v.Proficiency)
	}

	// Append education.
	pdf.Ln(24)
	writeHeading(pdf, "Education")
	pdf.Ln(8)
	for _, v := range p.Education {
		pdf.Ln(8)
		pdf.SetFontSize(fontSize)
		pdf.SetFontStyle("B")
		pdf.SetTextColor(30, 30, 30)
		pdf.MultiCell(0, fontSize, v.Title, "", "", false)
		pdf.Ln(6)
		writeKV(pdf, "School", v.Organization)
		writeKV(pdf, "Duration", v.From+" to "+v.To)
	}

	// Append interests.
	pdf.Ln(24)
	writeHeading(pdf, "Interests")
	pdf.Ln(8)
	for _, v := range p.Interests {
		pdf.Ln(6)
		pdf.SetFontSize(fontSize)
		pdf.SetFontStyle("")
		pdf.SetTextColor(50, 50, 50)
		pdf.MultiCell(0, fontSize, "- "+v, "", "", false)
	}

	// Append hobbies.
	pdf.Ln(24)
	writeHeading(pdf, "Hobbies")
	pdf.Ln(8)
	for _, v := range p.Hobbies {
		pdf.Ln(6)
		pdf.SetFontSize(fontSize)
		pdf.SetFontStyle("")
		pdf.SetTextColor(50, 50, 50)
		pdf.MultiCell(0, fontSize, "- "+v, "", "", false)
	}

	// Append links.
	pdf.AddPage()
	pdf.Ln(24)
	writeHeading(pdf, "Links")
	pdf.Ln(8)
	for _, v := range p.Links {
		pdf.Ln(8)
		pdf.SetFontSize(fontSize)
		pdf.SetFontStyle("B")
		pdf.CellFormat(50, fontSize, v.Label, "", 0, "", false, 0, "")
		pdf.SetFontStyle("U")
		pdf.CellFormat(0, fontSize, v.URL, "", 2, "", false, 0, "https://"+v.URL)
	}

	// Append contact.
	pdf.Ln(24)
	writeHeading(pdf, "Contact")
	addContactLink(pdf, "Email address", p.Contact.EmailAddress, "mailto:"+p.Contact.EmailAddress)
	if p.Contact.PGP != "" {
		addContactLink(pdf, "PGP key", p.Contact.PGP, "https://"+p.Contact.PGP)
	}
	// Write whole PDF.
	return pdf.Output(w)
}

func writeHeading(pdf fpdf.Pdf, heading string) {
	pdf.Bookmark(heading, 0, -1)
	pdf.SetFontSize(fontSizeHeading)
	pdf.SetFontStyle("B")
	pdf.SetTextColor(10, 10, 10)
	pdf.MultiCell(0, fontSizeHeading, heading, "", "", false)
}

func writeKV(pdf fpdf.Pdf, k, v string) {
	pdf.SetFontStyle("")
	pdf.SetFontSize(fontSize)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(colLeftSize, fontSize, k, "", 0, "", false, 0, "")
	pdf.SetTextColor(100, 100, 100)
	pdf.MultiCell(0, fontSize, v, "", "", false)
	pdf.Ln(4)
}

func addContactLink(pdf fpdf.Pdf, k, v, url string) {
	pdf.Ln(8)
	pdf.SetFontSize(fontSize)
	pdf.SetFontStyle("B")
	pdf.CellFormat(100, fontSize, k, "", 0, "", false, 0, "")
	pdf.SetFontStyle("U")
	pdf.CellFormat(0, fontSize, v, "", 2, "", false, 0, url)
}