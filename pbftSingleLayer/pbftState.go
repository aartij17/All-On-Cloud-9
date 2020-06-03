package pbftSingleLayer

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/config"
	"math"
	"time"
)

type pbftState struct {
	viewNumber        int
	currentTimestamp  int
	viewChangeCounter int
	timerIsRunning    bool
	candidateNumber   int
	suffix            string
	counter           map[reducedMessage]int
	timeoutTimer      *time.Timer
	//localLog          []common.Message
	messageOut chan common.Transaction
}

func newPbftState(suffix string) *pbftState {
	newState := pbftState{
		viewNumber:        0,
		currentTimestamp:  -1,
		viewChangeCounter: 0,
		timerIsRunning:    false,
		candidateNumber:   1,
		suffix:            suffix,
		counter:           make(map[reducedMessage]int),
		timeoutTimer:      nil,
		//localLog:          make([]common.Message, 0),
		messageOut: make(chan common.Transaction),
	}

	return &newState
}

func (state *pbftState) setTimer() {
	if !state.timerIsRunning {
		state.timeoutTimer.Reset(TIMEOUT * time.Second)
		state.timerIsRunning = true
	}
}

func (state *pbftState) stopTimer() {
	state.timeoutTimer.Stop()
	state.timerIsRunning = false
}

func (state *pbftState) hasQuorum(message reducedMessage) bool {
	//debugTxt := ""
	//for key, val := range state.counter {
	//	if val > 0 {
	//		debugTxt += key.messageType + ", " + strconv.Itoa(key.appId) + " : " + strconv.Itoa(val) + "\n"
	//	}
	//}
	//println(debugTxt)
	for i := 0; i < config.GetAppCnt(); i++ {
		message.appId = i
		if config.IsByzantineTolerant(config.GetAppName(i)) {
			left := config.GetAppNodeCntInt(i) - state.counter[message]
			if float64(left) >= float64(state.counter[message])/float64(2) {
				return false
			}
		} else {
			left := config.GetAppNodeCntInt(i) - state.counter[message]
			if left >= state.counter[message] {
				return false
			}
		}
	}

	return true
}

func (state *pbftState) markAsUsed(message reducedMessage) bool {
	for i := 0; i < config.GetAppCnt(); i++ {
		message.appId = i
		state.counter[message] = -math.MaxInt16
	}

	return true
}

func (state *pbftState) handleMessage(
	message common.Message,
	broadcast func(common.Message),
	isLeader func() bool,
	isSuggestedLeader func(int) bool,
	getId func() int,
) {
	_txn := *message.Txn
	_message := message
	_message.Txn = &_txn

	switch _message.MessageType {
	case NEW_VIEW:
		state.viewChangeCounter = 0
		state.stopTimer()
		state.viewNumber = _message.Timestamp
		state.candidateNumber = state.viewNumber + 1
	case VIEW_CHANGE:
		//if isSuggestedLeader(_message.Timestamp) {
		//	state.viewChangeCounter++
		//	//println("inside VIEW_CHANGE", getId(), state.viewChangeCounter, 2*state.failureTolerance+1)
		//	if state.viewChangeCounter == 2*state.failureTolerance+1 {
		//		//println("broadcasting NEW_VIEW")
		//		go broadcast(common.Message{
		//			MessageType: NEW_VIEW,
		//			Timestamp:   _message.Timestamp,
		//			FromNodeNum: getId(),
		//			Txn:         &dummyTxn,
		//		})
		//	}
		//}
	case NEW_MESSAGE:
		if isLeader() {
			state.currentTimestamp++
			go broadcast(common.Message{
				MessageType: PRE_PREPARE,
				Timestamp:   state.currentTimestamp,
				FromNodeNum: getId(),
				Txn:         _message.Txn,
			})
		} else {
			state.setTimer()
		}
	case PRE_PREPARE:
		state.stopTimer()
		go broadcast(common.Message{
			MessageType: PREPARE,
			Timestamp:   _message.Timestamp,
			FromNodeNum: getId(),
			Txn:         _message.Txn,
		})
	case PREPARE:
		reduced := reducedMessage{
			messageType: PREPARE,
			Txn:         newReducedTransaction(*_message.Txn),
			appId:       config.GetAppId(message.FromApp),
		}

		state.counter[reduced]++
		if state.hasQuorum(reduced) {
			state.markAsUsed(reduced)
			go broadcast(common.Message{
				MessageType: COMMIT,
				Timestamp:   _message.Timestamp,
				FromNodeNum: getId(),
				Txn:         _message.Txn,
			})
		}
	case COMMIT:
		reduced := reducedMessage{
			messageType: COMMIT,
			Txn:         newReducedTransaction(*_message.Txn),
			appId:       config.GetAppId(message.FromApp),
		}

		state.counter[reduced]++
		if state.hasQuorum(reduced) {
			state.markAsUsed(reduced)
			go broadcast(common.Message{
				MessageType: COMMITED,
				Timestamp:   _message.Timestamp,
				FromNodeNum: getId(),
				Txn:         _message.Txn,
			})
		}
	case COMMITED:
		reduced := reducedMessage{
			messageType: COMMITED,
			Txn:         newReducedTransaction(*_message.Txn),
			appId:       config.GetAppId(message.FromApp),
		}

		state.counter[reduced]++
		if state.hasQuorum(reduced) {
			state.markAsUsed(reduced)
			state.currentTimestamp = _message.Timestamp
			state.messageOut <- *_message.Txn
		}
	}
}
