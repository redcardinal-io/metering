from base_test import BaseAPITest
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
#
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

    def test_delete_meter(self, meter_id: str):
        """Test deleting a meter"""
        self.logger.info(f"Testing: Delete Meter - {meter_id}")
        
        response = self.client.delete(f"/v1/meters/{meter_id}")
        self.assert_response_success(response, 204)
        
        response = self.client.get(f"/v1/meters/{meter_id}")
        self.assert_response_error(response, 404)
        
        self.logger.info("✓ Deleted meter successfully")
    

