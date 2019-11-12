package main

import (
    "bufio"
    "fmt"
    "io"
    "encoding/json"
    "log"
    "os"
    "strings"
    "github.com/fsnotify/fsnotify"
)

func main() {
    fileName := "log_lines"
    
    if len( os.Args ) < 2 {
        fmt.Println("specify a log to view / tail:\n  wdaproxy\n  stf_device_ios\n  device_trigger\n  video_enabler\n  stf_provider\n  ffmpeg\n")
        os.Exit( 0 )
    }
    
    findProc := os.Args[1]
    
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    
    fh, err := os.Open( fileName )
    if err != nil {
        panic(err)
    }
    defer fh.Close()
    
    size := fileSize( fh )
    //fh.Seek( size, io.SeekStart )
    
    scanner  := bufio.NewScanner( fh )
    for scanner.Scan() {
        checkLine( []byte( scanner.Text() ), findProc )
    }
    
    err = watcher.Add( fileName )
    if err != nil {
        log.Fatal(err)
    }
    for {
        select {
            case event := <-watcher.Events:
                if event.Op & fsnotify.Write == fsnotify.Write {
                    //fmt.Println("modify")
                    newSize := fileSize( fh )
                    
                    newBytes := newSize - size
                    
                    if newBytes > 0 {
                        //fmt.Printf("  dif: %d\n", newBytes )
                        
                        //f.Seek(pos, io.SeekStart)
                        buf := make( []byte, newBytes )
                        fh.Read( buf )
                        //fmt.Printf("  \"%s\"\n", string( buf ) )
                        
                        checkLine( buf, findProc )
                        
                        size = newSize
                    }
                }
        }
    }
}

func checkLine( data []byte, findProc string ) {
    var dat map[string]interface{}
    
    startJ := strings.Index( string(data), "{" )
    endJ := strings.LastIndex( string(data), "}" )
    
    part := string(data)[ startJ : (endJ + 1) ]
    
    decoder := json.NewDecoder( strings.NewReader( part ) )
    for {
        err := decoder.Decode( &dat )
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }
        
        proc := dat["proc"].(string)
        if proc == findProc {
            //fmt.Println(dat)
            line := dat["line"].(string)
            fmt.Println( line )
        }
    }
}

func fileSize( fh *os.File ) (int64) {
    newinfo, err := fh.Stat()
    if err != nil {
        panic(err)
    }
    return newinfo.Size()
}