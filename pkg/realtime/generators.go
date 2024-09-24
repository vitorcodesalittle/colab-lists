package realtime

type Generator struct {
	InitialValue int64
	CurrentValue int64
	NextBuffer   chan int64
}

func NewGenerator(initialValue int64) *Generator {
	g := &Generator{
		InitialValue: initialValue,
		CurrentValue: initialValue,
		NextBuffer:   make(chan int64, 30),
	}
	g.start()
	return g
}

func (g *Generator) Next() int64 {
	return <-g.NextBuffer
}

func (g *Generator) start() {
	go func() {
		val := g.InitialValue
		for {
			g.NextBuffer <- val
			val = val + 1
		}
	}()
}
