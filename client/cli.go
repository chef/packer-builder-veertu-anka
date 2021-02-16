package client

// type CLI interface {
// 	RunCommand(args ...string) (MachineReadableOutput, error)
// 	RunCommandStreamer(outputStreamer chan string, args ...string) (MachineReadableOutput, error)
// }

// type AnkaCLI struct {
// }

// func (cli *AnkaCLI) RunCommand(args ...string) (MachineReadableOutput, error) {
// 	return cli.RunCommandStreamer(nil, args...)
// }

// func (cli *AnkaCLI) RunCommandStreamer(outputStreamer chan string, args ...string) (MachineReadableOutput, error) {
// 	if outputStreamer != nil {
// 		args = append([]string{"--debug"}, args...)
// 	}

// 	cmdArgs := append([]string{"--machine-readable"}, args...)
// 	log.Printf("Executing anka %s", strings.Join(cmdArgs, " "))
// 	cmd := exec.Command("anka", cmdArgs...)

// 	outPipe, err := cmd.StdoutPipe()
// 	if err != nil {
// 		log.Println("Err on stdoutpipe")
// 		return MachineReadableOutput{}, err
// 	}

// 	if outputStreamer == nil {
// 		cmd.Stderr = cmd.Stdout
// 	}

// 	if err = cmd.Start(); err != nil {
// 		log.Printf("Failed with an error of %v", err)
// 		return MachineReadableOutput{}, err
// 	}
// 	outScanner := bufio.NewScanner(outPipe)
// 	outScanner.Split(customSplit)

// 	for outScanner.Scan() {
// 		out := outScanner.Text()
// 		log.Printf("%s", out)

// 		if outputStreamer != nil {
// 			outputStreamer <- out
// 		}
// 	}

// 	scannerErr := outScanner.Err() // Expecting error on final output
// 	if scannerErr == nil {
// 		return MachineReadableOutput{}, errors.New("missing machine readable output")
// 	}
// 	if _, ok := scannerErr.(customErr); !ok {
// 		return MachineReadableOutput{}, err
// 	}

// 	finalOutput := scannerErr.Error()
// 	log.Printf("%s", finalOutput)

// 	parsed, err := parseOutput([]byte(finalOutput))
// 	if err != nil {
// 		return MachineReadableOutput{}, err
// 	}
// 	if err := cmd.Wait(); err != nil {
// 		return MachineReadableOutput{}, err
// 	}

// 	if err = parsed.GetError(); err != nil {
// 		return MachineReadableOutput{}, err
// 	}

// 	return parsed, nil
// }

// type MachineReadableError struct {
// 	*MachineReadableOutput
// }

// func (ae MachineReadableError) Error() string {
// 	return ae.Message
// }

// type MachineReadableOutput struct {
// 	Status        string `json:"status"`
// 	Body          json.RawMessage
// 	Message       string `json:"message"`
// 	Code          int    `json:"code"`
// 	ExceptionType string `json:"exception_type"`
// }

// func (parsed *MachineReadableOutput) GetError() error {
// 	if parsed.Status != statusOK {
// 		return MachineReadableError{parsed}
// 	}
// 	return nil
// }

// func parseOutput(output []byte) (MachineReadableOutput, error) {
// 	var parsed MachineReadableOutput
// 	if err := json.Unmarshal(output, &parsed); err != nil {
// 		return parsed, err
// 	}

// 	return parsed, nil
// }

// func dropCR(data []byte) []byte {
// 	if len(data) > 0 && data[len(data)-1] == '\r' {
// 		return data[0 : len(data)-1]
// 	}
// 	return data
// }

// func customSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
// 	// A tiny spin off on ScanLines

// 	if atEOF && len(data) == 0 {
// 		return 0, nil, nil
// 	}
// 	if i := bytes.IndexByte(data, '\n'); i >= 0 {
// 		return i + 1, dropCR(data[0:i]), nil
// 	}
// 	if atEOF { // Machine readable data is parsed here
// 		out := dropCR(data)
// 		return len(data), out, customErr{data: out}
// 	}
// 	return 0, nil, nil
// }

// type customErr struct {
// 	data []byte
// }

// func (e customErr) Error() string {
// 	return string(e.data)
// }
