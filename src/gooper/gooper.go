package main

import (
        "fmt"
        "code.google.com/p/portaudio-go/portaudio"
        "time"
)

func main() {
        portaudio.Initialize()
        defer portaudio.Terminate()

        l := NewLooper(1 * time.Second, 256)
        cmd := make(chan string)
        go l.DoLoop(cmd)
        time.Sleep(5 * time.Second)
        cmd <- "pause"
        time.Sleep(5 * time.Second)
        cmd <- "resume"
        time.Sleep(5 * time.Second)
        cmd <- "clear"
        time.Sleep(15 * time.Second)
}

type looper struct {
    in_stream, out_stream *portaudio.Stream
    in_buffer, out_buffer []float32
    loop_buffer []chunk
    count, chunkSize int
    loopLength time.Duration
}

type chunk struct {
    buffer []float32
}

func NewLooper(loopLength time.Duration, chunkSize int) *looper {
    // create buffers for input and output samples
    in := make([]float32, chunkSize)
    out := make([]float32, chunkSize)

    in_stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
    if err != nil {
        panic(err)
    }
    out_stream, err := portaudio.OpenDefaultStream(0, 1, 44100, len(out), out)
    if err != nil {
        panic(err)
    }

    in_stream.Start()
    out_stream.Start()

    loop_buffer := make([]chunk, 441000/chunkSize * int(loopLength.Seconds()))
    for i := range loop_buffer {
        loop_buffer[i].buffer = make([]float32, chunkSize)
    }

    l := &looper{in_stream: in_stream, out_stream: out_stream,
                 in_buffer: in, out_buffer: out,
                 loop_buffer: loop_buffer, count: 0, chunkSize: chunkSize,
                 loopLength: loopLength}

    return l

}

func (l *looper) DoLoop(cmd chan string) {
    // keep track of if the looper is active
    running := true

    // keep track of if the looper is active
    recording := true

    for {
        // check for a command
        command := ""
        select {
            case command = <-cmd:
                fmt.Println("got cmd: " + command)
            default:
                command = "none"
        }

        // perform command actions
        switch (command) {
            case "clear":
                clear_loop_buffer := make([]chunk, 441000/l.chunkSize * 
                                          int(l.loopLength.Seconds()))
                for i := range clear_loop_buffer {
                    clear_loop_buffer[i].buffer = make([]float32, l.chunkSize)
                }
                l.loop_buffer = clear_loop_buffer
            case "pause":
                running = false
            case "resume":
                running = true
        }

        // get new data from the input stream
        l.in_stream.Read()

        if running {
            // iterate over all samples in the buffers
            for i := range l.in_buffer {
                // put data from loop buffer into the output
                l.out_buffer[i] = l.loop_buffer[l.count].buffer[i]
                // add data from input into the loop buffer
                if recording {
                    l.loop_buffer[l.count].buffer[i] += l.in_buffer[i]
                }
            }
            // write to the output stream
            l.out_stream.Write()
            // make the buffer circular
            l.count = (l.count + 1) % (44100/l.chunkSize *
                                       int(l.loopLength.Seconds()))
        } else {
            // write silence to the output stream
            for i := range l.out_buffer {
                l.out_buffer[i] = 0
            }
            l.out_stream.Write()
        }
    }
}
