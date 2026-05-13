#!/bin/bash
set -euo pipefail

# ---------------------------------------------------------------------------
# Environment: source .env if present, fall back to defaults
# ---------------------------------------------------------------------------
if [[ -f .env ]]; then
    set -a
    source .env
    set +a
fi

export AIRTEL_API_ENDPOINT="${AIRTEL_API_ENDPOINT:-https://south.cloud.airtel.in}"
export AIRTEL_API_KEY="${AIRTEL_API_KEY:-}"
export AIRTEL_API_SECRET="${AIRTEL_API_SECRET:-}"
export AIRTEL_REGION="${AIRTEL_REGION:-south}"
export AIRTEL_ORGANIZATION="${AIRTEL_ORGANIZATION:-perftest}"
export AIRTEL_PROJECT_NAME="${AIRTEL_PROJECT_NAME:-cell-1}"
export AIRTEL_USERNAME="${AIRTEL_USERNAME:-}"
export AIRTEL_TEST_NETWORK_ID="${AIRTEL_TEST_NETWORK_ID:-}"
export AIRTEL_TEST_AVAILABILITY_ZONE="${AIRTEL_TEST_AVAILABILITY_ZONE:-S2}"

# ---------------------------------------------------------------------------
# Colors
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
info()  { echo -e "${CYAN}${BOLD}==> ${NC}${BOLD}$*${NC}"; }
pass()  { echo -e "${GREEN}${BOLD}PASS${NC} $*"; }
fail()  { echo -e "${RED}${BOLD}FAIL${NC} $*"; }
warn()  { echo -e "${YELLOW}$*${NC}"; }

# ---------------------------------------------------------------------------
# Integration test: service -> -run pattern
# ---------------------------------------------------------------------------
integration_pattern() {
    local service="$1"
    case "$service" in
        vpc)              echo "TestVPCIntegration" ;;
        subnet)           echo "TestSubnetIntegration" ;;
        volume)           echo "TestVolumeIntegration" ;;
        dns-zone)         echo "TestDNSZoneIntegration" ;;
        dns-record)       echo "TestDNSRecordIntegration" ;;
        dns)              echo "TestDNS" ;;
        nfs)              echo "TestFileStorage" ;;
        object-storage)   echo "TestObjectStorageIntegration" ;;
        security-group)   echo "TestSecurityGroupIntegration" ;;
        all)              echo "Integration|TestFileStorage" ;;
        *)
            fail "Unknown integration service: $service" >&2
            return 1
            ;;
    esac
}

# ---------------------------------------------------------------------------
# Unit test: service -> -run pattern
# ---------------------------------------------------------------------------
unit_pattern() {
    local service="$1"
    case "$service" in
        compute)          echo "Test(Create|Get|Update|Delete|List)(Compute|Computes|Flavors|Images|Keypairs|SecurityGroups)" ;;
        volume)           echo "Test(Create|Get|Update|Delete|Attach|Detach|List)Volum" ;;
        vpc)              echo "TestListVPCs" ;;
        subnet)           echo "TestListSubnets" ;;
        security-group)   echo "Test(Create|Get|Delete|List)SecurityGroup|TestDebugRuleCreate" ;;
        all|"")           echo "." ;;
        *)
            fail "Unknown unit service: $service" >&2
            return 1
            ;;
    esac
}

# ---------------------------------------------------------------------------
# Acceptance test: service -> -run pattern
# ---------------------------------------------------------------------------
acceptance_pattern() {
    local service="$1"
    case "$service" in
        volume)           echo "TestAccVolumeResource" ;;
        vm)               echo "TestAccVMResource" ;;
        security-group)   echo "TestAccSecurityGroup" ;;
        all|"")           echo "TestAcc" ;;
        *)
            fail "Unknown acceptance service: $service" >&2
            return 1
            ;;
    esac
}

# ---------------------------------------------------------------------------
# Run integration tests
# ---------------------------------------------------------------------------
run_integration() {
    local service="${1:-all}"
    local operation="${2:-}"
    local pattern
    pattern=$(integration_pattern "$service") || exit 1

    # Apply operation filter if provided
    if [[ -n "$operation" ]]; then
        case "$operation" in
            list)
                if [[ "$service" == "all" ]]; then
                    pattern="_List"
                else
                    pattern="${pattern}_List"
                fi
                ;;
            get)
                if [[ "$service" == "all" ]]; then
                    pattern="_Get"
                else
                    pattern="${pattern}_Get"
                fi
                ;;
            *)
                fail "Unknown operation filter: $operation (use 'list' or 'get')"
                exit 1
                ;;
        esac
    fi

    info "Running integration tests: $service${operation:+ ($operation only)} (pattern: $pattern)"
    echo -e "${CYAN}go test -tags=integration -v ./internal/client/... -run $pattern${NC}"
    echo ""
    go test -tags=integration -v ./internal/client/... -run "$pattern"
}

# ---------------------------------------------------------------------------
# Run unit tests (no integration tag)
# ---------------------------------------------------------------------------
run_unit() {
    local service="${1:-all}"
    local pattern
    pattern=$(unit_pattern "$service") || exit 1

    local run_flag=""
    if [[ "$pattern" != "." ]]; then
        run_flag="-run $pattern"
    fi

    info "Running unit tests: ${service:-all}"
    echo -e "${CYAN}go test -v ./internal/client/... $run_flag${NC}"
    echo ""
    if [[ -n "$run_flag" ]]; then
        go test -v ./internal/client/... -run "$pattern"
    else
        go test -v ./internal/client/...
    fi
}

# ---------------------------------------------------------------------------
# Run acceptance tests
# ---------------------------------------------------------------------------
run_acceptance() {
    local service="${1:-all}"
    local pattern
    pattern=$(acceptance_pattern "$service") || exit 1

    export TF_ACC=1

    info "Running acceptance tests: ${service:-all} (pattern: $pattern)"
    echo -e "${CYAN}TF_ACC=1 go test -v ./tests/resources/... -run $pattern${NC}"
    echo ""
    go test -v ./tests/resources/... -run "$pattern"
}

# ---------------------------------------------------------------------------
# List available services and test counts
# ---------------------------------------------------------------------------
list_services() {
    echo -e "${BOLD}Integration Tests${NC} (./internal/client/...  -tags=integration)"
    echo "  (use 'list' or 'get' as 3rd arg to filter by operation)"
    echo "  vpc              4 tests   TestVPCIntegration_{CreateGetDelete,List,Update,GetNonExistent}"
    echo "  subnet           5 tests   TestSubnetIntegration_{CreateGetDelete,ListSubnets,UpdateSubnet,GetNonExistent,CreateWithLabels}"
    echo "  volume          13 tests   TestVolumeIntegration_{CreateGetDelete,List,Update,...}"
    echo "  dns-zone         4 tests   TestDNSZoneIntegration_{CreateGetDelete,List,Update,GetNonExistent}"
    echo "  dns-record      11 tests   TestDNSRecordIntegration_{CreateGetDelete,Update,MultipleTypes,...}"
    echo "  dns             15 tests   All DNS zone + record tests"
    echo "  nfs              6 tests   TestFileStorage{Volume,ExportPath}Integration_{...}"
    echo "  object-storage   4 tests   TestObjectStorageIntegration_{CreateGetDelete,List,Update,GetNonExistent}"
    echo "  security-group  13 tests   TestSecurityGroupIntegration_{CreateGetDelete,List,...}"
    echo ""
    echo -e "${BOLD}Unit Tests${NC} (./internal/client/...)"
    echo "  compute          9 tests   TestCreate/Get/Update/Delete/ListCompute(s), Flavors, Images, ..."
    echo "  volume           7 tests   TestCreate/Get/Update/Delete/Attach/Detach/ListVolume(s)"
    echo "  vpc              1 test    TestListVPCs"
    echo "  subnet           1 test    TestListSubnets"
    echo "  security-group   9 tests   TestCreate/Get/Delete/ListSecurityGroup(s), Rules, DebugRuleCreate"
    echo ""
    echo -e "${BOLD}Acceptance Tests${NC} (./tests/resources/...)"
    echo "  volume           2 tests   TestAccVolumeResource{,WithAttachment}"
    echo "  vm               1 test    TestAccVMResource"
    echo "  security-group   3 tests   TestAccSecurityGroup{Resource,RuleResource,WithMultipleRules}"
}

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------
show_help() {
    echo "Usage: $0 <command> [service] [options]"
    echo ""
    echo "Commands:"
    echo "  integration <service> [list|get]  Run integration tests (optionally filter by operation)"
    echo "  unit [service]         Run unit tests (optionally filtered by service)"
    echo "  acceptance [service]   Run acceptance tests under tests/resources/"
    echo "  list                   List available services and test counts"
    echo "  help                   Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 integration vpc          # All VPC integration tests"
    echo "  $0 integration volume       # All volume integration tests"
    echo "  $0 integration dns          # All DNS (zone + record) tests"
    echo "  $0 integration dns-zone     # Only DNS zone tests"
    echo "  $0 integration all          # All integration tests"
    echo "  $0 integration vpc list     # Only VPC list tests"
    echo "  $0 integration vpc get      # Only VPC get tests"
    echo "  $0 integration all list     # All list tests across services"
    echo "  $0 integration all get      # All get tests across services"
    echo "  $0 unit                     # All unit tests"
    echo "  $0 unit compute             # Compute unit tests only"
    echo "  $0 acceptance               # All acceptance tests"
    echo "  $0 acceptance vm            # VM acceptance tests only"
    echo "  $0 list                     # List services and test counts"
}

# ---------------------------------------------------------------------------
# Main dispatch
# ---------------------------------------------------------------------------
command="${1:-help}"
shift || true

case "$command" in
    integration)
        [[ $# -eq 0 ]] && { fail "Service name required. Use 'all' for everything, or '$0 list' to see options."; exit 1; }
        run_integration "$1" "${2:-}"
        ;;
    unit)
        run_unit "${1:-all}"
        ;;
    acceptance)
        run_acceptance "${1:-all}"
        ;;
    list)
        list_services
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        fail "Unknown command: $command"
        echo ""
        show_help
        exit 1
        ;;
esac
