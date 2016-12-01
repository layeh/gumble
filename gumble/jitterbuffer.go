package gumble

import (
	"container/heap"
	"encoding/binary"
	_ "fmt"
	_ "log"
	"math"
	_ "os"
	"sync"
	"time"
)

const (
	// JitterStartDelay is the starting delay we will start our buffer with
	jitterStartDelay = 100 * time.Millisecond
)

// jbAudioPacket holds pre-decoded audio samples
type jbAudioPacket struct {
	Sequence int64
	Client   *Client
	Sender   *User
	Target   *VoiceTarget
	Samples  int
	Opus     []byte
	Length   int

	HasPosition bool
	X, Y, Z     float32
	IsLast      bool
}

type jitterBufferHeap []*jbAudioPacket

func (h jitterBufferHeap) Len() int           { return len(h) }
func (h jitterBufferHeap) Less(i, j int) bool { return h[i].Sequence < h[j].Sequence }
func (h jitterBufferHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *jitterBufferHeap) Push(x interface{}) {
	*h = append(*h, x.(*jbAudioPacket))
}

func (h *jitterBufferHeap) Pop() interface{} {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

// jitterBuffer struct holds all information to run the jitter buffer
// seq is the current sequence number of the audio packets we have sent to be played
//...
type jitterBuffer struct {
	seq           int64
	delay         *time.Timer
	jitter        time.Duration
	heap          jitterBufferHeap
	bufferSamples int64
	running       bool
	user          *User
	target        *VoiceTarget
	client        *Client
	HeapLock      *sync.Mutex
	RunningLock   *sync.Mutex
}

// Creates a new jitterBuffer, starts the go routine to handle all packets
func newJitterBuffer() *jitterBuffer {
	jb := &jitterBuffer{
		running:     false,
		seq:         -1,
		jitter:      jitterStartDelay,
		HeapLock:    &sync.Mutex{},
		RunningLock: &sync.Mutex{},
	}
	//heap.Init(jb.heap)
	//jb.process() ???
	return jb
}

// AddPacket adds a packet to the jitter buffer
func (j *jitterBuffer) AddPacket(ap *jbAudioPacket) error {
	samples, err := j.user.decoder.SampleSize(ap.Opus)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%s %v\n", err, ap.Opus)
		return err
	}
	ap.Samples = samples
	j.bufferSamples += int64(ap.Samples)
	j.HeapLock.Lock()
	heap.Push(&j.heap, ap)
	j.HeapLock.Unlock()
	if !j.running {
		j.RunningLock.Lock()
		j.running = true
		j.RunningLock.Unlock()
		if j.seq == -1 || len(j.heap) == 1 { // set our sequence to first received audio packet's sequence if buffer is empty (-1)
			//println("Heap empty, or -1, sequence is now at ", ap.Sequence)
			j.seq = ap.Sequence
		}
		j.client = ap.Client
		go j.process()
	}
	return nil
}

// TODO Kill stream if no audio for X seconds, don't depend on is last
// Should only be called once internally, handles all packets.
func (j *jitterBuffer) process() {
	time.Sleep(j.jitter)
	var chans []chan *AudioPacket
	j.client.volatileLock.Lock()
	j.client.volatileWg.Wait()
	for item := j.client.Config.AudioListeners.head; item != nil; item = item.next {
		j.client.volatileLock.Unlock()
		ch := make(chan *AudioPacket)
		defer close(ch)
		chans = append(chans, ch)
		event := AudioStreamEvent{
			Client: j.client,
			User:   j.user,
			C:      ch,
		}
		item.listener.OnAudioStream(&event)
		j.client.volatileLock.Lock()
		j.client.volatileWg.Wait()
	}
	j.client.volatileLock.Unlock()
	for {

		// Handle jitter buffer/*|| float64((j.bufferSamples/(AudioSampleRate/1000))) < j.delayMs.Seconds()*1000*/
		/*if len(j.heap) == 0  {
			j.RunningLock.Lock()
			j.running = false
			j.RunningLock.Unlock()
			break // TODO handle better (could run out w/ more audio on the way)
		}*/
		if len(j.heap) == 0 {
			continue // TODO Not right? :\
		}
		j.HeapLock.Lock()
		if j.target != j.heap[0].Target {
			j.target = j.heap[0].Target
		}
		j.HeapLock.Unlock()
		if j.heap[0].Sequence < j.seq { // Throw the packet out if we have passed it due to a delay
			j.HeapLock.Lock()
			_ = heap.Pop(&j.heap)
			j.HeapLock.Unlock()
			continue
		}
		var pcm []int16
		var nextPacket *jbAudioPacket

		j.HeapLock.Lock()
		//println(j.seq, " ", j.heap[0].Sequence)
		if j.seq+1 < j.heap[0].Sequence { // Send a null packet with the missing sequence, used for loss concealment
			//println("Delayed packet, expected: ", j.seq+1, "got: ", j.heap[0].Sequence)
			var err error
			j.seq = j.heap[0].Sequence
			pcm, err = j.user.decoder.Decode(nil, 30) // TODO length of silence
			if err != nil {
				//fmt.Fprintf(os.Stderr, "%s\n", err)
				j.HeapLock.Unlock()
				continue // skip!
			}
		} else {
			var err error
			nextPacket = heap.Pop(&j.heap).(*jbAudioPacket)
			pcm, err = j.user.decoder.Decode(nextPacket.Opus[:nextPacket.Length], AudioMaximumFrameSize)
			if err != nil {

				//fmt.Fprintf(os.Stderr, "%s %v\n", err, nextPacket.Opus[:nextPacket.Length])
				frames, _ := j.user.decoder.CountFrames(nextPacket.Opus)
				j.seq = nextPacket.Sequence + int64(frames)
				j.HeapLock.Unlock()
				continue
			}
		}
		j.HeapLock.Unlock()
		if nextPacket != nil {
			frames, _ := j.user.decoder.CountFrames(nextPacket.Opus)
			j.seq = nextPacket.Sequence + int64(frames)
		}

		event := AudioPacket{
			Client:      j.client,
			Sender:      j.user,
			Target:      j.target,
			AudioBuffer: AudioBuffer(pcm),
		}

		if nextPacket != nil && len(nextPacket.Opus)-nextPacket.Length == 3*4 {
			// the packet has positional audio data; 3x float32
			event.X = math.Float32frombits(binary.LittleEndian.Uint32(nextPacket.Opus))
			event.Y = math.Float32frombits(binary.LittleEndian.Uint32(nextPacket.Opus[4:]))
			event.Z = math.Float32frombits(binary.LittleEndian.Uint32(nextPacket.Opus[8:]))
			event.HasPosition = true
		}
		for _, ch := range chans {
			ch <- &event
		}
		if nextPacket != nil {
			time.Sleep(time.Duration(nextPacket.Samples/(AudioSampleRate/1000)) * time.Millisecond)
		} else {
			time.Sleep(30 * time.Millisecond)
		}
		if (nextPacket != nil && nextPacket.IsLast) || (len(j.heap) == 0) { // we can wait for our next packets now
			j.RunningLock.Lock()
			j.running = false
			j.RunningLock.Unlock()
			break
		}
	}
}
