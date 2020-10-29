#!/bin/bash
#
# Runs the Kubernetes conformance suite against an OpenShift cluster
#
# Test prerequisites:
#
# * all nodes that users can run workloads under marked as schedulable
#
source "$(dirname "${BASH_SOURCE[0]}")/lib/init.sh"

# Check inputs
if [[ -z "${KUBECONFIG-}" ]]; then
  os::log::fatal "KUBECONFIG must be set to a root account"
fi
test_report_dir="${ARTIFACT_DIR}"
mkdir -p "${test_report_dir}"

cat <<END > "${test_report_dir}/README.md"
This conformance report is generated by the OpenShift CI infrastructure. The canonical source location for this test script is located at https://github.com/openshift/kubernetes/blob/master/openshift-hack/conformance-k8s.sh

This file was generated by:

  Commit $( git rev-parse HEAD || "<commit>" )
  Tag    $( git describe || "<tag>" )

To recreate these results

1. Install an [OpenShift cluster](https://docs.openshift.com/container-platform/)
2. Retrieve a \`.kubeconfig\` file with administrator credentials on that cluster and set the environment variable KUBECONFIG

    export KUBECONFIG=PATH_TO_KUBECONFIG

3. Clone the OpenShift source repository and change to that directory:

    git clone https://github.com/openshift/kubernetes.git
    cd kubernetes

4. Place the \`oc\` binary for that cluster in your PATH
5. Run the conformance test:

    openshift-hack/conformance-k8s.sh

Nightly conformance tests are run against release branches and reported https://openshift-gce-devel.appspot.com/builds/origin-ci-test/logs/periodic-ci-origin-conformance-k8s/
END

version="$(grep k8s.io/kubernetes go.sum | awk '{print $2}' | sed s+/go.mod++ )"
os::log::info "Running Kubernetes conformance suite for ${version}"

# Execute OpenShift prerequisites
# Disable container security
oc adm policy add-scc-to-group privileged system:authenticated system:serviceaccounts
oc adm policy add-scc-to-group anyuid system:authenticated system:serviceaccounts
unschedulable="$( ( oc get nodes -o name -l 'node-role.kubernetes.io/master'; ) | wc -l )"
# TODO: undo these operations

# Execute Kubernetes prerequisites
make WHAT=cmd/kubectl
make WHAT=test/e2e/e2e.test
make WHAT=vendor/github.com/onsi/ginkgo/ginkgo
PATH="${OS_ROOT}/_output/local/bin/$( os::build::host_platform ):${PATH}"
export PATH

kubectl version  > "${test_report_dir}/version.txt"
echo "-----"    >> "${test_report_dir}/version.txt"
oc version      >> "${test_report_dir}/version.txt"

# Run the test, serial tests first, then parallel

rc=0

e2e_test="$( which e2e.test )"

# shellcheck disable=SC2086
ginkgo \
  -nodes 1 -noColor '-focus=(\[Conformance\].*\[Serial\]|\[Serial\].*\[Conformance\])' \
  ${e2e_test} -- \
  -report-dir "${test_report_dir}" \
  -allowed-not-ready-nodes ${unschedulable} \
  2>&1 | tee -a "${test_report_dir}/e2e.log" || rc=1

rename -v junit_ junit_serial_ "${test_report_dir}"/junit*.xml

# Skip Serial (those were run above) and...
# any matching 'session affinity timeout' as those will fail.
# on a cluster using OVNKubernetes which is the default CNI in 4.12+.
# The same is done for the more extensive suite of tests run in
# test-kubernetes-e2e.sh.
TEST_SKIPS="\[Serial\]| session affinity timeout "

# shellcheck disable=SC2086
ginkgo \
  -nodes 4 -noColor "-skip=${TEST_SKIPS}" '-focus=\[Conformance\]' \
  ${e2e_test} -- \
  -report-dir "${test_report_dir}" \
  -allowed-not-ready-nodes ${unschedulable} \
  2>&1 | tee -a "${test_report_dir}/e2e.log" || rc=1

echo
echo "Run complete, results in ${test_report_dir}"

exit $rc
