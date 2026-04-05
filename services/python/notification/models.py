from datetime import datetime, timezone
import uuid

from pydantic import BaseModel, Field


class NotificationIn(BaseModel):
    user_id: str
    title: str
    body: str


class NotificationStored(NotificationIn):
    notification_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    created_at: str = Field(
        default_factory=lambda: datetime.now(timezone.utc).isoformat()
    )
    read: bool = False


class NotificationList(BaseModel):
    user_id: str
    notifications: list[NotificationStored] = []
    total: int = 0
