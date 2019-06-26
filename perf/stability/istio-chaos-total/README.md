# Istio Chaos Total Test

This test creates a cronjob that runs every `chaosIntervalMinutes` and does the following:

1. Selects a number (between 1 and `chaosLevel`) of components to simultaneously scale to zero.
2. Scales those components to zero
3. Sleeps for `chaosDurationMinutes`
4. Scales those components to one

This is designed to test total failure of the different control plane components on the data plane.
