#!/usr/bin/env python3
"""
Integration Test Suite for Go Fiber API
Comprehensive testing framework with logging, fixtures, and utilities
"""

import json
import logging
import os
import sys
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Union
from dataclasses import dataclass
from enum import Enum

import requests
import pytest
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

# Configuration
@dataclass
class TestConfig:
    """Test configuration settings"""
    base_url: str = "http://localhost:8000"
    tenant_slug: str = "test-tenant"
    timeout: int = 30
    max_retries: int = 3
    backoff_factor: float = 0.3
    log_level: str = "INFO"
    log_file: str = "api_tests.log"


class TestStatus(Enum):
    """Test execution status"""
    PASSED = "PASSED"
    FAILED = "FAILED"
    SKIPPED = "SKIPPED"
    ERROR = "ERROR"


# Logging Setup
def setup_logging(config: TestConfig) -> logging.Logger:
    """Setup comprehensive logging configuration"""
    logger = logging.getLogger("api_integration_tests")
    logger.setLevel(getattr(logging, config.log_level.upper()))
    
    # Clear existing handlers
    logger.handlers.clear()
    
    # Create formatters
    detailed_formatter = logging.Formatter(
        '%(asctime)s | %(levelname)-8s | %(name)s | %(funcName)s:%(lineno)d | %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S'
    )
    
    simple_formatter = logging.Formatter(
        '%(asctime)s | %(levelname)-8s | %(message)s',
        datefmt='%H:%M:%S'
    )
    
    # File handler for detailed logs
    file_handler = logging.FileHandler(config.log_file, mode='w')
    file_handler.setLevel(logging.DEBUG)
    file_handler.setFormatter(detailed_formatter)
    logger.addHandler(file_handler)
    
    # Console handler for important messages
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(logging.INFO)
    console_handler.setFormatter(simple_formatter)
    logger.addHandler(console_handler)
    
    return logger


# HTTP Client with retry logic
class APIClient:
    """Enhanced HTTP client with retry logic and logging"""
    
    def __init__(self, config: TestConfig, logger: logging.Logger):
        self.config = config
        self.logger = logger
        self.session = requests.Session()
        
        # Setup retry strategy
        retry_strategy = Retry(
            total=config.max_retries,
            backoff_factor=config.backoff_factor,
            status_forcelist=[429, 500, 502, 503, 504],
        )
        
        adapter = HTTPAdapter(max_retries=retry_strategy)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)
        
        # Default headers
        self.session.headers.update({
            'Content-Type': 'application/json',
            'X-Tenant-Slug': config.tenant_slug
        })
    
    def _log_request(self, method: str, url: str, **kwargs):
        """Log HTTP request details"""
        self.logger.debug(f"→ {method.upper()} {url}")
        if 'json' in kwargs:
            self.logger.debug(f"  Request Body: {json.dumps(kwargs['json'], indent=2)}")
        if 'params' in kwargs:
            self.logger.debug(f"  Query Params: {kwargs['params']}")
    
    def _log_response(self, response: requests.Response, duration: float):
        """Log HTTP response details"""
        self.logger.debug(f"← {response.status_code} {response.reason} ({duration:.3f}s)")
        try:
            if response.text:
                response_data = response.json()
                self.logger.debug(f"  Response Body: {json.dumps(response_data, indent=2)}")
        except json.JSONDecodeError:
            self.logger.debug(f"  Response Body: {response.text[:200]}...")
    
    def request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make HTTP request with logging and error handling"""
        url = f"{self.config.base_url}{endpoint}"
        
        self._log_request(method, url, **kwargs)
        
        start_time = time.time()
        try:
            response = self.session.request(
                method=method,
                url=url,
                timeout=self.config.timeout,
                **kwargs
            )
            duration = time.time() - start_time
            self._log_response(response, duration)
            return response
            
        except requests.exceptions.RequestException as e:
            duration = time.time() - start_time
            self.logger.error(f"Request failed after {duration:.3f}s: {str(e)}")
            raise
    
    def get(self, endpoint: str, **kwargs) -> requests.Response:
        return self.request('GET', endpoint, **kwargs)
    
    def post(self, endpoint: str, **kwargs) -> requests.Response:
        return self.request('POST', endpoint, **kwargs)
    
    def put(self, endpoint: str, **kwargs) -> requests.Response:
        return self.request('PUT', endpoint, **kwargs)
    
    def delete(self, endpoint: str, **kwargs) -> requests.Response:
        return self.request('DELETE', endpoint, **kwargs)


# Test Data Factories
class TestDataFactory:
    """Factory for generating test data"""
    
    @staticmethod
    def create_feature_data(feature_type: str = "static") -> Dict[str, Any]:
        """Create test feature data"""
        unique_id = str(uuid.uuid4())[:8]
        return {
            "name": f"Test Feature {unique_id}",
            "slug": f"test-feature-{unique_id}",
            "description": f"Test feature for integration testing - {unique_id}",
            "type": feature_type,
            "created_by": "integration-test",
            "config": {"enabled": True, "priority": 1}
        }
    
    @staticmethod
    def create_meter_data() -> Dict[str, Any]:
        """Create test meter data"""
        unique_id = str(uuid.uuid4())[:8]
        return {
            "name": f"Test Meter {unique_id}",
            "slug": f"test-meter-{unique_id}",
            "description": f"Test meter for integration testing - {unique_id}",
            "event_type": "api_call",
            "aggregation": "count",
            "properties": ["user_id", "endpoint"],
            "populate": True,
            "created_by": "integration-test",
            "value_property": "count"
        }
    
    @staticmethod
    def create_plan_data(plan_type: str = "standard") -> Dict[str, Any]:
        """Create test plan data"""
        unique_id = str(uuid.uuid4())[:8]
        return {
            "name": f"Test Plan {unique_id}",
            "slug": f"test-plan-{unique_id}",
            "description": f"Test plan for integration testing - {unique_id}",
            "type": plan_type,
            "created_by": "integration-test"
        }
    
    @staticmethod
    def create_event_data(event_type: str = "api_call", user_id: str = None) -> Dict[str, Any]:
        """Create test event data"""
        return {
            "type": event_type,
            "user": user_id or f"test-user-{uuid.uuid4().hex[:8]}",
            "organization": f"test-org-{uuid.uuid4().hex[:8]}",
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "source": "integration-test",
            "properties": {
                "endpoint": "/api/test",
                "method": "GET",
                "status_code": 200,
                "response_time": 0.15
            }
        }
    
    @staticmethod
    def create_plan_assignment_data(plan_id: str, user_id: str = None, org_id: str = None) -> Dict[str, Any]:
        """Create test plan assignment data"""
        now = datetime.utcnow()
        valid_from = now.isoformat() + "Z"
        valid_until = (now + timedelta(days=30)).isoformat() + "Z"
        
        assignment = {
            "plan_id_or_slug": plan_id,
            "valid_from": valid_from,
            "valid_until": valid_until,
            "created_by": "integration-test"
        }
        
        if user_id:
            assignment["user_id"] = user_id
        if org_id:
            assignment["organization_id"] = org_id
            
        return assignment
# Base Test Class
class BaseAPITest:
    """Base class for API integration tests"""
    
    def __init__(self, client: APIClient, logger: logging.Logger):
        self.client = client
        self.logger = logger
        self.test_data_factory = TestDataFactory()
    
    def assert_response_success(self, response: requests.Response, expected_status: int = 200):
        """Assert response is successful"""
        if response.status_code != expected_status:
            error_msg = f"Expected status {expected_status}, got {response.status_code}"
            if response.text:
                try:
                    error_data = response.json()
                    error_msg += f" - {error_data.get('message', response.text)}"
                except:
                    error_msg += f" - {response.text[:200]}"
            raise AssertionError(error_msg)
    
    def assert_response_error(self, response: requests.Response, expected_status: int):
        """Assert response is an error"""
        if response.status_code != expected_status:
            raise AssertionError(f"Expected error status {expected_status}, got {response.status_code}")
    
    def extract_id_from_response(self, response: requests.Response) -> str:
        """Extract ID from API response"""
        data = response.json()
        if "data" in data and "id" in data["data"]:
            return data["data"]["id"]
        raise ValueError("Could not extract ID from response")


# Plan Tests
class PlanTests(BaseAPITest):
    """Integration tests for Plans API"""
    
    def test_create_plan(self) -> Dict[str, Any]:
        """Test creating a plan"""
        self.logger.info("Testing: Create Plan")
        
        plan_data = self.test_data_factory.create_plan_data()
        response = self.client.post("/v1/plans", json=plan_data)
        
        self.assert_response_success(response, 201)
        response_data = response.json()
        
        assert response_data["data"]["name"] == plan_data["name"]
        
        plan_id = response_data["data"]["id"]
        self.logger.info(f"✓ Created plan with ID: {plan_id}")
        
        return {"plan_id": plan_id, "plan_data": plan_data}
    
    def test_plan_assignment_lifecycle(self, plan_id: str):
        """Test complete plan assignment lifecycle"""
        self.logger.info(f"Testing: Plan Assignment Lifecycle - {plan_id}")
        
        # Create assignment
        assignment_data = self.test_data_factory.create_plan_assignment_data(
            plan_id=plan_id,
            user_id=f"test-user-{uuid.uuid4().hex[:8]}"
        )
        
        response = self.client.post("/v1/plans/assignments", json=assignment_data)
        self.assert_response_success(response, 201)
        
        assignment_id = response.json()["data"]["id"]
        self.logger.info(f"✓ Created plan assignment: {assignment_id}")
        
        # List assignments
        response = self.client.get("/v1/plans/assignments")
        self.assert_response_success(response, 200)
        
        # Update assignment
        update_data = {
            "plan_id_or_slug": plan_id,
            "user_id": assignment_data["user_id"],
            "valid_until": (datetime.utcnow() + timedelta(days=60)).isoformat() + "Z",
            "updated_by": "integration-test"
        }
        
        response = self.client.put("/v1/plans/assignments", json=update_data)
        self.assert_response_success(response, 200)
        self.logger.info("✓ Updated plan assignment")
        
        # Terminate assignment
        terminate_data = {
            "plan_id_or_slug": plan_id,
            "user_id": assignment_data["user_id"]
        }
        
        response = self.client.delete("/v1/plans/assignments", json=terminate_data)
        self.assert_response_success(response, 204)
        self.logger.info("✓ Terminated plan assignment")
# Meter Tests
class MeterTests(BaseAPITest):
    """Integration tests for Meters API"""
    
    def test_create_meter(self) -> Dict[str, Any]:
        """Test creating a meter"""
        self.logger.info("Testing: Create Meter")
        
        meter_data = self.test_data_factory.create_meter_data()
        response = self.client.post("/v1/meters", json=meter_data)
        
        self.assert_response_success(response, 201)
        response_data = response.json()
        
        assert response_data["data"]["name"] == meter_data["name"]
        assert response_data["data"]["aggregation"] == meter_data["aggregation"]
        
        meter_id = response_data["data"]["id"]
        self.logger.info(f"✓ Created meter with ID: {meter_id}")
        
        return {"meter_id": meter_id, "meter_data": meter_data}
    
    def test_query_meter(self, meter_slug: str):
        """Test querying meter data"""
        self.logger.info(f"Testing: Query Meter - {meter_slug}")
        
        query_data = {
            "meter_slug": meter_slug,
            "from": (datetime.utcnow() - timedelta(days=7)).isoformat() + "Z",
            "to": datetime.utcnow().isoformat() + "Z",
            "window_size": "day"
        }
        
        response = self.client.post("/v1/meters/query", json=query_data)
        self.assert_response_success(response, 200)
        
        data = response.json()
        assert "data" in data["data"]
        assert "window_size" in data["data"]
        
        self.logger.info("✓ Queried meter data successfully")

# Event Tests
class EventTests(BaseAPITest):
    """Integration tests for Events API"""
    
    def test_publish_single_event(self):
        """Test publishing a single event"""
        self.logger.info("Testing: Publish Single Event")
        
        event_data = {
            "events": [self.test_data_factory.create_event_data()],
            "allow_partial_success": False
        }
        
        # Use X-Tenant header instead of X-Tenant-Slug for events
        headers = {"X-Tenant": self.client.config.tenant_slug}
        response = self.client.post("/v1/events", json=event_data, headers=headers)
        
        self.assert_response_success(response, 200)
        
        data = response.json()
        assert data["data"]["success_count"] == 1
        
        self.logger.info("✓ Published single event successfully")
    
    def test_publish_batch_events(self):
        """Test publishing batch events"""
        self.logger.info("Testing: Publish Batch Events")
        
        events = [self.test_data_factory.create_event_data() for _ in range(5)]
        event_data = {
            "events": events,
            "allow_partial_success": True
        }
        
        headers = {"X-Tenant": self.client.config.tenant_slug}
        response = self.client.post("/v1/events", json=event_data, headers=headers)
        
        self.assert_response_success(response, 200)
        
        data = response.json()
        assert data["data"]["success_count"] == 5
        
        self.logger.info("✓ Published batch events successfully")


# Feature Tests
class FeatureTests(BaseAPITest):
    """Integration tests for Features API"""
    
    def test_create_feature_static(self) -> Dict[str, Any]:
        """Test creating a static feature"""
        self.logger.info("Testing: Create Static Feature")
        
        feature_data = self.test_data_factory.create_feature_data("static")
        response = self.client.post("/v1/features", json=feature_data)
        
        self.assert_response_success(response, 201)
        response_data = response.json()
        
        assert response_data["data"]["name"] == feature_data["name"]
        assert response_data["data"]["type"] == "static"
        
        feature_id = response_data["data"]["id"]
        self.logger.info(f"✓ Created static feature with ID: {feature_id}")
        
        return {"feature_id": feature_id, "feature_data": feature_data}
    
    def test_create_feature_metered(self) -> Dict[str, Any]:
        """Test creating a metered feature"""
        self.logger.info("Testing: Create Metered Feature")
        
        feature_data = self.test_data_factory.create_feature_data("metered")
        response = self.client.post("/v1/features", json=feature_data)
        
        self.assert_response_success(response, 201)
        response_data = response.json()
        
        assert response_data["data"]["type"] == "metered"
        feature_id = response_data["data"]["id"]
        self.logger.info(f"✓ Created metered feature with ID: {feature_id}")
        
        return {"feature_id": feature_id, "feature_data": feature_data}
    
    def test_list_features(self):
        """Test listing features"""
        self.logger.info("Testing: List Features")
        
        response = self.client.get("/v1/features")
        self.assert_response_success(response, 200)
        
        data = response.json()
        assert "data" in data
        assert isinstance(data["data"], list)
        
        self.logger.info(f"✓ Retrieved {len(data['data'])} features")
    
    def test_get_feature_by_id(self, feature_id: str):
        """Test getting feature by ID"""
        self.logger.info(f"Testing: Get Feature by ID - {feature_id}")
        
        response = self.client.get(f"/v1/features/{feature_id}")
        self.assert_response_success(response, 200)
        
        data = response.json()
        assert data["data"]["id"] == feature_id
        
        self.logger.info("✓ Retrieved feature by ID successfully")
    
    def test_update_feature(self, feature_id: str):
        """Test updating a feature"""
        self.logger.info(f"Testing: Update Feature - {feature_id}")
        
        update_data = {
            "name": "Updated Feature Name",
            "description": "Updated feature description for testing",
            "updated_by": "integration-test"
        }
        
        response = self.client.put(f"/v1/features/{feature_id}", json=update_data)
        self.assert_response_success(response, 200)
        
        data = response.json()
        assert data["data"]["name"] == update_data["name"]
        
        self.logger.info("✓ Updated feature successfully")
    
    def test_delete_feature(self, feature_id: str):
        """Test deleting a feature"""
        self.logger.info(f"Testing: Delete Feature - {feature_id}")
        
        response = self.client.delete(f"/v1/features/{feature_id}")
        self.assert_response_success(response, 204)
        
        # Verify feature is deleted
        response = self.client.get(f"/v1/features/{feature_id}")
        self.assert_response_error(response, 404)
        
        self.logger.info("✓ Deleted feature successfully")
    
    def test_feature_validation_errors(self):
        """Test feature validation errors"""
        self.logger.info("Testing: Feature Validation Errors")
        
        # Test missing required fields
        invalid_data = {"name": "Test"}  # Missing required fields
        response = self.client.post("/v1/features", json=invalid_data)
        self.assert_response_error(response, 400)
        
        # Test invalid feature type
        invalid_type_data = self.test_data_factory.create_feature_data()
        invalid_type_data["type"] = "invalid_type"
        response = self.client.post("/v1/features", json=invalid_type_data)
        self.assert_response_error(response, 400)
        
        self.logger.info("✓ Validation errors handled correctly")



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
        
        # Plan Tests
        self.logger.info("\n📋 Running Plan Tests...")
        plan = self.run_test("Create Plan", self.plan_tests.test_create_plan)
        
        if plan:
            self.run_test("Plan Assignment Lifecycle", self.plan_tests.test_plan_assignment_lifecycle, plan["plan_id"])
        
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


# CLI Entry Point
def main():
    """Main entry point for running tests"""
    import argparse
    
    parser = argparse.ArgumentParser(description="API Integration Test Suite")
    parser.add_argument("--base-url", default="http://localhost:8000", help="API base URL")
    parser.add_argument("--tenant", default="test-tenant", help="Tenant slug")
    parser.add_argument("--log-level", default="INFO", choices=["DEBUG", "INFO", "WARNING", "ERROR"])
    parser.add_argument("--timeout", type=int, default=30, help="Request timeout in seconds")
    
    args = parser.parse_args()
    
    config = TestConfig(
        base_url=args.base_url,
        tenant_slug=args.tenant,
        log_level=args.log_level,
        timeout=args.timeout
    )
    
    runner = IntegrationTestRunner(config)
    runner.run_all_tests()


if __name__ == "__main__":
    main()
