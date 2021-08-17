Features
========

This is an attempt at building a dynamic and configurable system for adding features. The general idea is to create game specific trigger functions in code and then configure them so that they activate
standard features that can be reused.


Configure Features
==================

(this is work in progress)

Add a Feature node in to an engine configuration that describes a tree of features. Leaves are features that will be added to the gamestate when a non-leaf(trigger) features
Trigger method determines that it is active. When a trigger activates it runs all its branches as well, until a final list of activated leaf features are generated.


Create a new function in the games config file, for example:

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


This tree has a root feature called TriggerSupaCrew and only performs one random number generation that it adds to its configured Params. The config file already contains a RandomRange node that can be used to control the random number generation. The root trigger has 3 child triggers:
TriggerSupaCrewActionSymbol, TriggerSupaCrewFatTileChance and TriggerSupaCrewFatTileReel. Each on does additional calculations on the generated random number (however, they do not generate any additional random numbers) and activates if game specific conditions are met. The action
symbol trigger, for example, will then finally activate a ReplaceTile generic feature that generates a gamestate reply:

	Features: [ {Def: {Id: 2, Type: "ReplaceTile"}, Data: {TileId: 9, ReplaceWithId: 4}}]

Parameters handling
===================

Parameters are first read from the config file and available as a map when writing the feature. The feature can then, when activating branches, update the parameters dynamically from the code and those will be added/override the branch features set of parameters. This is to enable
dynamic configuring of features and reuse.

TODO & BUGS
===========

The engine config function is not sufficient to activate features everywhere because the system calls the BaseRound function during init. So instead of relying on a special feature activated function the base function should probably be modified to run the feature system.
The PostActivate in the example above does not work, and more functionality is needed to handle priority and cases where one feature is exclusive to another feature.
Lots of possible crashes because the system is not type safe and relies on keeping track of the map datatypes of the parameters manually. A serializeStruct type function could help with checking types at the cost of writing lots of small data passing structs.