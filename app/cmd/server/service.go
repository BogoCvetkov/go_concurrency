package main

import (
	"log"
	"sync"
	"time"

	"github.com/signintech/gopdf"
)

func generatePDf(cont string, n string, wg *sync.WaitGroup) error {
	defer wg.Done()

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	pdf.AddTTFFont("font", "font.ttf")
	pdf.SetFont("font", "", 14)

	pdf.AddPage()

	pdf.Cell(nil, cont)
	err := pdf.WritePdf(n)

	if err != nil {
		log.Print(err.Error())
		return err
	}

	time.Sleep(time.Second * 2)
	return nil
}
