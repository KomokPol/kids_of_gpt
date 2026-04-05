from __future__ import annotations

from fastapi import APIRouter, HTTPException, Query

from .engine import RecommendationEngine
from .models import RecommendationItem, RecommendationResponse

router = APIRouter()


def _get_engine() -> RecommendationEngine:
    from .main import app_state

    return app_state["engine"]


def _to_item(d: dict, include_score: bool = True) -> RecommendationItem:
    data = {
        "item_id": str(d["item_id"]),
        "title": str(d["title"]),
        "description": str(d["description"]),
        "category": str(d["category"]),
        "price": int(d.get("price", 0)),
        "min_level": int(d.get("min_level", 0)),
    }
    if include_score:
        data["score"] = float(d.get("score", 0.0))
        data["reason"] = str(d.get("reason", ""))
    return RecommendationItem(**data)


@router.get("/recommendations/{user_id}", response_model=RecommendationResponse)
async def get_recommendations(
    user_id: str,
    level: int = Query(0, ge=0, le=99),
    category: str | None = Query(None, description="Фильтр по категории: barygi, kinoshka, balanda"),
    limit: int = Query(10, ge=1, le=100),
) -> RecommendationResponse:
    engine = _get_engine()
    rows = engine.recommend(user_id=user_id, level=level, category=category, limit=limit)
    items = [_to_item(r) for r in rows]
    return RecommendationResponse(
        user_id=user_id,
        level=level,
        recommendations=items,
        total=len(items),
    )


@router.get("/items", response_model=list[RecommendationItem])
async def list_items() -> list[RecommendationItem]:
    engine = _get_engine()
    out: list[RecommendationItem] = []
    for raw in engine.all_items():
        out.append(_to_item(raw, include_score=False))
    return out


@router.get("/items/{item_id}", response_model=RecommendationItem)
async def get_item(item_id: str) -> RecommendationItem:
    engine = _get_engine()
    for raw in engine.all_items():
        if str(raw.get("item_id")) == item_id:
            return _to_item(raw, include_score=False)
    raise HTTPException(status_code=404, detail="item_not_found")
