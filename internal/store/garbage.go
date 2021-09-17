package store

import (
	"time"
	"unsafe"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type gcdata struct {
	expireTs int64
}

type gcstring struct {
	gcdata
	str string
}

type gcPlayerStore struct {
	gcdata
	ps PlayerStore
}

type gcTransactionStore struct {
	gcdata
	ts TransactionStore
}

func (gc *gcdata) stamp(ttl int64) {
	if ttl == 0 {
		panic("gcdata with a ttl of zero")
	}
	if config.GlobalConfig.DevMode == true {
		if config.GlobalConfig.Local == true {
			ttl = 60
		} else {
			ttl = 3600
		}
	}
	gc.expireTs = gcTs + ttl
}

func NewGcString(s string, ttl int64) gcstring {
	var gc gcstring
	gc.stamp(ttl)
	gc.str = s
	return gc
}

func NewGcPlayerStore(ps PlayerStore, ttl int64) gcPlayerStore {
	var gc gcPlayerStore
	gc.stamp(ttl)
	gc.ps = ps
	return gc
}

func NewGcTransactionStore(ts TransactionStore, ttl int64) gcTransactionStore {
	var gc gcTransactionStore
	gc.stamp(ttl)
	gc.ts = ts
	return gc
}

var gcStartTime time.Time
var gcNowTime time.Time
var gcExecTime time.Duration
var gcPassTime time.Duration
var gcWorkIndex int
var gcTs int64 = time.Now().Unix()

func gcStart() {
	time.Sleep(100000000)
	gcStartTime = time.Now()
	gcNowTime = gcStartTime
	gcWorkIndex = 0
	ld.Lock.Lock()
}

func gcStop() {
	ld.Lock.Unlock()
	gcNowTime := time.Now()
	duration := gcNowTime.Sub(gcStartTime)
	gcExecTime += duration
	if duration > gcPassTime {
		gcPassTime = duration
	}
}

func gcRest() {
	gcWorkIndex++
	if gcWorkIndex > 1000 {
		gcStop()
		gcStart()
	}
}

func garbageCollector() {
	for true {
		gcTs = time.Now().Unix()
		gcExecTime = 0
		gcPassTime = 0

		var numBytes uintptr = 0
		numTokens := 0
		numPlayers := 0
		numMessages := 0
		numTransactions := 0
		numTransactionsByPlayerGame := 0

		gcStart()
		for k, gc := range ld.Token {
			gcRest()
			if gc.expireTs <= gcTs {
				numBytes += unsafe.Sizeof(ld.Token[k])
				numTokens++
				delete(ld.Token, k)
			}
		}
		for k, gc := range ld.Player {
			gcRest()
			if gc.expireTs <= gcTs {
				numBytes += unsafe.Sizeof(ld.Player[k])
				numPlayers++
				delete(ld.Player, k)
			}
		}
		for k, gc := range ld.Message {
			gcRest()
			if gc.expireTs <= gcTs {
				numBytes += unsafe.Sizeof(ld.Message[k])
				numMessages++
				delete(ld.Message, k)
			}
		}
		for k, gc := range ld.Transaction {
			gcRest()
			if gc.expireTs <= gcTs {
				numBytes += unsafe.Sizeof(ld.Transaction[k])
				numTransactions++
				delete(ld.Transaction, k)
			}
		}
		for k, gc := range ld.TransactionByPlayerGame {
			gcRest()
			if gc.expireTs <= gcTs {
				numBytes += unsafe.Sizeof(ld.TransactionByPlayerGame[k])
				numTransactionsByPlayerGame++
				delete(ld.TransactionByPlayerGame, k)
			}
		}
		gcStop()

		if numBytes > 0 {
			logger.Infof("store.garbageCollector freed %d bytes, %d tokens, %d players, %d messages, %d tx, %d txbygame in %.4fms(longest lock %.4fms)",
				numBytes, numTokens, numPlayers, numMessages, numTransactions, numTransactionsByPlayerGame, float64(gcExecTime)/1000000.0, float64(gcPassTime)/1000000.0)
		}
	}
}
