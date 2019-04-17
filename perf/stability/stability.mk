WD = ./perf/stability
STABILITY = $(WD)/setup_test.sh

# Standard set of stability tests to run
stable_tests = http10 graceful-shutdown gateway-bouncer mysql redis rabbitmq

# Tests that need no special setup
standard_tests = http10 graceful-shutdown redis rabbitmq

# Tests that have a special ./setup script in their folder
extra_setup_tests = mysql sds-certmanager gateway-bouncer allconfig

$(standard_tests):
	$(STABILITY) $@

$(extra_setup_tests):
	$(WD)/$@/setup.sh

stability: $(stable_tests)

# Extra tests that may be unstable or require additional configuration
stability_all: stability sds-certmanager allconfig

clean-stability:
	kubectl get namespaces -oname | grep "istio-stability-" | xargs kubectl delete
