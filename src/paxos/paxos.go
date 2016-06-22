package paxos

import (
	"net"
	"net/rpc"
//	"strings"
	"log"
	
//	"os"
//	"syscall"
	"sync"
//	"sync/atomic"
	"fmt"
	"time"
	"math/rand"
)

var totalServer int

type PaxosInstance struct {

	lock sync.Mutex
	PrepareN int
	ProposeN int
	ProposeV interface{}
	decided bool
	proposed bool 
	decidedValue interface{}
}

type Paxos struct{
	me int
	peers []string
	minimumSeq int
	maximumSeq int
	instance []*PaxosInstance
	l  net.Listener
	lock sync.Mutex
	done	int
	isdead bool
	hasDecided int
}

type PrepareMessage struct{
	Seq int
	PrepareN int
}

type PrepareACK struct{
	Mes string
	ProposeN int
	ProposeV interface{}
}

type ProposeMessage struct{
	Seq int
	ProposeN int
	ProposeV interface{}
}

type ProposeACK struct{
	Mes string
	ProposeN int
}

type DecidedMessage struct{
	Seq int
	V interface{}
}

type TmpValue struct{
	Value int
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b 
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b 
}

func (px *Paxos) Max() int{
	return px.maximumSeq
}

func (px *Paxos) Min() int {
	return px.minimumSeq
}

func (px *Paxos) PrepareHandler(args *PrepareMessage, reply *PrepareACK) error {
	
	px.lock.Lock()
	defer px.lock.Unlock()
	if args.Seq > px.maximumSeq {
		tmp := args.Seq - px.maximumSeq
		for i:= 1; i<=tmp; i++ {
			px.instance = append(px.instance, new(PaxosInstance))
		}
		px.maximumSeq = args.Seq
	}
	
	
	p := px.instance[args.Seq]
	//fmt.Println("prepare", args.PrepareN, p.PrepareN)
	
	p.lock.Lock()
	defer p.lock.Unlock()
	if args.PrepareN > p.PrepareN {
		p.PrepareN = args.PrepareN
		reply.Mes = "ok"
		reply.ProposeN = p.ProposeN
		reply.ProposeV = p.ProposeV	
	} else {
		reply.Mes = "reject"
		reply.ProposeN = p.PrepareN
	}
	return nil
}

func (px *Paxos) ProposeHandler(args *ProposeMessage, reply *ProposeACK) error {
	px.lock.Lock()
	defer px.lock.Unlock()
	if args.Seq > px.maximumSeq {
		tmp := args.Seq - px.maximumSeq
		for i:= 1; i<=tmp; i++ {
			px.instance = append(px.instance, new(PaxosInstance))
		}
		px.maximumSeq = args.Seq
	}
	
	//fmt.Println("propose", px.me, px.peers[px.me])
	p := px.instance[args.Seq]
	
	p.lock.Lock()
	defer p.lock.Unlock()
	
	if args.ProposeN >= p.PrepareN {
		p.PrepareN = args.ProposeN
		p.ProposeN = args.ProposeN
		p.ProposeV = args.ProposeV
		reply.Mes = "ok"
		reply.ProposeN = p.PrepareN
	} else {
		reply.Mes = "reject"
		reply.ProposeN = p.PrepareN
	}

	return nil
}
func (px *Paxos) Name(args *DecidedMessage, reply *string) error {
	fmt.Println(px.me)
	return nil
}

func (px *Paxos) updateDecided() {
	t := len(px.instance)
	
	for (px.hasDecided < t-1 && px.instance[px.hasDecided+1].decided) {
		px.hasDecided++
	}
}

func (px *Paxos) DecidedHandler(args *DecidedMessage, reply *string) error {
	px.lock.Lock()
	defer px.lock.Unlock()
	
	//fmt.Println("decide", px.me, px.peers[px.me])
	if args.Seq > px.maximumSeq {
		tmp := args.Seq - px.maximumSeq
		for i:= 1; i<=tmp; i++ {
			px.instance = append(px.instance, new(PaxosInstance))
		}
		px.maximumSeq = args.Seq
	}
	p := px.instance[args.Seq]
	
	p.lock.Lock()
	defer p.lock.Unlock()
	
	p.decided = true
	p.decidedValue = args.V
	str := "ok"
	reply = &str
	px.updateDecided()
	return nil
}

func (px *Paxos) HasDecidedHandler(args *int, reply *TmpValue) error {
	px.lock.Lock()
	defer px.lock.Unlock()
	//fmt.Println("has decided")
	reply.Value = px.hasDecided
	return nil
}

func (px *Paxos) Propose(seq int, value interface{}, p *PaxosInstance) {
	
	currentMaxN := 0
	
	for true {
	
		p.lock.Lock()
		if p.decided {
			p.lock.Unlock()
			return 
		}
		p.lock.Unlock()
		
		px.lock.Lock()
		if px.isdead {
			px.lock.Unlock()
			return 
		}
		px.lock.Unlock()


		//fmt.Println("currentMaxN", currentMaxN, "totalServer",totalServer)
		chooseN := (currentMaxN / totalServer + 1) * totalServer + px.me
		mes := &PrepareMessage{seq, chooseN}
		var prepareReply []*PrepareACK = make([]*PrepareACK, totalServer, totalServer)
		for i := 0; i < totalServer; i++ {
			prepareReply[i] = new(PrepareACK)
		}
		
		wg:=new(sync.WaitGroup)	
		
		for i := 0; i< totalServer; i++ {
			if i != px.me {
				wg.Add(1)
				go Call(wg, px.peers[i], "Paxos.PrepareHandler", mes, prepareReply[i])
			} else {
				px.PrepareHandler(mes, prepareReply[i])
				
			}
		}
		wg.Wait()
		cnt := 0
		tmpMaxN := 0
		var tmpValue interface{}
		for _, reply := range(prepareReply){
			if reply.Mes == "ignore" || reply.Mes == "" {
				cnt ++
			} else if reply.Mes == "reject" {
				cnt ++
				currentMaxN = max(currentMaxN, reply.ProposeN)
			} else if reply.Mes == "ok" {
				cnt --
				currentMaxN = max(currentMaxN, reply.ProposeN)
				if reply.ProposeN > tmpMaxN {
					tmpMaxN = reply.ProposeN
					tmpValue = reply.ProposeV
				}
			}
		}
		
		//fmt.Println(px.me, cnt, tmpMaxN, tmpValue)
		if cnt >= 0 {
			continue;
			//break;
		}
		
		
		if tmpMaxN == 0 {
			tmpValue = value
		}
		
		var proposeReply []*ProposeACK = make([]*ProposeACK, totalServer, totalServer)
		for i := 0; i < totalServer; i++ {
			proposeReply[i] = new(ProposeACK)
		}
		mes1 := &ProposeMessage{seq, chooseN, tmpValue}
		wg = new(sync.WaitGroup)	
		for i := 0; i < totalServer; i++ {
			if i != px.me {
				wg.Add(1)
				go Call(wg, px.peers[i], "Paxos.ProposeHandler", mes1, proposeReply[i])
			} else {
				px.ProposeHandler(mes1, proposeReply[i])
			}
		}
		wg.Wait()
		
		cnt = 0
		for i := 0; i < totalServer; i++ {
			if proposeReply[i].Mes == "ok" {
				cnt --
			} else if proposeReply[i].Mes == "reject" || 
					proposeReply[i].Mes == "ignore"  || proposeReply[i].Mes == "" {
				cnt ++
			} 
		}
		
		if cnt >= 0 {
			continue
		}
		//fmt.Println(px.me)
		
		p.lock.Lock()
		p.decided = true
		p.decidedValue = tmpValue
		p.lock.Unlock()
		
		px.updateDecided()
		
		mes2 := &DecidedMessage{seq, tmpValue}
		
		for i := 0; i < totalServer; i++ { //is not neccessary to ensure other server receive this message 
			if i != px.me {
				wg.Add(1)
				go Call(wg, px.peers[i], "Paxos.DecidedHandler", mes2, new(string))
			}
		}
		
		break
		
	} 
}

func (px *Paxos) Start(seq int, value interface{}){
	px.lock.Lock()
	
	if seq > px.maximumSeq {
		tmp := seq - px.maximumSeq
		for i:= 1; i<=tmp; i++ {
			px.instance = append(px.instance, new(PaxosInstance))
		}
		px.maximumSeq = seq
	}
	
	p := px.instance[seq]
	
	if p.decided || p.proposed {
		px.lock.Unlock()
		return
	}
	
	p.proposed = true
	px.lock.Unlock()
	go px.Propose(seq, value, p)
	
}

func (px *Paxos) Status(seq int) (decided bool, v interface{}){
	px.lock.Lock()
	defer px.lock.Unlock()
	if seq > px.maximumSeq {
		return false, nil 
	}
	p := px.instance[seq]
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.decided, p.decidedValue
}

func (px *Paxos) Done(seq int) {
	px.lock.Lock()
	px.done = max(px.done, seq)
	px.lock.Unlock()
}

func (px *Paxos) forget() {
	px.lock.Lock()
	tmp := min(px.done, px.hasDecided)
	px.lock.Unlock()

	if tmp <= -1 {
		return
	}
	
	wg := new(sync.WaitGroup)
	var Reply *TmpValue
	
	Reply = new(TmpValue)
	
	for i := 0; i < totalServer; i++ {
		if i != px.me {
			wg.Add(1)
			Call(wg, px.peers[i], "Paxos.HasDecidedHandler", &tmp, Reply)
			tmp = min(Reply.Value, tmp)
		}
		
	//	fmt.Println(px.me, i)
	}
	
	px.lock.Lock()
	tmp = min(px.maximumSeq, tmp)
	//fmt.Println(tmp)
	if tmp >= px.minimumSeq {
		for i := px.minimumSeq; i <= tmp; i++ {
			px.instance[i] = nil
		} 
		px.minimumSeq = tmp + 1
	}
	px.lock.Unlock()
}

func (px *Paxos) Kill(){
	px.lock.Lock()
	px.isdead = true
	if px.l != nil {
		px.l.Close()
	}
	px.lock.Unlock()
}

func Make(peers []string, me int) *Paxos { 
	px := new(Paxos)
	px.me = me
	px.peers = peers
	px.maximumSeq = -1
	px.instance = make([]*PaxosInstance, 0, 0)
	
	totalServer = len(peers)

	px.hasDecided = -1
	px.done = -1
	
	newServer := rpc.NewServer()
	newServer.Register(px)
	//fmt.Println(peers[me])
	l, e := net.Listen("tcp", peers[me])
	if e != nil {
			log.Fatal("listen error: ", e)
		}
	px.l = l
	go func() {
			for true {
				px.lock.Lock()
				if px.isdead == true {
					px.lock.Unlock()
					break
				}
				px.lock.Unlock()
				conn, err := px.l.Accept()
				if err == nil{
					px.lock.Lock()
					if px.isdead == true {
						px.lock.Unlock()
						conn.Close()
						break
					}
					//fmt.Println(px.me, conn)
					px.lock.Unlock()
					go newServer.ServeConn(conn)
				}
				if err != nil && px.isdead == false {
					fmt.Printf("Paxos(%v) accept the RPC connection: %v\n", me, err.Error())
				}
			}
		}()	
		
	go func() {
	
			for {
				time.Sleep(100*time.Millisecond)
				px.forget()
			}
		}()
	return px	
}

func Call(wg *sync.WaitGroup, srv string, name string, args interface{}, reply interface{}) bool {
	
	defer wg.Done()
	
	if rand.Int() % 100 > 100 {
		return false
	}
	
	c, err := rpc.Dial("tcp", srv)
	
	if err != nil {
		//err1 := err.(*net.OpError)
	//	if err1.Err != syscall.ENOENT && err1.Err != syscall.ECONNREFUSED {
			//fmt.Printf("paxos Dial() failed: %v\n", err1)
	//	}
		return false
	}
	defer c.Close()

	//fmt.Println(srv, name)
	err = c.Call(name, args, reply)
	//c.Call("Paxos.Name", args, reply)
	if err == nil {
		return true
	}
	fmt.Println(err)
	return false
}

func main() {
/*	s := []string{"127.0.0.1:1234", "127.0.0.2:1235", "127.0.0.3:1236", "127.0.0.4:1237", "127.0.0.5:1238","127.0.0.6:1238","127.0.0.7:1238","127.0.0.8:1238","127.0.0.9:1238",
	"127.0.0.10:1238","127.0.0.11:1238","127.0.0.12:1238","127.0.0.13:1238","127.0.0.14:1238"}
*/
	s := []string{"127.0.0.1:1234","127.0.0.2:1234","127.0.0.3:1234", "127.0.0.4:1234", "127.0.0.5:1234"}
	var px [25]*Paxos
	T := 1
	totalServer = len(s)
	for i:= totalServer-1; i >= 0; i-- {
		px[i] = Make(s, i)
	}
	
	//statusTest 
	fmt.Println("status Test")
	for i:= T-1; i >= 0; i-- {
		for j:= 0; j < totalServer; j++ {
			go px[j].Start(i, i*totalServer+j)
		}
	}
	
	time.Sleep(time.Second)
		for i:= 0; i < T; i++ {
			for j:= 0; j < totalServer; j++ {
				fmt.Println(px[j].Status(i))
			}
		}

	//doneTest
	fmt.Println("done Test")
	px[1].Done(0)
	
	for i:= 1; i >= 1; i-- {
		for j:= 0; j < totalServer; j++ {
			go px[j].Start(i, i*totalServer+j)
		}
	}
	
	time.Sleep(time.Second)
		for i:= 1; i < 2; i++ {
			for j:= 0; j < totalServer; j++ {
				fmt.Println(px[j].Status(i))
			}
		}
	
	
	//maximum-minimum Test
	fmt.Println("maximum-minimum Test")
	
	fmt.Println(px[1].Max(), px[1].Min())
	
	//killTest
	fmt.Println("kill Test")
	px[0].Kill()
	
	
	for i:= 2; i >= 2; i-- {
		for j:= 0; j < totalServer; j++ {
			go px[j].Start(i, i*totalServer+j)
		}
	}
	
	time.Sleep(time.Second)
		for i:= 2; i < 3; i++ {
			for j:= 0; j < totalServer; j++ {
				fmt.Println(px[j].Status(i))
			}
		}
	//startTheSameSEQ test
	fmt.Println("propose the same seq Test")
	fmt.Println(px[1].Status(1))
	px[1].Start(1, 5)
	time.Sleep(time.Second)
	fmt.Println(px[1].Status(1))
	
	//Disorder Propose Test
	fmt.Println("disorder propose Test")
	px[1].Start(100, 100)
	time.Sleep(time.Second)
	fmt.Println(px[2].Status(100))
	
}
