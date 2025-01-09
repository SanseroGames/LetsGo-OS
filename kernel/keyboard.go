package kernel

type Keystate struct {
	Keycode uint8
}

// -------------------
// Keyboard Ringbuffer
// -------------------

type KeyboardRing struct {
	Ring   GenericRing
	Buffer [32]Keystate
}

func (r *KeyboardRing) Init() {
	r.Ring.Cap = r.Cap()
}

func (r *KeyboardRing) Len() int {
	return r.Ring.Len()
}

func (r *KeyboardRing) Cap() int {
	return len(r.Buffer)
}

func (r *KeyboardRing) Push(s Keystate) {
	// Not thread safe
	if i := r.Ring.Push(); i != -1 {
		r.Buffer[i] = s
	}
}

func (r *KeyboardRing) Pop() *Keystate {
	// Not thread safe
	if i := r.Ring.Pop(); i != -1 {
		return &r.Buffer[i]
	}
	return nil
}

// End keyboard ring buffer

const (
	keyboardInputPort = 0x60
)

var buffer KeyboardRing = KeyboardRing{
	Ring: GenericRing{}, // Important to prevent initialization at runtime
}

var tempKeystate Keystate = Keystate{}

//go:nospilt
func handleKeyboard() {
	keycode := Inb(keyboardInputPort) // TODO: constant Where to get this?
	tempKeystate.Keycode = keycode
	buffer.Push(tempKeystate)
}

func InitKeyboard() {
	RegisterPICHandler(1, handleKeyboard)
	EnableIRQ(1)
	buffer.Init()
}
