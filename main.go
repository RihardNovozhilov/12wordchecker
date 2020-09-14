// sudo apt-get install -y wv
package main
import (

	"log"
	"os"
	"path/filepath"
	"strings"
	"net/http"
	"fmt"
	"io/ioutil"
	"github.com/h2non/filetype"
	"code.sajari.com/docconv"
	"regexp"
	"sync"
	"runtime"

)

func getContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}

func scan_recursive(dir_path string, ignore []string) ([]string, []string) {

	folders := []string{}
	files := []string{}

	
	filepath.Walk(dir_path, func(path string, f os.FileInfo, err error) error {

		_continue := false

		// Loop : Ignore Files & Folders
		for _, i := range ignore {

			// If ignored path
			if strings.Index(path, i) != -1 {

				// Continue
				_continue = true
			}
		}

		if _continue == false {

			f, err = os.Stat(path)

			// If no error
			if err != nil {
				log.Fatal(err)
			}

			// File & Folder Mode
			f_mode := f.Mode()

			// Is folder
			if f_mode.IsDir() {

				// Append to Folders Array
				folders = append(folders, path)

				// Is file
			} else if f_mode.IsRegular() {

				// Append to Files Array
				files = append(files, path)
			}
		}

		return nil
	})

	return folders, files
}


func getWords() string{

	if data, err := getContent("https://raw.githubusercontent.com/bitcoin/bips/master/bip-0039/english.txt"); err != nil {
		panic(err)
	} else {

		s := string(data)
		log.Println("Words parsed succsefuly")
		return s
	}

}

func scanForWords(wg *sync.WaitGroup, fileArray []string, wordArray []string){
	defer wg.Done()
	var inFileFoundArray []string
	for _, i := range fileArray {
		inFileFoundArray = nil
		counter := 0
		isValidForCheck := false
		file, err := ioutil.ReadFile(i)
		fileString := string(file)
		// f, err := os.Open(i)
		// defer f.Close()

		if err != nil {
			panic(err)
		}

		kind, _ := filetype.Match(file)

		switch kind.Extension {
			case "doc":

				doc, errDoc := os.Open(i)
				defer doc.Close()

				if errDoc != nil {
					panic(errDoc)
				}

				resp, _, _ := docconv.ConvertDoc(doc)

				fileString = resp
				isValidForCheck = true

			case "docx":

				
				docx, errDocx := os.Open(i)
				defer docx.Close()

				if errDocx != nil {
					panic(errDocx)
				}

				resp, _, _ := docconv.ConvertDocx(docx)

				fileString = resp
				isValidForCheck = true
			case "rtf":
				isValidForCheck = true
			
		}

		if kind == filetype.Unknown {
			isValidForCheck = true
		} 

		if !isValidForCheck {
			continue
		}
		// Matching words 

			for _, element := range wordArray {
				checkFlag, _ := regexp.MatchString("\\b"+element+"\\b", fileString)
				if checkFlag && element != "" {
							// log.Println("File", i , "contains:", element)
							counter = counter + 1
							inFileFoundArray = append(inFileFoundArray,element)
						}
						if counter == findLimit {
							log.Println("In File", i , "found:", inFileFoundArray)
							break
						}

			}
	 

	}

	
}

var (
	findLimit = 12
	numCPU = runtime.NumCPU()
	divided [][]string
	wg sync.WaitGroup
)

func main() {


		s := getWords()
		wordArray := strings.Split(s, "\n")

		fmt.Print("Enter Dir: ")   //Print function is used to display output in same line
        var userDir string    
        fmt.Scanln(&userDir) 

		_, files := scan_recursive(userDir, []string{".DS_Store", ".idea", "/.idea/", "/.idea"})

		chunkSize := (len(files) + numCPU - 1) / numCPU
		
		for i := 0; i < len(files); i += chunkSize {
			end := i + chunkSize
	
			if end > len(files) {
				end = len(files)
			}
	
			divided = append(divided, files[i:end])
		}

		for i := 0; i < len(divided); i++ {
			fmt.Println("Main: Starting worker")
			wg.Add(1)
			go scanForWords(&wg, divided[i], wordArray)
		}


	fmt.Println("Main: Waiting for workers to finish")
	wg.Wait()
	fmt.Println("Main: Completed")

	

}