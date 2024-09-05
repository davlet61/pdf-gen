package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/signintech/gopdf"
)

func createPDFPart(inputImage string, start, end int, wg *sync.WaitGroup, output string) {
	defer wg.Done()

	// Create a new PDF for this part
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Get page size (A4)
	pageWidth := 595.28
	pageHeight := 841.89

	// Manually define image size (adjust as necessary)
	imgWidth := 500.0
	imgHeight := 500.0

	// Calculate scaling factor to fit image on page
	scaleW := pageWidth / imgWidth
	scaleH := pageHeight / imgHeight
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}

	// Calculate new image dimensions
	newWidth := imgWidth * scale
	newHeight := imgHeight * scale

	// Calculate position to center image on page
	x := (pageWidth - newWidth) / 2
	y := (pageHeight - newHeight) / 2

	// Add the image to the PDF for the range of pages
	for i := start; i < end; i++ {
		pdf.AddPage()

		err := pdf.Image(inputImage, x, y, &gopdf.Rect{W: newWidth, H: newHeight})
		if err != nil {
			fmt.Printf("Error adding image to page: %v\n", err)
			return
		}
	}

	// Save the partial PDF
	err := pdf.WritePdf(output)
	if err != nil {
		fmt.Printf("Error writing partial PDF: %v\n", err)
		return
	}

	fmt.Printf("Partial PDF %s created successfully.\n", output)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run script.go <input_image> <number_of_repetitions> <output_pdf>")
		os.Exit(1)
	}

	inputImage := os.Args[1]
	repetitions, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Error: Invalid number of repetitions")
		os.Exit(1)
	}
	outputPDF := os.Args[3]

	// Number of Go routines (number of workers)
	numWorkers := 4

	// Calculate how many repetitions each Go routine will handle
	pagesPerWorker := repetitions / numWorkers
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Start workers to generate parts of the PDF
	for i := 0; i < numWorkers; i++ {
		start := i * pagesPerWorker
		end := start + pagesPerWorker
		partialOutput := fmt.Sprintf("partial_%d.pdf", i)
		go createPDFPart(inputImage, start, end, &wg, partialOutput)
	}

	// Wait for all workers to complete
	wg.Wait()

	// Merge the partial PDFs
	partials := []string{}
	for i := 0; i < numWorkers; i++ {
		partials = append(partials, fmt.Sprintf("partial_%d.pdf", i))
	}

	// Call MergeCreateFile with correct number of arguments
	err = api.MergeCreateFile(partials, outputPDF, false, nil)
	if err != nil {
		fmt.Printf("Error merging PDFs: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Final PDF created successfully:", outputPDF)
}
