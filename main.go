package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
	"log"
	"bufio"
	"sync"
	"archive/zip"
	"strconv"
)

var sem = make(chan int, 50)
var channel = make(chan string, 100)
var wg sync.WaitGroup
var mutex = &sync.Mutex{}

func main() {
	inputFile, err := os.Open("urllist.txt")
	if err != nil {
		log.Fatal("Error opening input file:", err)
	}

	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	urls := [100]string{}

	i := 0
	for scanner.Scan() {
		urls[i] = scanner.Text()
		i++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(scanner.Err())
	}

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 3 //Maximal 3 Verbindungen zu gleicher Domain


	wg.Add(100)
	//https://golang.org/doc/effective_go.html?h=concurrency#goroutines - entspricht der Serve-Funktion
	for _, url := range urls {
		url := url // Create new instance of req for the goroutine.
		sem <- 1
		go func(url string){
			getContentOfUrl(url)
			<-sem
		}(url)
	}

	wg.Wait()
	print("Fetching complete")
	var bound = cap(channel)/4
	core_1 := make([]string, bound)
	core_2 := make([]string, bound)
	core_3 := make([]string, bound)
	core_4 := make([]string, bound)

	for i=0; i<bound; i++{
		w:= <- channel
		x:= <- channel
		y:= <- channel
		z:= <- channel
		core_1[i] = w
		core_2[i] = x
		core_3[i] = y
		core_4[i] = z
	}
	output_name := "crawled.zip"
	//Create Zip-File
	zipFile, err := os.Create(output_name)
	if err!=nil {
		log.Fatal(err)
	}
	zipWriter := zip.NewWriter(zipFile)
	wg.Add(4)
	go writeContentToZip(core_1, 0, zipWriter)
	go writeContentToZip(core_2, 1, zipWriter)
	go writeContentToZip(core_3, 2, zipWriter)
	go writeContentToZip(core_4, 3, zipWriter)
	wg.Wait()

	err = zipWriter.Close()
	if err != nil {
		log.Fatal(err)
	}

}

func getContentOfUrl(url string){
	defer wg.Done()
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s", err)
		fmt.Printf("%s", url);
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		channel <- string(contents)
	}

}

func writeContentToZip(input[] string, slice int, zipWriter *zip.Writer){
	defer wg.Done()
	from := slice*cap(input)
	to := (slice+1)*cap(input)
	j := 0
	for i := from; i<to; i++{
		filename := strconv.Itoa(i)+".html"
		mutex.Lock()
		f, err := zipWriter.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write([]byte(input[j]))
		if err != nil {
			log.Fatal(err)
		}
		mutex.Unlock()
		j++
	}
}


