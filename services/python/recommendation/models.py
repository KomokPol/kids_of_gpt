from pydantic import BaseModel, Field


class RecommendationItem(BaseModel):
    item_id: str
    title: str
    description: str
    category: str
    price: int = 0
    min_level: int = 0
    score: float = 0.0
    reason: str = ""


class RecommendationResponse(BaseModel):
    user_id: str
    level: int
    recommendations: list[RecommendationItem] = []
    total: int = 0
