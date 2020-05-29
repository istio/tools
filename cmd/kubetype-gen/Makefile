
define kubetype-gen =
$(shell go run ./main.go -i $$(go list ${1}) -p "$$(go list .)/testdata/test_output/${2}" -h "./boilerplate.go.txt" 2>&1)
endef

test: run-negative-tests run-positive-tests

check-env:
	@if ! git diff --quiet HEAD -- tests; then \
	    echo "FAIL: tests contain uncommitted changes. Please commit changes in ./testdata."; \
	    false; \
	fi

run-positive-tests: run-positive-defaults-test run-positive-types-test

run-positive-defaults-test: check-env
	@echo "Running $@"
	$(eval $@_output = $(call kubetype-gen,./testdata/test_input/positive/defaults,defaults))
	@echo "${$@_output}" | grep -q "Errors occurred while scanning input.  See previous output for details."; test $$? || (echo "kubetype-gen command failed" && echo "${$@_output}" && false)
	@if ! git diff --quiet HEAD -- tests || [ -n "$$(git ls-files -o -- tests)" ]; then \
	    echo "FAIL: output files have changed"; \
	    false; \
	fi

run-positive-types-test: check-env
	@echo "Running $@"
	$(eval $@_output = $(call kubetype-gen,./testdata/test_input/positive/types,types))
	@echo "${$@_output}" | grep -q "Errors occurred while scanning input.  See previous output for details."; test $$? || (echo "kubetype-gen command failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -q "Invalid value specified for +kubetype-gen:kubeType in type .*/test_input/positive/types.EmptyKubeType.  Using default name EmptyKubeType."; test $$? || (echo "Missing warning for types.EmptyKubeType" && echo "${$@_output}" && false)
	@if ! git diff --quiet HEAD -- tests || [ -n "$$(git ls-files -o -- tests)" ]; then \
	    echo "FAIL: output files have changed"; \
	    false; \
	fi


run-negative-tests: run-negative-defaults-empty-group-test run-negative-defaults-invalid-group-version-test run-negative-types-test

run-negative-defaults-empty-group-test: check-env
	@echo "Running $@"
	$(eval $@_output = $(call kubetype-gen,./testdata/test_input/negative/defaults/emptygroup,fail))
	@echo "${$@_output}" | grep -q "Errors occurred while scanning input.  See previous output for details."; test $$? || (echo "kubetype-gen command should have failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -q "Invalid Group/Version for package istio.io/tools/cmd/kubetype-gen/testdata/test_input/negative/defaults/emptygroup, Group not specified for Group/Version: groupversion"; test $$? || (echo "negative/defaults/emptygroup test failed" && echo "${$@_output}" && false)

run-negative-defaults-invalid-group-version-test: check-env
	@echo "Running $@"
	$(eval $@_output = $(call kubetype-gen,./testdata/test_input/negative/defaults/invalidgroupversion,fail))
	@echo "${$@_output}" | grep -q "Errors occurred while scanning input.  See previous output for details."; test $$? || (echo "kubetype-gen command should have failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -q "Could not calculate Group/Version for package .*/test_input/negative/defaults/invalidgroupversion: invalid group version 'group/version/version' specified: unexpected GroupVersion string: group/version/version"; test $$? || (echo "negative/defaults/invalidgroupversion test failed" && echo "${$@_output}" && false)

run-negative-types-test: check-env
	@echo "Running $@"
	$(eval $@_output = $(call kubetype-gen,./testdata/test_input/negative/types,fail))
	@echo "${$@_output}" | grep -q "Errors occurred while scanning input.  See previous output for details."; test $$? || (echo "kubetype-gen command should have failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -qE "duplicate kube types specified for .*/test_output/fail/group/version.AGoodType.  duplicated by: \[istio.io/tools/cmd/kubetype-gen/testdata/test_input/negative/types.(AGoodType|DuplicateKubeType) istio.io/tools/cmd/kubetype-gen/testdata/test_input/negative/types.(AGoodType|DuplicateKubeType)\]"; test $$? || (echo "types.DuplicateKubeType test failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -q "Invalid Group/Version for type .*/test_input/negative/types.EmptyGroup: groupversion"; test $$? || (echo "types.EmptyGroup test failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -q "Could not calculate Group/Version for type .*/test_input/negative/types.InvalidGroupVersion: invalid group version 'group/version/version' specified"; test $$? || (echo "types.InvalidGroupVersion test failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -q "Invalid Group/Version for type .*/test_input/negative/types.NoGroupVersion: <nil>"; test $$? || (echo "types.NoGroupVersion test failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -qE "Overlapping packages for Group/Versions group2(.name.io)?/version and group2(.name.io)?/version"; test $$? || (echo "types.OverlappingPackageGroupType test failed" && echo "${$@_output}" && false)
	@echo "${$@_output}" | grep -q "Could not create metadata for type: .*/test_input/negative/types.OverlappingPackageGroupType."; test $$? || (echo "types.OverlappingPackageGroupType test failed" && echo "${$@_output}" && false)
