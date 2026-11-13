# go_playground
A playground of learning golang, It just seems so interesting.



## Victoria Transit Visualizer /protobuffer

Built a real-time data visualizer in Go that consumes BC Transit’s GTFS Realtime feed, decoding Protobuf messages into live vehicle positions and trip updates
Parsed and transformed Protobuf payloads using generated Go bindings, rendering routes  in a web page refreshed every 15 seconds, written in JavaScript, CSS, and HTML


<img width="1907" height="925" alt="victoria-bus" src="https://github.com/user-attachments/assets/e6c2ba68-125f-4618-b610-10aba9e1fab8" />

### Run this project
```bash
git clone git@github.com:samnne/go_playground.git
cd go_playground/protobuffer
go run main.go

# Terminal Two
cd /protobuffer
cd client
bun index.ts
```

Check out http://localhost:3000/ to see the visualizer
