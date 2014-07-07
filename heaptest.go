package main

import (
    "log"
    "container/heap"
)


type MeshiEvent struct {
    User string
    Message string
    Time int64
}

type MeshiEvents []MeshiEvent

func (h MeshiEvents) Len() int {
    return len(h)
}

func (h MeshiEvents) Less(i, j int) bool {
    return h[i].Time < h[j].Time
}

func (h MeshiEvents) Swap(i, j int) {
    //fmt.Printf("swap: %v <> %v\n", h[i], h[j])
    h[i], h[j] = h[j], h[i]
}

func (h *MeshiEvents) Push(x interface{}) {
    *h = append(*h, x.(MeshiEvent))
}

func (h *MeshiEvents) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}




func main() {
    log.Println("Hello!")

    m := &MeshiEvents{}

    heap.Init(m)
    
    heap.Push(m, MeshiEvent{"u","m1",123})
    heap.Push(m, MeshiEvent{"u","m2",13})
    heap.Push(m, MeshiEvent{"u","m3",999})

    log.Println((*m)[0].Message)

    log.Println(heap.Pop(m).(MeshiEvent).Message)
    log.Println(heap.Pop(m).(MeshiEvent).Message)
    log.Println(heap.Pop(m).(MeshiEvent).Message)

}
