from __future__ import annotations

from fastapi import APIRouter, Depends

from .aggregator import Aggregator
from .models import DashboardStats, EventIn, EventStored, UserStats

router = APIRouter()


def get_aggregator() -> Aggregator:
    from .main import app_state
    return app_state["aggregator"]


@router.post("/events", response_model=EventStored, status_code=201)
async def ingest_event(event: EventIn) -> EventStored:
    stored = EventStored(**event.model_dump())
    agg = get_aggregator()
    await agg.ingest(stored.model_dump())
    return stored


@router.get("/stats/{user_id}", response_model=UserStats)
async def user_stats(user_id: str) -> UserStats:
    agg = get_aggregator()
    data = await agg.user_stats(user_id)
    return UserStats(**data)


@router.get("/dashboard", response_model=DashboardStats)
async def dashboard() -> DashboardStats:
    agg = get_aggregator()
    data = await agg.dashboard()
    return DashboardStats(**data)
