# test_runner.py
from base_test import TestConfig, setup_logging, APIClient, TestStatus
from test_plans import PlanTests
from test_features import FeatureTests
from test_meters import MeterTests
from test_events import EventTests
import time
from datetime import datetime
from typing import Dict, List, Optional, Any

# Test Result Tracking
class TestResult:
    """Track test execution results"""

    def __init__(self):
        self.results: List[Dict[str, Any]] = []
        self.start_time = time.time()

    def add_result(self, test_name: str, status: TestStatus, duration: float, 
                   error_message: str = None, details: Dict[str, Any] = None):
        """Add test result"""
        self.results.append({
            "test_name": test_name,
            "status": status.value,
            "duration": duration,
            "error_message": error_message,
            "details": details or {},
            "timestamp": datetime.utcnow().isoformat()
        })

    def get_summary(self) -> Dict[str, Any]:
        """Get test execution summary"""
        total_duration = time.time() - self.start_time
        status_counts = {}

        for status in TestStatus:
            status_counts[status.value] = sum(1 for r in self.results if r["status"] == status.value)

        return {
            "total_tests": len(self.results),
            "total_duration": total_duration,
            "status_counts": status_counts,
            "success_rate": (status_counts.get("PASSED", 0) / len(self.results) * 100) if self.results else 0
        }

# Main Test Runner
class IntegrationTestRunner:
    """Main test runner for API integration tests"""

    def __init__(self, config: TestConfig):
        self.config = config
        self.logger = setup_logging(config)
        self.client = APIClient(config, self.logger)
        self.results = TestResult()

        # Test classes
        self.feature_tests = FeatureTests(self.client, self.logger)
        self.meter_tests = MeterTests(self.client, self.logger)
        self.event_tests = EventTests(self.client, self.logger)
        self.plan_tests = PlanTests(self.client, self.logger)

    def run_test(self, test_name: str, test_func, *args, **kwargs):
        """Run a single test with error handling and result tracking"""
        self.logger.info(f"\n{'='*60}")
        self.logger.info(f"Running: {test_name}")
        self.logger.info(f"{'='*60}")

        start_time = time.time()
        try:
            result = test_func(*args, **kwargs)
            duration = time.time() - start_time

            self.results.add_result(test_name, TestStatus.PASSED, duration, details=result)
            self.logger.info(f"✅ {test_name} - PASSED ({duration:.3f}s)")
            return result

        except AssertionError as e:
            duration = time.time() - start_time
            self.results.add_result(test_name, TestStatus.FAILED, duration, str(e))
            self.logger.error(f"❌ {test_name} - FAILED ({duration:.3f}s): {str(e)}")
            return None

        except Exception as e:
            duration = time.time() - start_time
            self.results.add_result(test_name, TestStatus.ERROR, duration, str(e))
            self.logger.error(f"💥 {test_name} - ERROR ({duration:.3f}s): {str(e)}")
            return None

    def run_all_tests(self):
        """Run complete test suite"""
        self.logger.info("🚀 Starting API Integration Test Suite")
        self.logger.info(f"Base URL: {self.config.base_url}")
        self.logger.info(f"Tenant: {self.config.tenant_slug}")

        # Feature Tests
        self.logger.info("\n🔧 Running Feature Tests...")
        static_feature = self.run_test("Create Static Feature", self.feature_tests.test_create_feature_static)
        metered_feature = self.run_test("Create Metered Feature", self.feature_tests.test_create_feature_metered)

        self.run_test("List Features", self.feature_tests.test_list_features)

        if static_feature:
            self.run_test("Get Feature by ID", self.feature_tests.test_get_feature_by_id, static_feature["feature_id"])
            self.run_test("Update Feature", self.feature_tests.test_update_feature, static_feature["feature_id"])
            self.run_test("Delete Feature", self.feature_tests.test_delete_feature, static_feature["feature_id"])

        self.run_test("Feature Validation Errors", self.feature_tests.test_feature_validation_errors)

        # Meter Tests
        self.logger.info("\n📊 Running Meter Tests...")
        meter = self.run_test("Create Meter", self.meter_tests.test_create_meter)

        if meter:
            self.run_test("Query Meter", self.meter_tests.test_query_meter, meter["meter_data"]["slug"])

        # Event Tests
        self.logger.info("\n📡 Running Event Tests...")
        self.run_test("Publish Single Event", self.event_tests.test_publish_single_event)
        self.run_test("Publish Batch Events", self.event_tests.test_publish_batch_events)

        if meter:
            self.run_test("Delte Meter", self.meter_tests.test_delete_meter, meter["meter_id"])



        # Plan Tests
        self.logger.info("\n📋 Running Plan Tests...")
        plan = self.run_test("Create Plan", self.plan_tests.test_create_plan)

        if plan:
            self.run_test("Plan Assignment Lifecycle", self.plan_tests.test_plan_assignment_lifecycle, plan["plan_id"])

        if plan:
            self.run_test("Delte Plan", self.plan_tests.test_delete_plan, plan["plan_id"])


        # Test Summary
        self.print_summary()

    def print_summary(self):
        """Print test execution summary"""
        summary = self.results.get_summary()

        self.logger.info("\n" + "="*60)
        self.logger.info("🏁 TEST EXECUTION SUMMARY")
        self.logger.info("="*60)
        self.logger.info(f"Total Tests: {summary['total_tests']}")
        self.logger.info(f"Duration: {summary['total_duration']:.2f}s")
        self.logger.info(f"Success Rate: {summary['success_rate']:.1f}%")
        self.logger.info("")

        for status, count in summary['status_counts'].items():
            if count > 0:
                emoji = {"PASSED": "✅", "FAILED": "❌", "ERROR": "💥", "SKIPPED": "⏭️"}.get(status, "❓")
                self.logger.info(f"{emoji} {status}: {count}")

        # Failed tests details
        failed_tests = [r for r in self.results.results if r["status"] in ["FAILED", "ERROR"]]
        if failed_tests:
            self.logger.info("\n📋 Failed Tests:")
            for test in failed_tests:
                self.logger.info(f"  • {test['test_name']}: {test['error_message']}")


if __name__ == "__main__":
    config = TestConfig()
    IntegrationTestRunner(config).run_all_tests()

