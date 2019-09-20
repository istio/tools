# Istio Chaos Partial Test

This test creates a cronjob that runs every `chaosIntervalMinutes` and does the following:

1. Selects a component to kill instances of.
1. Kills all but one instance of the chosen component (or the single instance if there is only one).

This is designed to test partial failure of the different control plane components on the data plane.
