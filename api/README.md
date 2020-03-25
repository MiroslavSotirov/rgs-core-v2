API DOCUMENTATION V2

MV Game Client V2 communication spec with RGS V2
================================================

1. Game client sends init call to rgs

/init2
Authorization: Maverick-Host-Token aaabbbcccddd
game : req (the-year-of-zhu)
mode : req (demo or real)
operator : req (mav)
currency : required only if launching in demo mode


GET /init2/?game={ gameSlug }&mode={[ demo | real ]}&operator=mav

Response includes all game init information
- game information like wins with indices, win lines, reels, etc.
- previous gamestate for recovery
- player and wallet information (balance, currency)
- stakeValues and defaultStake, plus any other game-specific parameters
- encrypted token for next request
- links for gameplay, player history, banking lobby, casino lobby


2. Game client sends post to gameplay link returned
POST /play2/{gamestateID}

Auth Header must be set to retrieve session:
Authorization: MAVERICK-Host-Token abc-1234-def


request body may include the following (* = required)
{"stake" : 10000, 					// this is in Fixed notation, so the number represents the financial amount * 1000000 (this example is equivalent to 0.01). on base rounds, the stake must be one of the values returned in the init call stake_values, or else the call will be rejected. in bonus rounds the stake is most often inferred from the triggering round. currency is inferred from the player's wallet currency
"game" : "the-year-of-zhu"			// *
"wallet" : "demo"                   // * demo or dashur for now, this will be returned in the init response
"action" : "base", 					// * this is required to ensure that the client and rgs are on the same page. this will be validated against the available options, of which there is most often only one.
"selectedWinLines" : [0,1,2,3], 	// this is only required in variable line games like Seasons, otherwise it may be omitted
"selectedFeature" : "freespins15", 	// this is only required in the case the previous action required player input to select one of several features, otherwise it may be omitted
"respinReel" : 2  					// the zero-indexed reel to respin, unless the action is "respin", this should be omitted
}




Response:
{"host/verified-token": "abcd"		// the token to use for the Auth header in the next rgs call made
"stake": 10000						// fixed notation, see stake in request for details
"win": 20000						// "", win for this spin
"cumulativeWin": 300000             // used for freespins/bonus rounds, total win amount since bonus started
"freeSpinsRemaining": 3				// number of remaining free spins, omitted if not in free spins
"balance": {
	"amount": {
		"amount": 500000,			// player balance in Fixed notation
		"currency": CNY				// player balance currency
		}
	"freeGames": 5					// the number of free promotional games remaining in player's account
	}
"view": [[1,5,4],[2,3,1],[0,1,3]]	// first dimension is reels
"wins": [{"payout": {"symbol": 1
                     "count": 3
                     "multiplier": 100}, // the base multiplier
          "index": "1:3",
          "multiplier": 2,
          "symbol_positions": [0,2,2],
          "winline": ,
          "win": 10000},],			// fixed notation
"nextAction": "base"
}