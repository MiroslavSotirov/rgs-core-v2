
Maverick calls to dashur:

AUTHORIZATION:
Send: [token, gameId]
receive : [account info (corresponds to player type), refresh token, auth token, account balance, latest gameplay struct for account+gameId (information in gameplay.session)]

GET_BALANCE:
Should not really be necessary if it’s possible to return balance in Authorization call and after each transaction

TRANSACTION:
Send: [Accountid, Gameplay (in current state, gameplay with same ID can be saved on later transaction. Most recent takes precedence), ]
Receive: [OK or error]

transaction types : [“WAGER”,  -- stake settling, initial gameplay with bet amount sent
“PAYOUT”,  -- win settling, gameplay results with engine output and all gamestates sent
“ENDROUND”]  -- client state complete, all results animated


// Error codes

// Json examples

// websockets?

// refunds (nondeterministic errors)



// Relevant Datatypes
```
type Money struct {
  Amount float64
  Currency string
}
```

```
// payout used in engine definitions
type Payout struct {
  symbol int
  count  int
  Multiplier float64
}
```

```
// prize used to convey win information
type Prize struct {
  Payout Payout //
  Index string // either the line number, "ways", or prize name (i.e. "chooseFreeSpins")
  Multiplier float64 // includes additional multipliers from wilds or freespins etc.
}
```
```
// gamestate represents one engine action round--one spin, whether free or paid
type Gamestate struct {
  Engine string // which engine or sub-engine is this action being performed on
  Complete bool // complete is true if it has been animated by the game client
  Action string // e.g. "spin", "freespin", "prizeselect"
  SymbolGrid [][]int // first dimension is reels, second dimension is position. dimensions should match engine definition's ViewSize
  ReelsetId int // for engines with multiple reelsets todo: do we need this?
  Prizes []Prize // any award (i.e. linewin, freespins, prizeselect, etc.
  SelectedWinLines []int // a slice of integers representing the positions of winlines
  RelativePayout float64 // == win / bet == sum (relativeWin)
  Multiplier float64 // includes multipliers on entire gamestate
}
```
```
// gameplay returned by play function
// one gameplay represents one bet, it can contain numerous gamestates
// saved in localstorage until Complete == true
type Gameplay struct {
  Complete bool // complete is true if all child gamestates are complete
  Bet Money // the amount of money deducted from the player's wallet
  Win Money // the amount of money paid into the player's wallet
  Gamestates []Gamestate // the slice of gamestates encountered in this play
  Id string
  SessionId string
}
```
```
type player struct {
  id string // can be dashur aid or randomly generated for devwallet
  progress map[string]string // info about game progress
}
```
```
//deleted from localstorage as soon as it is expired
type session struct {
  id string
  player player
  gameId string
}
```
