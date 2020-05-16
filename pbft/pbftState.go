package pbft

import (
	"All-On-Cloud-9/common"
	"time"
)

type pbftState struct {
	viewNumber        int
	failureTolerance  int
	totalNodes        int
	currentTimestamp  int
	viewChangeCounter int
	timerIsRunning    bool
	candidateNumber   int
	suffix            string
	counter           map[reducedMessage]int
	timeoutTimer      *time.Timer
	localLog          []common.Message
}

func newPbftState(failureTolerance int, totalNodes int, suffix string) *pbftState {
	newState := pbftState{
		viewNumber:        0,
		failureTolerance:  failureTolerance,
		totalNodes:        totalNodes,
		counter:           make(map[reducedMessage]int),
		currentTimestamp:  -1,
		viewChangeCounter: 0,
		timerIsRunning:    false,
		candidateNumber:   1,
		timeoutTimer:      nil,
		suffix:            suffix,
		localLog:          make([]common.Message, 0),
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

func (state *pbftState) handleMessage(
	message common.Message,
	broadcast func(common.Message),
	isLeader func() bool,
	isSuggestedLeader func(int) bool,
	getId func() int,
) *common.Transaction {
	switch message.MessageType {
	case NEW_VIEW:
		state.viewChangeCounter = 0
		state.candidateNumber = state.viewNumber + 1
		state.stopTimer()
		state.viewNumber = message.Timestamp
	case VIEW_CHANGE:
		if isSuggestedLeader(message.Timestamp) {
			state.viewChangeCounter++
			//println("inside VIEW_CHANGE", getId(), state.viewChangeCounter, 2*state.failureTolerance+1)
			if state.viewChangeCounter == 2*state.failureTolerance+1 {
				//println("broadcasting NEW_VIEW")
				broadcast(common.Message{
					MessageType: NEW_VIEW,
					Timestamp:   message.Timestamp,
					FromNodeNum: getId(),
					Txn:         &dummyTxn,
				})
			}
		}
	case NEW_MESSAGE:
		if isLeader() {
			state.currentTimestamp++
			broadcast(common.Message{
				MessageType: PRE_PREPARE,
				Timestamp:   state.currentTimestamp,
				FromNodeNum: getId(),
				Txn:         message.Txn,
			})
		} else {
			state.setTimer()
		}
	case PRE_PREPARE:
		// TODO: Check if is from leader
		state.stopTimer()
		broadcast(common.Message{
			MessageType: PREPARE,
			Timestamp:   message.Timestamp,
			FromNodeNum: getId(),
			Txn:         message.Txn,
		})
	case PREPARE:
		reduced := reducedMessage{
			messageType: PREPARE,
			Txn:         message.Txn,
		}

		state.counter[reduced]++
		if state.counter[reduced] >= 2*state.failureTolerance+1 {
			broadcast(common.Message{
				MessageType: COMMIT,
				Timestamp:   message.Timestamp,
				FromNodeNum: getId(),
				Txn:         message.Txn,
			})
		}
	case COMMIT:
		reduced := reducedMessage{
			messageType: COMMIT,
			Txn:         message.Txn,
		}

		state.counter[reduced]++
		if state.counter[reduced] >= 2*state.failureTolerance+1 {
			broadcast(common.Message{
				MessageType: COMMITED,
				Timestamp:   message.Timestamp,
				FromNodeNum: getId(),
				Txn:         message.Txn,
			})
		}
	case COMMITED:
		reduced := reducedMessage{
			messageType: COMMITED,
			Txn:         message.Txn,
		}

		state.counter[reduced]++
		if state.counter[reduced] >= state.failureTolerance+1 {
			state.currentTimestamp = message.Timestamp
			return message.Txn
		}
	}
	return nil
}
