PARAMETER SELECTOR MODULE
=========================

The Parameter Selector works as follows:

1. The selection of potential stake values starts from a list of base integers, which are as follows:
  - 1
  - 2
  - 3
  - 5
  - 10
  - 20
  - 30
  - 50
  - 100
  - 200
  - 300
  - 500
  - 1000
  - 2000
  - 3000
  - 5000
2. The list of integers is multiplied by a multiplier depending on the currency:
  - USD: 0.01
  - CNY: 0.04
  - EUR: 0.01
  - GBP: 0.01
  - VND: 100
  - KRW: 10
  - ZAR: 0.1
  - JPY: 1
  - THB: 0.2
  - MYR: 0.02
  - IDR: 100
  - XBT: 0.004
3. Certain bet profiles have been configured in parameterConfig.yml. More can be added. The values that may be set are:
	- min: the starting index to take from the base list
	- max: the ending index to take from the base list (N.B. this will result in an error if it is less than min)
	- default: the index of the unsliced list to use as the default bet (N.B. this will be overridden if the value is not contained in the valid list of stakevalues
4. HostProfiles can be set to use different preconfigured profiles. Multiple hosts will need to be created for single operators wishing to cater to multiple client groups.
5. The parameter selection works as follows:
	- The defaultStake is selected based on the initial list of all possible stakeValues
	- The initial list is sliced based on the max value for the selected profile
	- The initial list is sliced based on the min value for the selected profile
	- If the previous bet is in the list of remaining stakeValues, it is used as the defaultBet
	- Otherwise, the fallback defaultStake is used, as long as it is contained within the valid remaining stakeValues
	- Otherwise, the min or max value from the list of valid stakeValues is used, depending on which is closer to the fallback default
	- Finally, some engines need to be handled specially. This is done with a specific function at the end. Any time a new game is added with special bet settings, the method should be updated.
