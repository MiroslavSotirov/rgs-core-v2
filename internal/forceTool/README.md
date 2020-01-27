FORCE TOOL
##########

The force tool is used in demo mode for development or marketing purposes. The force tool allows the user to simulate outcomes.


Activating the Force Tool
*************************
In the general rgs application configuration should be a setting called `devmode`. If this is set to `true`, the force tool will be enabled on project startup.
Please note that this necessarily also disables communication with a remote service in the store module.
`devmode` is set to `true` on dev.maverick-ops.com, but `false` on staging and production environments.
The force tool is activated by navigating to the following url pattern:
```
dev.maverick-ops.com/v2/rgs/force/<game_name>/<token>/<force_id>
```

<force_id> can be chosen from the following list:
- 3scatter
- 4scatter
- 5scatter
- scatterAndPay
- 5wilds

* not all engines allow all force types. allowance is subject to specifics of engine math

