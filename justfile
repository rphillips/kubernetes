_default:
  @just --list

[group('runner')]
local-up-cluster:
  hack/local-up-cluster.sh

[group('build')]
build-e2e:
  make WHAT=test/e2e/e2e.test

[group('test')]
test-e2e focus *args:
  _output/bin/e2e.test --provider=local --kubeconfig=/home/rphillips/.kube/config --ginkgo.focus="{{focus}}" {{args}}

[group('test')]
watch-test-e2e focus *args:
  watchexec -e go -w test/e2e -- just test-e2e {{focus}} {{args}}
