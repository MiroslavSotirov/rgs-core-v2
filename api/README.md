API DOCUMENTATION V2

MV Game Client V2 communication spec with RGS V2
================================================

1. Game client sends init call to rgs
GET /launch/?game={gameSlug}&mode={[demo | real]}&operator={operatorID}&token={token}

Response includes all game init information
- game information like wins with indices, win lines, reels, etc.
- previous gamestate for recovery
- player and wallet information (balance, currency)
- stakeValues and defaultStake, plus any other game-specific parameters
- encrypted token for next request
- links for gameplay, player history, banking lobby, casino lobby


2. Game client sends post to gameplay link returned
POST /play
Auth Header must be set to retrieve session:
Authorization: MAVERICK-Host-Token abc-1234-def
form info can include stake, selectedWinLines, etc.

Response includes gameplay result
- json output for gameplay






Active Endpoints
================
/healthcheck GET
returns 200


/