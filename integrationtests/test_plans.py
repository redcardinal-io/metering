from base_test import BaseAPITest
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
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
            "valid_from": (datetime.utcnow() + timedelta(days=15)).isoformat() + "Z",
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


