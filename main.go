package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func screenshot(host string) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("ignore-certificate-errors", "1"))
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	var imageBuf []byte
	if err := chromedp.Run(ctx, ScreenshotTasks("http://"+host+"/", &imageBuf)); err != nil {
		log.Printf("Error occured while trying to capture screenshot: %s", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(host+".png", imageBuf, 0644); err != nil {
		log.Printf("Error occured while trying to write screenshot to disk: %s", err)
		os.Exit(1)
	}
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

func main() {
	timeout := flag.String("timeout", "1m", "specify a timeout like 1m for one minute or 1s for one second")
	host := flag.String("host", "MISSING", "specify a host like 1.1.1.1")
	file := flag.String("file", "NOFILE", "specify a file for screenshotting in bulk")

	flag.Parse()

	timeoutPrefix := time.Minute
	timeoutAmount := 1

	if *host == "MISSING" && *file == "NOFILE" {
		fmt.Println("Too few arguments! Please supply file or host!")
		os.Exit(1)
	}
	if *host != "MISSING" && *file != "NOFILE" {
		fmt.Println("Too many arguments! Please ONLY supply file OR host!")
		os.Exit(1)
	}

	// Responsible for parsing the timeout time
	replacer := strings.NewReplacer("m", "", "s", "", "ms", "", "h", "", "ns", "")
	switch {
	case strings.Contains(*timeout, "m"):
		timeoutPrefix = time.Minute
	case strings.Contains(*timeout, "s"):
		timeoutPrefix = time.Second
	case strings.Contains(*timeout, "ms"):
		timeoutPrefix = time.Millisecond
	case strings.Contains(*timeout, "h"):
		timeoutPrefix = time.Hour
	case strings.Contains(*timeout, "ns"):
		timeoutPrefix = time.Nanosecond
	}
	timeoutAmount, err := strconv.Atoi(replacer.Replace(*timeout))
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		os.Exit(1)
	}

	if *file != "NOFILE" {

		// FILE READING SNIPPET FROM:
		// https://gist.github.com/kendellfab/7417164

		inFile, err := os.Open(*file)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
		}
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			host := scanner.Text()
			if host == "" {
				continue
			}
			go screenshot(host)
			select {
			case <-time.After(time.Duration(timeoutAmount) * timeoutPrefix):
				fmt.Printf("Timeout at %s!\n", host)
			}
		}
	} else {
		go screenshot(*host)
		select {
		case <-time.After(time.Duration(timeoutAmount) * timeoutPrefix):
			fmt.Printf("Timeout at %s!\n", *host)
		}
	}
	os.Exit(0)

}
