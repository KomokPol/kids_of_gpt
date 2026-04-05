from __future__ import annotations

import uuid
from datetime import datetime, timezone
from typing import Any

from pydantic import BaseModel, Field


class EventIn(BaseModel):
    event_type: str = Field(..., examples=["burmalda.spin_completed"])
    aggregate_type: str = Field(..., examples=["spin"])
    aggregate_id: str = Field(..., examples=["spin-42"])
    user_id: str = Field(..., examples=["user-1"])
    producer: str = Field("analytics-api", examples=["burmalda-service"])
    correlation_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    payload: dict[str, Any] = Field(default_factory=dict)


class EventStored(EventIn):
    event_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    occurred_at: str = Field(
        default_factory=lambda: datetime.now(timezone.utc).isoformat()
    )


class UserStats(BaseModel):
    user_id: str
    total_events: int = 0
    event_counts: dict[str, int] = Field(default_factory=dict)
    first_seen: str | None = None
    last_seen: str | None = None
    streak_days: int = 0


class DashboardStats(BaseModel):
    total_events: int = 0
    total_users: int = 0
    event_type_counts: dict[str, int] = Field(default_factory=dict)
    top_users: list[dict[str, Any]] = Field(default_factory=list)
    recent_events: list[dict[str, Any]] = Field(default_factory=list)
