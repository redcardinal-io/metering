from base_test import BaseAPITest
import uuid
from datetime import datetime
from typing import Dict, List, Optional, Any

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


