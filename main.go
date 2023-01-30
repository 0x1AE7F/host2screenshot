package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) <= 0 {
		log.Println("Missing argument IP!")
		os.Exit(1)
	}
	if argsWithoutProg[0] == "" {
		log.Println("Argument IP cant be empty!")
		os.Exit(1)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("ignore-certificate-errors", "1"))
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	log.Printf("Current IP: %s", argsWithoutProg[0])

	var imageBuf []byte
	if err := chromedp.Run(ctx, ScreenshotTasks("http://"+argsWithoutProg[0]+"/", &imageBuf)); err != nil {
		log.Printf("Error occured while trying to capture screenshot: %s", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(argsWithoutProg[0]+".png", imageBuf, 0644); err != nil {
		log.Printf("Error occured while trying to write screenshot to disk: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func ScreenshotTasks(url string, imageBuf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			*imageBuf, err = page.CaptureScreenshot().WithQuality(90).Do(ctx)
			return err
		}),
	}
}
