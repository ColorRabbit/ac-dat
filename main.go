package main

import (
	"awesomeProject/dict"
	"bufio"
	"os"
	"unsafe"
)

const (
	BufferSize = 256 * 1024
)

var (
	Cache      []rune
	CacheBytes []byte
	Base       []int
	Check      []int
	Fail       map[int]int
	Info       map[int][]rune
)

func main() {
	ac := dict.HandleDict("./demo_dict.txt")
	Base = ac.Base
	Check = ac.Check
	Fail = ac.Failure
	Info = ac.Output

	Exec("./input1.txt", "./output2.txt")
}

func Exec(inputPath, outputPath string) {
	// input
	inputFile, err := os.Open(inputPath)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()
	reader := bufio.NewReaderSize(inputFile, BufferSize)

	// output
	outputFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	writer := bufio.NewWriterSize(outputFile, BufferSize)

	// process loop
	var line []byte
	for {
		line, err = reader.ReadSlice('\n')
		if err != nil {
			writer.Write(HandleFileLine(line))
			if err.Error() == "EOF" {
				break // 到达文件末尾
			}
		}
		writer.Write(HandleFileLine(line))
	}

	writer.Flush()
}

func HandleFileLine(input []byte) (output []byte) {
	var matchBegin, matchEnd []int
	state := 1
	checkLength := len(Check)
	Cache = make([]rune, len(input))
	CacheBytes = make([]byte, len(input))

	for k, c := range *(*string)(unsafe.Pointer(&input)) {
		Cache[k] = c
		CacheBytes[k] = byte(c)

	reMatch:
		if t := state + int(c) + dict.RootState; t > 0 {
			if t > checkLength {
				if state > len(Fail) {
					state = 1
					goto reMatch
				}
				state = Fail[state]
				if state <= 0 {
					state = 1
				}
				goto reMatch
			}
			if state > 1 && state != Check[t] {
				if state > len(Fail) {
					state = 1
					goto reMatch
				}
				state = Fail[state]
				if state <= 0 {
					state = 1
				}
				goto reMatch
			}
			if state == Check[t] {
				state = Base[t]

				if info := Info[state]; len(info) > 0 {
					infoLength := len(info)
					if k < infoLength {
						continue
					}
					for kk, vv := range Cache[k-infoLength+1 : k+1] {
						if vv != info[kk] {
							continue
						}
					}

					// he hers 优先匹配到he，但是需要确定hers
					begin := k - infoLength + 1
					end := k + 1
					if len(matchBegin) > 0 && matchBegin[len(matchBegin)-1] == begin {
						matchEnd[len(matchEnd)-1] = end
						continue
					}
					matchBegin = append(matchBegin, begin)
					matchEnd = append(matchEnd, end)
				}
			}
		}
	}

	if len(matchBegin) == 0 {
		return input
	}
	startTag := []uint8{227, 128, 144}
	endTag := []uint8{227, 128, 145}
	matchIndex := 0
	for k, v := range matchBegin {
		output = append(output, CacheBytes[matchIndex:v]...)
		output = append(output, startTag...)
		output = append(output, CacheBytes[v:matchEnd[k]]...)
		output = append(output, endTag...)
		matchIndex = matchEnd[k]
	}
	if matchIndex < len(input) {
		output = append(output, CacheBytes[matchIndex:len(input)]...)
	}
	return
}
