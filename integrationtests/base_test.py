# base_test.py
import json
import logging
import sys
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from dataclasses import dataclass
from enum import Enum

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry


@dataclass
class TestConfig:
    base_url: str = "http://localhost:8000"
    tenant_slug: str = "test-tenant"
    timeout: int = 30
    max_retries: int = 3
    backoff_factor: float = 0.3
    log_level: str = "INFO"
    log_file: str = "api_tests.log"


class TestStatus(Enum):
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
