package store

import (
	"runtime"
	"time"

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
	if config.GlobalConfig.LocalDataTtl > 0 {
		ttl = config.GlobalConfig.LocalDataTtl
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

const gcKeepAmount int = 10000
const gcReadAmount int = 10000
const gcDeleteAmount int = 2000
const gcSleepTime time.Duration = time.Duration(10000000)

var gcStartTime time.Time
var gcNowTime time.Time
var gcExecTime time.Duration
var gcPassTime time.Duration
var gcDeleteTime time.Duration
var gcWorkIndex int
var gcTs int64 = time.Now().Unix()
var gcExpired []string = make([]string, gcDeleteAmount)
var gcNumExpired int = 0

func gcStart() {
	time.Sleep(gcSleepTime)
	gcStartTime = time.Now()
	gcNowTime = gcStartTime
	gcWorkIndex = 0
	ld.Lock.RLock()
}

func gcStop() {
	ld.Lock.RUnlock()
	gcNowTime := time.Now()
	duration := gcNowTime.Sub(gcStartTime)
	gcExecTime += duration
	if duration > gcPassTime {
		gcPassTime = duration
	}
}

func gcRest() {
	gcWorkIndex++
	if gcWorkIndex > gcReadAmount {
		gcStop()
		gcStart()
	}
}

type deletefunc func(k string)

func gcMark(k string, del deletefunc) {
	gcExpired[gcNumExpired] = k
	gcNumExpired++
	if gcNumExpired == gcDeleteAmount {
		gcDelete(del)
	}
}

func gcDelete(del deletefunc) {
	if gcNumExpired > 0 {
		gcStop()
		time.Sleep(gcSleepTime)
		gcStartTime := time.Now()
		ld.Lock.Lock()
		for i := 0; i < gcNumExpired; i++ {
			del(gcExpired[i])
		}
		ld.Lock.Unlock()
		gcNowTime := time.Now()
		duration := gcNowTime.Sub(gcStartTime)
		gcExecTime += duration
		if duration > gcDeleteTime {
			gcDeleteTime = duration
		}
		gcStart()
		gcNumExpired = 0
	}
}

func garbageCollector() {
	for true {
		gcTs = time.Now().Unix()
		gcExecTime = 0
		gcPassTime = 0
		gcDeleteTime = 0

		//		var numBytes uintptr = 0
		numTokens := 0
		numPlayers := 0
		numMessages := 0
		numTransactions := 0
		numTransactionsByPlayerGame := 0

		deltokenfn := func(key string) { delete(ld.Token, Token(key)) }
		delplayerfn := func(key string) { delete(ld.Player, key) }
		delmessagefn := func(key string) { delete(ld.Message, key) }
		deltransactionfn := func(key string) { delete(ld.Transaction, key) }
		deltransactionbpgfn := func(key string) { delete(ld.TransactionByPlayerGame, key) }

		gcStart()
		if len(ld.Token) > gcKeepAmount {
			for k, gc := range ld.Token {
				gcRest()
				if gc.expireTs <= gcTs {
					gcMark(string(k), deltokenfn)
					numTokens++
				}
			}
			gcDelete(deltokenfn)
		}
		if len(ld.Player) > gcKeepAmount {
			for k, gc := range ld.Player {
				gcRest()
				if gc.expireTs <= gcTs {
					gcMark(k, delplayerfn)
					numPlayers++
				}
			}
			gcDelete(delplayerfn)
		}
		if len(ld.Message) > gcKeepAmount {
			for k, gc := range ld.Message {
				gcRest()
				if gc.expireTs <= gcTs {
					gcMark(k, delmessagefn)
					numMessages++
				}
			}
			gcDelete(delmessagefn)
		}
		if len(ld.Transaction) > gcKeepAmount {
			for k, gc := range ld.Transaction {
				gcRest()
				if gc.expireTs <= gcTs {
					gcMark(k, deltransactionfn)
					numTransactions++
				}
			}
			gcDelete(deltransactionfn)
		}
		if len(ld.TransactionByPlayerGame) > gcKeepAmount {
			for k, gc := range ld.TransactionByPlayerGame {
				gcRest()
				if gc.expireTs <= gcTs {
					gcMark(k, deltransactionbpgfn)
					numTransactionsByPlayerGame++
				}
			}
		}
		gcDelete(deltransactionbpgfn)
		gcStop()

		if numTokens > 0 || numMessages > 0 || numTransactions > 0 || numTransactionsByPlayerGame > 0 {
			logger.Infof("Collected %d tokens, %d players, %d messages, %d tx, %d txbygame in %.4fms(max rlock, lock: %.4fms, %.4fms)",
				numTokens, numPlayers, numMessages, numTransactions, numTransactionsByPlayerGame,
				float64(gcExecTime)/1000000.0, float64(gcPassTime)/1000000.0, float64(gcDeleteTime)/1000000.0)
			time.Sleep(1000000000)
		} else {
			time.Sleep(5000000000)
		}
		runtime.GC()
	}
}
