from base_test import BaseAPITest
from typing import Dict, List, Optional, Any
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


