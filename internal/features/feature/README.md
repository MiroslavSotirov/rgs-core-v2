Features
========

The feature system is functionally a tree of configurable computational functions that is executed after the spin is generated, and generate the feature array of the game state.
Features are used to construct dynamic and resuable game mechanics. The tree structure is evaluated recursively such that if a node activates then execution is continued to its subtree.


Configure Features
==================

Create a new function in the games config yml file, for example:

EngineDefs:
  - name: base
    WinType: lines
    StakeDivisor: 10
    function: FeatureRound
    ...
    Features:
      - Type: "TriggerSupaCrew"
        Params:
          RandomRange: 4536
        Features:
          - Type: "TriggerSupaCrewActionSymbol"
            Params:
              TileId: 9
            Features:
              - Id: 2
                Type: "ReplaceTile"
          - Type: "TriggerSupaCrewFatTileChance"
            Params:
              TileId: 11
              PostActivate: "stop"
            Features:
              - Id: 0
                Type: "FatTile"
          - Type: "TriggerSupaCrewFatTileReel"
            Params:
              W: 2
              H: 2
              TileId: 10
            Features:
            - Id: 0
              Type: "FatTile"
            - Id: 1
              Type: "InstaWin"


This feature tree has a root feature called TriggerSupaCrew and only performs one random number generation that it adds to the params. The root the has tree features in its subtree:
TriggerSupaCrewActionSymbol, TriggerSupaCrewFatTileChance and TriggerSupaCrewFatTileReel. Each one does additional calculations on the generated random number and activates if implementation conditions are met. 
The TriggerSupaCrewActionSymbol, for example, when activated will trigger a ReplaceTile feature that is added to the feature array:

	Features: [ {Def: {Id: 2, Type: "ReplaceTile"}, Data: {TileId: 9, ReplaceWithId: 4}}]

Parameters handling
===================

Parameters are first read from the config file Params field and available as a map in the feature code. The feature can update the parameters dynamically. This has two modes, in the default mode any changes by a feature in a subtree is discarded when execution return to a node.
If the boolean parameter Collated is set to true then subtrees changes to the parameters are restored to enable return parameters.
